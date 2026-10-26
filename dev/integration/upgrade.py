#!/usr/bin/env python3
"""Upgrade through a real terminal and resume the interrupted read after unlock."""
import errno
import os
import pty
import select
import signal
import sys
import time

binary, *args = sys.argv[1:]
password = os.environ['BWB_TEST_PASSWORD']
expected = os.environ['BWB_TEST_EXPECT'].encode()
pid, fd = pty.fork()
if pid == 0:
    os.execve(binary, [binary, *args], os.environ)
output = bytearray()
status = None
sent = False
try:
    deadline = time.monotonic() + 60
    while status is None:
        assert time.monotonic() < deadline, 'upgrade command timed out'
        if select.select([fd], [], [], .05)[0]:
            try:
                output.extend(os.read(fd, 65536))
            except OSError as error:
                if error.errno != errno.EIO:
                    raise
        if b'Password: ' in output and not sent:
            os.write(fd, (password + '\n').encode())
            sent = True
        child, code = os.waitpid(pid, os.WNOHANG)
        if child:
            status = os.waitstatus_to_exitcode(code)
    if os.environ.get('BWB_TEST_FAILURE'):
        assert status != 0 and not sent and expected not in output, 'failed preflight changed the session'
    else:
        assert status == 0 and sent and expected in output, 'upgrade/read/unlock failed'
    assert password.encode() not in output, 'password echoed'
finally:
    if status is None:
        os.kill(pid, signal.SIGKILL)
        os.waitpid(pid, 0)
    os.close(fd)
