# SPEC-EVIDENCE-CITATION-CANON-001 — 진행 기록

카드: t375 · 워크트리 `.claude/worktrees/t375` · 브랜치 `WT-state-evidence-canon` · 기준 HEAD `b64043481`

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (조건부 — plan.md §B. M2가 §C.3을 결정한 직후 재확인 의무 있음)
- 산출물: spec.md · plan.md · acceptance.md · progress.md
- REQ 13건 (상한 16) / AC 15건 (상한 16)
- SPEC ID 정규식 검사: Bash 실행, 출력 `PASS`
- ID 중복 검사: `find .moai/specs -maxdepth 1 -type d -name 'SPEC-EVIDENCE*'` → `SPEC-EVIDENCE-CLAIM-INVARIANT-001` 1건만, 충돌 없음

### iter1 감사 수리 (2026-08-31)

`.moai/reports/t375/plan-audit-iter1.md` — FAIL 0.70, blocking 11건. 전부 수리했다.

| 결함 | 처리 |
|---|---|
| D1 하한 7이 모집단보다 두 자릿수 작음 | 트리별 하한 300으로 교체, 도출 명령을 AC 본문에 기재 (acceptance AC-ECC-010, plan §D.2) |
| D2 REQ-ECC-008 판정 AC 부재 | **AC-ECC-015 신설** — 방문 트리 목록 + 트리별 하한 + 미러 정합 + 방문 뮤테이션 |
| D3 131 미귀속 | 스크립트 도출값 124로 교체, 경위를 spec §1.1.1에 기록 |
| D4 §1.1 명령 2행이 자기 값을 못 만듦 | 출처를 커밋된 스크립트 2개로 교체, "532 정정 노트" 삭제, 추적 집합 사용 이유 명시 |
| D5 carve-out 누락 | `internal/web/events.go:29` 추가, §1.4 "하나"→"둘", AC-ECC-005·006 확장 |
| D6 AC-ECC-002가 오늘 통과 | 고정 문구 3종 grep으로 교체 (오늘 전부 0회 확인) |
| D7 허용목록 단위 미지정 | 파일 + 정확한 리터럴로 규정, "파일 전체 면제 금지" 단언 추가 (REQ-ECC-009, AC-ECC-013) |
| D8 `.gitkeep` 귀결 미명시 | **판정 변경** — `.gitkeep`·`!` 예외 줄 모두 두지 않음 (spec §4.2) |
| D9 판별식 미적용 | **적용했고 감사 제안과 다른 결론** — navigator 아래 전부 무시하지 않음 (spec §4.3) |
| D10 AC 3건 요구 층 공백 | REQ-ECC-002 주어를 doctrine 표면 문서로 확대 + REQ-ECC-013 신설 |
| D11 인용 넓이 상한 부재 | REQ-ECC-004에 "파일 하나를 이름 붙인다" 상한 통합 |
| D12 Tier 최대 케이스 미계산 | 11–18 조건부 + M2 재확인 의무 (plan §B) |
| D13 세션 디렉터리 124 | "최상위 124 = 세션 123 + snapshots 1"로 정정 |
| D15 부채 수치 오염 | 124 / 231 / 189로 갱신, `mcp_glm.go:110` 후속 후보 추가 |

**감사 소견 중 따르지 않은 것 1건 — D9.** 감사는 `fix-drafts/`의 잔여가 "요청됐으나 완료되지 않은 위임"을 뜻하므로 t373의 `chain/`과 같은 모양이라고 제안했다. §4.1 판별식의 숨은 전제(성공 경로에서 처분하는 코드가 있어야 잔존이 실패를 뜻한다)를 검사한 결과 그 전제가 성립하지 않는다:

```
grep -rn 'RemoveAll' internal/navigator/fix/ internal/cli/navigator_fix.go   (테스트 제외) → 0행
```

처분 코드가 없으므로 잔존은 성공한 실행에서도 남고 아무 신호가 아니다. 다만 결론은 감사가 원한 방향과 같다 — 무시하지 않는다 — 근거가 다르다(내용 기반 신호 `applied.json` + 존재-부재 논거 대칭). 상세는 spec.md §4.3.

### iter2 감사 종결 (2026-08-31) — PASS-WITH-DEBT 0.85

`.moai/reports/t375/plan-audit-iter2.md` — Tier M 임계 0.80 상회, iter1 대비 +0.15 단조. 재감사 상한 도달로 종결 판정. iter1 13건 중 11 종결 / 2 부분 종결 / 미변경 0. **감사가 iter1 D9를 철회했다** — §4.3의 반박이 옳다고 판정.

부채 4건을 이 판에서 닫았다.

| 항목 | 처리 |
|---|---|
| N1 (blocking) | 리드가 스크립트에 `SPEC_OWN_DIR` 배제를 추가·검증. SPEC 쪽 몫은 §1.1.2 신설 — 배제 이유(추적 집합 논거의 유효 기간)와 불변성 실측표(188 in → 184/515/346/124/231 무변). "커밋된"을 "이 카드가 plan-close로 함께 추적한다"로 정정 |
| D1 잔여 | §D.2.1 + AC-ECC-015 #2 — 하위트리 **집합 상등** 단언. 하한 300은 agents(21)·output-styles(3) 소실을 통과시키고, 그 둘이 `manager-lead.md`와 배너 3지점을 담는다. 독립성 시연(하한 통과 + 상등 실패) 포함 |
| D7 잔여 | AC-ECC-013 #1에 판정 명령 부여 — grep이 아니라 **빈 리터럴 항목 거부 뮤테이션 서브테스트**. grep은 오늘 목록에 대해서만 참이라 내일 추가될 항목을 막지 못한다 |
| N2 | §1.1에 반영 — 두 스크립트가 판별식을 각자 리터럴로 정의하므로 "갈라질 수 없다"가 아니라 "함께 고치고 둘 다 돌린다" |
| N3 | AC-ECC-010·plan §D.2의 미러 하한 도출을 4개 하위트리 열거로 교정(340 → **338**). 범위 밖 2개 명시 |
| N4 | §1.1.1에 단위 다리 신설 — 532는 맨 문자열, 515는 뒤에 `/`가 오는 형태, 차이 17 |
| N5 | REQ 재번호 — 013(경계 사례 기록)을 011로, 011·012(ignore 2건)를 012·013으로. 정의 순서가 001…013 단조가 되고 절 구성(§2.2 가드 / §2.3 ignore)도 유지된다. acceptance 매트릭스·plan 참조 동반 갱신 |

시연 방향이 4 → **6**으로 늘었다(하위트리 방문 · 허용목록 단위 추가).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
