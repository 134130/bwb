# bwb

Bitwarden를 CLI와 TUI로 사용하는 Go 클라이언트. 현재는 Cobra CLI → go-daemonkit RPC → 잠금 메모리 세션·영속 cache → Bitwarden 직접 API까지 구현했다. TUI는 후속이다.

- [요구사항과 인터페이스 제안](docs/design.md)
- [Full integration 개발 환경](dev/README.md)
- [예제 config](dev/config.example.toml)

최종 설계 방향: CLI + TUI, go-daemonkit daemon, Bitwarden API 직접 호출.
서버 login과 로컬 unlock을 분리하고, 명령이 필요한 인증을 거쳐 재개된다.
풀린 키는 daemon 메모리, refresh token은 OS credential store, 암호화된 vault는 로컬 SQLite에 보관한다.
BW_SESSION이나 bw/rbw 실행 파일에 의존하지 않는다. 세부 설계와 구현 순서는 위 문서를 참고한다.

필수 [integration snapshot](testdata/integration/README.md)은 8종·19개 항목을 포함한다. `dev/integration/run.py`가 전용 서버에 복원해 실제 API를 검사하고, Lefthook pre-commit이 매 커밋마다 실행한다. 실행 준비는 개발 가이드를 따른다.

## 현재 실행 범위

```sh
go build -o .dev/bwb .
.dev/bwb login --profile personal --server SERVER_URL --email EMAIL --password-stdin < PASSWORD_INPUT
.dev/bwb read --profile personal ITEM_ID login/password
.dev/bwb sync --profile personal
.dev/bwb item list --profile personal
.dev/bwb item list --profile personal --search github --sort type,name --reverse
.dev/bwb organization list --profile personal
.dev/bwb collection list --profile personal --organization ORGANIZATION_ID
.dev/bwb folder list --profile personal
.dev/bwb item list --profile personal --collection COLLECTION_ID --format json
.dev/bwb item list --profile personal --group-by collection
.dev/bwb status --profile personal
.dev/bwb lock --profile personal
.dev/bwb unlock --profile personal --password-stdin < PASSWORD_INPUT
.dev/bwb daemon stop
.dev/bwb unlock --profile personal --password-stdin < PASSWORD_INPUT
.dev/bwb logout --profile personal
```

`PASSWORD_INPUT`은 안전하게 공급한 stdin을 뜻한다. 비밀번호를 argv에 넣거나 echo가 켜진 터미널에 직접 입력하지 않는다. 마지막 LF/CRLF 한 개만 제거하며 공백은 유지한다. `read`는 줄바꿈을 추가하지 않는다. `name`, `notes`, `fields/NAME`(슬래시 포함), `TYPE/FIELD` 지원. 현재 RPC 응답 제한 때문에 필드 하나는 최대 10,000 bytes다.

첫 login이 같은 바이너리의 `daemon run`을 자동 시작한다. profile 기본값은 `default`. login은 API 인증·첫 sync 후 암호화 snapshot과 KDF/보호된 키/device ID를 daemonkit SQLite에 저장하고, refresh token은 OS keychain에 저장한다. 비밀번호·복호화 키·access token은 디스크에 저장하지 않는다. 같은 profile을 다른 서버·계정에 연결하려면 먼저 logout해야 한다.

`read`는 별도 CLI 프로세스에서도 HTTP 없이 cached ciphertext를 복호화한다. `unlock`은 저장된 KDF·보호된 키로 비밀번호를 검증하므로 재시작 후에도 네트워크와 keychain 접근 없이 동작한다. `sync`는 필요할 때 keychain의 refresh token을 사용하며 401 후 refresh는 한 번만 시도한다. 대화형 `read/list/sync`는 인증 필요 오류를 받으면 CLI에서 인증한 뒤 원래 작업을 한 번만 재시도한다. token 만료는 cached read를 막지 않는다.

`lock`은 진행 중 작업을 취소하고 메모리 키·토큰·snapshot을 파기한다. 암호화 cache와 keychain 로그인은 남는다. `logout`은 선택 profile의 로컬 cache와 keychain 로그인도 삭제한다. 원격 vault나 원격 로그인은 삭제/폐기하지 않는다. daemon 종료는 전체 profile을 잠근다. 이미 반환한 평문·전송 중 응답은 회수하지 않는다.

