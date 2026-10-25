# bwb

Bitwarden를 CLI와 TUI로 사용하는 Go 클라이언트. 현재는 Cobra CLI → go-daemonkit RPC → 잠금 메모리 세션·영속 cache → Bitwarden 직접 API까지 구현했다. TUI는 후속이다.

- [요구사항과 인터페이스 제안](docs/design.md)
- [Full integration 개발 환경](dev/README.md)
- [제안 config](dev/config.example.toml)

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
.dev/bwb status --profile personal
.dev/bwb lock --profile personal
.dev/bwb unlock --profile personal --password-stdin < PASSWORD_INPUT
.dev/bwb daemon stop
.dev/bwb unlock --profile personal --password-stdin < PASSWORD_INPUT
.dev/bwb logout --profile personal
```

`PASSWORD_INPUT`은 안전하게 공급한 stdin을 뜻한다. 비밀번호를 argv에 넣거나 echo가 켜진 터미널에 직접 입력하지 않는다. 마지막 LF/CRLF 한 개만 제거하며 공백은 유지한다. `read`는 줄바꿈을 추가하지 않는다. `name`, `notes`, `fields/NAME`(슬래시 포함), `TYPE/FIELD` 지원. 현재 RPC 응답 제한 때문에 필드 하나는 최대 10,000 bytes다.

첫 login이 같은 바이너리의 `daemon run`을 자동 시작한다. profile 기본값은 `default`. login은 API 인증·첫 sync 후 암호화 snapshot과 KDF/보호된 키/device ID를 daemonkit SQLite에 저장하고, refresh token은 OS keychain에 저장한다. 비밀번호·복호화 키·access token은 디스크에 저장하지 않는다. 같은 profile을 다른 서버·계정에 연결하려면 먼저 logout해야 한다.

`read`는 별도 CLI 프로세스에서도 HTTP 없이 cached ciphertext를 복호화한다. `unlock`은 저장된 KDF·보호된 키로 비밀번호를 검증하므로 재시작 후에도 네트워크와 keychain 접근 없이 동작한다. `sync`는 필요할 때 keychain의 refresh token을 사용하며 401 후 refresh는 한 번만 시도한다. 대화형 `read/sync`는 인증 필요 오류를 받으면 CLI에서 인증한 뒤 원래 작업을 한 번만 재시도한다. token 만료는 cached read를 막지 않는다.

`lock`은 진행 중 작업을 취소하고 메모리 키·토큰·snapshot을 파기한다. 암호화 cache와 keychain 로그인은 남는다. `logout`은 선택 profile의 로컬 cache와 keychain 로그인도 삭제한다. 원격 vault나 원격 로그인은 삭제/폐기하지 않는다. daemon 종료는 전체 profile을 잠근다. 이미 반환한 평문·전송 중 응답은 회수하지 않는다.

`status`/`lock`은 daemon을 시작하지 않는다. status 출력은 `stopped`, `logged_out`, `locked`, `unlocked`, `logout_pending`. logout 삭제 실패는 오류로 반환하고 저장된 표시로 unlock을 막는다. 같은 profile에서 logout을 재시도하면 정리를 완료한다.

SQLite와 keychain은 분산 트랜잭션이 아니므로 credential 참조와 정리 목록을 먼저 기록한다. 인증 실패는 이전 cache를 유지하며, 미완료 credential 쓰기와 이전 credential은 다음 login/sync/logout에서 정리한다. 이미 커밋된 저장은 취소로 되돌리지 않는다. 저장된 로그인과 메모리 세션의 세대가 달라지면 sync는 재잠금·unlock을 요구해 이전 token으로 새 credential을 덮어쓰지 않는다. keychain 장애 시 평문 파일 fallback은 없다. OS keychain 호출은 진행 중 취소할 수 없어 lock/logout이 그 호출을 기다릴 수 있다. 서버 token 회전 직후 프로세스가 중단되면 재로그인이 필요할 수 있다. logout은 논리적 삭제이며 SQLite 잔여 페이지나 백업의 물리적 소거를 보장하지 않는다.

새 바이너리를 빌드한 뒤 `.dev/bwb daemon update`로 daemonkit이 선택한 실행 파일을 갱신한다. 다른 내용의 바이너리면 daemon이 재시작되어 unlock이 필요하다. 동일 바이너리는 no-op이다. config 파일·idle timeout은 후속이다.

대화형 `login/unlock`은 `--password-stdin` 없이도 비밀번호를 입력받는다. `read/sync`의 자동 인증은 stdin이 TTY이고 `--non-interactive`가 아닐 때만 실행한다. stdout이 pipe여도 prompt는 controlling terminal에 표시하므로 출력 데이터와 섞이지 않는다. 비TTY 실행은 입력을 소비하지 않고 인증 필요 오류를 반환한다. 최초 로그인 정보가 없는 profile은 `login --server ... --email ...`로 먼저 준비한다.

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

조직 vault는 RSA-OAEP SHA-1/SHA-256(`4.`/`3.`)으로 감싼 조직 키와 선택적 item key를 지원한다. 조직·collection 소속과 권한은 암호화 cache에 보존하며 재시작 후 offline unlock/read가 가능하다. 조직 키는 작업 중 보호 메모리에서 사용한 뒤 파기하고, RSA private key 파싱 중 Go 임시 할당의 즉시 zeroize는 보장하지 않는다. `viewPassword=false` 항목은 현재 이름만 읽을 수 있고 나머지 필드는 명시적 제한 오류다. Collection 권한을 합산하지 않고 서버가 계산한 항목 권한을 사용한다. 이 제한은 클라이언트 정책이며 서버가 암호문·키를 제공하는 사용자의 다른 프로그램까지 막는 보안 경계는 아니다. 조직/collection 관리·목록 명령은 아직 없다.
