#!/usr/bin/env python3
"""Drive the real Explorer through a PTY; assert the rendered screen, not stale output."""
import errno
import re
import hmac
import hashlib
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

from pathlib import Path

import pyte


def main():
    binary, scenario, *args = sys.argv[1:]
    password = os.environ['BWB_TEST_PASSWORD']
    secret = 'DEMO-only-password-1!'
    # pyte lacks xterm CSI S/T scrolling; advertise its supported VT100 capabilities.
    pid, fd = pty.fork()
    if pid == 0:
        os.execve(binary, [binary, *args], {**os.environ, 'TERM': 'vt100', 'COLORTERM': 'truecolor', 'NO_COLOR': '0'})
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

    def until(predicate, timeout=15):
        deadline = time.monotonic() + timeout
        while True:
            if predicate():
                pause()  # A terminal frame may span several PTY reads.
                if predicate():
                    return
            assert status is None, f'TUI exited: {status}'
            assert time.monotonic() < deadline, 'TUI timed out: ' + view().replace(password, '[password]').replace(secret, '[secret]')
            poll()

    def send(text):
        os.write(fd, text.encode())

    def pause():
        deadline = time.monotonic() + .15
        while time.monotonic() < deadline:
            poll()

    def deadline(signum, frame):
        raise TimeoutError('TUI deadline: ' + view().replace(password, '[password]').replace(secret, '[secret]'))

    signal.signal(signal.SIGALRM, deadline)
    signal.alarm(55)
    try:
        resize(120, 34)
        until(lambda: 'Password:' in view())
        if scenario == 'cancel':
            send('\x1b')
            until(lambda: 'Authentication canceled' in view())
            send('q')
        else:
            send(password + '\r')
            if scenario == 'otp':
                until(lambda: 'seconds remaining' in view())
                def shown_code():
                    return re.search(r'(?m)^│ *([0-9]{6}) *│$', view())
                until(lambda: shown_code() is not None)
                initial = shown_code().group(1)
                def expected_code(step):
                    digest = hmac.new(b'12345678901234567890', struct.pack('>Q', step), hashlib.sha1).digest()
                    offset = digest[-1] & 15
                    return f'{(int.from_bytes(digest[offset:offset+4], "big") & 0x7fffffff) % 1000000:06d}'
                step = int(time.time()) // 30
                assert initial in [expected_code(step-1), expected_code(step), expected_code(step+1)]
                assert 'GEZDGNBV' not in view()
                until(lambda: shown_code() is not None and shown_code().group(1) != initial, timeout=35)
                subprocess.run([binary, 'daemon', 'stop'], check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=12)
                until(lambda: 'Daemon disconnected' in view())
                assert shown_code() is None
                send('\r')
                until(lambda: 'Password:' in view())
                send(password + '\r')
                until(lambda: 'seconds remaining' in view())
                subprocess.run([binary, 'lock'], check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
                until(lambda: 'Vault locked' in view())
                assert shown_code() is None
                send('q')
                until(lambda: status is not None)
                assert status == 0 and password.encode() not in output
                print('PASS: TOTP authentication, generation, rollover and lock purge')
                return
            until(lambda: 'GitHub (demo)' in view())
            assert secret not in view()
            if scenario == 'login':
                subprocess.run([binary, 'daemon', 'stop'], check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=12)
                until(lambda: 'Daemon disconnected' in view())
                send('\r')
                until(lambda: 'Password:' in view())
                send(password + '\r')
                until(lambda: 'GitHub (demo)' in view())
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
            def find_item(query, name):
                send('/' + query)
                pause()
                send('\r')
                pause()
                for _ in range(20):
                    if any('>' in line[:40] and name in line[:40] for line in screen.display):
                        break
                    send('\x1b[B')
                    pause()
                else:
                    raise AssertionError('item not reachable: ' + name)
                send('\r')
                until(lambda: 'Secrets hidden' in view() and 'Loading fields' not in view())

            def select_field(name):
                for _ in range(20):
                    if any('> ' + name in line[41:] for line in screen.display):
                        return
                    send('\x1b[B')
                    pause()
                raise AssertionError('field not reachable: ' + name)

            find_item('OTP', 'OTP (public')
            until(lambda: 'Authentication code' in view())
            assert 'GEZDGNBV' not in view()
            send('v')
            until(lambda: re.search(r'Authentication code +[0-9]{6}', view()) is not None)
            assert 'Secrets revealed' in view()
            send('v')
            until(lambda: 'Secrets hidden' in view())
            assert 'GEZDGNBV' not in view()
            send('\t\x1b')
            pause()
            find_item('GitHub', 'GitHub (demo)')
            select_field('Password')
            for rgb, active, inactive in [('ffff/ffff/ffff', '252a31', 'cbd0d6'), ('1717/1919/1d1d', 'dce0e5', '424953')]:
                send('\x1b]11;rgb:' + rgb + '\x1b\\')
                until(lambda: screen.buffer[6][41].fg == active and screen.buffer[6][0].fg == inactive)
            assert secret not in view()
            clipboard_file = Path(os.environ['BWB_CLIPBOARD_FILE'])
            clipboard_file.write_text('')
            send('y')
            until(lambda: 'Copied' in view())
            assert secret not in view() and clipboard_file.read_text() == secret
            assert 'v reveal item' in view(), 'copy message replaced key hints'
            send('v')
            until(lambda: secret in view())
            send('\x1b[B')  # Field navigation retains reveal.
            pause()
            assert secret in view()
            send('\t')      # Pane navigation retains reveal.
            pause()
            assert secret in view()
            send('/')
            until(lambda: secret not in view())
            send('\r\t')
            pause()
            send('v')
            until(lambda: secret in view())
            send('r')
            until(lambda: secret not in view())
            pause()
            until(lambda: 'Working' not in view() and 'Secrets hidden' in view())
            assert 'GitHub' in screen.display[3]
            until(lambda: any('>' in line[:40] and 'GitHub (demo)' in line[:40] for line in screen.display))
            send('v')
            until(lambda: secret in view())
            subprocess.run([binary, 'daemon', 'stop'], check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=12)
            until(lambda: 'Daemon disconnected' in view() and secret not in view() and 'GitHub' not in view())
            send('\r')
            until(lambda: 'Password:' in view())
            send(password + '\r')
            until(lambda: 'GitHub (demo)' in view())
            assert secret not in view()
            resize(60, 20)
            pause()
            subprocess.run([binary, '--profile', 'default', 'lock'], check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            until(lambda: 'Vault locked' in view() and secret not in view() and 'GitHub' not in view())
            send('\r')
            until(lambda: 'Password:' in view())
            send(password + '\r')
            until(lambda: 'GitHub (demo)' in view())
            resize(120, 34)
            send('\x10')  # Ctrl+P: Profiles.
            until(lambda: 'Applied timeout:' in view() and 'Config:' in view())
            send('\r')
            until(lambda: 'GitHub (demo)' in view())
            send('q')
        until(lambda: status is not None)
        assert status == 0, status
        assert password.encode() not in output, 'master password echoed'
        print('PASS: TUI ' + scenario)
    except BaseException:
        import traceback
        traceback.print_exc()
        print(view().replace(password, '[password]').replace(secret, '[secret]'), flush=True)
        raise
    finally:
        signal.alarm(0)
        if status is None:
            os.kill(pid, signal.SIGKILL)
            os.close(fd)
            os.waitpid(pid, 0)
        else:
            os.close(fd)


if __name__ == '__main__':
    main()
