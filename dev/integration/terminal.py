#!/usr/bin/env python3
"""Exercise the real CLI with a controlling TTY and separately piped stdout."""
import errno
import os
import pty
import select
import signal
import subprocess
import sys
import termios
import time


class Command:
    def __init__(self, binary, args):
        out_read, out_write = os.pipe()
        self.pid, self.tty = pty.fork()
        if self.pid == 0:
            os.close(out_read)
            os.dup2(out_write, 1)
            os.close(out_write)
            os.execve(binary, [binary, *args], os.environ)
        os.close(out_write)
        self.out = out_read
        self.output = bytearray()
        self.terminal = bytearray()
        self.status = None
        self.open_fds = {self.tty, self.out}

    def poll(self):
        if self.open_fds:
            ready, _, _ = select.select(list(self.open_fds), [], [], 0.02)
            for fd in ready:
                try:
                    data = os.read(fd, 65536)
                except OSError as error:
                    if error.errno != errno.EIO:
                        raise
                    data = b''
                if not data:
                    self.open_fds.discard(fd)
                elif fd == self.tty:
                    self.terminal.extend(data)
                else:
                    self.output.extend(data)
        if self.status is None:
            child, status = os.waitpid(self.pid, os.WNOHANG)
            if child:
                self.status = os.waitstatus_to_exitcode(status)

    def has_prompt(self):
        return b'Password: ' in self.terminal

    def send(self, value):
        os.write(self.tty, value)

    def close(self):
        if self.status is None:
            os.kill(self.pid, signal.SIGKILL)
            os.waitpid(self.pid, 0)
        os.close(self.tty)
        os.close(self.out)


def until(commands, predicate):
    deadline = time.monotonic() + 15
    while not predicate():
        assert time.monotonic() < deadline, 'terminal scenario timed out'
        for command in commands:
            command.poll()


