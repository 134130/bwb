# bwb 설계

상태: 목표 설계와 구현 계획. Cobra item read/totp, daemon/영속 cache, 인증 재개까지 구현했다. 조직 vault/collection 복호화·권한 제한·snapshot 검증도 구현했다. 목록·검색 RPC/CLI, config와 profile별 자동 잠금도 구현했다. TUI Explorer도 구현했다. 아래 계약에는 미구현 내용이 포함되며 현재 사용법은 README를 따른다.

## 1. 방향과 첫 범위

CLI와 TUI가 go-daemonkit 기반 daemon을 공유하고, daemon이 Bitwarden Password Manager API를 직접 호출한다. 공식 bw/rbw 실행 파일, bw serve, BW_SESSION 가져오기는 사용하지 않는다.

핵심은 **서버 로그인과 로컬 vault unlock의 분리**다. 명령은 필요한 상태만 확보하고 이어서 실행한다. 반복 조회는 네트워크와 외부 프로세스 실행 없이 처리한다.

첫 범위는 macOS/Linux, 개인·조직 vault 읽기, 자체 로그인/로그아웃, unlock/lock, sync, 목록/상세/필드 조회, 읽기 전용 TUI다. master-password 로그인과 PBKDF2/Argon2id, TOTP 2FA를 검증 대상으로 한다. 조직 vault/collection 읽기는 목록 구현의 선행 작업이다. SSO/WebAuthn, 조직·collection 관리, 편집, attachments, SSH agent, 환경 주입은 후속이다. 지원하지 않는 인증·키·암호문 형식은 명시적 오류로 처리한다.

Full integration은 실제 서버의 인증·동기화·복호화·조회·재인증을 통과하는 것이다. 버전 관리된 필수 SQL dump snapshot을 매번 독립 서버에 복원하고 Lefthook pre-commit에서 매 커밋마다 검증한다. 수동/선택적 import와 HTTP mock만으로 완료를 선언하지 않는다.

## 2. 상태와 저장 위치

인증 상태(`unauthenticated`, `authenticated`, `reauth_required`)와 vault 상태(`empty`, `locked`, `unlocked`)는 독립적이다. authenticated는 인증 자료가 있다는 뜻이며 원격 서버가 지금도 허용한다는 보장은 아니다.

| 자료 | 저장 위치 | lock / daemon 종료 |
| --- | --- | --- |
| 마스터 비밀번호, 중간 KDF 키 | 필요한 동안만 메모리 | 작업 종료 즉시 파기 |
| 풀린 user/item/organization key | daemon의 보호된 메모리 | 파기 |
| 검색 index, 복호화 값 | daemon/요청의 메모리 | 파기 |
| API access token | daemon 메모리 | lock에서는 유지, 종료 시 파기 |
| API refresh token 또는 개인 API Key | OS credential store | 유지; logout 시 삭제 |
| 암호화된 vault, protected key, KDF 설정 | daemonkit SQLite | 유지; logout 시 삭제 |
| 서버·계정 ID, device ID, last_sync | daemonkit SQLite | 유지; logout 시 계정 데이터 삭제 |
| profile 설정 | TOML config | 유지 |

refresh token과 개인 API Key는 daemonkit `app.Secrets()`를 사용한다. plaintext 파일 fallback은 없다. credential store 사용 불가 시 로그인 저장에 실패했음을 알리고 완료로 보고하지 않는다. access token·refresh token·API Key는 vault 복호화 키가 아니다. 보호된 키와 암호문은 비밀의 평문과 구분한다.

