# Usage and implementation reference

Bitwarden를 CLI와 TUI로 사용하는 Go 클라이언트. Cobra CLI와 TUI → go-daemonkit RPC → 잠금 메모리 세션·영속 cache → Bitwarden 직접 API로 동작한다.

- [요구사항과 인터페이스 제안](./design.md)
- [Full integration 개발 환경](../dev/README.md)
- [예제 config](../dev/config.example.toml)

최종 설계 방향: CLI + TUI, go-daemonkit daemon, Bitwarden API 직접 호출.
서버 login과 로컬 unlock을 분리하고, 명령이 필요한 인증을 거쳐 재개된다.
풀린 키는 daemon 메모리, refresh token 또는 개인 API Key는 OS credential store, 암호화된 vault는 로컬 SQLite에 보관한다.
BW_SESSION이나 bw/rbw 실행 파일에 의존하지 않는다. 세부 설계와 구현 순서는 위 문서를 참고한다.

필수 [integration snapshot](../testdata/integration/README.md)은 8종·21개 항목을 포함한다. `dev/integration/run.py`가 전용 서버에 복원해 실제 API를 검사하고, Lefthook pre-commit이 매 커밋마다 실행한다. 실행 준비는 개발 가이드를 따른다.

## 현재 실행 범위

```sh
go build -o .dev/bwb .
.dev/bwb login --profile personal --server SERVER_URL --email EMAIL --password-stdin < PASSWORD_INPUT
.dev/bwb item read --profile personal ITEM_ID
.dev/bwb item read --profile personal ITEM_ID --field password
.dev/bwb item totp --profile personal ITEM_ID
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

`PASSWORD_INPUT`은 안전하게 공급한 stdin을 뜻한다. 비밀번호를 argv에 넣거나 echo가 켜진 터미널에 직접 입력하지 않는다. 마지막 LF/CRLF 한 개만 제거하며 공백은 유지한다.

### 항목 찾기 → 상세 보기 → 값 읽기

```sh
bwb item list --search titan
bwb item read 'titan-otp'
bwb item read ITEM_ID --field password
bwb item totp ITEM_ID
```

`ITEM`은 UUID 또는 대소문자를 무시한 정확한 이름이다. 동명 항목이 있으면 후보 ID와 소속을 출력하고 실패하므로 ID로 다시 선택한다. 부분 검색은 `item list --search`에서 한다. 조회는 기존 cache를 사용하며, 최신 서버 상태가 필요하면 각 명령에 `--sync`를 붙인다.

`item read ITEM`은 이름·타입·소속과 `FIELD / LABEL / VALUE` 표를 보여준다. 사용자명·URL·일반 custom field는 표시하고 비밀번호·TOTP seed·notes·카드/신원 정보·hidden custom field는 `[hidden]`으로 표시한다. 카드 brand는 일반 정보다. 화면에 나온 필드 키를 그대로 `--field`에 전달한다. 예: `username`, `password`, `notes`, `url/1`, `custom/token`. Custom 이름의 특수문자는 percent encode하고 동명 필드는 고유 suffix로 구분한다. 키를 직접 조합할 필요는 없다.

`item read ITEM --field KEY`는 그 값만 줄바꿈 추가 없이 출력한다. 빈 문자열도 성공이며 없는 키는 사용 가능한 키와 함께 오류를 반환한다. `--reveal`은 상세의 숨겨진 값을 표시한다. 두 옵션은 함께 사용하지 않는다. 항목별 재확인이나 조직의 조회 제한은 두 경우 모두 유지하며, 제한된 상세는 metadata와 제한 사유만 표시한다.

`item totp ITEM`은 현재 인증 코드만 출력한다. Seed를 읽으려면 `item read ITEM --field totp-secret`을 사용한다. Base32 및 `otpauth://totp`의 SHA1/SHA256/SHA512, 6/8자리 코드를 지원하며 기본 주기는 30초다. HOTP와 Steam 코드는 지원하지 않는다.

`item read/totp --format json`은 `{"version":1,"data":{"item":{...},"fields":[...]}}` 구조다. 각 필드는 `key`, `label`, `sensitive`, `state`, `value`를 갖는다. 숨김은 `state: "hidden", value: null`, 빈 값은 `state: "empty", value: ""`이며 JSON 자체는 값을 공개하지 않는다. 상세가 RPC JSON 응답 한도(48,000 bytes)를 넘으면 `--field`로 하나씩 읽는다. 단일 필드도 한도를 넘으면 실패한다.

기존 루트 `bwb read`와 `login/password` 같은 CLI 필드 경로는 제거했다. 호환 alias는 없다. 새 바이너리로 `.dev/bwb daemon update`를 실행한 뒤 사용한다.

첫 login이 같은 바이너리의 `daemon run`을 자동 시작한다. profile 기본값은 `default`. login은 API 인증·첫 sync 후 암호화 snapshot과 KDF/보호된 키/device ID를 daemonkit SQLite에 저장하고, refresh token 또는 개인 API Key를 OS keychain에 저장한다. 비밀번호·복호화 키·access token은 디스크에 저장하지 않는다. 같은 profile을 다른 서버·계정에 연결하려면 먼저 logout해야 한다.

