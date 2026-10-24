# Synthetic integration snapshot v1

- `vaultwarden.sql`: 실제 Vaultwarden API로 등록·암호화 저장한 계정과 vault의 SQLite logical dump.
- `manifest.json`: 서버 image digest, 공개 테스트 자격정보, fixture ID→서버 ID, 타입 목록, dump/fixture SHA-256.
- `../bitwarden.json`: 복호화 결과의 기대값. 서버 응답 mock이 아니다.

재생성/실행 절차는 [개발 가이드](../../dev/README.md)를 따른다. 일반 실행은 dump를 복원하며 seed/import하지 않는다. snapshot은 테스트 자격정보와 해시를 포함하므로 합성 계정만 허용한다. 서버의 RSA signing key는 dump에 포함하지 않는다. 복원된 서버가 새 signing key를 만들고 테스트에서 fresh login/refresh를 수행한다. 이전 JWT 재사용에 의존하지 않는다.

## Cipher type coverage

| ID | Payload | 현재 케이스 |
| --- | --- | --- |
| 1 | login | URI, 중복 이름, hidden custom field, 빈 값, Unicode, 공개 TOTP seed |
| 2 | secureNote | 여러 줄 notes |
| 3 | card | 테스트 카드 번호, 만료일, 보안 코드 |
| 4 | identity | 이름/주소/이메일, nullable 필드 |
| 5 | sshKey | 유효한 공개 테스트 Ed25519 private/public key와 fingerprint |
| 6 | bankAccount | 은행/계좌/라우팅 번호/PIN/IBAN 등 |
| 7 | driversLicense | 이름/생년월일/면허번호/발급·만료일 등 |
| 8 | passport | 여권번호/국적/발급국/생년월일/발급·만료일 등 |

13 items, 2 folders, 1 personal account, PBKDF2-SHA256 600000 iterations, AES-CBC-256 + HMAC-SHA256 (`2.` cipher string).

타입 목록 근거: [Bitwarden CipherType](https://github.com/bitwarden/clients/blob/c0ca72a0c83cc8b7f990138ac04961cf63d985bc/libs/common/src/vault/enums/cipher-type.ts).
추가 타입 필드는 같은 revision의 `libs/common/src/vault/models/data/`를 확인했다.
서버 근거: [Vaultwarden 1.37.4 CipherData](https://github.com/dani-garcia/vaultwarden/blob/1.37.4/src/api/core/ciphers.rs).
`Bitwarden-Client-Version: 2026.10.0`을 보내며, 헤더가 없으면 이 서버는 SSH key를 sync에서 제외한다.

이 matrix는 타입별 기본 round-trip coverage다. 모든 optional field·키 계층·인증 방식이 검증되었다는 의미는 아니다. 후속 사례는 기존 fixture에 추가하고 snapshot을 재생성한다. 테스트 crypto helper의 round-trip은 독립적인 공식 client 상호운용성 검증을 대신하지 않는다.

일반 restore 검증은 Go 제품 코드(`internal/bitwarden`)와 Cobra 명령(`internal/cli`)의 integration test가 수행한다. Python crypto는 dump 생성 및 독립 golden 값의 출처이며, Go 구현이 Python 결과와 일치하는지를 비교한다.
