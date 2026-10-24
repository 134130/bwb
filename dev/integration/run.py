"""Disposable real-server contract tests. Test credentials/keys are intentionally public.

This is test tooling, not bwb's production API or cryptography implementation.
"""
import argparse
import base64
import hashlib
import hmac
import json
import os
from pathlib import Path
import sqlite3
import subprocess
import tempfile
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid

from cryptography.hazmat.primitives import hashes, padding
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.primitives.kdf.hkdf import HKDFExpand

ROOT = Path(__file__).resolve().parents[2]
SNAPSHOT = ROOT / 'testdata/integration'
IMAGE = 'vaultwarden/server:1.37.4@sha256:efb3cde962015fcc036b2ea625242248611943b27212ebf4392841fb30fad055'
EMAIL = 'bwb-integration@example.invalid'
PASSWORD = 'PUBLIC-INTEGRATION-ONLY-not-a-real-password!'
ITERATIONS = 600000
TYPES = {1: 'login', 2: 'secureNote', 3: 'card', 4: 'identity', 5: 'sshKey',
         6: 'bankAccount', 7: 'driversLicense', 8: 'passport'}


def require(condition, message):
    if not condition:
        raise AssertionError(message)


def b64(value):
    return base64.b64encode(value).decode()


def master_keys():
    master = hashlib.pbkdf2_hmac('sha256', PASSWORD.encode(), EMAIL.encode(), ITERATIONS)
    stretched = b''.join(HKDFExpand(algorithm=hashes.SHA256(), length=32, info=info).derive(master)
                         for info in (b'enc', b'mac'))
    password_hash = b64(hashlib.pbkdf2_hmac('sha256', master, PASSWORD.encode(), 1))
    return stretched, password_hash


def encrypt(value, key):
    iv = os.urandom(16)
    padder = padding.PKCS7(128).padder()
    padded = padder.update(value) + padder.finalize()
    cipher = Cipher(algorithms.AES(key[:32]), modes.CBC(iv)).encryptor()
    ciphertext = cipher.update(padded) + cipher.finalize()
    mac = hmac.digest(key[32:], iv + ciphertext, 'sha256')
    return '2.' + '|'.join(map(b64, (iv, ciphertext, mac)))


def decrypt(value, key):
    require(value.startswith('2.'), 'unsupported cipher type')
    iv, ciphertext, mac = [base64.b64decode(x, validate=True) for x in value[2:].split('|')]
    require(hmac.compare_digest(mac, hmac.digest(key[32:], iv + ciphertext, 'sha256')),
            'cipher authentication failed')
    cipher = Cipher(algorithms.AES(key[:32]), modes.CBC(iv)).decryptor()
    padded = cipher.update(ciphertext) + cipher.finalize()
    unpadder = padding.PKCS7(128).unpadder()
    return unpadder.update(padded) + unpadder.finalize()


def transform(value, fn):
    # Fixture payload strings are encrypted; enums, nulls and booleans retain their types.
    if isinstance(value, str):
        return fn(value)
    if isinstance(value, list):
        return [transform(x, fn) for x in value]
    if isinstance(value, dict):
        return {k: transform(v, fn) for k, v in value.items()}
    return value


def fixture():
    return json.loads((ROOT / 'testdata/bitwarden.json').read_text())


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def docker(*args):
    return subprocess.check_output(['docker', *args], text=True).strip()


