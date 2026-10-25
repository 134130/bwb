# bwb 설계

상태: 목표 설계. 첫 구현은 Cobra read + Go API/PBKDF2 복호화와 실제 snapshot integration까지다. 아래 daemon/profile/TUI 계약은 아직 미구현이며 현재 사용법은 README를 따른다.

## 1. 방향과 첫 범위

CLI와 TUI가 go-daemonkit 기반 daemon을 공유하고, daemon이 Bitwarden Password Manager API를 직접 호출한다. 공식 bw/rbw 실행 파일, bw serve, BW_SESSION 가져오기는 사용하지 않는다.

핵심은 **서버 로그인과 로컬 vault unlock의 분리**다. 명령은 필요한 상태만 확보하고 이어서 실행한다. 반복 조회는 네트워크와 외부 프로세스 실행 없이 처리한다.

첫 범위는 macOS/Linux, 개인 vault, 자체 로그인/로그아웃, unlock/lock, sync, 목록/상세/필드 조회, 읽기 전용 TUI다. master-password 로그인과 PBKDF2/Argon2id, TOTP 2FA를 첫 검증 대상으로 한다. 조직 vault, SSO/WebAuthn, 편집, attachments, SSH agent, 환경 주입은 후속이다. 지원하지 않는 인증·키·암호문 형식은 명시적 오류로 처리한다.

Full integration은 실제 서버의 인증·동기화·복호화·조회·재인증을 통과하는 것이다. 버전 관리된 필수 SQL dump snapshot을 매번 독립 서버에 복원하고 Lefthook pre-commit에서 매 커밋마다 검증한다. 수동/선택적 import와 HTTP mock만으로 완료를 선언하지 않는다.

## 2. 상태와 저장 위치

인증 상태(`unauthenticated`, `authenticated`, `reauth_required`)와 vault 상태(`empty`, `locked`, `unlocked`)는 독립적이다. authenticated는 인증 자료가 있다는 뜻이며 원격 서버가 지금도 허용한다는 보장은 아니다.

| 자료 | 저장 위치 | lock / daemon 종료 |
| --- | --- | --- |
| 마스터 비밀번호, 중간 KDF 키 | 필요한 동안만 메모리 | 작업 종료 즉시 파기 |
| 풀린 user/item/organization key | daemon의 보호된 메모리 | 파기 |
| 검색 index, 복호화 값 | daemon/요청의 메모리 | 파기 |
| API access token | daemon 메모리 | lock에서는 유지, 종료 시 파기 |
| API refresh token | OS credential store | 유지; logout 시 삭제 |
| 암호화된 vault, protected key, KDF 설정 | daemonkit SQLite | 유지; logout 시 삭제 |
| 서버·계정 ID, device ID, last_sync | daemonkit SQLite | 유지; logout 시 계정 데이터 삭제 |
| profile 설정 | TOML config | 유지 |

refresh token은 daemonkit `app.Secrets()`를 사용한다. plaintext 파일 fallback은 없다. credential store 사용 불가 시 로그인 저장에 실패했음을 알리고 완료로 보고하지 않는다. access token과 refresh token은 vault 복호화 키가 아니다. 보호된 키와 암호문은 비밀의 평문과 구분한다.

이 저장 정책은 재로그인 없이 재시작 후 refresh하고, 오프라인에서도 비밀번호로 cached vault를 unlock하기 위한 제안이다. rbw의 token JSON 저장 대신 이미 사용 중인 daemonkit의 OS credential store를 재사용한다. 지원 grant가 refresh token을 제공하지 않으면 필요한 때 재로그인하며 임의로 API secret 저장을 추가하지 않는다.

`lock`은 로컬 비밀 접근을 막는다. `logout`은 lock + 로컬 인증 자료 및 cached vault 제거다. 둘 다 이미 출력된 값이나 다른 장치의 세션을 회수하지 않는다. 서버 token 폐기는 해당 서버가 지원하고 검증된 경우에만 별도 동작으로 명시한다.