`item read`는 별도 CLI 프로세스에서도 HTTP 없이 cached ciphertext를 복호화한다. `unlock`은 저장된 KDF·보호된 키로 비밀번호를 검증하므로 재시작 후에도 네트워크와 keychain 접근 없이 동작한다. `sync`는 필요할 때 keychain의 refresh token 또는 API Key로 access token을 발급하며 401 후 갱신은 한 번만 시도한다. 대화형 `item read/item totp/list/sync`는 인증 필요 오류를 받으면 CLI에서 인증한 뒤 원래 작업을 한 번만 재시도한다. API Key 재로그인은 명시적 `login --apikey`가 필요하다. token 만료는 cached read를 막지 않는다.

`lock`은 진행 중 작업을 취소하고 메모리 키·토큰·snapshot을 파기한다. 암호화 cache와 keychain 로그인은 남는다. `logout`은 선택 profile의 로컬 cache와 keychain 로그인도 삭제한다. 원격 vault나 원격 로그인은 삭제/폐기하지 않는다. daemon 종료는 전체 profile을 잠근다. 이미 반환한 평문·전송 중 응답은 회수하지 않는다.

`status`/`lock`은 daemon을 시작하지 않는다. status 출력은 `stopped`, `logged_out`, `locked`, `unlocked`, `logout_pending`. logout 삭제 실패는 오류로 반환하고 저장된 표시로 unlock을 막는다. 같은 profile에서 logout을 재시도하면 정리를 완료한다.

SQLite와 keychain은 분산 트랜잭션이 아니므로 credential 참조와 정리 목록을 먼저 기록한다. 인증 실패는 이전 cache를 유지하며, 미완료 credential 쓰기와 이전 credential은 다음 login/sync/logout에서 정리한다. 이미 커밋된 저장은 취소로 되돌리지 않는다. 저장된 로그인과 메모리 세션의 세대가 달라지면 sync는 재잠금·unlock을 요구해 이전 token으로 새 credential을 덮어쓰지 않는다. keychain 장애 시 평문 파일 fallback은 없다. OS keychain 호출은 진행 중 취소할 수 없어 lock/logout이 그 호출을 기다릴 수 있다. 서버 token 회전 직후 프로세스가 중단되면 재로그인이 필요할 수 있다. logout은 논리적 삭제이며 SQLite 잔여 페이지나 백업의 물리적 소거를 보장하지 않는다.

새 바이너리를 빌드한 뒤 `.dev/bwb daemon update`로 daemonkit이 선택한 실행 파일을 갱신한다. 다른 내용의 바이너리면 daemon이 재시작되어 unlock이 필요하다. 동일 바이너리는 no-op이다. config 파일과 profile별 idle timeout을 지원한다.

대화형 `login/unlock`은 `--password-stdin` 없이도 비밀번호를 입력받는다. `item read/item totp/list/sync`의 자동 인증은 stdin이 TTY이고 `--non-interactive`가 아닐 때만 실행한다. stdout이 pipe여도 prompt는 controlling terminal에 표시하므로 출력 데이터와 섞이지 않는다. 비TTY 실행은 입력을 소비하지 않고 인증 필요 오류를 반환한다. 최초 로그인 정보가 없는 profile은 `login`에서 서버 URL과 이메일을 입력한다. 명시적 `--server`/`--email`이나 저장된 값이 있으면 묻지 않는다. 주소와 이메일은 보이는 입력이며 잘못된 값은 다시 입력받는다. 비TTY/`--non-interactive`에서는 빠진 필드를 명시하도록 안내하며 stdin의 비밀번호를 소비하지 않는다. 서버는 로그인에 성공한 profile에만 저장하고 config에서는 지정하지 않는다.

자동 인증은 profile당 prompt 소유자 하나를 예약하고 다른 CLI는 결과를 기다린다. 소유자의 취소·연결 종료·lock/logout은 예약을 폐기하며 대기자에게 새 prompt를 연쇄 생성하지 않는다. 응답 직후 연결 감시를 시작하기 전 이탈 같은 경우를 위해 미사용 예약은 2분 뒤 만료된다. 비밀번호는 daemon 예약에 보관하지 않는다. 암호화 실패·네트워크·keychain 장애는 자동 로그인으로 바꾸지 않는다. 잠금과 서버 로그인 만료가 겹치면 1회 복구 한도 때문에 명시적 login이 필요할 수 있다.

현재 대화형 입력은 macOS/Linux 지원. 다른 OS에서는 명시적 `login/unlock --password-stdin`을 사용한다. 입력 중 Ctrl-C/SIGINT/lock을 처리한 뒤 터미널 echo를 복원한다.

### 개인 API Key 로그인

```sh
bwb login --apikey --profile work --server SERVER_URL
```

