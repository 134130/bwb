"""Fast checks for the test harness and committed snapshot; no server required."""
import base64
import copy
import json
import sqlite3
import unittest

import run


class ContractTest(unittest.TestCase):
    def test_authenticated_encryption_rejects_tampering_and_wrong_key(self):
        key = bytes(range(64))
        for plain in (b'', b'hello', '여러 줄\n값'.encode(), b'x' * 64):
            cipher = run.encrypt(plain, key)
            self.assertEqual(run.decrypt(cipher, key), plain)
            with self.assertRaisesRegex(AssertionError, 'authentication'):
                run.decrypt(cipher, bytes(reversed(key)))
            parts = cipher[2:].split('|')
            mac = bytearray(base64.b64decode(parts[2]))
            mac[0] ^= 1
            parts[2] = run.b64(mac)
            with self.assertRaisesRegex(AssertionError, 'authentication'):
                run.decrypt('2.' + '|'.join(parts), key)
        with self.assertRaisesRegex(AssertionError, 'unsupported'):
            run.decrypt('0.not-supported', key)

    def test_comparison_requires_every_fixture_field(self):
        expected = {'password': '', 'fields': [{'value': 'secret'}]}
        run.compare(expected, dict(expected, optional=None), 'item')
        for actual in ({'password': ''}, {'password': '', 'fields': []},
                       {'password': None, 'fields': [{'value': 'secret'}]}):
            with self.assertRaises(AssertionError):
                run.compare(expected, actual, 'item')
        changed = copy.deepcopy(expected)
        changed['fields'][0]['value'] = 'different'
        with self.assertRaises(AssertionError):
            run.compare(expected, changed, 'item')

    def test_snapshot_matches_fixture_and_restores(self):
        manifest = json.loads((run.SNAPSHOT / 'manifest.json').read_text())
        self.assertEqual(manifest['version'], 1)
        self.assertEqual(manifest['image'], run.IMAGE)
        self.assertEqual(manifest['types'], {str(k): v for k, v in run.TYPES.items()})
        self.assertEqual(manifest['fixture_sha256'], run.digest(run.ROOT / 'testdata/bitwarden.json'))
        self.assertEqual(manifest['dump_sha256'], run.digest(run.SNAPSHOT / 'vaultwarden.sql'))
        self.assertEqual(manifest['organizations_sha256'], run.digest(run.ROOT / 'testdata/organizations.json'))
        organizations = json.loads((run.ROOT / 'testdata/organizations.json').read_text())
        self.assertGreaterEqual(len(organizations['organizations']), 2)
        self.assertTrue(all(len(o['collections']) >= 2 for o in organizations['organizations']))
        self.assertEqual(len(manifest['organizations']['entries']), len(organizations['organizations']))
        org_items = {i['id'] for o in manifest['organizations']['entries'] for i in o['items']}
        self.assertEqual(len(org_items), len(organizations['organizations'])*len(organizations['items']))
        data = run.fixture()
        self.assertEqual(set(manifest['items']), {x['id'] for x in data['items']})
        self.assertEqual(set(manifest['folders']), {x['id'] for x in data['folders']})
        with sqlite3.connect(':memory:') as db:
            db.executescript((run.SNAPSHOT / 'vaultwarden.sql').read_text())
            self.assertEqual(db.execute('PRAGMA integrity_check').fetchone()[0], 'ok')
            self.assertEqual(db.execute('PRAGMA foreign_key_check').fetchall(), [])
            self.assertEqual(db.execute('SELECT count(*) FROM users').fetchone()[0], 2)
            self.assertEqual(db.execute('SELECT count(*) FROM ciphers').fetchone()[0], len(data['items']) + len(org_items))
            self.assertEqual(db.execute('SELECT count(*) FROM organizations').fetchone()[0], len(organizations['organizations']))
            self.assertEqual(db.execute('SELECT count(*) FROM collections').fetchone()[0],
                             sum(len(o['collections']) for o in organizations['organizations']))


if __name__ == '__main__':
    unittest.main()
