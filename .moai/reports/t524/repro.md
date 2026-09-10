# t524 재현 — plan-auditor 추적성 검사의 두 겹 약점 (develop 트리)

- 카드: t524 (Class B, 재현 먼저). 짝 카드: t518 (`SPEC-SPEC-LINT-BLIND-AXES-001`, 도구 쪽)
- 워크트리: `.claude/worktrees/t524`, 브랜치 `WT-auditor-req-shorthand`
- 기반 정렬: `origin/develop` `d060e0d13`에서 생성했다. 이어서 `git merge --no-ff 6d228ea19`로 기반을 맞춰 병합 커밋 `a6c9e406c`를 만들었다(`HEAD^2` = `6d228ea19`, tree `e906eab95` = `6d228ea19^{tree}`).
- 측정 트리: HEAD `a6c9e406c73ff99fbbef20599d9c5ac4603f593d`
- lint 계기: 이 트리에서 빌드한 바이너리(`go build -o <scratchpad>/moai-t524 ./cmd/moai`, build_exit=0)를 경로로 직접 호출했다. `version` 출력은 `v3.1.3 none built unknown`이다. ldflags 없이 빌드해서 커밋 스탬프가 없으므로, 계기 좌표는 위 HEAD로 귀속한다. 설치된 `~/go/bin/moai`는 쓰지 않았다.

## 원본 사례 (t499)

`.moai/reports/t499/plan-audit-iter2.md` 178-184행 원문 요지는 다음과 같다.

- iter1은 `REQ-CPW-001, 002` 꼴의 매핑을 전체 ID로 세지 않았다.
- iter1은 같은 문장에서 `moai spec lint`의 `CoverageIncomplete` 침묵을 방증으로 인용했다(`plan-audit-iter1.md:29`, `:212`).
- 그 침묵은 lint가 해당 SPEC의 REQ를 하나도 수집하지 못해서 생긴 것이었다. 감사자는 이를 "순환"이라고 스스로 기록했다.

## 1. 문서 쪽 — 규칙 부재 (정정 전)

증거: `repro/doctrine-grep.txt`

```text
internal/template/templates/.claude/agents/moai/plan-auditor.md silence_or_corroboration_clause=0 shorthand_normalization=0 control_traceability=6 control_AC-4=1
.claude/agents/moai/plan-auditor.md silence_or_corroboration_clause=0 shorthand_normalization=0 control_traceability=6 control_AC-4=1
grep_exit=0
```

- 두 사본 모두 다음 두 규칙이 0건이다.
  - 다른 도구의 침묵을 방증으로 인용하는 것에 관한 규칙
  - 축약 REQ 표기를 정규화하는 규칙
- 같은 grep이 `Traceability` 6건과 `AC-4` 1건을 잡으므로, 0건은 파일을 못 읽어서 나온 결과가 아니다.
- 추적성 검사는 AC-4/AC-5 체크리스트 두 줄과 Group A의 `^### REQ-` grep이 전부다. 매핑 열을 전체 ID로 세는 명령은 없다.

## 2. 도구 쪽 — spec lint 추적성 규칙의 수집 범위

코드 근거:
- `internal/spec/lint.go:1085` `CoverageRule.Check`는 `len(doc.REQs) == 0`이면 `nil`을 반환한다.
- `internal/spec/ears.go:128` `ExtractRequirementMappings`는 `maps REQ-…(, REQ-…)*` 형식에서만 ID를 모은다. 각 항목에 `REQ-` 접두사가 있어야 한다.

| 픽스처 | REQ 정의 형식 | 매핑 형식 | 실제 누락 | lint 추적성 결과 | 증거 |
|---|---|---|---|---|---|
| A | `- REQ-X-00N:` 목록 | `(maps REQ-FIXA-001, 002)` 축약 | 없음 | `CoverageIncomplete` `REQ-FIXA-002` — 축약 두 번째 항목을 읽지 못함(거짓 경고) | `repro/lint-fixA.txt` |
| B (대조) | 목록 | `(maps REQ-FIXB-001, REQ-FIXB-002)` 전체 ID | 없음 | 추적성 경고 0건 | `repro/lint-fixB.txt` |
| C | 목록 | 표 `\| AC \| REQ-X-001, REQ-X-002 \|` (`maps` 없음) | 없음 | `CoverageIncomplete` 3건 — 표 매핑을 전혀 읽지 않음(거짓 경고) | `repro/lint-fixC.txt` |
| D | `### REQ-X-00N` 제목 | `(maps REQ-FIXD-001, 002)` 축약 | **`REQ-FIXD-003` 무매핑(진짜 누락)** | **추적성 경고 0건 — 침묵** | `repro/lint-fixD.txt` |

네 번 모두 `lint_exit=0`이다(`*.exit`). 추적성 경고는 권고 등급이라 종료 코드를 바꾸지 않는다.

**결론:**
- **D가 순환 인용의 현재 형태다.** plan-auditor Group A는 REQ를 `^### REQ-` 제목으로 세고, lint는 바로 그 제목 형식의 정의를 수집하지 못한다. 감사자는 REQ 3개를 보는데 lint는 0개를 보고 침묵한다. 이 침묵을 방증으로 인용하면 진짜 누락(`REQ-FIXD-003`)이 통과한다.
- A와 C는 반대 방향이다. lint가 거짓 경고를 내므로, 경고가 늘 뜨는 SPEC에서는 lint의 추적성 결과를 신호로 쓸 수 없다.

## 확인하지 못한 것

- plan-auditor 에이전트 자체를 픽스처에 돌려 PASS를 내는지는 실행하지 않았다. LLM 판정이라 Go 테스트로 재현할 수 없고, 이번 재현은 문서·도구 두 층의 기계적 사실만 쟀다.
- 코퍼스 규모: 축약 표기가 있는 acceptance.md는 724개 중 28개로 셌다. 몇 곳이 실제 누락을 가리는지는 재지 않았다.
- t499 당시 트리의 lint 동작은 재측정하지 않았다(원본 기록 인용).