`client_id`, `client_secret`, 마스터 비밀번호를 숨김 입력으로 받는다. `--email`은 선택 사항이다. 최초 로그인에서 이메일을 생략하면 API Key 인증 후 서버 프로필에서 조회하고 계정 ID를 검증한다. config 또는 기존 로그인에 이메일이 있으면 그 값을 사용하며, 명시한 이메일도 실제 계정과 일치해야 한다. 저장된 로그인에 서버가 있으면 `--server`를 생략할 수 있다. 최초 대화형 로그인에서는 빠진 서버 URL을 입력받는다. 기존 login처럼 첫 sync와 unlock까지 성공해야 로그인을 저장한다. API Key는 서버 인증에 사용하고, 마스터 비밀번호는 로컬 복호화에만 사용한다. 개인 키(`user.<account UUID>`)만 지원하며 조직 API Key와 Secrets Manager 토큰은 받지 않는다.

자동화에서는 `BWB_CLIENTID`와 `BWB_CLIENTSECRET`을 안전하게 환경변수로 공급하고 다음 명령을 실행한다. `--password-stdin`에는 마스터 비밀번호만 전달한다.

```sh
bwb login --apikey --profile work --non-interactive --password-stdin < PASSWORD_INPUT
```

API Key 환경변수는 `--apikey`일 때만 읽고, daemon 시작 전에 현재 CLI 환경에서 제거한다. 비대화형 실행에서 값이 부족하면 stdin을 읽기 전에 실패한다. TOML이나 비밀값을 받는 명령행 옵션은 사용하지 않는다.

keychain에는 서버·이메일·계정 ID에 묶인 `client_id`와 `client_secret`을 저장한다. 메모리의 secret은 `memguard`로 보호하며 lock/종료 시 파기한다. unlock은 오프라인으로 가능하고, 다음 sync에서 keychain의 API Key로 재인증한다. 키가 폐기되면 `login --apikey`를 안내하며 비밀번호 로그인으로 전환하지 않는다. cached read는 계속 가능하다. logout은 로컬 키를 삭제하며 서버의 API Key 자체는 폐기하지 않는다.

마스터 비밀번호가 있는 PBKDF2 계정만 지원한다. Key Connector·마스터 비밀번호 없는 계정·Argon2id는 지원하지 않는다. 실행 중인 구 daemon은 `daemon update`로 갱신해야 한다. API Key profile은 v2 레코드를 사용해 구 바이너리가 거부하며, 기존 비밀번호 profile의 v1 레코드는 계속 읽는다. 구 버전으로 내리려면 새 버전에서 API Key profile을 먼저 logout한다.

### Okta / 조직 SSO 로그인

기존 API Key profile도 일반 `login`에서 SSO가 발견되거나 `login --sso`를 지정하면 같은 계정의 SSO로 전환할 수 있다. 성공 전에는 기존 cache와 자격증명을 유지하며, 취소·복호화·저장 실패는 기존 로그인 방식을 바꾸지 않는다. SSO가 발견되지 않으면 API Key 재로그인은 `login --apikey`로 진행한다.

```sh
bwb login --profile work --server https://vault.example.com --email you@example.com
# 자동 탐색이 지원되지 않거나 조직 도메인이 등록되지 않은 경우
bwb login --sso --profile work --server https://vault.example.com --identifier YOUR_ORG
# 첫 로그인 성공 이후
bwb login --profile work
```