class Server:
    def __init__(self, directory):
        self.directory = directory
        self.container = 'bwb-integration-' + uuid.uuid4().hex

    def __enter__(self):
        try:
            docker('run', '-d', '--name', self.container, '-p', '127.0.0.1::80',
                   '-v', f'{self.directory}:/data',
                   '-e', 'SIGNUPS_ALLOWED=true', '-e', 'INVITATIONS_ALLOWED=false',
                   '-e', 'PUSH_ENABLED=false', '-e', 'LOGIN_RATELIMIT_MAX_BURST=100',
                   '-e', 'ROCKET_LOG_LEVEL=off', IMAGE)
            port = docker('port', self.container, '80/tcp').split(':')[-1]
            self.url = 'http://127.0.0.1:' + port
            deadline = time.monotonic() + 60
            while time.monotonic() < deadline:
                try:
                    self.request('GET', '/alive')
                    return self
                except (urllib.error.URLError, TimeoutError, ConnectionError):
                    time.sleep(0.25)
            raise RuntimeError('Vaultwarden did not become ready')
        except BaseException:
            self.__exit__(None, None, None)
            raise

    def __exit__(self, *_):
        # Only this invocation's randomly named disposable container is removed.
        subprocess.run(['docker', 'rm', '-f', self.container], check=False, capture_output=True)

    def request(self, method, path, data=None, token=None, form=False):
        headers = {'Accept': 'application/json', 'Bitwarden-Client-Version': '2026.10.0'}
        if token:
            headers['Authorization'] = 'Bearer ' + token
        body = None
        if data is not None:
            body = (urllib.parse.urlencode(data) if form else json.dumps(data)).encode()
            headers['Content-Type'] = ('application/x-www-form-urlencoded' if form else 'application/json')
        request = urllib.request.Request(self.url + path, body, headers, method=method)
        with urllib.request.urlopen(request, timeout=10) as response:
            raw = response.read()
            return json.loads(raw) if raw else None

    def login(self, password_hash=None):
        return self.request('POST', '/identity/connect/token', {
            'grant_type': 'password', 'client_id': 'web', 'scope': 'api offline_access',
            'username': EMAIL, 'password': password_hash or master_keys()[1],
            'deviceType': 8, 'deviceIdentifier': '30000000-0000-4000-8000-000000000001',
            'deviceName': 'bwb-integration'}, form=True)


def seed(server):
    data = fixture()
    master, password_hash = master_keys()
    user_key = os.urandom(64)
    verification = server.request('POST', '/identity/accounts/register/send-verification-email',
                                  {'email': EMAIL, 'name': 'BWB integration'})
    server.request('POST', '/identity/accounts/register/finish', {
        'emailVerificationToken': verification,
        'email': EMAIL, 'masterPasswordHash': password_hash,
        'key': encrypt(user_key, master), 'kdf': 0, 'kdfIterations': ITERATIONS})
    auth = server.login()
    token = auth['access_token']
    sync = server.request('GET', '/api/sync', token=token)
    account_id = sync['profile']['id']
    folders = {}
    for folder in data['folders']:
        result = server.request('POST', '/api/folders',
                                {'name': encrypt(folder['name'].encode(), user_key)}, token)
        folders[folder['id']] = result['id']
    items = {}
    for item in data['items']:
        payload = {k: item[k] for k in ('type', 'favorite', 'reprompt')}
        payload.update({'organizationId': None, 'encryptedFor': account_id,
                        'folderId': folders.get(item['folderId'])})
        for field in ('name', 'notes', 'fields', TYPES[item['type']]):
            payload[field] = transform(item[field], lambda s: encrypt(s.encode(), user_key))
        result = server.request('POST', '/api/ciphers', payload, token)
        items[item['id']] = result['id']
    return {'version': 1, 'image': IMAGE, 'email': EMAIL, 'password': PASSWORD,
            'kdf': {'type': 0, 'iterations': ITERATIONS}, 'account_id': account_id,
            'folders': folders, 'items': items,
            'types': {str(k): v for k, v in TYPES.items()}}