이 저장 정책은 재로그인 없이 재시작 후 refresh하고, 오프라인에서도 비밀번호로 cached vault를 unlock하기 위한 것이다. rbw의 token JSON 저장 대신 이미 사용 중인 daemonkit의 OS credential store를 재사용한다. 명시적 `login --apikey`는 개인 `client_id`·`client_secret`을 저장하고, refresh token 없이 `client_credentials` grant로 재발급한다. 첫 로그인에서 마스터 비밀번호로 복호화와 sync를 검증한 뒤 저장한다. API Key profile은 인증 방식을 포함한 v2 레코드를 사용하며 기존 비밀번호 profile의 v1 레코드를 계속 읽는다. 키가 폐기되면 명시적 API Key 재로그인을 요구하고 자동 비밀번호 로그인으로 전환하지 않는다. Key Connector·마스터 비밀번호 없는 계정·Argon2id는 이 인증 경로에서 지원하지 않는다.

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
| item list/read/totp, TUI 탐색 | 기존 cache가 있으면 아니오 | 예 |
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
bwb item list --profile work --search github
bwb item list --profile work --search github --sync --format json
bwb item read ITEM --profile work
bwb item read ITEM --profile work --field password
bwb item read ITEM --profile work --reveal --format json
bwb item totp ITEM --profile work
bwb tui --profile work
bwb status --profile work
bwb login --profile work
bwb unlock --profile work
bwb lock --profile work
bwb logout --profile work
bwb sync --profile work
bwb profile list
bwb profile show work
bwb config path
bwb config validate
bwb daemon status
bwb daemon update
bwb daemon stop
```

- 탐색·상세·값 읽기는 `item` 아래로 통일한다. 루트 `read`, `item get`, positional field, `bwb://` reference와 타입 접두사 필드 경로는 제공하지 않는다. 호환 alias도 없다.
- 인자 없음은 help. TUI는 명시적으로 시작한다. status/daemon status는 daemon을 시작하지 않는다. 다른 작업은 필요하면 시작한다.
- profile 우선순위는 `--profile > BWB_PROFILE > default_profile > default`다.
- `ITEM`은 UUID 또는 대소문자를 무시한 정확한 이름이다. 부분 검색은 `item list --search`에서 수행한다. 동명 항목은 후보 ID와 소속을 표시하고 오류로 끝낸다. 첫 결과를 자동 선택하지 않는다.
- `item read ITEM`은 소속 metadata와 필드 키·label·값을 함께 표시한다. 사용자명·URL·카드 brand·일반 custom field는 표시하고 password/TOTP seed/hidden field/notes/나머지 카드·신원 정보는 기본 마스킹한다. 상세 하단에서 단일 값·reveal·TOTP의 다음 명령을 안내한다.
- 필드 키는 `username`, `password`, `notes`, `url/1`, `custom/token` 등이다. Custom 이름은 percent encode하며 동명 custom field는 suffix로 구분한다. 사용자에게 키를 추측하게 하지 않고 상세 화면에 표시한 키를 그대로 받는다.
- `--field KEY`는 원문 값 하나를 줄바꿈 추가 없이 출력한다. 빈 값은 성공, 없는 필드는 사용 가능한 키와 함께 오류다. `--reveal`은 상세 전체의 숨겨진 값을 표시한다. 두 옵션은 함께 사용하지 않는다.
- `item totp ITEM`은 현재 인증 코드만 출력한다. `--field totp-secret`은 별도의 명시적 seed 추출이다. RFC 6238 SHA1/SHA256/SHA512, 6/8자리, Base32 및 otpauth URI를 지원한다. HOTP/Steam은 범위 밖이다.
- 조회는 cache가 기본이다. `--sync`를 붙인 경우에만 먼저 sync한다. 목록 0건은 cache 기준의 결과와 sync 힌트를 표시하고 성공한다.
- 상세/TOTP JSON은 `{version: 1, data: {item, fields, restriction?}}`다. 필드는 `key`, `label`, `sensitive`, `state`, `value`를 갖는다. 숨김은 `hidden/null`, 빈 값은 `empty/""`, 표시값은 `visible/string`으로 구분한다. JSON 선택 자체는 reveal을 뜻하지 않는다. 목록 JSON의 data는 배열이다.
- 상세는 daemon의 단일 profile 작업 안에서 같은 snapshot으로 조회한다. JSON 응답이 48,000 bytes를 넘으면 부분 출력 없이 실패하고 `--field`를 안내한다. 단일 필드도 한도를 넘으면 실패한다.
- reprompt/조직의 조회 제한은 field/reveal/TOTP에도 적용한다. 전체 상세는 metadata와 제한 사유만 표시하며 명시적 비밀 조회는 오류다.
- stdout은 결과, stderr는 진단이다. 현재 오류는 텍스트와 종료 코드 1을 사용한다. 오류별 구조화 JSON/종료 코드 세분화는 아직 없다.
- logout은 선택 profile의 로컬 cache와 credential을 삭제한다. 원격 vault item은 삭제하지 않는다.

