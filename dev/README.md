# 재현 가능한 Vaultwarden integration

초기 데이터는 필수다. 버전 관리되는 [SQL snapshot](../testdata/integration/vaultwarden.sql)을 매번 새 서버에 복원한다. 개인 vault, 수동 계정 생성, 선택적인 import에 의존하지 않는다.

## 실행

Docker, Python 3.11+, mise.toml에 고정된 Go, 사용 가능한 OS credential store가 필요하다. macOS는 cgo/Keychain, Linux는 session D-Bus의 Secret Service가 필요하며 native store가 없으면 테스트를 실패시킨다. 저장소 루트에서:

```sh
python3 -m venv .dev/venv
.dev/venv/bin/python -m pip install -r dev/integration/requirements.txt
mise run init
mise run integration
```

`mise run init`은 Git hook을 설치한다. integration task는 항상 `.dev/venv/bin/python`을 사용한다. 새 checkout마다 위 준비를 한 번 수행한다.

runner는 SHA-256 검사 → SQLite 복원·무결성 검사 → digest로 고정된 Vaultwarden 실행 → Go 제품 API·Cobra integration 검증 → 전용 컨테이너·임시 데이터 정리를 수행한다. 계정 등록이나 seed는 일반 테스트 실행에서 하지 않는다.

매 실행마다 임의의 `bwb-integration-*` 이름과 loopback 임시 포트를 사용한다. 기존 `dev/compose.yaml`의 개발 volume이나 사용자 bw 설정에는 접근하지 않는다. 동시에 여러 실행이 가능하다. 강제 kill/호스트 중단 시 임시 컨테이너가 남을 수 있으며 해당 실행의 컨테이너만 정리한다.

## 수동 CLI / TUI 개발 환경

```sh
mise r dev:up
mise r dev:run -- tui
mise r dev:run -- item list
mise r dev:run -- --profile restricted item list
mise r dev:down
```

`dev:up`은 Python venv·고정 requirements를 준비하고, integration과 같은 checksum 검사와 SQL snapshot 복원으로 실제 Vaultwarden을 시작한다. Docker와 native credential store는 미리 실행되어 있어야 한다. `owner`(기본)와 `restricted` profile을 실제 로그인한다. 테스트 비밀번호는 `PUBLIC-INTEGRATION-ONLY-not-a-real-password!`다.

서버는 loopback 임시 포트를 사용하며 `.dev/demo/`의 config/state, 고유 `BWB_INSTANCE`와 keychain namespace로 개인 환경 및 integration 실행과 격리한다. `dev:up` 재실행은 기존 서버와 설정을 유지하며 잠긴 profile을 자동 unlock하지 않는다. snapshot 변경이나 서버 중단 시 `dev:down` 후 `dev:up`으로 새로 준비한다. 동일 checkout의 환경 준비·정리는 순차 실행한다.

`dev:run -- ...`은 준비된 환경에서 `go run . ...`을 실행하고 인자·터미널 입출력을 그대로 전달한다. 서버 준비나 전체 integration 테스트는 실행하지 않는다. 직접 실행하려면 별도 shell에서 다음을 사용한다:

```sh
source .dev/demo/env.sh
go run . tui
```

TUI를 종료해도 환경은 유지된다. Go daemon 코드를 수정했다면 `mise r dev:run -- daemon update`로 실행 중인 daemon도 갱신한다. `dev:down`은 준비한 두 profile을 logout해 keychain을 정리한 뒤 daemon·전용 컨테이너·개발 데이터를 삭제한다. 정리가 실패하면 환경 정보를 보존하므로 문제 해결 후 다시 실행한다. 개발 환경에서 별도 profile을 직접 로그인했다면 먼저 해당 profile을 logout한다. `env.sh`를 source한 shell은 정리 후 닫는다.

## 현재 검증 범위