## 3. 명령 실행과 인증 hook

```text
CLI / TUI → daemon에 작업 요청
  ├─ 목록/조회 + cache 있음
  │    └─ locked → unlock_required → 비밀번호 입력 → 로컬 unlock → 조회
  ├─ 목록/조회 + cache 없음
  │    └─ login_required → 로그인 → 최초 sync → unlock → 조회
  └─ sync
       └─ access token 없음/만료 → refresh
            └─ refresh 불가 → login_required → 로그인 → sync
```

| 명령 | API 로그인 필요 | vault unlock 필요 |
| --- | --- | --- |
| status / profile list | 아니오 | 아니오 |
| item list/get/read, TUI 탐색 | 기존 cache가 있으면 아니오 | 예 |
| sync | 예 | 아니오 |
| login | 예 | 아니오; 입력 비밀번호를 장기 보관하지 않음 |
| unlock | cache가 있으면 아니오 | 예 |
| lock / logout | 아니오 | 아니오 |

로그인은 protected key와 KDF metadata를 확보한다. 기존 cache의 키 세대와 일치하지 않으면 새 metadata만 덮어쓰지 않고 후속 sync를 요구한다. 최초 조회 흐름은 한 인증 작업 안에서 받은 비밀번호로 login·sync·unlock을 마친 뒤 비밀번호를 파기해 중복 prompt를 피한다. 현재 명시적 login은 sync와 메모리 unlock까지 수행한다.

hook은 임의 shell script가 아니라 **CLI/TUI의 내장 인증 조정 흐름**이다. daemon은 `login_required`, `unlock_required`, 2FA challenge를 반환하고 UI가 입력을 받는다. daemon이 직접 터미널 prompt를 띄우지 않는다.

- profile당 인증 흐름 하나. daemon이 작업 ID와 소유 연결을 정하고 나머지 요청은 결과를 기다린다. 입력은 해당 작업·연결·profile·generation에 묶는다.
- prompt 소유 연결은 별도 대기 RPC로 감시하며 끊기거나 취소되면 challenge를 폐기한다. 예약 응답 후 감시 시작 전 이탈은 2분 만료로 정리한다. 다른 대기자가 자동으로 prompt를 연쇄 생성하지 않고 취소 결과를 받는다. 대기자 하나의 취소는 다른 작업을 취소하지 않는다.
- 인증 완료 후 원래 읽기 작업을 한 번 재개한다. 재인증이 또 필요하면 오류로 끝내며 무한 loop하지 않는다. 잘못된 비밀번호/2FA/사용자 취소도 원래 작업으로 실패를 전달한다.
- 비TTY 또는 `--non-interactive`에서는 prompt하지 않고 인증 필요 오류를 반환한다. stdout이 pipe여도 controlling terminal이 있고 명시적으로 non-interactive가 아니면 그 terminal에서 입력한다. 비밀 입력은 echo하지 않는다.
- 자동 refresh는 profile당 하나만 수행한다. 401에는 refresh 후 한 번 재시도하고, invalid/revoked refresh는 credential store에서 제거한 뒤 login을 요청한다. 네트워크 실패는 로그인 만료로 오인하지 않는다.
- refresh token rotation은 새 token을 저장한 뒤 성공 처리한다. 저장 실패/크래시로 이전 token이 무효화될 수 있으며 다음 작업은 재로그인으로 회복한다.
- 원격 인증 실패 자체가 기존 로컬 cache의 복호화 가능 여부를 바꾸지는 않는다. 오프라인 조회는 허용하며 원격 삭제·권한 회수는 다음 sync까지 반영되지 않는다.
- 후속 쓰기 명령은 결과가 불명확할 때 자동 재실행하지 않는다.

## 4. CLI 계약

