# t683 Verdict — Security Guardian weak-crypto 산문 오탐 (ISSUE #1708)

Date: 2026-09-13 · Lane: lane-1 · Branch: `WT-guardian-md5-prose` (develop `74d872aaf` 기점) · Tier S, Class B (run→sync)

## Claim

Security Guardian의 weak-crypto 클래스가 산문 속 맨 단어 md5/sha1/ECB에 발화하던 결함을, 카드 [HARD]가 요구한 방향 — **암호 API 호출 문맥으로의 축소** (산문 파일 전면 제외 아님) — 로 수리하고, 양방향 회귀와 true-positive 보존을 테스트로 확보한다.

## Changed files

- `internal/hook/security/patterns.go` — weak-crypto 클래스: 맨 단어 패턴 2개(`(?i)\b(MD5|SHA1)\b`, `(?i)\bECB\b`) → 콜-컨텍스트 패턴 4개로 교체 (총 패턴 수 28→30, AC-SG-003 상한 30 이내)
- `internal/hook/security/weak_crypto_test.go` (신규) — 산문 침묵 / 코드 검출 / 전후 보존 3개 테스트
- `.moai/reports/t683/verdict.md` (본 문서)

`scan.go`는 무변경 — 카드가 지적한 "파일 형식 필터 부재"는 사실이지만, [HARD]가 지정한 수리 축은 패턴 측 축소이며 전면 제외는 금지됐다.

## New weak-crypto patterns (4)

1. `(?i)\b(md5|sha1)(\(|::|\.[A-Za-z_][A-Za-z0-9_]*\()` — 호출/한정 접근: `md5(password)`, `md5.New()`, `SHA1.hexdigest(`, `Digest::MD5`, `Md5::new()`. 토큰과 문장부호 사이 공백 불허 → 산문의 "MD5 (RFC 6151)" 형태 침묵.
2. `(?i)\b(createHash|getInstance)\(\s*["'\x60](md5|sha1)["'\x60]\s*\)` — 명명 알고리즘 생성자: `crypto.createHash('md5')`, `MessageDigest.getInstance("MD5")`.
3. `(?i)crypto/(md5|sha1)\b` — Go import 경로.
4. `(?i)(\bMODE_ECB\b|\bCipherMode\s*\.\s*ECB\b|["'\x60][^"'\x60\n]*\bECB\b[^"'\x60\n]*["'\x60])` — 코드 문맥의 ECB만: `AES.MODE_ECB`, `CipherMode.ECB`, 인용된 cipher spec(`"AES/ECB/PKCS5Padding"`).

## Evidence (본 run의 실측)

| # | Command | Observed |
|---|---------|----------|
| E1 | RED-first: 신규 3테스트를 **구형 패턴 트리에서** 실행 | `TestWeakCryptoProseSilent` FAIL — 산문 7줄 전부 오탐(결함 재현) · `TestWeakCryptoCodeDetected` FAIL — `AES.MODE_ECB` 미검출(구형 패턴의 별도 사각: `_`는 단어 문자라 `\bECB\b`가 MODE_ECB를 못 봄) |
| E2 | 패턴 교체 후 `go test ./internal/hook/security/ -count=1` | **ok** — 신규 3테스트 + 기존 전체(prefilter·scanner·ast_grep·guardian·coverage) 전부 GREEN |
| E3 | `TestWeakCryptoTruePositivesPreserved` | PASS — 구형 맨단어 baseline(12픽스처 중 11) 대비 신규 콜-컨텍스트 검출 12 — **true-positive 감소 없음, 1건 증가(MODE_ECB)** |
| E4 | 변이 프로브: 구형 `(?i)\b(MD5|SHA1)\b` 재삽입 → ProseSilent 실행 | FAIL(오탐 5건 재발) → 원복 후 GREEN 재확인 — 침묵 테스트의 RED가 관측됨 |
| E5 | `gofmt -l`(빈 출력) + `go vet ./internal/hook/security/`(무소식) | 클린 |
| E6 | 실시간 dogfood 관측: 패턴 편집 직후 Guardian이 **본 카드의 테스트 파일을 29건 weak-crypto 오탐** | 맨단어 패턴이 테스트 픽스처·주석의 md5/sha1/ECB를 건 전점 — 패턴 교체 후 동일 파일 잔여 1건(아래 Residual) |

## 이슈 원문 정합 (t663 판정서 오탐 재현)

이슈가 보고한 표면 — "파일 동일성 확인을 기록한 판정서·증거 로그가 finding으로 보고" — 는 E1의 산문 픽스처(`파일 동일성 확인은 md5 해시로 기록했다 (t663 판정서).` 등)로 문자 단위로 재현됐고, E2로 침묵이 확보됐다.

## Gaps

- `internal/hook/security` 외 패키지는 로컬 미실행(CI 몫). Layer 2(ScanDiff)·Layer 3(CrossFileScan)은 동일 패턴 테이블을 소비하므로 패키지 테스트에서 간접 커버되나, 각 레이어의 diff-hunk 단위 동작은 별도 재측정하지 않았다.
- Go/Python/JS 이외 언어의 ECB·해시 호출 형태는 픽스처 수준 커버(12건)이며 전수 아니다.
- 산문 중 **인용부호 안의 ECB**("the 'ECB' mode")는 여전히 적중한다 — 인용 cipher spec과 구분 불가한 축(Residual 기재).

## Residual risk

- **자기-스캔 잔여 1건**: patterns.go 자신의 패턴 리터럴 문자열(`createHash('md5')` 등)이 콜-컨텍스트에 적중한다 — 패턴 테이블 소스는 정의상 weak-crypto 코드 형태를 담는 유일한 정당 표면으로, 제외하려면 파일 수준 필터가 필요해 본 카드의 [HARD](전면 제외 금지)와 충돌한다. 그대로 두며 기록한다.
- 인용부호 산문의 ECB 언급은 잔여 오탐 축이다.
- 본 수리는 **릴리스 전까지 사용자에게 전달되지 않는다**(t680과 동일 축 — 배포가 전달 행위).

🗿 MoAI
