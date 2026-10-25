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

## 현재 검증 범위

- Go 제품 코드의 실제 prelogin, 잘못된 비밀번호 거부, 로그인, 명시적 refresh와 401 후 자동 refresh, sync.
- protected user key 복호화 후 8종·19개 항목의 필드를 평문 기대값과 비교.
- 조직 2개 × collection 2개, 다중 소속·동명 collection, folder 동시 소속, RSA 조직 키·item key, 읽기 전용/숨김 제한 계정, 실제 권한 회수 후 sync 반영.
- 서버 UUID, folder 관계, enum/boolean/null/빈 값/Unicode/여러 줄 값 유지.
- fixture/dump checksum 일치, SQL 무결성·foreign key 검사.
- Go 제품 코드의 암호문 변조·잘못된 키·padding 오류 거부, 독립 Python KDF/cipher vector 대조, 기대 필드 누락/값 변경 감지.
- 실제 바이너리의 daemon 자동 기동, 별도 CLI 프로세스의 비밀번호/빈 값/여러 줄 출력, offline read 시 HTTP 0회, 실패한 sync/재로그인 후 cache 유지.
- profile identity 변경 거부, 진행 중 login의 lock 취소, lock 후 조회 거부, 종료·재시작 후 오프라인 unlock과 HTTP 0회.
- 실제 OS keychain에 저장한 refresh token으로 재시작 후 sync, 반복 재시작, logout 후 unlock 거부·다른 profile 유지.
- cache version/KDF/계정/암호문 검증, native 쓰기 실패·취소·삭제 실패의 정리 목록 복구, SQLite chunk 원자 저장.
- 실제 PTY에서 stdout pipe와 prompt 분리, echo 비활성/복원, 잘못된·긴 비밀번호, Ctrl-C/SIGINT, lock, 동시 CLI의 단일 prompt·취소 전파.
- refresh 거부 후 자동 재로그인·sync 1회 재시도, 인증 만료 중 cached read 유지, 비대화형·네트워크 장애는 prompt하지 않음.
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