### 조직·folder·collection 계약 (목록 구현 전 확정)

- profile은 계정 경계다. 계정 안에서 item은 개인 소유 또는 한 organization 소유이며, 개인 정리용 folder는 0~1개, 조직 collection은 0~여러 개다. folder와 collection은 별도 관계이며 상호 배타적인 분류로 합치지 않는다.
- item 조회 모델은 `organizationId`, `folderId`, `collectionIds`를 보존한다. organization/collection은 ID로 식별하고 collection에 organization ID를 둔다. 동명 collection을 합치거나 이름의 `/`를 실제 부모 관계로 추론하지 않는다.
- `bwb organization list`, `bwb collection list --organization ID`, `bwb folder list`로 탐색한다. `item list`에는 `--organization ID`, `--collection ID`, `--folder ID` 필터를 두며 서로 다른 필터는 AND다. 잘못된 ID와 서로 맞지 않는 조직/collection 선택은 입력 오류다. 이름을 ID 대신 조용히 해석하지 않는다.
- 기본 목록은 item당 한 행이며 조직·folder·collection 소속을 표시한다. JSON은 ID 관계를 유지하는 평면 배열을 data에 담는다. 목록의 organization/folder는 `{id,name}`, collections는 같은 구조의 배열이며 backend item 모델의 ID 관계를 보존한다. `--group-by collection`은 다중 소속 item을 각 그룹에 표시하고 미소속 그룹도 둔다. 그룹 때문에 전체 고유 item 수가 늘어난 것으로 보고하지 않는다.
- 정렬은 `--sort name|type|folder|organization`의 쉼표 조합과 `--reverse`를 지원할 계획이다. 기본은 name, 동률은 ID로 고정한다. 여러 값을 가진 collection의 정렬은 첫 범위에서 제외한다. 표의 열 너비·다중 소속 표현은 실제 snapshot으로 확인한 뒤 확정한다.

선행 구현과 검증:

1. 실제 sync 응답과 upstream 구현에서 암호화된 사용자 private key → organization key → 선택적 item key 경로, collection 이름 암호화, 항목별 유효 권한 필드를 확인한다. 사용자 키 fallback으로 조직 항목을 복호화하지 않는다. 기존 `internal/bitwarden`이 키 선택·프로토콜 검증을 소유한다.
2. 암호화 snapshot에 조직·collection·보호된 키를 함께 보존한다. plaintext 키는 보호 메모리에만 두고 lock/logout에서 파기한다. 기존 개인 cache의 offline unlock 호환성을 유지하며, 불완전한 조직 cache를 빈 조직 목록으로 간주하지 않는다. 포맷 변경이 필요하면 version과 호환 정책을 명시한다.
3. 서버가 내려준 접근 범위와 유효 제한을 서비스 조회에 적용한다. collection별 권한을 임의로 합산하지 않는다. read-only는 읽기를 막는 권한이 아니며, hide-password 등의 제한은 명시적 read에도 적용한다. 지원하지 못하는 제한은 우회하지 않고 명시적 오류로 처리한다. CLI/TUI에만 제한을 두지 않는다.
4. sync 성공 시 항목·조직·collection·키 metadata를 같은 generation으로 교체한다. 삭제·탈퇴·권한 변경은 성공한 sync 후 반영한다. offline cache는 마지막 sync의 권한 상태이며 원격 회수를 즉시 알 수 없다. 손상된 키/관계 때문에 실패한 sync는 기존 cache를 유지하고 실패를 알린다.
5. 재현 가능한 SQL snapshot에 두 조직, 조직 간 동명 collection, 다중 collection 소속, folder와 collection 동시 소속, 개인·조직 동명 item, 제한 권한 계정을 추가한다. 공개 합성 키와 manifest 기대값을 함께 관리한다. 기존 8종 coverage는 유지하고 조직 키/item key 경로는 대표 항목으로 검증한다.
6. 실제 API 기반 read, 재시작 후 offline unlock/read, profile·조직 키 격리, lock과 조회 경합, sync 후 권한 회수·삭제를 integration에서 검증한다. 비동기 경합 단위 테스트는 synctest를 사용한다. Lefthook의 두 번 독립 snapshot 복원 검증에 포함한다.