```sh
bwb --help
bwb tui --profile work
bwb status --profile work
bwb login --profile work
bwb unlock --profile work
bwb lock --profile work
bwb logout --profile work
bwb sync --profile work
bwb item list --profile work --search github --format json
bwb item get ITEM_ID --profile work
bwb item get ITEM_ID --profile work --reveal --format json
bwb read 'bwb://work/ITEM_ID/login/password'
bwb read 'bwb://work/ITEM_ID/fields/api_token'
bwb profile list
bwb profile show work
bwb profile use work
bwb config path
bwb config validate
bwb daemon status
bwb daemon stop
```

- 인자 없음은 help. TUI는 명시적으로 시작한다. status/daemon status는 daemon을 시작하지 않는다. 다른 작업은 필요하면 시작한다.
- profile 우선순위는 `--profile > BWB_PROFILE > default_profile`. 없는 profile은 오류. URI의 profile과 명시적 flag/env가 충돌하면 오류이며 default profile은 무시한다.
- ID는 UUID. 이름 검색은 탐색용이며 중복 이름을 첫 결과로 조용히 해결하지 않는다.
- reference는 `bwb://PROFILE/UUID/login/password`, `.../login/username`, `.../notes`, `.../fields/NAME`부터 지원한다. segment는 한 번만 percent decode하고 중복 custom field 이름은 ambiguity 오류다.
- 목록은 ID/name/type/folder/favorite. 상세는 password/TOTP seed/hidden field/notes/카드·신원 정보를 기본 마스킹한다. JSON 선택 자체는 reveal을 뜻하지 않는다.
- read는 명시적인 원문 출력 요청이다. 줄바꿈을 추가하지 않으며 빈 값은 성공, 없는 필드는 오류다.
- stdout은 결과만, stderr는 진단만. JSON 결과는 `{version: 1, data: ...}`, JSON 오류는 stderr의 `{code, message}`다. backend 오류 원문을 노출하지 않는다.
- 종료 코드: 0 성공, 1 backend/내부 실패, 2 입력/config 오류, 3 인증 필요/실패, 4 item/field 없음, 5 모호한 선택, 130 취소. RPC 오류 코드는 CLI에서 매핑한다.
- logout은 선택 profile의 로컬 cache만 삭제한다. 원격 vault item은 삭제하지 않는다. 복수 profile 전체 logout/purge는 첫 범위에서 제외한다.

## 5. Profile / config

[예제](../dev/config.example.toml)는 비밀을 포함하지 않는 TOML이다.

- profile = 서버 + 계정 + 해당 인증/cache. folder/collection이나 dev/prod override 집합이 아니다. profile 상속/합성과 프로젝트 config 자동 탐색은 없다.
- config 경로는 `${XDG_CONFIG_HOME:-$HOME/.config}/bwb/config.toml`. `--config > BWB_CONFIG > 기본 경로` 순서로 한 파일만 읽는다.
- 이름은 `[a-z][a-z0-9_-]{0,31}`. 알 수 없는 config 키·schema version·잘못된 duration은 거부한다.
- config는 version/default_profile와 profile별 server/email/lock_timeout만 가진다. token/password/키는 금지한다.
- `lock_timeout = "0s"`는 명시적 lock 또는 daemon 종료까지 복호화 키를 유지한다. 양수는 마지막 성공한 사용자 비밀 조회부터 계산한다. status, sync, TUI polling은 연장하지 않는다.
- 한 OS 사용자당 daemonkit instance 하나. profile별 storage/secrets namespace를 사용하고 검증된 server URL/account ID를 저장된 identity와 대조한다. config 파일을 바꾸어 같은 이름을 다른 identity에 연결하면 충돌 오류다.
- server/email이 바뀐 profile은 기존 세션/cache를 재사용하지 않는다. 먼저 기존 identity로 logout하거나 새 profile을 만든다. profile use는 기본 선택만 바꾸며 다른 profile의 상태를 바꾸지 않는다.
- cloud US/EU는 검증된 API/Identity 주소로 매핑한다. self-hosted는 base URL의 `/api`, `/identity`를 사용한다. 분리 custom endpoint는 후속이다.
- HTTPS 필수, localhost/loopback 개발 서버만 HTTP 허용. URL userinfo/query/fragment를 거부하고 다른 origin으로 인증정보를 redirect하지 않는다.
- 로컬 디렉터리 0700, 파일 0600. config는 원자 교체하고 변경 충돌을 감지한다. SQLite cache와 OS credential store는 별도 저장소이므로 한 transaction이라고 가정하지 않는다.

