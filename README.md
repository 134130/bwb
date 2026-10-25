# bwb

Bitwarden를 CLI와 TUI로 사용하는 Go 클라이언트. 현재는 Cobra CLI → go-daemonkit RPC → 메모리 세션 → Bitwarden 직접 API까지 구현했다. TUI는 후속이다.

- [요구사항과 인터페이스 제안](docs/design.md)
- [Full integration 개발 환경](dev/README.md)
- [제안 config](dev/config.example.toml)

최종 설계 방향: CLI + TUI, go-daemonkit daemon, Bitwarden API 직접 호출.
서버 login과 로컬 unlock을 분리하고, 명령이 필요한 인증을 거쳐 재개된다.
풀린 키는 daemon 메모리, refresh token은 OS credential store, 암호화된 vault는 로컬 SQLite에 보관한다.
BW_SESSION이나 bw/rbw 실행 파일에 의존하지 않는다. 세부 설계와 구현 순서는 위 문서를 참고한다.

필수 [integration snapshot](testdata/integration/README.md)은 8종·13개 항목을 포함한다. `dev/integration/run.py`가 전용 서버에 복원해 실제 API를 검사하고, Lefthook pre-commit이 매 커밋마다 실행한다. 실행 준비는 개발 가이드를 따른다.

## 현재 실행 범위

```sh
go build -o .dev/bwb .
.dev/bwb login --profile personal --server SERVER_URL --email EMAIL --password-stdin < PASSWORD_INPUT
.dev/bwb read --profile personal ITEM_ID login/password
.dev/bwb sync --profile personal
.dev/bwb status --profile personal
.dev/bwb lock --profile personal
.dev/bwb daemon stop
```

`PASSWORD_INPUT`은 안전하게 공급한 stdin을 뜻한다. 비밀번호를 argv에 넣거나 echo가 켜진 터미널에 직접 입력하지 않는다. 마지막 LF/CRLF 한 개만 제거하며 공백은 유지한다. `read`는 줄바꿈을 추가하지 않는다. `name`, `notes`, `fields/NAME`(슬래시 포함), `TYPE/FIELD` 지원. 현재 RPC 응답 제한 때문에 필드 하나는 최대 10,000 bytes다.

첫 login이 같은 바이너리의 `daemon run`을 자동 시작한다. profile 기본값은 `default`. login은 API 인증과 첫 sync를 함께 수행하고, 이후 별도 CLI 프로세스의 read는 HTTP 없이 cached ciphertext를 복호화한다. sync 실패나 잘못된 재로그인은 기존 snapshot/세션을 유지한다. refresh는 sync 중 401에 한 번 시도한다. 자동 login hook은 아직 없다.

`status`와 `lock`은 daemon을 시작하지 않는다. lock은 선택 profile의 진행 중 작업을 취소하고 키·토큰·snapshot을 제거한다. 이미 반환한 평문이나 전송 중인 RPC 응답은 회수하지 않는다. daemon 종료는 전체 profile을 잠그며 재시작 후 login이 필요하다. profile은 현재 메모리 이름이며 config 파일 해석·keychain·디스크 vault cache·idle timeout은 후속이다. device ID는 로그인마다 생성한다.

```text
main.go → internal/cli (Cobra)
              ↓ internal/vaultapi (typed RPC)
          internal/daemon (Manager, App, handlers)
              ↓
          internal/vault (profile, cancellation, session lifetime)
              ↓
          internal/bitwarden (HTTP, crypto, encrypted snapshot)
```

키와 access/refresh token의 장기 보관은 `memguard` locked memory를 사용한다. 평문 snapshot은 보관하지 않는다. crypto·HTTP·JSON·RPC 처리 중 임시 Go 문자열/버퍼는 생기며 모든 복사본의 즉시 zeroize를 보장하지 않는다. 메모리 보호 실패 시 평문 fallback은 없다. 같은 OS 사용자 프로세스는 daemon 접근 가능하다. daemonkit의 실행 파일·로그·SQLite 인프라는 디스크에 생기지만 vault·비밀번호·token을 거기에 기록하지 않는다. Unix 격리 경로는 `XDG_STATE_HOME`으로 지정한다.

지원 범위는 PBKDF2(100000–2000000 iterations), 개인 vault, AES-CBC/HMAC `2.` 암호문, 8종의 기본 필드다. Argon2id, 2FA/SSO, 조직 vault, item별 key, reprompt, passkey는 아직 지원하지 않는다. 공식 Bitwarden 서버와의 상호운용성은 별도 검증이 필요하다.

```sh
go test ./...
mise run integration
```

integration은 실제 Vaultwarden에서 Go API 코드와 빌드된 CLI·daemon을 실행해 snapshot 기대값, profile 격리, offline read의 HTTP 0회, lock 중 로그인 취소, 재시작 잠금을 검사한다. Python은 서버 복원과 snapshot 생성에만 쓰며 제품의 암호화 구현은 Go 표준 라이브러리를 사용한다.