현재 조직 키와 item key 복호화, 조직 metadata/collection의 암호화 cache 보존, 서버의 viewPassword 제한, 두 조직·네 collection snapshot 검증을 구현했다. 제한 항목은 우선 이름만 읽도록 보수적으로 처리한다. 조직 키는 작업 중에만 보호 메모리에 두며 파싱된 RSA private key의 Go 임시 할당은 즉시 zeroize를 보장하지 않는다. 목록·검색 RPC와 표/JSON 표현, 정렬·필터·그룹 표시도 구현했다. 조직/collection 생성·수정, 초대·멤버 관리 API는 이번 범위가 아니다.

## 5. Profile / config

[예제](../dev/config.example.toml)는 비밀을 포함하지 않는 TOML이다. 읽기·검증과 profile list/show를 구현했으며 편집 명령은 후속이다.

- profile = 서버 + 계정 + 해당 인증/cache. folder/collection이나 dev/prod override 집합이 아니다. profile 상속/합성과 프로젝트 config 자동 탐색은 없다.
- config 경로는 `${XDG_CONFIG_HOME:-$HOME/.config}/bwb/config.toml`. `--config > BWB_CONFIG > 기본 경로` 순서로 한 파일만 읽는다.
- 이름은 `[a-z][a-z0-9_-]{0,31}`. 알 수 없는 config 키·schema version·잘못된 duration은 거부한다.
- config는 version/default_profile와 profile별 server/email/lock_timeout만 가진다. token/password/키는 금지한다.
- `lock_timeout = "0s"`는 명시적 lock 또는 daemon 종료까지 복호화 키를 유지한다. 양수는 login/unlock 완료부터 시작하고 성공한 read/list에서 갱신한다. status와 sync는 연장하지 않는다. 후속 TUI의 자동 polling은 사용자 조회와 구분해야 한다.
- 한 OS 사용자당 daemonkit instance 하나. profile별 storage/secrets namespace를 사용하고 검증된 server URL/account ID를 저장된 identity와 대조한다. config 파일을 바꾸어 같은 이름을 다른 identity에 연결하면 충돌 오류다.
- server/email이 바뀐 profile은 기존 세션/cache를 재사용하지 않는다. 먼저 기존 identity로 logout하거나 새 profile을 만든다. profile use는 기본 선택만 바꾸며 다른 profile의 상태를 바꾸지 않는다.
- cloud US/EU는 검증된 API/Identity 주소로 매핑한다. self-hosted는 base URL의 `/api`, `/identity`를 사용한다. 분리 custom endpoint는 후속이다.
- HTTPS 필수, localhost/loopback 개발 서버만 HTTP 허용. URL userinfo/query/fragment를 거부하고 다른 origin으로 인증정보를 redirect하지 않는다.
- 로컬 디렉터리 0700, 파일 0600. config는 원자 교체하고 변경 충돌을 감지한다. SQLite cache와 OS credential store는 별도 저장소이므로 한 transaction이라고 가정하지 않는다.

## 6. Daemon 소유권과 보안