## 6. Daemon 소유권과 보안

```text
CLI / TUI ─ typed local RPC ─ vault service
                                ├─ 보호된 메모리: 풀린 키, index, access token
                                ├─ daemonkit Secrets: refresh token
                                ├─ daemonkit Storage: 암호화 snapshot + metadata
                                └─ 재사용 HTTP client: Identity / Vault API
```

- go-daemonkit의 프로세스 수명주기/IPC/typed RPC/취소/storage를 재사용한다. 별도 process manager, HTTP daemon, 범용 backend plugin은 만들지 않는다.
- UI는 입력/표현, service는 상태·권한·마스킹·동시성, API 코드는 protocol validation, crypto 코드는 키 계층과 암호문 검증을 소유한다. 잠금/인증 규칙을 CLI와 TUI에 복제하지 않는다.
- secure string은 Go string 별칭이 아니다. 잠긴 메모리와 명시적 파기를 지원하는 구현을 선택한다. 보호된 메모리 확보 실패 시 unlock을 실패시키며 평문 장기 보관으로 fallback하지 않는다.
- 비밀번호/KDF 중간 키는 인증 작업 후 파기하고, 풀린 키는 lock 때 파기한다. 평문 index도 민감 자료로 취급한다. RPC/HTTP/Go 문자열의 임시 복사까지 무복사나 완전한 zeroization을 보장하지 않는다.
- secret을 argv/log/panic/status/SQLite에 기록하지 않는다. 전체 request/response 로깅 금지. refresh token의 OS credential store 저장은 명시적인 예외이며 OS 보호 정책을 따른다.
- lock은 vault generation을 증가시키고 진행 중 복호화를 취소/회수한 뒤 key/index를 파기한다. lock 응답 이후 이전 generation의 비밀 응답을 발행하지 않는다. sync는 암호문을 갱신할 수 있지만 lock 뒤 index를 되살리지 않는다.
- logout은 profile 전체 작업을 취소하고 admission을 막는다. SQLite에 logout 진행 상태를 먼저 기록하고 credential 삭제와 cache 제거를 완료한다. 중간 실패 시 잠긴 상태를 유지하고 재시작 때 정리를 재개하며 성공이라고 보고하지 않는다. keychain에서 남은 token을 재사용하지 않는다.
- 재시작 후 vault는 locked다. credential store와 DB의 identity/완료 상태가 맞지 않는 부분 로그인은 폐기하거나 재인증한다. 키가 없는데 unlocked로 표시하지 않는다.
- OS 사용자 IPC 경계를 신뢰한다. root나 같은 사용자 권한의 악성 프로세스에 대한 완전한 격리는 제공하지 않는다.
- HTTP deadline/응답 크기/KDF CPU·메모리 상한/TLS 검증을 적용한다. MAC/암호문 검증 실패는 해당 조회를 실패시키며 부분 plaintext를 반환하지 않는다. 429는 Retry-After를 존중하고 무한 재시도하지 않는다.
- 암호 primitive를 자체 발명하지 않는다. Go 호환 구현의 Password Manager 지원·라이선스·유지보수와 공식 test vector를 확인한 뒤 의존성을 선택한다. Secrets Manager SDK를 vault SDK로 간주하지 않는다.

## 7. Sync / 조회 / TUI