def verify(server, manifest):
    prelogin = server.request('POST', '/identity/accounts/prelogin', {'email': EMAIL})
    require(prelogin['kdf'] == 0 and prelogin['kdfIterations'] == ITERATIONS, 'KDF mismatch')
    try:
        server.login('invalid-password-hash')
    except urllib.error.HTTPError as e:
        require(e.code in (400, 401), 'unexpected wrong-password response')
    else:
        raise AssertionError('wrong password accepted')
    auth = server.login()
    key = decrypt(auth['Key'] if 'Key' in auth else auth['key'], master_keys()[0])
    require(len(key) == 64, 'invalid user key')
    refreshed = server.request('POST', '/identity/connect/token', {
        'grant_type': 'refresh_token', 'client_id': 'web',
        'refresh_token': auth['refresh_token']}, form=True)
    sync = server.request('GET', '/api/sync', token=refreshed['access_token'])
    require(sync['profile']['id'] == manifest['account_id'], 'account changed')
    data = fixture()
    actual = {x['id']: x for x in sync['ciphers']}
    require(set(actual) == set(manifest['items'].values()), 'item IDs/count changed')
    require({x['type'] for x in actual.values()} == set(TYPES), 'missing cipher type')
    folders = {x['id']: decrypt(x['name'], key).decode() for x in sync['folders']}
    require(folders == {manifest['folders'][x['id']]: x['name'] for x in data['folders']},
            'folder mismatch')
    for expected in data['items']:
        item = actual[manifest['items'][expected['id']]]
        for field in ('type', 'favorite', 'reprompt'):
            require(item[field] == expected[field], f'{expected["id"]}: {field} mismatch')
        require(item['folderId'] == manifest['folders'].get(expected['folderId']), 'folder ID mismatch')
        for field in ('name', 'notes', 'fields', TYPES[expected['type']]):
            got = transform(item[field], lambda s: decrypt(s, key).decode())
            # Server may add optional null/default fields; all fixture fields must survive.
            compare(expected[field], got, f'{expected["id"]}/{field}')
    print(f'PASS: real login, bad password, refresh, sync, decrypt; {len(actual)} items; types 1–8')


def compare(expected, actual, path):
    if isinstance(expected, dict):
        require(isinstance(actual, dict), f'{path}: expected object')
        for k, value in expected.items():
            require(k in actual, f'{path}/{k}: missing')
            compare(value, actual[k], path + '/' + k)
    elif isinstance(expected, list):
        require(isinstance(actual, list) and len(expected) == len(actual), f'{path}: list mismatch')
        for i, (a, b) in enumerate(zip(expected, actual)):
            compare(a, b, f'{path}/{i}')
    else:
        require(expected == actual, f'{path}: value mismatch')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--rebuild', action='store_true', help='replace the committed synthetic SQL snapshot')
    args = parser.parse_args()
    with tempfile.TemporaryDirectory(prefix='bwb-integration-') as directory:
        db_path = Path(directory) / 'db.sqlite3'
        if args.rebuild:
            with Server(directory) as server:
                manifest = seed(server)
                verify(server, manifest)
            with sqlite3.connect(db_path) as db:
                sql = '\n'.join(db.iterdump()) + '\n'
            dump = SNAPSHOT / 'vaultwarden.sql'
            dump.write_text(sql)
            manifest['dump_sha256'] = digest(dump)
            manifest['fixture_sha256'] = digest(ROOT / 'testdata/bitwarden.json')
            (SNAPSHOT / 'manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')
            print('Snapshot rebuilt. Run again without --rebuild to verify restoration.')
        else:
            manifest = json.loads((SNAPSHOT / 'manifest.json').read_text())
            require(manifest['version'] == 1 and manifest['image'] == IMAGE, 'snapshot version/image mismatch')
            require(manifest['dump_sha256'] == digest(SNAPSHOT / 'vaultwarden.sql'), 'dump checksum mismatch')
            require(manifest['fixture_sha256'] == digest(ROOT / 'testdata/bitwarden.json'),
                    'fixture changed: rebuild and review snapshot')
            with sqlite3.connect(db_path) as db:
                db.executescript((SNAPSHOT / 'vaultwarden.sql').read_text())
                require(db.execute('PRAGMA integrity_check').fetchone()[0] == 'ok', 'corrupt snapshot')
            with Server(directory) as server:
                verify(server, manifest)


if __name__ == '__main__':
    main()
