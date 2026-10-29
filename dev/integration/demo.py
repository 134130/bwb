"""Persistent, isolated development vault using the integration snapshot."""
import json
import os
import shlex
import shutil
import subprocess
import sys
import uuid

import run

HOME = run.ROOT / '.dev/demo'
STATE = HOME / 'environment.json'


def environment(state):
    env = {k: v for k, v in os.environ.items() if not k.startswith('BWB_')}
    env.update(BWB_INSTANCE=state['instance'], BWB_CONFIG=str(HOME / 'config.toml'),
               XDG_STATE_HOME=str(HOME / 'state'))
    return env


def cli(state, *args, password=None):
    subprocess.run(['go', 'run', '.', *args], cwd=run.ROOT, env=environment(state),
                   input=password, text=True, check=True)


def save(state):
    temporary = STATE.with_suffix('.tmp')
    temporary.write_text(json.dumps(state))
    temporary.replace(STATE)


def up():
    manifest = run.snapshot_manifest()
    run.docker('info')
    if STATE.exists():
        state = json.loads(STATE.read_text())
        if state['snapshot'] != run.digest(run.SNAPSHOT / 'manifest.json'):
            raise RuntimeError('Snapshot changed; run mise r dev:down, then mise r dev:up')
        if run.docker('inspect', '-f', '{{.State.Running}}', state['container']) != 'true':
            raise RuntimeError('Development server stopped; run mise r dev:down, then mise r dev:up')
    else:
        HOME.mkdir(parents=True, exist_ok=True)
        data = HOME / 'server'
        data.mkdir()
        server = run.Server(str(data))
        state = dict(instance='demo-' + uuid.uuid4().hex, container=server.container,
                     snapshot=run.digest(run.SNAPSHOT / 'manifest.json'), ready=False)
        save(state)
        run.restore_snapshot(data / 'db.sqlite3')
        server.__enter__()
        state['url'] = server.url
        save(state)
    if not state.get('url'):
        raise RuntimeError('Incomplete setup; run mise r dev:down, then mise r dev:up')
    if not state['ready']:
        profiles = {'owner': manifest['email'], 'restricted': manifest['organizations']['restricted_email']}
        config = 'version = 1\ndefault_profile = "owner"\n'
        for profile, email in profiles.items():
            config += f'\n[profiles.{profile}]\nemail = {json.dumps(email)}\nlock_timeout = "0s"\n'
        (HOME / 'config.toml').write_text(config)
        for profile in profiles:
            cli(state, '--profile', profile, 'login', '--server', state['url'], '--password-stdin', '--non-interactive',
                password=manifest['password'] + '\n')
        state['ready'] = True
        save(state)
    env = environment(state)
    (HOME / 'env.sh').write_text('unset BWB_PROFILE BWB_CLIENTID BWB_CLIENTSECRET\n' + ''.join(
        f'export {key}={shlex.quote(env[key])}\n' for key in ('BWB_INSTANCE', 'BWB_CONFIG', 'XDG_STATE_HOME')))
    print('Development vault ready. Run: mise r dev:run -- tui')
    print('Direct execution: source .dev/demo/env.sh && go run . tui')
    print('Public test password: ' + manifest['password'])


def down():
    if not STATE.exists():
        print('No development environment to remove.')
        return
    state = json.loads(STATE.read_text())
    # Keep state on cleanup failure so native credentials remain recoverable.
    if (HOME / 'config.toml').exists():
        for profile in ('owner', 'restricted'):
            cli(state, '--profile', profile, 'logout', '--non-interactive')
        cli(state, 'daemon', 'stop')
    containers = run.docker('ps', '-aq', '--filter', 'name=^/' + state['container'] + '$')
    if containers:
        run.docker('rm', '-f', state['container'])
    shutil.rmtree(HOME)
    print('Development environment removed.')


def main():
    action, *args = sys.argv[1:]
    if action == 'up':
        up()
    elif action == 'down':
        down()
    elif action == 'run':
        if not STATE.exists() or not json.loads(STATE.read_text())['ready']:
            raise RuntimeError('Run mise r dev:up first')
        state = json.loads(STATE.read_text())
        os.chdir(run.ROOT)
        os.execvpe('go', ['go', 'run', '.', *args], environment(state))
    else:
        raise RuntimeError('Expected up, down or run')


if __name__ == '__main__':
    try:
        main()
    except (RuntimeError, AssertionError, subprocess.CalledProcessError) as error:
        sys.exit(str(error))