- Go 제품 코드의 실제 prelogin, 잘못된 비밀번호 거부, 로그인, 명시적 refresh와 401 후 자동 refresh, sync.
- protected user key 복호화 후 8종·21개 항목의 필드를 평문 기대값과 비교.
- 조직 2개 × collection 2개, 다중 소속·동명 collection, folder 동시 소속, RSA 조직 키·item key, 읽기 전용/숨김 제한 계정, 실제 권한 회수 후 sync 반영.
- 서버 UUID, folder 관계, enum/boolean/null/빈 값/Unicode/여러 줄 값 유지.
- fixture/dump checksum 일치, SQL 무결성·foreign key 검사.
- Go 제품 코드의 암호문 변조·잘못된 키·padding 오류 거부, 독립 Python KDF/cipher vector 대조, 기대 필드 누락/값 변경 감지.
- 실제 목록 RPC/CLI의 21개 항목·조직·folder·collection 발견, AND 필터·검색·정렬·그룹 표시, 비밀 필드 비노출, 제한 profile 격리, 재시작 후 HTTP 없는 목록과 TTY 자동 unlock. Byte-budget pagination과 stale cursor 거부, 제어문자 escape, lock의 index 폐기는 단위 검사로 검증한다.
- 실제 바이너리의 daemon 자동 기동, 별도 CLI 프로세스의 비밀번호/빈 값/여러 줄 출력, offline read 시 HTTP 0회, 실패한 sync/재로그인 후 cache 유지.
- `item read`의 ID/정확한 이름 선택, 동명 후보 오류, 기본 마스킹·명시적 reveal·JSON 상태·발견한 필드 추출, `item totp`의 코드/seed 분리, `--sync` 검색. 기존 루트 read와 타입 접두사 필드 경로 거부도 검증한다. TOTP 계산은 RFC 6238의 SHA1/SHA256/SHA512 vector로 교차 검증한다.
- profile identity 변경 거부, 진행 중 login의 lock 취소, lock 후 조회 거부, 종료·재시작 후 오프라인 unlock과 HTTP 0회.
- 실제 OS keychain에 저장한 refresh token으로 재시작 후 sync, 반복 재시작, logout 후 unlock 거부·다른 profile 유지.
- cache version/KDF/계정/암호문 검증, native 쓰기 실패·취소·삭제 실패의 정리 목록 복구, SQLite chunk 원자 저장.
- Reprompt 항목의 CLI/TUI 재확인, 잘못된 비밀번호·취소, 항목별 권한과 sync 폐기, `insecure_skip_reprompt`의 profile 격리·명시적 reload.
- 실제 PTY에서 stdout pipe와 prompt 분리, echo 비활성/복원, 잘못된·긴 비밀번호, Ctrl-C/SIGINT, lock, 동시 CLI의 단일 prompt·취소 전파.
- refresh 거부 후 자동 재로그인·sync 1회 재시도, 인증 만료 중 cached read 유지, 비대화형·네트워크 장애는 prompt하지 않음.
- Config의 정책과 로그인 시 입력하는 서버·이메일의 identity 변경 차단, profile별 자동 잠금과 0s 비활성화, 재시작 후 TOML timeout 재적용, 명시적 reload 전후 정책 분리·잘못된 reload의 기존 정책 유지·잘못된 TOML 중 lock/logout, 만료 후 TTY 자동 unlock을 실제 CLI로 검증한다. 시간·조회 연장·정책 단축·취소·shutdown 경합은 synctest로 검사한다.
- 입력 제한, API redirect 차단·응답 크기·KDF 비용 상한, synctest 기반 작업 취소와 shutdown 검증.

Python runner는 복원된 서버의 URL을 테스트 자식 프로세스에 전달하고 `go test -tags=integration -count=1`을 실행한다. integration 태그 없이 `go test ./...`를 실행하면 Docker 없는 단위 검사만 실행한다. integration 태그가 있는데 서버 URL이 없으면 skip하지 않고 실패한다.

Python crypto helper는 snapshot 생성과 복원된 데이터의 독립 교차 검증에 사용한다. Go 제품 API와 CLI integration도 매번 실행한다. 공유 서버의 권한 변경 테스트 때문에 패키지는 `-p=1`로 순차 실행하고, 서로 다른 runner 실행은 독립 서버를 사용한다. CLI 테스트는 임시 state 경로와 고유 `BWB_INSTANCE`에 daemon을 실행한다. 정상 종료 시 각 profile을 logout해 전용 keychain 항목을 지우고 daemon/state를 정리한다. 강제 종료 시 해당 instance의 `logout`으로 남은 credential 정리 목록을 처리할 수 있다. 키·토큰은 memguard로 보관하며 명시적 파기는 단위 테스트로 검사한다. TUI·Argon2id·2FA·첨부파일은 후속이다. 현재 offline read는 살아 있는 세션이나 저장된 cache를 비밀번호로 unlock한 뒤 지원한다. 네이티브 keychain 경로는 실제 integration으로 검사한다. 장애 복구 단위 테스트와 서버 앞 proxy의 refresh 거부/연결 장애만 의도적으로 실패를 주입하며 로그인·sync·암호문은 실제 서버를 사용한다. `terminal.py`는 실제 바이너리의 TTY 동작을 검증하는 테스트 도구다.

