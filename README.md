# bwb

Bitwarden를 CLI와 TUI로 사용하는 Go 클라이언트. 현재는 Cobra 기반 `read` 명령과 직접 API 로그인·sync·복호화까지 구현했다. daemon/TUI는 다음 단계다.

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
go run . --help
go run . read ITEM_ID login/password --server SERVER_URL --email EMAIL --password-stdin
```

`read`의 비밀번호는 stdin으로 공급한다. 비밀번호를 argv에 넣거나 echo가 켜진 터미널에 직접 입력하지 않는다. 마지막 LF/CRLF 한 개만 제거하며 비밀번호의 공백은 유지한다. 출력에는 줄바꿈을 추가하지 않는다. `name`, `notes`, `fields/NAME`, `TYPE/FIELD`를 읽을 수 있다.

이 첫 구현은 호출마다 login/sync하고 종료 시 key buffer를 지운다. 저장된 인증·profile/config 해석·대화형 prompt·URI reference·daemon cache는 아직 없다. device ID도 호출마다 새로 생성한다. 현재 오류 종료 코드는 1이며 세부 코드 분류는 후속이다.

지원 범위는 PBKDF2(100000–2000000 iterations), 개인 vault, AES-CBC/HMAC `2.` 암호문, 8종의 기본 필드다. Argon2id, 2FA/SSO, 조직 vault, item별 key, reprompt, passkey는 아직 지원하지 않는다. 메모리는 일반 Go buffer이며 daemon용 locked memory 보장을 하지 않는다. API 서버와의 공식 Bitwarden 상호운용성은 별도 검증이 필요하다.

```sh
go test ./...
mise run integration
```

integration은 실제 Vaultwarden에서 Go API 코드와 Cobra 명령을 실행해 snapshot 기대값과 비교한다. Python은 서버 복원과 snapshot 생성에만 쓰며 제품의 암호화 구현은 Go 표준 라이브러리를 사용한다.