`status`/`lock`은 daemon을 시작하지 않는다. status 출력은 `stopped`, `logged_out`, `locked`, `unlocked`, `logout_pending`. logout 삭제 실패는 오류로 반환하고 저장된 표시로 unlock을 막는다. 같은 profile에서 logout을 재시도하면 정리를 완료한다.

SQLite와 keychain은 분산 트랜잭션이 아니므로 credential 참조와 정리 목록을 먼저 기록한다. 인증 실패는 이전 cache를 유지하며, 미완료 credential 쓰기와 이전 credential은 다음 login/sync/logout에서 정리한다. 이미 커밋된 저장은 취소로 되돌리지 않는다. 저장된 로그인과 메모리 세션의 세대가 달라지면 sync는 재잠금·unlock을 요구해 이전 token으로 새 credential을 덮어쓰지 않는다. keychain 장애 시 평문 파일 fallback은 없다. OS keychain 호출은 진행 중 취소할 수 없어 lock/logout이 그 호출을 기다릴 수 있다. 서버 token 회전 직후 프로세스가 중단되면 재로그인이 필요할 수 있다. logout은 논리적 삭제이며 SQLite 잔여 페이지나 백업의 물리적 소거를 보장하지 않는다.

새 바이너리를 빌드한 뒤 `.dev/bwb daemon update`로 daemonkit이 선택한 실행 파일을 갱신한다. 다른 내용의 바이너리면 daemon이 재시작되어 unlock이 필요하다. 동일 바이너리는 no-op이다. config 파일과 profile별 idle timeout을 지원한다.

대화형 `login/unlock`은 `--password-stdin` 없이도 비밀번호를 입력받는다. `read/list/sync`의 자동 인증은 stdin이 TTY이고 `--non-interactive`가 아닐 때만 실행한다. stdout이 pipe여도 prompt는 controlling terminal에 표시하므로 출력 데이터와 섞이지 않는다. 비TTY 실행은 입력을 소비하지 않고 인증 필요 오류를 반환한다. 최초 로그인 정보가 없는 profile은 `login --server ... --email ...`로 먼저 준비한다.

자동 인증은 profile당 prompt 소유자 하나를 예약하고 다른 CLI는 결과를 기다린다. 소유자의 취소·연결 종료·lock/logout은 예약을 폐기하며 대기자에게 새 prompt를 연쇄 생성하지 않는다. 응답 직후 연결 감시를 시작하기 전 이탈 같은 경우를 위해 미사용 예약은 2분 뒤 만료된다. 비밀번호는 daemon 예약에 보관하지 않는다. 암호화 실패·네트워크·keychain 장애는 자동 로그인으로 바꾸지 않는다. 잠금과 서버 로그인 만료가 겹치면 1회 복구 한도 때문에 명시적 login이 필요할 수 있다.

현재 대화형 입력은 macOS/Linux 지원. 다른 OS에서는 명시적 `login/unlock --password-stdin`을 사용한다. 입력 중 Ctrl-C/SIGINT/lock을 처리한 뒤 터미널 echo를 복원한다.

```text
main.go → internal/cli (Cobra + 인증 실행부)
              ↓ internal/vaultapi (typed RPC)
          internal/daemon (Manager, App, handlers, RPC 오류 변환)
              ↓
          internal/vault (profile, 인증 예약, 도메인 오류, session lifetime)
              ↓
          internal/bitwarden (HTTP, crypto, encrypted snapshot)
```

키와 access/refresh token의 장기 보관은 `memguard` locked memory를 사용한다. 평문 snapshot은 보관하지 않는다. crypto·HTTP·JSON·RPC 처리 중 임시 Go 문자열/버퍼는 생기며 모든 복사본의 즉시 zeroize를 보장하지 않는다. 메모리 보호 실패 시 평문 fallback은 없다. 같은 OS 사용자 프로세스는 daemon 접근 가능하다. daemonkit SQLite에는 암호화 vault와 계정 metadata만 기록하며 비밀번호·token은 넣지 않는다. cache record는 512 KiB 단위로 원자적으로 저장하며 최대 64 MiB다. Unix 경로는 `XDG_STATE_HOME`, 실행·keychain namespace는 `BWB_INSTANCE`로 격리한다. 테스트는 둘 모두 독립 값으로 설정한다.

