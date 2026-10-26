#!/usr/bin/env python3
"""Drive the real Explorer through a PTY; assert the rendered screen, not stale output."""
import errno
import fcntl
import os
import pty
import select
import signal
import struct
import subprocess
import sys
import termios
import time

import pyte


def main():
    binary, scenario, *args = sys.argv[1:]
    password = os.environ['BWB_TEST_PASSWORD']
    secret = 'DEMO-only-password-1!'
    pid, fd = pty.fork()
    if pid == 0:
        os.execve(binary, [binary, *args], {**os.environ, 'TERM': 'xterm-256color'})
    screen = pyte.Screen(120, 34)
    stream = pyte.ByteStream(screen)
    output = bytearray()
    status = None

    def resize(width, height):
        screen.resize(lines=height, columns=width)
        fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack('HHHH', height, width, 0, 0))
        os.kill(pid, signal.SIGWINCH)

    def view():
        return '\n'.join(screen.display)

    def poll():
        nonlocal status
        if select.select([fd], [], [], 0.03)[0]:
            try:
                data = os.read(fd, 65536)
            except OSError as error:
                if error.errno != errno.EIO:
                    raise
                data = b''
            output.extend(data)
            stream.feed(data)
        if status is None:
            child, code = os.waitpid(pid, os.WNOHANG)
            if child:
                status = os.waitstatus_to_exitcode(code)

    def until(predicate):
        deadline = time.monotonic() + 15
        while not predicate():
            assert status is None, f'TUI exited: {status}'
            assert time.monotonic() < deadline, 'TUI timed out: ' + view().replace(password, '[password]').replace(secret, '[secret]')
            poll()

    def send(text):
        os.write(fd, text.encode())

    def pause():
        deadline = time.monotonic() + .15
        while time.monotonic() < deadline:
            poll()

    try:
        resize(120, 34)
        until(lambda: 'Password:' in view())
        if scenario == 'cancel':
            send('\x1b')
            until(lambda: 'Authentication canceled' in view())
            send('q')
        else:
            send(password + '\r')
            until(lambda: 'GitHub (demo)' in view())
            assert secret not in view()
            if scenario == 'login':
                send('\x10')
                until(lambda: 'Applied timeout:' in view())
                send('r')
                pause()
                until(lambda: 'Applied timeout:' in view() and 'Working' not in view())
                send('x')
                until(lambda: 'Remove saved authentication for steady?' in view())
                send('y')
                until(lambda: 'State: logged_out' in view())
                send('u')
                until(lambda: 'Password:' in view())
                send(password + '\r')
                until(lambda: 'GitHub (demo)' in view())
                send('q')
                until(lambda: status is not None)
                assert status == 0 and password.encode() not in output
                print('PASS: TUI configured login, reload and logout')
                return
            send('/GitHub')
            until(lambda: 'GitHub' in screen.display[2])
            pause()
            send('\r')  # Leave search input, keeping the filtered tree.
            pause()
            # If arrows were processed by the search input, now navigate the tree.
            for _ in range(12):
                if any('>' in line[:60] and 'GitHub (demo)' in line[:60] for line in screen.display):
                    break
                send('\x1b[B')
                pause()
            send('\r')
            until(lambda: 'Enter reveals selected field' in view())
            for _ in range(10):
                if any('> Password ' in line[60:] for line in screen.display):
                    break
                send('\x1b[B')
                pause()
            else:
                raise AssertionError('password field not reachable')
            assert secret not in view()
            send('\r')
            until(lambda: secret in view())
            send('r')
            until(lambda: secret not in view())
            pause()
            until(lambda: 'Working' not in view() and 'Enter reveals selected field' in view())
            assert 'GitHub' in screen.display[2]
            assert any('>' in line[:60] and 'GitHub (demo)' in line[:60] for line in screen.display)
            # Refresh resets field selection; explicitly select password again.
            for _ in range(10):
                if any('> Password ' in line[60:] for line in screen.display):
                    break
                send('\x1b[B')
                pause()
            else:
                raise AssertionError('password field missing after sync')
            send('\r')
            until(lambda: secret in view())
            send('\t')
            until(lambda: secret not in view())
            send('\r\r')
            until(lambda: secret in view())
            resize(60, 20)
            pause()
            subprocess.run([binary, '--profile', 'default', 'lock'], check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            until(lambda: 'Vault locked' in view())
            assert secret not in view() and 'GitHub' not in view()
            send('\r')
            until(lambda: 'Password:' in view())
            send(password + '\r')
            until(lambda: 'GitHub (demo)' in view())
            resize(120, 34)
            send('\x10')  # Ctrl+P: Profiles.
            until(lambda: 'Applied timeout:' in view())
            assert 'Config:' in view()
            send('\r')
            until(lambda: 'GitHub (demo)' in view())
            send('q')
        until(lambda: status is not None)
        assert status == 0, status
        assert password.encode() not in output, 'master password echoed'
        print('PASS: TUI ' + scenario)
    finally:
        if status is None:
            os.kill(pid, signal.SIGKILL)
            os.waitpid(pid, 0)
        os.close(fd)


if __name__ == '__main__':
    main()
