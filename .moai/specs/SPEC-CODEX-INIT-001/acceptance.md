# SPEC-CODEX-INIT-001 — 인수 기준

모든 기준은 기계 판정이다. codex 실 바이너리에 의존하는 항목은 없다 — 생성기 호출 seam · exec seam · 프롬프트 응답 seam 을 스텁해 판정한다.

**공통 fixture — 배선 상태 5종** (아래 여러 AC 가 이 표를 참조한다):

| 이름 | `.codex/` 상태 | 기대 판정 |
|---|---|---|
| S1 | 디렉터리 없음 | 미배선 |
| S2 | 디렉터리 있고 **비어 있음** | 미배선 |
| S3 | `hooks.json` 만 | 미배선(부분) |
| S4 | `config.toml` 만 | 미배선(부분) |
| S5 | 둘 다 | 배선됨 |

디렉터리 존재만 보는 구현은 S2 를 배선으로 읽고 훅 0개 상태로 기동한다 — S2·S3·S4 가 그 구현을 떨어뜨리는 행이다.

## AC-CI-001 — 상태 × 동사 교차곱 (REQ-CI-001, REQ-CI-002)

**5상태 × 2동사 = 10칸 전부** 를 표 시험으로 돈다. 한 동사만 시험하면 나머지 동사가 미배선으로 기동하는 구현이 통과한다.

- **Given** S1~S5 각각 + 생성기 호출·exec 호출을 세는 seam
- **When** `moai codex cli` 와 `moai codex app` 을 각각 실행하면
- **Then** S1~S4 에서는 **두 동사 모두** 제안이 뜨고 제안 전 exec 0회다.
- **And** S5 에서는 **두 동사 모두** 제안 없이 기동하고 생성기 호출 0회다.
- **And** 10칸의 결과가 동사에 따라 갈리지 않는다 (같은 상태면 같은 판정).

## AC-CI-002 — 판정 로직 단일성 (REQ-CI-001)

- **Given** 런처의 배선 상태 판정 함수 호출 횟수를 세는 seam
- **When** S1~S5 각각에서 제안 여부를 결정하면
- **Then** 그 함수 호출이 각 실행마다 ≥1회다 — 이 SPEC 이 두 번째 판정 로직을 갖지 않았다는 실행 증거.
- **And** 이 SPEC 이 추가하는 코드에 `.codex/hooks.json` / `.codex/config.toml` 존재 검사를 직접 수행하는 경로가 없다 (파일 검사 grep 0건).

## AC-CI-003 — 거절 경로 (REQ-CI-003)

- **Given** S1~S4 각각 + 프로젝트 트리 전체의 파일 목록·mtime 스냅샷
- **When** 두 동사 각각에서 제안을 **거절** 하면
- **Then** 생성기 호출 0회, exec 0회, 스냅샷 **무변경** 이다 (8칸 전부).
- **And** 종료 코드는 오류가 아니라 취소를 뜻하는 값이고, 출력은 미배선 상태와 조치를 명명한다.

## AC-CI-004 — 수락 경로와 생성기 위임 (REQ-CI-004)

- **Given** S1~S4 각각 + 생성기 호출의 **인자까지** 포착하는 seam
- **When** 두 동사 각각에서 제안을 **수락** 하면
- **Then** 생성기가 **정확히 1회** 호출되고 그 인자에 codex 에이전트 선택이 포함된다 (8칸 전부).
- **And** 이 SPEC 이 추가하는 코드에 배선 파일을 쓰는 호출이 0건이다 (`hooks.json` / `config.toml` 쓰기 grep 0건).
- **And Given** 비대화형 실행 (프롬프트 불가)
- **Then** 제안 대신 미배선 상태와 조치를 보고하고, 생성기 호출 0회 · exec 0회로 끝난다 — 자동화가 프롬프트에서 멈추지 않는다.

## AC-CI-005 — 지시 계약: 연결 생성 (REQ-CI-005, REQ-CI-006)

**판정 대상 형태를 먼저 고정한다.** `CLAUDE.md` 는 `@AGENTS.md` 한 줄로 `AGENTS.md` 를 가져오며, 이 줄이 셈의 대상이다 (이 저장소 `CLAUDE.md` 가 쓰는 형태와 동일).

fixture 5종을 각각 초기화한다:

| 이름 | 초기 상태 | 초기화 후 기대 |
|---|---|---|
| I1 | 둘 다 없음 | 둘 다 생성, `CLAUDE.md` 에 `@AGENTS.md` 줄 1건 |
| I2 | `AGENTS.md` 만 (사용자 내용) | `AGENTS.md` **바이트 무변경**, `CLAUDE.md` 생성 + 줄 1건 |
| I3 | `CLAUDE.md` 만 (사용자 내용) | 기존 내용 **보존** + 줄만 추가, `AGENTS.md` 생성 |
| I4 | 둘 다 있는데 **줄이 없음** | 두 파일 내용 보존 + `CLAUDE.md` 에 줄만 추가 |
| I5 | 둘 다 있고 줄도 있음 | 두 파일 **바이트 무변경** |