def main():
    scenario, binary, *args = sys.argv[1:]
    password = os.environ['BWB_TEST_PASSWORD'].encode()
    expected = os.environ['BWB_TEST_EXPECT'].encode()
    if scenario in ('apikey', 'apikey_setup'):
        os.environ.pop('BWB_CLIENTID', None)
        os.environ.pop('BWB_CLIENTSECRET', None)
    commands = []
    try:
        commands.append(Command(binary, args))
        if scenario in ('setup_cancel', 'setup_signal'):
            owner = commands[0]
            until(commands, lambda: b'Server URL: ' in owner.terminal or owner.status is not None)
            assert owner.status is None, 'missing server prompt: ' + repr(bytes(owner.terminal))
            if scenario == 'setup_cancel':
                owner.send(b'\x03')
            else:
                os.kill(owner.pid, signal.SIGINT)
            until(commands, lambda: owner.status is not None and not owner.open_fds)
            assert owner.status != 0 and not owner.output and not owner.has_prompt()
            assert termios.tcgetattr(owner.tty)[3] & termios.ECHO, 'terminal echo not restored'
            print('PASS: TTY scenario', scenario)
            return
        if scenario in ('setup', 'sso_setup'):
            owner = commands[0]
            for label, value in ((b'Server URL: ', os.environ['BWB_TEST_SERVER'].encode()),
                                 (b'Email: ', os.environ['BWB_TEST_EMAIL'].encode())):
                until(commands, lambda: label in owner.terminal or owner.status is not None)
                assert owner.status is None, 'missing identity prompt: ' + repr(bytes(owner.terminal))
                owner.send(b'invalid-value\n')
                until(commands, lambda: owner.terminal.count(label) == 2 or owner.status is not None)
                assert owner.status is None, 'invalid identity was not retried'
                owner.send(value + b'X\x7f\n')
                until(commands, lambda: value in owner.terminal)
            scenario = 'sso' if scenario == 'sso_setup' else 'success'
        if scenario in ('concurrent', 'cancel_waiters'):
            commands.append(Command(binary, args))
        if scenario in ('apikey', 'apikey_setup'):
            owner = commands[0]
            fields = [(b'Client ID: ', os.environ['BWB_INTEGRATION_CLIENTID'].encode()),
                      (b'Client secret: ', os.environ['BWB_INTEGRATION_CLIENTSECRET'].encode())]
            if scenario == 'apikey_setup':
                fields.append((b'Server URL: ', os.environ['BWB_TEST_SERVER'].encode()))
            fields.append((b'Password: ', password))
            for label, value in fields:
                until(commands, lambda: label in owner.terminal or owner.status is not None)
                assert label in owner.terminal, 'missing API key prompt'
                assert not termios.tcgetattr(owner.tty)[3] & termios.ECHO, 'credential echo enabled'
                owner.send(value + b'\n')
            until(commands, lambda: owner.status is not None and not owner.open_fds)
            assert owner.status == 0 and bytes(owner.output) == expected, 'API key login failed'
            assert b'Email: ' not in owner.terminal, 'API key login unnecessarily requested email'
            assert termios.tcgetattr(owner.tty)[3] & termios.ECHO, 'terminal echo not restored'
            assert os.environ['BWB_INTEGRATION_CLIENTSECRET'].encode() not in owner.terminal, 'API key echoed'
        elif scenario == 'sso_browser_cancel':
            owner = commands[0]
            until(commands, lambda: owner.has_prompt() or owner.status is not None)
            assert owner.status is None and b'Complete SSO login' not in owner.terminal
            owner.send(password + b'\n')
            until(commands, lambda: b'Complete SSO login' in owner.terminal or owner.status is not None)
            assert owner.status is None, 'SSO failed before browser cancellation'
            owner.send(b'\x03')
            until(commands, lambda: owner.status is not None and not owner.open_fds)
            assert owner.status != 0 and not owner.output
            assert termios.tcgetattr(owner.tty)[3] & termios.ECHO, 'terminal echo not restored'
        elif scenario == 'sso':
            import urllib.request
            owner = commands[0]
            for count, value in ((1, b'incorrect-password'), (2, password)):
                until(commands, lambda: owner.terminal.count(b'Password: ') == count or owner.status is not None)
                assert owner.status is None, 'SSO failed before password: ' + repr(bytes(owner.terminal))
                if count == 1:
                    assert b'Complete SSO login' not in owner.terminal, 'browser opened before password input'
                with urllib.request.urlopen(os.environ['BWB_TEST_SSO_CHECK']) as response:
                    expected_exchanges = int(os.environ['BWB_TEST_SSO_EXCHANGES']) - (1 if count == 1 else 0)
                    assert int(response.read()) == expected_exchanges, 'SSO exchange ran before initial password input or repeated on retry'
                assert not termios.tcgetattr(owner.tty)[3] & termios.ECHO, 'password echo enabled'
                owner.send(value + b'\n')
            until(commands, lambda: owner.status is not None and not owner.open_fds)
            assert owner.status == 0 and bytes(owner.output) == expected, 'SSO login failed: ' + repr(bytes(owner.terminal))
            assert owner.terminal.count(b'Complete SSO login') == 1, 'SSO opened twice'
            assert termios.tcgetattr(owner.tty)[3] & termios.ECHO, 'terminal echo not restored'
        elif scenario == 'no_prompt':
            until(commands, lambda: all(c.status is not None and not c.open_fds for c in commands))
            assert commands[0].status != 0 and not commands[0].has_prompt()
            assert not commands[0].output, 'failed command wrote stdout'
        else:
            until(commands, lambda: any(c.has_prompt() for c in commands) or all(c.status is not None for c in commands))
            owners = [c for c in commands if c.has_prompt()]
            assert len(owners) == 1, 'expected one prompt: ' + repr([bytes(c.terminal) for c in commands])
            owner = owners[0]
            assert not termios.tcgetattr(owner.tty)[3] & termios.ECHO, 'password echo enabled'
            if len(commands) > 1:
                # Both clients have time to reach the reservation before it completes.
                deadline = time.monotonic() + 0.2
                while time.monotonic() < deadline:
                    for command in commands:
                        command.poll()
                assert sum(c.has_prompt() for c in commands) == 1, 'parallel prompts'
            if scenario == 'lock':
                subprocess.run([binary, 'lock'], stdin=subprocess.DEVNULL, stdout=subprocess.PIPE,
                               stderr=subprocess.PIPE, check=True, timeout=10)
            elif scenario in ('cancel', 'cancel_waiters'):
                owner.send(b'\x03')
            elif scenario == 'long':
                owner.send(b'x' * 4097 + b'OVERLONG-PASSWORD-END\n')
            elif scenario == 'signal':
                os.kill(owner.pid, signal.SIGINT)
            else:
                owner.send((b'incorrect-password' if scenario == 'wrong' else password) + b'\n')
            until(commands, lambda: all(c.status is not None and not c.open_fds for c in commands))
            success = scenario in ('success', 'concurrent')
            for command in commands:
                assert (command.status == 0) == success, 'unexpected CLI exit status'
                assert bytes(command.output) == (expected if success else b''), 'unexpected stdout'
                assert termios.tcgetattr(command.tty)[3] & termios.ECHO, 'terminal echo not restored'
            assert sum(c.terminal.count(b'Password: ') for c in commands) == 1, 'authentication retried or cascaded'
        for command in commands:
            assert password not in command.terminal, 'password echoed to terminal'
            assert b'incorrect-password' not in command.terminal, 'rejected password echoed'
            assert b'OVERLONG-PASSWORD-END' not in command.terminal, 'overlong input leaked'
        print('PASS: TTY scenario', scenario)
    finally:
        for command in commands:
            command.close()


if __name__ == '__main__':
    main()