```text
CLI / TUI ─ typed local RPC ─ vault service
                                ├─ 보호된 메모리: 풀린 키, index, access token
                                ├─ daemonkit Secrets: refresh token 또는 개인 API Key
                                ├─ daemonkit Storage: 암호화 snapshot + metadata
                                └─ 재사용 HTTP client: Identity / Vault API
```

- go-daemonkit의 프로세스 수명주기/IPC/typed RPC/취소/storage를 재사용한다. 별도 process manager, HTTP daemon, 범용 backend plugin은 만들지 않는다.
- UI는 입력/표현, service는 상태·권한·마스킹·동시성, API 코드는 protocol validation, crypto 코드는 키 계층과 암호문 검증을 소유한다. 잠금/인증 규칙을 CLI와 TUI에 복제하지 않는다.
- secure string은 Go string 별칭이 아니다. 잠긴 메모리와 명시적 파기를 지원하는 구현을 선택한다. 보호된 메모리 확보 실패 시 unlock을 실패시키며 평문 장기 보관으로 fallback하지 않는다.
- 비밀번호/KDF 중간 키는 인증 작업 후 파기하고, 풀린 키는 lock 때 파기한다. 평문 index도 민감 자료로 취급한다. RPC/HTTP/Go 문자열의 임시 복사까지 무복사나 완전한 zeroization을 보장하지 않는다.
- secret을 argv/log/panic/status/SQLite에 기록하지 않는다. 전체 request/response 로깅 금지. refresh token과 명시적 개인 API Key 로그인의 OS credential store 저장은 예외이며 OS 보호 정책을 따른다.
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
- list/search는 unlock 이후 첫 목록 요청에서 만든 이름·소속 metadata index를 재사용한다. 원문 비밀 필드는 복호화하거나 index에 담지 않는다. 성공한 login/unlock/sync와 lock/logout/shutdown에서 index 참조를 폐기한다. 상세는 필요한 항목을 메모리의 암호문 snapshot에서 복호화한다. cache-only 조회는 HTTP와 subprocess 0회가 기준이다.
- RPC 메시지 64 KiB에 맞춰 목록은 byte budget으로 페이지 처리한다. page cursor는 snapshot generation에 묶고 변경 시 재시작을 요구한다. 한도를 넘는 단일 값은 명시적 오류이며 비밀을 잘라 성공 처리하지 않는다.
- TUI는 profile/auth/vault 상태·last_sync, 검색 목록, 상세, 일시적 reveal, lock/sync/quit를 제공한다. profile/item/focus 변경과 lock에서 reveal을 지운다. TUI 종료는 daemon lock과 다르다.
- 키보드 조작, 색상 외 상태 표시, 작은 화면/Unicode/긴 값 처리를 지원한다. clipboard는 후속이다.

## 8. Full integration과 구현 순서

[개발 환경](../dev/README.md)은 로컬 Vaultwarden를 사용한다. 공식 Bitwarden와의 호환성은 별도 실제 서버에서 검증한다. [SQL snapshot과 manifest](../testdata/integration/README.md)를 반드시 복원한다. [합성 데이터](../testdata/bitwarden.json)는 암호문 seed의 원본이자 복호화 기대값이며 API 응답을 대신하지 않는다. 서버 8종 타입은 현재 19개 항목으로 검증하고 새 타입/필드는 fixture와 dump를 함께 갱신한다.

1. **직접 API 수직 구현**: 전용 계정 login → 실제 sync → 키 복호화 → 항목 read. PBKDF2/Argon2id와 변조/잘못된 키를 검증한다. crypto 단위 테스트는 공식 vector와 대조한다.
2. **daemon과 저장 수명**: CLI를 연결하고 profile 격리, keychain token, transaction cache, lock/unlock/logout, 재시작 후 오프라인 unlock을 검증한다.
3. **인증 재개**: token 만료→refresh, refresh 무효→login→원래 읽기 재개, 2FA/취소/비TTY/동시 요청/연결 끊김을 검증한다.
4. **조직 vault/collection**: 조직·item 키와 유효 읽기 권한, 암호화 cache, 확장 snapshot을 검증한다. 목록의 선행 작업이다.
5. **목록·검색**: 개인·조직 항목을 함께 조회하고 정렬, folder/collection 필터·그룹 표시, HTTP 없는 조회를 검증한다.
6. **TUI**: 같은 RPC로 CLI와 동일 값·마스킹·잠금 상태를 확인한다.