- 최초 cache 준비와 명시적 sync에서만 네트워크를 사용한다. 자동 background sync는 첫 범위에서 제외하고 last_sync와 오프라인 상태를 보여준다.
- sync는 profile당 하나. 새 암호문 snapshot/키 metadata/last_sync를 SQLite transaction으로 함께 교체한다. 구조·identity 검증 실패 시 이전 snapshot을 보존한다. locked 상태의 sync는 암호학적 복호화 검증까지 완료했다고 주장하지 않는다.
- unlocked sync는 새 snapshot으로 임시 index를 만든 뒤 publish한다. 키 변경이나 복호화 실패 시 이전 index를 새 snapshot과 섞지 않고 vault를 잠근다. 잠긴 sync에서 저장된 암호문은 다음 unlock 때 MAC/복호화를 검증한다.
- list/search는 unlock 때 만든 최소 index를 사용한다. 상세는 필요한 항목을 메모리의 암호문 snapshot에서 복호화한다. cache-only 조회는 HTTP와 subprocess 0회가 기준이다.
- RPC 메시지 64 KiB에 맞춰 목록은 byte budget으로 페이지 처리한다. page cursor는 snapshot generation에 묶고 변경 시 재시작을 요구한다. 한도를 넘는 단일 값은 명시적 오류이며 비밀을 잘라 성공 처리하지 않는다.
- TUI는 profile/auth/vault 상태·last_sync, 검색 목록, 상세, 일시적 reveal, lock/sync/quit를 제공한다. profile/item/focus 변경과 lock에서 reveal을 지운다. TUI 종료는 daemon lock과 다르다.
- 키보드 조작, 색상 외 상태 표시, 작은 화면/Unicode/긴 값 처리를 지원한다. clipboard는 후속이다.

## 8. Full integration과 구현 순서

[개발 환경](../dev/README.md)은 로컬 Vaultwarden를 사용한다. 공식 Bitwarden와의 호환성은 별도 실제 서버에서 검증한다. [SQL snapshot과 manifest](../testdata/integration/README.md)를 반드시 복원한다. [합성 데이터](../testdata/bitwarden.json)는 암호문 seed의 원본이자 복호화 기대값이며 API 응답을 대신하지 않는다. 서버 8종 타입은 현재 13개 항목으로 검증하고 새 타입/필드는 fixture와 dump를 함께 갱신한다.

1. **직접 API 수직 구현**: 전용 계정 login → 실제 sync → 키 복호화 → 항목 read. PBKDF2/Argon2id와 변조/잘못된 키를 검증한다. crypto 단위 테스트는 공식 vector와 대조한다.
2. **daemon과 저장 수명**: CLI를 연결하고 profile 격리, keychain token, transaction cache, lock/unlock/logout, 재시작 후 오프라인 unlock을 검증한다.
3. **인증 재개**: token 만료→refresh, refresh 무효→login→원래 읽기 재개, 2FA/취소/비TTY/동시 요청/연결 끊김을 검증한다.
4. **TUI**: 같은 RPC로 CLI와 동일 값·마스킹·잠금 상태를 확인한다.

추가 acceptance:

- 서버 item 변경 후 sync 반영, 오프라인 기존 cache 조회, 잘못된 비밀번호와 손상 cache에서 실패.
- 서로 다른 profile의 키/token/cache가 섞이지 않음. server/account 변경 거부.
- lock과 read/sync 경합, logout 도중 credential store 오류/크래시 후 정리 재개.
- warm 조회 HTTP/subprocess 0회. cold KDF/login/sync와 warm latency를 분리 측정하며 수치는 측정 후 확정.
- DB/config/log/argv에 plaintext 비밀번호·복호화 키·token·vault 값 없음. refresh token은 OS credential store에만 존재. logout 뒤 cache/credential 접근 불가.
- 중복 이름, 빈 값, Unicode, 여러 줄 notes, hidden field, card/identity 기본 마스킹, RPC 크기 초과.