지원 범위는 PBKDF2(100000–2000000 iterations), 개인·조직 vault, AES-CBC/HMAC `2.` 암호문, 8종의 기본 필드다. Argon2id, 2FA/SSO, reprompt, passkey는 아직 지원하지 않는다. 공식 Bitwarden 서버와의 상호운용성은 별도 검증이 필요하다.

```sh
go test ./...
mise run integration
```

integration은 실제 Vaultwarden에서 Go API 코드와 빌드된 CLI·daemon을 실행해 snapshot 기대값, profile 격리, offline read/unlock의 HTTP 0회, lock 중 로그인 취소, 재시작 후 keychain refresh, logout 영속성을 검사한다. 실제 PTY로 비밀 비노출·취소·동시 prompt 억제·자동 재로그인도 확인한다. Python은 서버 복원과 snapshot 생성에만 쓰며 제품의 암호화 구현은 Go 표준 라이브러리를 사용한다.

조직 vault는 RSA-OAEP SHA-1/SHA-256(`4.`/`3.`)으로 감싼 조직 키와 선택적 item key를 지원한다. 조직·collection 소속과 권한은 암호화 cache에 보존하며 재시작 후 offline unlock/read가 가능하다. 조직 키는 작업 중 보호 메모리에서 사용한 뒤 파기하고, RSA private key 파싱 중 Go 임시 할당의 즉시 zeroize는 보장하지 않는다. `viewPassword=false` 항목은 현재 이름만 읽을 수 있고 나머지 필드는 명시적 제한 오류다. Collection 권한을 합산하지 않고 서버가 계산한 항목 권한을 사용한다. 이 제한은 클라이언트 정책이며 서버가 암호문·키를 제공하는 사용자의 다른 프로그램까지 막는 보안 경계는 아니다. 조직/collection 관리는 아직 없다.

목록은 `item/folder/organization/collection list`로 조회한다. 기본은 테두리·색상 없는 표이며 item 열은 ID/TYPE/ORGANIZATION/FOLDER/COLLECTIONS/FAV/NAME이다. UUID와 이름을 생략하지 않고 제어문자는 escape한다. `--search`는 이름의 대소문자 무시 부분 검색, `--sort`는 name/type/folder/organization의 쉼표 조합(item 기준), `--reverse`는 정렬 기준의 역순이다. 동률은 ID 오름차순으로 고정한다. Folder/organization 목록은 name 정렬, collection 목록은 name/organization 정렬을 지원한다.

`--organization`, `--folder`, `--collection`은 UUID 필터이며 함께 사용하면 AND다. 없는 ID 또는 서로 맞지 않는 조직·collection은 오류다. `--group-by folder|organization|collection`은 표만 그룹화한다. Collection 다중 소속 항목은 각 그룹에 나타나며 동명 그룹은 ID로 구분한다. 미소속 항목도 별도 그룹으로 표시한다. 0건은 헤더만 출력하고 성공한다.

`--format json`은 `{"version":1,"data":[...]}` 형태의 평면 배열이다. 각 항목의 organization/folder는 `{id,name}`, collections는 같은 구조의 배열이며 없는 선택 필드는 생략한다. 비밀번호·메모·사용자명·URL·custom field는 목록에 포함하지 않는다. JSON은 원래 문자열을 보존하며 grouping으로 항목을 중복하지 않는다.

첫 목록 요청에서 이름·소속 metadata만 복호화해 profile 메모리에 index를 만든다. 원문 비밀 필드는 index에 넣지 않으며 후속 목록은 index를 재사용한다. 이름 자체는 민감할 수 있고 Go 메모리 index의 즉시 zeroize는 보장하지 않는다. 성공한 login/unlock/sync와 lock/logout/shutdown에서 index 참조를 폐기한다. RPC는 최대 128행·JSON 48 KiB씩 전달하며 cursor는 profile/query/index generation에 묶인다. 한 행이 한도를 넘으면 명시적으로 실패한다. CLI는 전 페이지를 받은 뒤 출력하며 도중 snapshot 변경은 부분 출력 없이 재실행을 요구한다. 자동 인증은 목록 전체를 최대 한 번 재시도한다.

Config는 `${XDG_CONFIG_HOME:-$HOME/.config}/bwb/config.toml`에서 읽는다(macOS도 동일). `--config > BWB_CONFIG > 기본 경로`로 한 파일만 선택하며 병합하지 않는다. 기본 파일이 없으면 기존 명령을 그대로 사용하고, 명시한 파일이 없으면 실패한다. TOML schema는 [예제](dev/config.example.toml)를 따른다. Version 1, 존재하는 default_profile, `[a-z][a-z0-9_-]{0,31}` profile 이름, 서버·이메일·비음수 lock_timeout을 검증한다. 생략된 timeout은 0s이며 알 수 없는 키(비밀번호/token 포함)는 거부한다. 파일을 만들거나 수정하는 명령은 아직 없다.