## 데이터 / 타입 추가

1. 개인 타입은 `testdata/bitwarden.json`, 조직·collection 사례는 `testdata/organizations.json`에 합성 기대값과 고유 fixture ID를 추가/수정한다. 실계정 dump나 실제 비밀은 사용하지 않는다.
2. 새 cipher type이면 `dev/integration/run.py`의 `TYPES`, `internal/bitwarden/vault.go`의 payload/암호화 필드 목록, fixture 구조 검사를 함께 갱신한다. 새로운 필드가 암호문인지 enum인지 공식 스키마를 확인한다. 현재 문자열 필드를 암호화하는 규칙에 맞지 않는 타입은 해당 payload 변환을 명시적으로 추가한다.
3. 전용 임시 서버에서 API로 계정·암호문 데이터를 생성해 snapshot을 갱신한다:

   ```sh
   .dev/venv/bin/python dev/integration/run.py --rebuild
   ```

4. `bitwarden.json`, `organizations.json`, SQL dump, manifest를 함께 리뷰한다. 기존 타입을 제거하거나 실패 케이스를 건너뛰어 테스트를 통과시키지 않는다. snapshot 포맷/서버 버전 변경은 migration 호환성을 확인하고 manifest version을 갱신한다.
5. 일반 integration을 두 번 실행해 각 fresh restore가 통과하는지 확인한다. 단위 검사도 실행한 뒤 fixture·dump·manifest를 함께 커밋한다.

fixture를 바꾸고 snapshot 재생성을 잊으면 checksum 검사에서 실패한다. normal test는 기준 dump를 수정하지 않는다. 재생성은 안전한 난수 키/IV와 서버 UUID/시간을 사용하므로 byte-identical하지 않지만, 커밋된 dump의 매번 복원은 같은 계정·ID·값을 보장한다.

## 지속 실행

`lefthook.yml`의 pre-commit integration job이 매 커밋마다 fixture·snapshot 단위 검사와 fresh restore 두 번을 실행한다. 파일 필터 없이 항상 실행하며 수동 실행은 `mise run integration`이다. Docker가 꺼져 있거나 venv/의존성이 없거나 검사가 실패하면 커밋을 막고 skip/fallback으로 성공 처리하지 않는다. 고정 버전을 자동으로 latest로 바꾸지는 않는다. upstream 타입 추가 시 스키마와 이미지 업그레이드를 리뷰해 반영한다.

현재 coverage와 upstream 근거는 [snapshot 설명](../testdata/integration/README.md)에 있다. 테스트 계정 비밀번호와 SSH private key는 의도적으로 공개된 합성 데이터다. 이 snapshot을 공개 서버나 실제 계정으로 사용하지 않는다.

`dev/compose.yaml`은 별도의 수동 개발 서버 구성이다. pre-commit integration runner는 이 장기 volume을 사용하지 않는다.

API Key 검증은 복원된 임시 서버에서 테스트 계정의 개인 키를 발급해 실행한다. SQL dump나 manifest에는 API Key를 추가하지 않는다. 이메일 생략 CLI 로그인과 프로필 자동 조회, 401 후 재발급, daemon 재시작 후 keychain 복구, 오프라인 unlock, 폐기된 키의 명시적 재로그인 안내, logout, 실제 PTY의 세 단계 숨김 입력을 검사한다. `testdata/apikey.json`은 단위 테스트용 합성 암호문이며 비밀번호와 API Key는 의도적으로 공개된 테스트 값이다. keychain 장애·계정 불일치·잠금과 자격증명 읽기의 경합은 단위 테스트로 검사한다.

SSO는 모의 인증 서버와 브라우저 opener를 사용해 실제 CLI·daemon·PTY·OS keychain 경로를 검사한다. 도메인 자동 탐색, PKCE, 브라우저 실행 전 비밀번호 입력, 입력·브라우저 대기 중 취소, 오입력 후 브라우저 없이 재시도, profile 로그인 정보 재사용, 재시작 후 `cli` OAuth client의 refresh를 검증한다. 긴 SSO token의 실제 OS keychain 분할 저장·갱신·재시작 복구를 검사하고, 부분 쓰기 실패·취소·누락된 조각·logout 정리는 단위 테스트로 검증한다. 인증 예약의 취소·잠금·로그아웃·만료·동시 대기는 synctest로 검사한다. 실제 Okta/Bitwarden IdP 연동은 별도 검증이 필요하다.

