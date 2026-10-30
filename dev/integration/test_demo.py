import json
import os
from pathlib import Path
import sqlite3
import tempfile
import unittest
from unittest.mock import patch

import demo
import run


class DemoTest(unittest.TestCase):
    def test_restore_shared_snapshot(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / 'db.sqlite3'
            manifest = run.restore_snapshot(path)
            self.assertEqual(manifest['email'], run.EMAIL)
            with sqlite3.connect(path) as db:
                self.assertEqual(db.execute('PRAGMA foreign_key_check').fetchall(), [])

    def test_run_preserves_arguments_and_isolates_environment(self):
        with tempfile.TemporaryDirectory() as directory:
            state = Path(directory) / 'environment.json'
            state.write_text(json.dumps({'instance': 'demo-test', 'ready': True}))
            with patch.object(demo, 'STATE', state), patch.dict(os.environ, {
                'BWB_PROFILE': 'personal', 'BWB_CLIENTSECRET': 'private', 'TERM': 'xterm-256color',
            }), patch('sys.argv', ['demo.py', 'run', 'item', 'read', 'name with spaces']), \
                    patch('os.chdir'), patch('os.execvpe') as execute:
                demo.main()
                command, args, env = execute.call_args.args
                self.assertEqual(command, 'go')
                self.assertEqual(args, ['go', 'run', '.', 'item', 'read', 'name with spaces'])
                self.assertEqual(env['BWB_INSTANCE'], 'demo-test')
                self.assertEqual(env['TERM'], 'xterm-256color')
                self.assertNotIn('BWB_PROFILE', env)
                self.assertNotIn('BWB_CLIENTSECRET', env)

    def test_down_logs_out_all_configured_profiles(self):
        for profiles in (('owner', 'restricted'), ('owner', 'restricted', 'argon2id')):
            with tempfile.TemporaryDirectory() as directory:
                home = Path(directory) / 'demo'
                home.mkdir()
                state = home / 'environment.json'
                state.write_text(json.dumps({'container': 'test-container'}))
                (home / 'config.toml').write_text('\n'.join(f'[profiles.{p}]' for p in profiles))
                with patch.object(demo, 'HOME', home), patch.object(demo, 'STATE', state), \
                        patch.object(demo, 'cli') as cli, patch.object(run, 'docker', return_value=''), patch('builtins.print'):
                    demo.down()
                    self.assertEqual([call.args[1:] for call in cli.call_args_list],
                                     [('--profile', p, 'logout', '--non-interactive') for p in profiles] + [('daemon', 'stop')])
                    self.assertFalse(home.exists())

    def test_failed_logout_retains_cleanup_state(self):
        with tempfile.TemporaryDirectory() as directory:
            home = Path(directory)
            state = home / 'environment.json'
            state.write_text('{}')
            (home / 'config.toml').touch()
            with patch.object(demo, 'HOME', home), patch.object(demo, 'STATE', state), \
                    patch.object(demo, 'cli', side_effect=RuntimeError('keychain unavailable')):
                with self.assertRaisesRegex(RuntimeError, 'keychain'):
                    demo.down()
                self.assertTrue(state.exists())