추가 acceptance:

- 서버 item 변경 후 sync 반영, 오프라인 기존 cache 조회, 잘못된 비밀번호와 손상 cache에서 실패.
- 서로 다른 profile의 키/token/cache가 섞이지 않음. server/account 변경 거부.
- lock과 read/sync 경합, logout 도중 credential store 오류/크래시 후 정리 재개.
- warm 조회 HTTP/subprocess 0회. cold KDF/login/sync와 warm latency를 분리 측정하며 수치는 측정 후 확정.
- DB/config/log/argv에 plaintext 비밀번호·복호화 키·token·API Key·vault 값 없음. refresh token과 API Key의 영속 저장은 OS credential store만 사용. logout 뒤 cache/credential 접근 불가.
- 중복 이름, 빈 값, Unicode, 여러 줄 notes, hidden field, card/identity 기본 마스킹, RPC 크기 초과.

현재 Go API의 login/refresh/sync/8종 복호화와 Cobra → daemon RPC를 snapshot 서버에서 검증한다. profile별 키·토큰은 memguard locked memory, snapshot은 암호문으로 메모리에 유지한다. read는 HTTP 없이 동작하며 lock/shutdown은 진행 중 작업을 취소하고 세션을 파기한다. 임시 crypto/HTTP/RPC 평문 복사본까지 완전히 zeroize하지는 않는다. `main.go`는 root, 패키지는 `cli`, `vaultapi`, `daemon`, `vault`, `bitwarden`으로 나눈다. keychain refresh token과 SQLite 암호화 cache 저장, HTTP 없는 unlock, 로컬 logout을 구현했다. `lock`은 메모리만 파기하고 재시작 후 비밀번호로 unlock한다. SQLite/keychain 간 작업은 credential 정리 목록으로 복구하며, logout 실패 시 표시를 저장하고 unlock을 막는다. CLI 공통 실행부의 인증 후 원래 read/list/sync 1회 재시도까지 구현했다. 서비스의 AuthRequired/ErrStateChanged는 daemon에서 RPC 코드로 바꾸며 일반 장애는 그대로 전달한다. 자동 인증 예약은 연결 감시·취소·lock 및 2분 만료로 정리하고, 대기자는 prompt를 중복 생성하지 않는다. Argon2id·2FA는 후속이다. 테스트는 전용 서버·계정만 사용한다.

## 9. 참고한 인터페이스와 근거