- 상태 조회 integration은 daemon 설정 경로, reload 전후 실제 timeout, 잘못된 TOML 중 조회, verbose 조회의 잠금 시각 불변과 stopped 조회의 무기동을 검사한다.

## Release와 hook 관리

개발 도구 버전과 실행 명령은 `mise.toml` 및 `.mise/tasks/`에서 관리한다. `mise run init`으로 lefthook을 설치한다. Lefthook은 `mise run tidy`, `mise run fmt -- <staged files>`, `mise run test`를 호출한다. Integration 진입점은 `.mise/tasks/pre-commit/integration`(mise task `pre-commit:integration`)이며 기존 `mise run integration`을 실행한다. Lefthook 2.2의 일반 run 작업은 빈 커밋에서 생략되므로 integration만 script 작업으로 연결해 매 커밋 검사를 유지한다.

```sh
mise run release:check
mise run release -- --snapshot --skip=publish
```

Snapshot은 `dist/`에만 생성되며 GitHub에 게시하지 않는다. Native macOS keychain이 CGO와 Apple framework를 사용하므로 전체 릴리스 빌드는 Xcode Command Line Tools가 설치된 macOS에서 실행한다. 대상은 macOS/Linux의 amd64·arm64이며 archive와 SHA-256 `checksums.txt`를 생성한다. 현재 Windows 대화형 인증은 지원하지 않아 릴리스 대상에서 제외한다. macOS 서명·공증은 아직 하지 않는다.

바이너리 버전은 Go 1.27의 build info를 사용한다. `go build`와 GoReleaser 모두 빌드한 HEAD의 `vA.B.C` 태그를 자동 기록하므로 별도 ldflags 주입은 없다. 변경된 작업 트리는 `+dirty`, 태그 없는 커밋은 Go pseudo-version을 표시한다. `bwb version`/`--version`으로 CLI, `bwb daemon status`의 `VERSION`으로 실행 중 daemon을 확인한다. `-buildvcs=false`나 Git 정보 없는 소스 빌드는 `dev`를 표시할 수 있다.

정식 게시 진입점은 `mise run release`다. 깨끗한 작업 트리의 `vX.Y.Z` 태그에서 `GITHUB_TOKEN`을 제공하면 게시한다. `vX.Y.Z-rc.1` 같은 태그는 prerelease로 분류한다. `.github/workflows/release.yml`은 `v*` 태그 push에서만 실행되며 macOS runner에 mise 도구를 설치하고 동일한 task를 호출한다. Workflow에 별도 빌드 명령은 없다. 일반 PR/push CI는 추가하지 않는다. Release task는 단위 테스트와 GoReleaser 설정 검사를 먼저 실행한다. Docker·OS keychain을 사용하는 full integration은 기존 로컬 pre-commit에서 계속 실행한다.

TUI 테스트는 `dev/integration/tui.py`에서 실제 controlling PTY와 pyte 화면 해석기를 사용한다. 의존성을 갱신하려면 `.dev/venv/bin/python -m pip install -r dev/integration/requirements.txt`를 실행한다. Snapshot integration마다 검색 tree·상세 표시/숨김·resize·외부 lock·unlock 복귀·Profiles 전환·서버 입력을 통한 최초 로그인·reload/logout/재로그인을 검증한다. Master password의 terminal echo 여부도 검사한다. Tree filter·layout 순수성·늦은 응답 폐기는 단위 테스트, background 조회/activity의 idle 경계는 synctest로 검증한다.

## 로컬 go-daemonkit 검증

go-daemonkit `8f5523de9203`은 `go.mod`와 `go.sum`으로 고정되어 있으며 일반 빌드에 workspace는 필요하지 않다.
로컬 프레임워크 변경을 함께 검증할 때만 인접 checkout을 준비하고 bwb 루트에서 workspace를 만든다.
`go.work`와 `go.work.sum`은 커밋하지 않는다.

```sh
go work init . ../go-daemonkit
go test ./...
mise run integration
```

이미 workspace가 있으면 `go work use ../go-daemonkit`을 사용한다.
의존성을 변경할 때는 `GOWORK=off go test ./...`로 workspace 없이도 검증하고
`go mod tidy`로 `go.sum`을 갱신한다.

launch 환경은 기존처럼 상속하되 `BWB_CONFIG`를 자식 환경에서만 교체한다.
환경 원문을 journal에 저장하지 않으며 다른 launcher의 crash recovery까지 복원하지 않는다.
reload는 파일 읽기부터 정책 publication까지 직렬화하고, 오류·commit 전 취소에는 기존 정책을 유지한다.