현재 Go API의 login/refresh/sync/8종 복호화와 Cobra → daemon RPC를 snapshot 서버에서 검증한다. profile별 키·토큰은 memguard locked memory, snapshot은 암호문으로 메모리에 유지한다. read는 HTTP 없이 동작하며 lock/shutdown은 진행 중 작업을 취소하고 세션을 파기한다. 임시 crypto/HTTP/RPC 평문 복사본까지 완전히 zeroize하지는 않는다. `main.go`는 root, 패키지는 `cli`, `vaultapi`, `daemon`, `vault`, `bitwarden`으로 나눈다. keychain refresh token과 SQLite 암호화 cache 저장, HTTP 없는 unlock, 로컬 logout을 구현했다. `lock`은 메모리만 파기하고 재시작 후 비밀번호로 unlock한다. SQLite/keychain 간 작업은 credential 정리 목록으로 복구하며, logout 실패 시 표시를 저장하고 unlock을 막는다. CLI 공통 실행부의 인증 후 원래 read/sync 1회 재시도까지 구현했다. 서비스의 AuthRequired/ErrStateChanged는 daemon에서 RPC 코드로 바꾸며 일반 장애는 그대로 전달한다. 자동 인증 예약은 연결 감시·취소·lock 및 2분 만료로 정리하고, 대기자는 prompt를 중복 생성하지 않는다. TUI·Argon2id·2FA는 후속이다. 테스트는 전용 서버·계정만 사용한다.

## 9. 참고한 인터페이스와 근거

| 도구 | 채택한 부분 |
| --- | --- |
| [rbw](https://github.com/doy/rbw#usage) | login/unlock 분리, agent의 키 보관, 명령 전 인증 확보 |
| [Vault](https://developer.hashicorp.com/vault/docs/commands) | 일관된 명령·구조화 출력 |
| [1Password](https://developer.1password.com/docs/cli/secret-references/) | item 탐색과 secret reference/read 분리 |
| [age](https://github.com/FiloSottile/age) | stdin/stdout 조합, 작은 명령 표면 |
| [fnox](https://fnox.jdx.dev/guide/profiles.html) | TOML, 명시적 profile 선택 |

rbw 검토 revision: `77464d414a9e41165d5b15f049bee1f0c0f2a982`.
[actions](https://github.com/doy/rbw/blob/77464d414a9e41165d5b15f049bee1f0c0f2a982/src/actions.rs)는 키 복호화와 unauthorized 시 refresh 재시도를,
[db](https://github.com/doy/rbw/blob/77464d414a9e41165d5b15f049bee1f0c0f2a982/src/db.rs)는 token/암호문 영속 저장을,
[locked](https://github.com/doy/rbw/blob/77464d414a9e41165d5b15f049bee1f0c0f2a982/src/locked.rs)는 region lock 시도와 zeroize를 보여준다.
rbw가 모든 refresh 실패를 자동 login으로 회복하거나 모든 secret을 메모리에만 둔다고 가정하지 않는다.

BW_SESSION은 [공식 CLI의 로컬 secure storage 키](https://github.com/bitwarden/clients/blob/main/apps/cli/src/platform/services/node-env-secure-storage.service.ts)이며 API token이 아니다. bwb에서는 이 환경변수를 인증 계약으로 사용하지 않는다.

의존성: Cobra/UUID, 터미널 상태 제어용 `golang.org/x/term`, `github.com/134130/go-daemonkit v0.0.0-20261024090218-d756ef3418d7`, `github.com/awnumar/memguard v0.23.0`. memguard는 직접 OS별 메모리 보호 코드를 만들지 않고 locked memory와 명시적 파기를 사용하기 위한 선택이다. daemonkit Storage와 Secrets를 vault 영속 저장에 사용하며 별도 DB·keychain 구현은 만들지 않는다.