```sh
bwb config path
bwb config validate
bwb profile list
bwb profile show personal
bwb --profile personal login                 # config의 서버·이메일 사용
bwb --config ./test.toml item list
```

Profile 선택은 `--profile > BWB_PROFILE > default_profile > default`다. 명시적 profile이나 BWB_PROFILE이 있으면 일반 명령은 TOML을 읽지 않는다. 기본 profile 결정과 `config/profile` 검사 명령만 파일을 읽는다. `profile list/show`는 디스크의 설정을 보여주며 실행 중 daemon 설정을 의미하지 않는다. Login의 생략된 server/email은 daemon에 적용된 profile에서 가져오며, 명시한 값이 해당 identity와 다르면 거부한다. 잘못된 TOML에서도 `--profile NAME lock/logout/status`와 daemon stop은 실행 가능하다.

Daemon은 시작할 때 설정 경로를 확정하고 server/email/lock_timeout을 읽는다. 자동 시작 시 `--config` 또는 BWB_CONFIG의 경로를 전달한다. 이미 실행 중인 daemon의 파일 경로는 이후 CLI의 `--config`로 변경되지 않는다. 다른 경로를 쓰려면 daemon을 종료하고 새 경로로 시작한다. 재시작에도 같은 경로를 사용하려면 BWB_CONFIG를 설정하거나 `--config`를 다시 전달한다. `daemon run --config ...`는 foreground 실행 경로를 지정한다.

직접 파일을 수정한 뒤 `bwb config validate`, `bwb daemon reload`를 실행한다. Reload는 daemon 시작 시 선택한 파일 전체를 검증하고 성공할 때만 정책을 교체한다. 문법 오류·누락 파일·로그인된 profile의 identity 변경이나 삭제는 기존 정책을 유지한 채 거부한다. Identity 변경/삭제는 먼저 해당 profile을 logout한다. Logout은 인증 상태만 지우며 TOML 정의는 유지한다. Watcher는 없고 일반 조회는 정책을 적용하지 않는다. 정책은 DB에 저장하지 않으며, 이전 개발 버전의 DB policy 필드는 무시한다(인증 캐시는 유지).

`lock_timeout = "0s"`는 자동 잠금을 끈다. 양수면 login/unlock 완료부터 계산하고 성공한 read/list에서 갱신한다. Status·sync·reload·실패한 조회는 연장하지 않는다. Reload로 시간이 줄어 이미 만료됐다면 잠금 처리하고 요청 진입에서도 deadline을 검사한다. 만료 시 진행 작업 취소, 키·목록 index 파기를 수행하며 cache/keychain 로그인은 유지한다. 다음 대화형 read/list는 기존 자동 unlock 후 재개한다. Native keychain 호출은 진행 중 취소할 수 없어 실제 파기 완료가 늦어질 수 있다.

실행 상태 확인:

```sh
bwb daemon status
bwb --profile personal status --verbose
```

`daemon status`는 Ready인 daemon의 선택된 config 절대 경로를 표시한다(기본 파일이 없더라도 선택된 경로 표시). `status --verbose`는 적용 중인 서버·계정·lock_timeout과 UTC 잠금 예정 시각 `LOCK_AT`을 표시한다. 잠겼거나 자동 잠금이 꺼졌으면 예정 시각은 `-`다. Config 없는 legacy profile은 저장된 계정과 timeout 0s를 표시한다. 두 명령은 daemon을 자동 시작하거나 활동 시간을 연장하지 않는다. 기본 `status`의 한 줄 출력은 유지한다. TOML이 잘못된 경우에도 `daemon status`와 명시적 `--profile NAME status --verbose`로 현재 적용 상태를 확인할 수 있다.

릴리스는 `mise run release`로 실행한다. 게시 없는 빌드 검증은 `mise run release -- --snapshot --skip=publish`다. GitHub Actions도 `v*` 태그 push에서 같은 task를 호출한다. 지원 대상과 환경 요구사항은 [개발 문서](dev/README.md#release와-hook-관리)를 참고한다.