- **And** I2 · I3 · I4 의 기존 사용자 내용은 초기화 후에도 **부분 문자열로 온전히 존재** 한다 (재작성되지 않았다는 증거).
- **And** 모든 fixture 에서 `@AGENTS.md` 줄 수는 정확히 1이다.

## AC-CI-006 — 멱등: 2회 실행 바이트 비교 (REQ-CI-007)

1회 실행 후 줄 수를 세는 것은 멱등의 판정이 아니다 — 2회차에 중복이 생기는 구현을 잡지 못한다.

- **Given** I1~I5 **그리고 아래 AC-CI-007 의 로컬 파일 fixture 2종** (총 7종)
- **When** 각 fixture 에 초기화를 **연속 2회** 실행하면
- **Then** 1회 후 스냅샷과 2회 후 스냅샷의 **모든 지시 파일이 바이트 동일** 하다 (`AGENTS.md` · `CLAUDE.md` · 존재한다면 `CLAUDE.local.md`).
- **And** 두 시점 모두 `@AGENTS.md` 줄 수가 1이다.

## AC-CI-007 — 로컬 지시 파일 도달성 (REQ-CI-008)

**파일명 grep 으로 판정하지 않는다** — 불활성 산문 한 줄도 파일명을 포함하므로 통과해 버린다. **sentinel 내용의 도달** 로 판정한다.

- **Given** `CLAUDE.local.md` 에 고유 문자열 `SENTINEL-LOCAL-7q7` 를 심은 fixture (L1)
- **When** 초기화한 뒤, 두 하네스 각각의 진입 파일(`AGENTS.md` / `CLAUDE.md`)에서 시작해 import 지시문을 따라가며 도달 가능한 내용을 모으면
- **Then** **양쪽 모두** 에서 `SENTINEL-LOCAL-7q7` 에 도달한다.
- **And** `CLAUDE.local.md` 자체는 바이트 무변경이다.
- **And** 그 파일을 **직접** 가리키는 지시문은 `AGENTS.md` 와 `CLAUDE.md` 를 통틀어 정확히 1건이다 — 두 곳에서 각각 가리키면 같은 내용이 두 번 로드된다.
- **And** 도달에 쓰인 지시문은 AC-CI-005 가 고정한 것과 같은 실행되는 import 형태다 (산문 언급이 아니다).
- **And Given** `CLAUDE.local.md` 가 없는 fixture (L2)
- **Then** 그 파일을 가리키는 지시문이 0건이다 — 없는 파일을 가리키지 않는다.

## AC-CI-008 — 프로젝트 밖 무쓰기 (REQ-CI-005 ~ REQ-CI-008)

- **Given** 임시 홈 디렉터리 + 임시 `CODEX_HOME` 의 파일 목록·mtime 스냅샷
- **When** I1~I5 · L1 · L2 전부에 초기화를 실행한 뒤 다시 스냅샷하면
- **Then** 두 스냅샷이 동일하다 — 초기화는 프로젝트 트리 안에만 쓴다.

## AC-CI-009 — 게이트 (전 REQ)

- **When** `go build ./...` · `go vet ./...` · `GOOS=windows go vet ./...` · `golangci-lint run` · `go test ./internal/cli/... -run 'Codex' -timeout 600s` 를 실행하면
- **Then** 전부 rc 0 이다.
- **And** 템플릿 중립성 가드(`MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...`)가 통과하고, 이 SPEC 이 추가한 사용자 노출 문안에 SPEC ID · 카드 id · 내부 날짜 · 커밋 SHA 가 0건이다.

---

## 판정 제외 (근거 명시)

- **실제 `moai init --agent codex` 실행 결과**: 생성기 자체의 산출물은 SPEC-CODEX-WIRING-001 의 인수 기준이 판정한다. 이 SPEC 은 **호출했는가** 까지만 본다 — 생성기가 무엇을 까는지를 여기서 다시 판정하면 두 SPEC 이 같은 것을 두 번 정의한다.
- **실제 Codex 세션에서의 지시 로딩**: codex 가 `AGENTS.md` 를 읽는다는 것은 외부 런타임의 동작이라 CI 에서 판정하지 않는다. 운영자 수동 확인 항목으로 `progress.md` 에 남긴다 — 확인 방법은 격리된 `CODEX_HOME` 에서 `codex debug prompt-input` 으로 지시 로딩 여부 관측 (SPEC-CODEX-SKILLS-CANONICAL-001 이 세운 방식).
