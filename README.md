# bwb

Bitwarden를 CLI와 TUI로 사용하는 Go 클라이언트. 현재는 요구사항 정립과 개발 데이터 준비 단계이며 실행 가능한 `bwb` 명령은 아직 없다.

- [요구사항과 인터페이스 제안](docs/design.md)
- [Full integration 개발 환경](dev/README.md)
- [제안 config](dev/config.example.toml)

설계 방향: CLI + TUI, go-daemonkit daemon, Bitwarden API 직접 호출.
서버 login과 로컬 unlock을 분리하고, 명령이 필요한 인증을 거쳐 재개된다.
풀린 키는 daemon 메모리, refresh token은 OS credential store, 암호화된 vault는 로컬 SQLite에 보관한다.
BW_SESSION이나 bw/rbw 실행 파일에 의존하지 않는다. 세부 설계와 구현 순서는 위 문서를 참고한다.

필수 [integration snapshot](testdata/integration/README.md)은 8종·13개 항목을 포함한다. `dev/integration/run.py`가 전용 서버에 복원해 실제 API를 검사하고, Lefthook pre-commit이 매 커밋마다 실행한다. 실행 준비는 개발 가이드를 따른다.