| 도구 | 채택한 부분 |
| --- | --- |
| [rbw](https://github.com/doy/rbw#usage) | login/unlock 분리, agent의 키 보관, 명령 전 인증 확보 |
| [Vault](https://developer.hashicorp.com/vault/docs/commands) | 일관된 명령·구조화 출력 |
| [1Password](https://developer.1password.com/docs/cli/secret-references/) | 명시적 비밀값 선택; CLI에서는 item 아래의 field 옵션으로 제공 |
| [age](https://github.com/FiloSottile/age) | stdin/stdout 조합, 작은 명령 표면 |
| [fnox](https://fnox.jdx.dev/guide/profiles.html) | TOML, 명시적 profile 선택 |

rbw 검토 revision: `77464d414a9e41165d5b15f049bee1f0c0f2a982`.
[actions](https://github.com/doy/rbw/blob/77464d414a9e41165d5b15f049bee1f0c0f2a982/src/actions.rs)는 키 복호화와 unauthorized 시 refresh 재시도를,
[db](https://github.com/doy/rbw/blob/77464d414a9e41165d5b15f049bee1f0c0f2a982/src/db.rs)는 token/암호문 영속 저장을,
[locked](https://github.com/doy/rbw/blob/77464d414a9e41165d5b15f049bee1f0c0f2a982/src/locked.rs)는 region lock 시도와 zeroize를 보여준다.
rbw가 모든 refresh 실패를 자동 login으로 회복하거나 모든 secret을 메모리에만 둔다고 가정하지 않는다.

BW_SESSION은 [공식 CLI의 로컬 secure storage 키](https://github.com/bitwarden/clients/blob/main/apps/cli/src/platform/services/node-env-secure-storage.service.ts)이며 API token이 아니다. bwb에서는 이 환경변수를 인증 계약으로 사용하지 않는다.

의존성: Cobra/UUID, 터미널 상태 제어용 `golang.org/x/term`, `github.com/134130/go-daemonkit v0.0.0-20261024090218-d756ef3418d7`, `github.com/awnumar/memguard v0.23.0`. memguard는 직접 OS별 메모리 보호 코드를 만들지 않고 locked memory와 명시적 파기를 사용하기 위한 선택이다. daemonkit Storage와 Secrets를 vault 영속 저장에 사용하며 별도 DB·keychain 구현은 만들지 않는다.

현재 config 적용 계약: TOML은 원하는 설정, SQLite/keychain은 인증·캐시 상태를 소유한다. CLI는 기본 profile 선택과 config/profile 검사에 파일을 읽는다. Daemon은 시작 시 선택한 파일의 profile identity/timeout을 적용하고 `daemon reload`에서 같은 파일 전체를 검증하여 교체한다. 일반 조회는 설정을 적용하지 않는다. Reload 실패 시 기존 정책 유지, 로그인된 profile identity 변경/삭제는 먼저 logout 필요. Logout은 TOML 정의를 삭제하지 않는다. 정책은 DB에 중복 저장하지 않으며 이전 개발 버전의 policy 필드는 무시한다. 설정 교체는 profile 작업과 배타적으로 진행하고, 검증/취소 확인 후에는 caller 취소와 무관하게 메모리 교체를 완료한다. Lock과 idle 만료는 reload 대기 중에도 진행 작업을 취소할 수 있다. 재적용은 마지막 활동 시각을 바꾸지 않는다. Timer generation과 admission deadline 검사로 오래된 타이머와 지연된 만료를 처리한다. 파일 watcher 및 config/profile 편집 명령은 구현하지 않는다.

## TUI 구현 경계

`internal/tui`는 Bubble Tea v2 기반 Explorer/Profiles와 인증 dialog를 담당한다. `Root`만 terminal size를 저장한다. 자식 계약은 `Update(tea.Msg, Size) tea.Cmd` / `View(Size) string`이며 부모의 순수 layout 함수가 content 크기를 계산한다. 레이아웃 변경 시 `LayoutMsg`로 scroll/input 상태를 먼저 동기화한다. View는 상태를 변경하지 않는다. 화면별 layout만 사용하며 범용 layout framework는 없다.

`internal/vaultclient`는 CLI/TUI가 공유하는 인증 예약·watch·한 번 재시도·페이지 수집을 소유한다. TUI의 RPC 결과는 UI epoch에 묶어 profile 전환·잠금·선택 변경 이후의 응답을 버린다. 비밀번호는 in-TUI 입력과 context 취소를 연결하며 기존 daemon의 단일 인증 소유자 규칙을 따른다. TUI는 CLI처럼 직접 서버에 접근하지 않는다.

추가 RPC: `vault.describe`(field label/path만), `vault.profiles`(설정 및 기존 저장 키에서 발견한 이름), `vault.activity`(명시적 사용자 활동). `vault.list`의 background 요청은 idle을 연장하지 않는다. 필드 metadata 조회는 값 복호화를 하지 않으며 기존 reprompt/조직 제한을 보존한다.