일반 `login`은 비밀번호를 읽기 전에 이메일로 [검증된 조직 도메인 탐색 API](https://github.com/bitwarden/clients/blob/main/libs/common/src/admin-console/services/organization-domain/org-domain-api.service.ts)를 호출한다. 한 조직이 일치하면 SSO로 진행하고 identifier를 채운다. 빈 결과나 미지원 응답(404/405/501)은 비밀번호 로그인으로 진행하며, 네트워크·서버 장애나 잘못된 응답은 오류로 반환한다. 복수 조직은 `--sso --identifier`로 선택해야 한다. 이 조회는 계정의 SSO 강제 여부 전체를 알려주는 API가 아니므로, 도메인이 등록되지 않은 SSO 조직은 `--sso`를 사용한다. 명시적 SSO나 저장된 SSO profile은 탐색을 생략한다.

Vaultwarden의 고정 더미 identifier는 SSO 활성 여부를 뜻하지 않으므로 `/identity/sso/prevalidate`로 추가 확인한다. 비활성 응답은 비밀번호 로그인으로 진행하고, 검증 중 장애는 오류로 반환한다. ([Vaultwarden 구현](https://github.com/dani-garcia/vaultwarden/blob/1.37.4/src/api/identity.rs))

`--server`는 Okta 주소가 아닌 Bitwarden 웹 금고 주소다. 터미널에서 Bitwarden 마스터 비밀번호를 먼저 입력한 뒤, 브라우저에서 Bitwarden 조직에 설정된 SSO를 통해 Okta 로그인과 조직의 MFA를 완료한다. Okta 비밀번호와 client secret은 bwb가 받지 않는다. `--identifier`는 Bitwarden 관리 콘솔의 Single sign-on에 설정한 SSO 식별자이며 생략하면 브라우저에서 입력한다. `--sso`에서는 `--email`도 선택 사항이며 인증 후 서버 프로필에서 확인한다. 서버는 저장된 로그인에서, 이메일은 config 또는 저장된 로그인에서 가져온다. 최초 대화형 로그인에서 서버가 없으면 입력받고, 다른 계정으로 인증하면 저장하지 않는다. 성공한 SSO profile에는 로그인 방식과 identifier도 저장해 다음 `login --profile work`에서 재사용한다.

[공식 CLI의 브라우저 SSO 흐름](https://github.com/bitwarden/clients/blob/main/apps/cli/src/auth/commands/login.command.ts)에 맞춰 Authorization Code + PKCE를 사용한다. 콜백은 로컬 8065–8070 포트 중 사용 가능한 하나에서 받고, state와 중복 응답을 검증한다. 마스터 비밀번호 입력을 마친 뒤 브라우저를 열고, 콜백을 받으면 코드를 토큰으로 교환한 다음 미리 받은 비밀번호로 금고를 연다. 브라우저 대기 중 비밀번호는 CLI의 잠긴 메모리에만 보관하며, 사용·실패·취소 후 파기한다. 잘못된 비밀번호는 브라우저 인증을 반복하지 않고 재입력할 수 있으며, `--password-stdin`은 한 번만 읽는다. profile당 인증 소유자 하나만 브라우저를 열고 다른 CLI는 결과를 기다린다. 브라우저·토큰 교환·비밀번호 입력을 포함한 예약은 5분 뒤 만료되며, 취소·연결 종료·lock/logout·daemon 종료 때 대기 토큰을 파기하고 늦은 콜백/입력을 거부한다.

브라우저가 열리지 않으면 stderr의 URL을 같은 컴퓨터에서 직접 연다. `--apikey`와 함께 사용할 수 없고, `--non-interactive`는 브라우저 실행 전에 오류를 반환한다. 비TTY 입력은 `--password-stdin`으로 마스터 비밀번호만 공급할 수 있지만 브라우저 인증은 여전히 필요하다.

마스터 비밀번호는 로컬 금고 복호화에만 사용한다. 로그인·복호화·첫 sync가 모두 성공해야 기존 keychain/cache 저장 흐름으로 반영한다. SSO refresh token은 keychain에 저장하고, 재시작 후에도 `cli` OAuth client로 갱신한다. unlock은 오프라인으로 가능하다. SSO 세션이 만료되면 `login --sso`를 안내하며 비밀번호 로그인으로 전환하지 않는다. TUI에서는 로그인 후 기존 profile을 선택해 사용한다.

마스터 비밀번호가 있는 PBKDF2 계정을 지원한다. Key Connector, 마스터 비밀번호 없는 Trusted Device 계정, Argon2id, Okta 인증 이후 Bitwarden 자체가 추가로 요구하는 two-step login은 지원하지 않는다. 실제 Okta/Bitwarden 서버 연동은 아직 검증하지 않았으며, 모의 서버로 코드 교환·계정 검증·토큰 갱신·실패 복구를 검사한다. SSO profile은 v3 레코드, 분할 credential은 v4 레코드로 구 daemon의 오해석을 막는다. 새 바이너리에서 `bwb daemon update`를 실행하고, SSO 미지원 버전으로 내리기 전에는 해당 profile을 logout한다.

```text
main.go → internal/cli (Cobra + 인증 실행부)
              ↓ internal/vaultapi (typed RPC)
          internal/daemon (Manager, App, handlers, RPC 오류 변환)
              ↓
          internal/vault (profile, 인증 예약, 도메인 오류, session lifetime)
              ↓
          internal/bitwarden (HTTP, crypto, encrypted snapshot)
```

키와 access/refresh token의 장기 보관은 `memguard` locked memory를 사용한다. 평문 snapshot은 보관하지 않는다. crypto·HTTP·JSON·RPC 처리 중 임시 Go 문자열/버퍼는 생기며 모든 복사본의 즉시 zeroize를 보장하지 않는다. 메모리 보호 실패 시 평문 fallback은 없다. 같은 OS 사용자 프로세스는 daemon 접근 가능하다. daemonkit SQLite에는 암호화 vault와 계정 metadata만 기록하며 비밀번호·token은 넣지 않는다. cache record는 512 KiB 단위로 원자적으로 저장하며 최대 64 MiB다. Credential JSON은 최대 64 KiB이며, 2,560바이트를 넘으면 여러 OS keychain 항목으로 나눠 저장한다. 토큰 갱신은 새 조각을 모두 저장한 뒤 참조를 바꾸고, 미완료 쓰기와 이전 조각은 정리 목록으로 복구한다. 기존 단일 항목도 계속 읽으며 logout은 모든 조각을 삭제한다. 분할 credential이 있는 profile은 구 버전으로 내리기 전에 logout한다. Unix 경로는 `XDG_STATE_HOME`, 실행·keychain namespace는 `BWB_INSTANCE`로 격리한다. 테스트는 둘 모두 독립 값으로 설정한다.

지원 범위는 PBKDF2(100000–2000000 iterations), 개인·조직 vault, AES-CBC/HMAC `2.` 암호문, 8종의 기본 필드다. 마스터 비밀번호 재확인(reprompt) 항목도 login/sync/offline unlock과 이름 목록은 지원한다. 필드 조회는 항목별 마스터 비밀번호 재확인 후 허용한다. SSO는 위의 마스터 비밀번호 기반 흐름을 지원하며, Argon2id, Bitwarden 자체의 추가 two-step login, passkey는 아직 지원하지 않는다. 공식 Bitwarden 서버와의 상호운용성은 별도 검증이 필요하다.

```sh
go test ./...
mise run integration
```

integration은 실제 Vaultwarden에서 Go API 코드와 빌드된 CLI·daemon을 실행해 snapshot 기대값, profile 격리, offline read/unlock의 HTTP 0회, lock 중 로그인 취소, 재시작 후 keychain refresh, logout 영속성을 검사한다. 실제 PTY로 비밀 비노출·취소·동시 prompt 억제·자동 재로그인도 확인한다. Python은 서버 복원과 snapshot 생성에만 쓰며 제품의 암호화 구현은 Go 표준 라이브러리를 사용한다.

조직 vault는 RSA-OAEP SHA-1/SHA-256(`4.`/`3.`)으로 감싼 조직 키와 선택적 item key를 지원한다. 조직·collection 소속과 권한은 암호화 cache에 보존하며 재시작 후 offline unlock/read가 가능하다. 조직 키는 작업 중 보호 메모리에서 사용한 뒤 파기하고, RSA private key 파싱 중 Go 임시 할당의 즉시 zeroize는 보장하지 않는다. `viewPassword=false` 항목은 현재 이름만 읽을 수 있고 나머지 필드는 명시적 제한 오류다. Collection 권한을 합산하지 않고 서버가 계산한 항목 권한을 사용한다. 이 제한은 클라이언트 정책이며 서버가 암호문·키를 제공하는 사용자의 다른 프로그램까지 막는 보안 경계는 아니다. 조직/collection 관리는 아직 없다.

목록은 `item/folder/organization/collection list`로 조회한다. 기본은 테두리·색상 없는 표이며 item 열은 ID/TYPE/ORGANIZATION/FOLDER/COLLECTIONS/FAV/NAME이다. UUID와 이름을 생략하지 않고 제어문자는 escape한다. `--search`는 이름의 대소문자 무시 부분 검색, `--sort`는 name/type/folder/organization의 쉼표 조합(item 기준), `--reverse`는 정렬 기준의 역순이다. 동률은 ID 오름차순으로 고정한다. Folder/organization 목록은 name 정렬, collection 목록은 name/organization 정렬을 지원한다.

`--organization`, `--folder`, `--collection`은 UUID 필터이며 함께 사용하면 AND다. 없는 ID 또는 서로 맞지 않는 조직·collection은 오류다. `--group-by folder|organization|collection`은 표만 그룹화한다. Collection 다중 소속 항목은 각 그룹에 나타나며 동명 그룹은 ID로 구분한다. 미소속 항목도 별도 그룹으로 표시한다. 0건은 cache에 일치하는 항목이 없다는 안내와 `--sync` 힌트를 출력하고 성공한다. JSON은 빈 배열이다.

`--format json`은 `{"version":1,"data":[...]}` 형태의 평면 배열이다. 각 항목의 organization/folder는 `{id,name}`, collections는 같은 구조의 배열이며 없는 선택 필드는 생략한다. 비밀번호·메모·사용자명·URL·custom field는 목록에 포함하지 않는다. JSON은 원래 문자열을 보존하며 grouping으로 항목을 중복하지 않는다.

첫 목록 요청에서 이름·소속 metadata만 복호화해 profile 메모리에 index를 만든다. 원문 비밀 필드는 index에 넣지 않으며 후속 목록은 index를 재사용한다. 이름 자체는 민감할 수 있고 Go 메모리 index의 즉시 zeroize는 보장하지 않는다. 성공한 login/unlock/sync와 lock/logout/shutdown에서 index 참조를 폐기한다. RPC는 최대 128행·JSON 48 KiB씩 전달하며 cursor는 profile/query/index generation에 묶인다. 한 행이 한도를 넘으면 명시적으로 실패한다. CLI는 전 페이지를 받은 뒤 출력하며 도중 snapshot 변경은 부분 출력 없이 재실행을 요구한다. 자동 인증은 목록 전체를 최대 한 번 재시도한다.

Config는 `${XDG_CONFIG_HOME:-$HOME/.config}/bwb/config.toml`에서 읽는다(macOS도 동일). `--config > BWB_CONFIG > 기본 경로`로 한 파일만 선택하며 병합하지 않는다. 기본 파일이 없으면 기존 명령을 그대로 사용하고, 명시한 파일이 없으면 실패한다. TOML schema는 [예제](../dev/config.example.toml)를 따른다. Version 1, 존재하는 default_profile, `[a-z][a-z0-9_-]{0,31}` profile 이름, 이메일·비음수 lock_timeout을 검증한다. `server` 필드는 제거되었으므로 기존 파일에서 삭제하고 최초 `login`에서 입력하거나 `--server`로 전달한다. 생략된 timeout은 0s이며 알 수 없는 키(비밀번호/token 포함)는 거부한다. 파일을 만들거나 수정하는 명령은 아직 없다.

```sh
bwb config path
bwb config validate
bwb profile list
bwb profile show personal
bwb --profile personal login                 # 최초 서버 입력, 이후 저장된 서버 사용
bwb --config ./test.toml item list
```

Profile 선택은 `--profile > BWB_PROFILE > default_profile > default`다. 명시적 profile이나 BWB_PROFILE이 있으면 일반 명령은 TOML을 읽지 않는다. 기본 profile 결정과 `config/profile` 검사 명령만 파일을 읽는다. `profile list/show`는 디스크의 설정을 보여주며 실행 중 daemon 설정을 의미하지 않는다. Login의 생략된 서버는 저장된 인증 profile에서, 이메일은 daemon에 적용된 config 또는 저장된 profile에서 가져온다. 최초 대화형 로그인에서 없는 값만 입력받으며, 명시한 값이 저장된 identity와 다르면 거부한다. 잘못된 TOML에서도 `--profile NAME lock/logout/status`와 daemon stop은 실행 가능하다.

Daemon은 시작할 때 설정 경로를 확정하고 email/lock_timeout을 읽는다. 자동 시작 시 `--config` 또는 BWB_CONFIG의 경로를 전달한다. 이미 실행 중인 daemon의 파일 경로는 이후 CLI의 `--config`로 변경되지 않는다. 다른 경로를 쓰려면 daemon을 종료하고 새 경로로 시작한다. 재시작에도 같은 경로를 사용하려면 BWB_CONFIG를 설정하거나 `--config`를 다시 전달한다. `daemon run --config ...`는 foreground 실행 경로를 지정한다.

직접 파일을 수정한 뒤 `bwb config validate`, `bwb daemon reload`를 실행한다. Reload는 daemon 시작 시 선택한 파일 전체를 검증하고 성공할 때만 정책을 교체한다. 문법 오류·누락 파일·로그인된 profile의 identity 변경이나 삭제는 기존 정책을 유지한 채 거부한다. Identity 변경/삭제는 먼저 해당 profile을 logout한다. Logout은 인증 상태만 지우며 TOML 정의는 유지한다. Watcher는 없고 일반 조회는 정책을 적용하지 않는다. 정책은 DB에 저장하지 않으며, 이전 개발 버전의 DB policy 필드는 무시한다(인증 캐시는 유지).

`lock_timeout = "0s"`는 자동 잠금을 끈다. 양수면 login/unlock 완료부터 계산하고 성공한 read/list에서 갱신한다. Status·sync·reload·실패한 조회는 연장하지 않는다. Reload로 시간이 줄어 이미 만료됐다면 잠금 처리하고 요청 진입에서도 deadline을 검사한다. 만료 시 진행 작업 취소, 키·목록 index 파기를 수행하며 cache/keychain 로그인은 유지한다. 다음 대화형 read/list는 기존 자동 unlock 후 재개한다. Native keychain 호출은 진행 중 취소할 수 없어 실제 파기 완료가 늦어질 수 있다.

버전과 실행 상태 확인:

```sh
bwb version
bwb --version
bwb daemon status
bwb --profile personal status --verbose
```

`version`과 `--version`은 현재 CLI 바이너리의 버전을 표시하며 config나 daemon 연결이 필요 없다. Go가 빌드 시 기록한 모듈 버전을 go-daemonkit의 `Options.Version`에도 전달한다. HEAD에 `vA.B.C` 태그가 있으면 해당 버전을 사용하고, 미커밋 변경은 `+dirty`, 태그 없는 커밋은 Go pseudo-version, 빌드 정보가 없으면 `dev`로 표시한다. 실행 시 Git을 호출하지 않는다. `daemon status`의 `VERSION`은 go-daemonkit이 조회한 실행 중 daemon의 버전이므로 CLI와 다를 수 있다.

`daemon status`는 Ready인 daemon의 선택된 config 절대 경로를 표시한다(기본 파일이 없더라도 선택된 경로 표시). `status --verbose`는 적용 중인 서버·계정·lock_timeout과 UTC 잠금 예정 시각 `LOCK_AT`을 표시한다. 잠겼거나 자동 잠금이 꺼졌으면 예정 시각은 `-`다. Config 없는 legacy profile은 저장된 계정과 timeout 0s를 표시한다. 두 명령은 daemon을 자동 시작하거나 활동 시간을 연장하지 않는다. 기본 `status`의 한 줄 출력은 유지한다. TOML이 잘못된 경우에도 `daemon status`와 명시적 `--profile NAME status --verbose`로 현재 적용 상태를 확인할 수 있다.

릴리스는 `mise run release`로 실행한다. 게시 없는 빌드 검증은 `mise run release -- --snapshot --skip=publish`다. GitHub Actions도 `v*` 태그 push에서 같은 task를 호출한다. 지원 대상과 환경 요구사항은 [개발 문서](../dev/README.md#release와-hook-관리)를 참고한다.

## TUI Explorer

```sh
bwb --profile personal tui
```

터미널 입력·출력이 모두 필요하다. Explorer는 개인 folder, 조직 collection, 항목을 같은 tree에 표시한다. `/`로 fuzzy 검색하면 tree를 유지하면서 매치 없는 가지를 숨긴다. 디렉터리가 맞으면 하위 전체를, 항목이 맞으면 상위 경로와 해당 항목을 표시한다. 여러 collection에 속한 항목은 각 경로에 표시한다. `/`로 구분된 컨테이너 이름은 화면상 계층으로 표현하며 권한 상속을 의미하지 않는다. 같은 이름의 컨테이너도 ID를 구분한다.

- `↑↓`/`j,k`: 이동. `←→`/`h,l`: 접기·펼치기. `s`: 디렉터리 우선 이름 정렬 방향 전환.
- `/`: 검색 입력. `Enter` 또는 `Esc`: 검색어를 유지하고 tree로 복귀. Tree에서 `Esc`: 검색 해제 및 원래 위치 복원.
- `Tab`/`Shift+Tab`: tree/상세 pane. 좁은 화면에서는 포커스 pane만 표시한다.
- 항목에서 `Enter`: 상세로 이동. `v`: item 전체 reveal/hide. 필드에서 `Enter`: 긴 값 펼치기(숨김 상태는 유지). `y`: 선택 필드 즉시 복사. `Esc`/`q`: 펼친 값 닫기. `PgUp/PgDown`: 스크롤.
- `r`: Explorer 수동 sync. 성공 시 검색·선택·펼침·정렬을 유지해 갱신한다. 실패 시 기존 목록을 유지하며 표시 중인 비밀은 즉시 숨긴다. 진행 중 잠금·종료는 가능하다.
- `Ctrl+P`: Profiles. `Tab`: 목록/설정 pane 전환, 설정 pane에서 `↑↓`/`PgUp/PgDown` 스크롤. `Enter`/`u`: profile 선택·인증. `l`: 선택 profile 잠금. `x`: 확인 후 logout. `r`: daemon이 선택한 TOML 전체 reload.
- `Ctrl+L`: 현재 profile 잠금. `?`: 도움말. `q`: 종료(daemon은 기존 timeout 정책 유지). 비밀번호 창은 `Esc`로 취소한다.

Search·Explorer·Item은 같은 두께의 테두리를 사용하며 비활성 패널의 테두리만 dim 처리한다. 터미널 background 응답으로 dark/light 색상을 선택하고, 응답이 없으면 dark를 사용한다. 선택 행과 포커스는 별도로 표시하며 상태 메시지가 단축키 안내를 덮지 않는다.

항목 선택은 Item RPC로 일반 값을 표시하고 민감한 값은 마스킹한다. `v`로 전체 민감 값을 가져오며, 필드 이동·pane 전환은 reveal을 유지한다. item 변경·검색 시작·도움말·profile 전환·sync는 민감 값을 제거한다. reprompt 항목은 Enter로 비밀번호 모달을 열며, 조직 권한 제한은 그대로 유지한다. 500ms 상태 조회로 잠금을 감지하면 tree·검색어·상세·비밀을 화면과 UI 상태에서 제거하고 다시 인증한 뒤 opaque ID로 위치를 복원한다. 연결 오류는 데이터 제거 후 Enter로 재연결한다. 중단된 daemon은 기존 CLI config 경로로 다시 시작하며, 필요 시 인증 후 목록을 다시 읽는다. 실패하면 오류를 표시하고 Enter로 재시도할 수 있다. 연결 대기는 Esc로 취소하며 이전 작업은 재실행하지 않는다. 이미 표시된 터미널 내용과 Go 문자열의 물리적 소거는 보장하지 않는다.

검색·탐색 같은 실제 사용자 조작은 최대 초당 한 번 Activity RPC로 idle 시간을 갱신한다. 백그라운드 목록·상태 polling·TOTP 갱신은 시간을 연장하지 않는다. 항목 상세·reveal·복사는 CLI와 같은 조회 활동 정책을 따른다. CLI에서는 상세에 표시된 `url/1` 같은 키로 URL을 읽는다.

Profiles는 TOML에 정의된 profile과 SQLite에 저장된 로그인 profile을 함께 보여준다. 파일 값과 적용값을 비교하고 파일 오류를 표시하며, 파일 오류 중에도 잠금·logout이 가능하다. 편집·자동 sync는 아직 제공하지 않는다.


### TOTP

`bwb item totp ITEM` (또는 `bwb totp ITEM`)는 stdin·stdout이 모두 터미널이면 코드와 남은 시간을 표시하고 만료 시 자동 갱신한다. `y`는 표시 중인 현재 코드를 즉시 복사하며 만료 임박 여부로 기다리지 않는다. `q`/Esc는 닫기, Ctrl+C는 종료, Ctrl+L은 잠금이다.

파이프·리다이렉션 또는 `--non-interactive`에서는 현재 코드 한 개만 stdout에 출력한다. 최소 유효시간 대기는 없으므로 만료 직전 코드가 나올 수 있다.

TUI에서 item을 `v`로 reveal하면 `Authentication code` 행에 코드와 남은 시간이 표시되고 자동 갱신한다. 숨긴 상태에서도 해당 행의 `y`로 현재 코드를 바로 복사한다. `TOTP seed`는 item 전체 reveal에 포함되는 시드 필드다. 기존 `item read ITEM --field totp-secret`도 시드 원문 조회를 유지한다.

코드 생성은 daemon에서 수행한다. TOTP 갱신 RPC는 코드·만료시각만 전달하며 idle 시간을 연장하지 않는다. 잠금·연결 종료 시 코드를 제거한다. OS 클립보드에 복사한 값은 자동 갱신·삭제하지 않는다. macOS는 pbcopy, Linux는 설치된 클립보드 도구와 그래픽 세션이 필요하며 지원되지 않으면 오류를 표시한다.

Base32 시드 및 표준 `otpauth://totp` URI, SHA1/SHA256/SHA512, 6·8자리와 1–86400초 주기를 지원한다. HOTP·Steam 형식은 명시적 오류다. 생성 시각은 daemon의 시스템 시계를 따른다.

TOTP 대화형 화면도 연결 종료 시 코드를 제거하고 Enter로 재연결한다. 필요 시 다시 인증하며 새 코드를 조회한다.


### Daemon version updates

릴리스 바이너리는 CLI와 daemon에 같은 버전을 주입한다. 일반 명령이 daemon에 연결할 때, stdin·stdout이 모두 터미널이고 CLI의 안정 릴리스 버전이 더 높으면 설치된 CLI 바이너리로 daemon을 자동 교체한다. 원격 다운로드는 없다. 같은 버전·더 최신 daemon·dev·prerelease·비교 불가능 버전은 자동 교체하지 않는다. 상태·버전 조회는 업데이트를 실행하지 않는다.

비대화형·파이프·리다이렉션·`--non-interactive`에서는 기존 daemon을 유지하고 stderr에 한 번만 안내한다. stdout 데이터는 바꾸지 않는다.

```text
bwb: daemon v0.3.0 is older than CLI v0.4.0. Run: bwb daemon update
```

자동 교체는 실행 중 daemon의 config 경로를 유지하고 파일을 먼저 검증한다. 모든 profile의 메모리 세션과 진행 작업은 재시작으로 종료된다. 현재 대화형 명령은 필요한 인증을 거쳐 계속하며 다른 TUI는 기존 재연결 흐름을 따른다. daemonkit의 실행 토큰으로 버전 확인 이후의 재시작·수동 업데이트 충돌을 감지한다. 충돌 시 버전과 config 검증부터 다시 판단하며, 세 번 충돌하면 오류를 반환한다. 명시적 `daemon update`는 버전과 무관한 수동 교체 동작을 유지한다. 교체 실패는 오류로 반환하며 daemonkit의 바이너리·프로세스 롤백을 사용한다. 데이터 마이그레이션 롤백은 포함하지 않는다. 재시작은 disk config를 다시 읽으며 마지막 적용 값을 복원하지 않는다. 기본 config의 선택적 로드 의미도 유지한다. 구버전 daemon이 선택적 여부를 제공하지 않고 파일도 없으면 안전하게 자동 교체를 중단한다.


필드는 TUI 상세에서 `y`로 reveal 없이 복사할 수 있다. CLI는 `bwb item read ITEM --field password --copy`를 사용한다. 성공 메시지 `Copied`는 stderr에만 출력하며 stdout은 비운다. `--copy`는 단일 `--field`가 필요하며 `--reveal`·JSON 출력과 함께 사용할 수 없다. 클립보드 실패 시 평문 출력으로 대체하지 않는다. TUI에서 필드·pane 이동, 잠금, 도움말 전환 등으로 취소된 복사 요청의 결과는 클립보드에 쓰지 않는다. 복사한 값은 기존 TOTP와 마찬가지로 자동 삭제하지 않는다.


### Item master password reprompt

`reprompt=1` 항목은 목록에서 이름만 표시한다. CLI 조회·복사·TOTP는 대화형 비밀번호 재확인 후 계속하며, 비대화형 요청은 reprompt 오류를 반환한다. TUI에서는 항목에서 Enter 또는 v로 재확인 모달을 연다. Esc는 현재 위치를 유지하고 취소하며, 오답은 재시도를 요구한다. 확인 후에도 민감 값은 기본 마스킹된다.

검증은 현재 vault의 KDF와 보호된 키를 사용하며 서버 로그인·세션 교체를 하지 않는다. 임시 capability는 profile·item에 한정된다. TUI는 같은 항목의 필드 이동·pane 전환·reveal/hide·TOTP 갱신에 재사용한다. 항목 이탈·검색·도움말·profile 전환·sync·잠금·연결 종료·config reload에서 폐기한다. CLI는 명령 종료 시 폐기한다. 연결 감시가 시작되지 않은 capability는 2분 뒤 만료된다. capability는 비밀번호를 포함하지 않는 임의 bearer token이며 디스크에 저장하지 않는다.

```toml
[profiles.personal]
insecure_skip_reprompt = false
```

기본값은 false. true는 unlocked vault에서 항목 재확인만 생략하며, login/unlock·조직 권한 제한·기본 마스킹은 유지한다. 파일 수정 후 `bwb daemon reload`로 적용한다. Profiles 화면에서 파일 값과 적용값을 확인할 수 있다.
