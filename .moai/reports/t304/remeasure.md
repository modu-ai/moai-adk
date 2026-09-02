# t304 재측정 증거 — codemaps 인용 경로 실재성 (lane-7 배차 런, 2026-09-02)

**Claim**: 카드 본문의 known-6 팬텀 전제를 현재 develop 기준으로 재측정했고, 부재 인용은 8건이며 전부 판별됐다. 양성 팬텀(수리 대상)은 `internal/factory` 1건 + `ListActive` 식별자 드리프트뿐이다.

**Baseline-attribution**:
- 측정 트리 A = `.claude/worktrees/t304` @ `65196a5a7` (= origin/develop, 카드 배차 시점 기준선)
- 측정 트리 B = `.claude/worktrees/t432` @ `7e1c4d94f` (WT-codemaps-refresh, 미병합 — 병합 순서 상 this card가 흡수할 상태)
- 모든 명령은 2026-09-02 이번 런에서 실행.

## 측정 1 — 트리 A (develop 65196a5a7)

명령:
```bash
grep -rhoE '\b(internal|pkg|cmd)/[A-Za-z0-9/_.-]+' .moai/project/codemaps/*.md \
  | sed -E 's/[.,;:)\*]+$//' | sort -u > /tmp/t304-cited-paths.txt
# 각 경로: test -e
```

관측 출력(요지):
- `cited: 86`
- `ABSENT: cmd/moai/main` / `internal/bodp` / `internal/design` / `internal/evaluator` / `internal/factory` / `internal/migrate` / `internal/research` / `internal/state`

## 측정 2 — 트리 B (t432 재생성 후)

동일 명령을 t432 codemaps 디렉터리에 적용: `cited: 102`, ABSENT 8건 동일 집합.

## 판별 (전부 이번 런에서 코드로 확인)

| 인용 | 판별 | 근거 |
|---|---|---|
| internal/design | 부정 각주 — 결함 아님 | modules.md:93 "> … 존재하지 않음" (dd817c44c, 2026-08-31 resync) |
| internal/migrate | 부정 각주 — 결함 아님 | modules.md:147, `internal/migration` 실재 |
| internal/state | 부정 각주 — 결함 아님 | modules.md:218, `internal/session` 실재 |
| internal/research | 부정 각주 — 결함 아님 | modules.md:232 |
| internal/evaluator | 부정 각주 — 결함 아님 | modules.md:258, SPEC-CLEANUP-EVALUATOR-001 제거 기록 |
| **internal/factory** | **양성 팬텀 — 수리 대상** | modules.md:158-162 양성 서술. **이름 변경**: `internal/kanban` 실재 — record.go(202 `validateSessionID`)·revision.go(167/174 `SuppressStep0551`)·integration_lock.go 확인. 인용 진입점 internal/cli/factory.go·launcher_blockcap_infinite.go는 트리 실재(t432 §1 표 15/17행과 일치) |
| **internal/bodp** | **양성 팬텀(트리 A) / 부정 각주(트리 B)** | 제거 커밋 `5792fc755` (#1278, `git log --all -- internal/bodp` 실측). 트리 A modules.md:95-97 양성 섹션 잔존. 트리 B는 dependencies.md:185 부정 각주로 이동 — 병합 후 각주 처우가 판단 자료 |
| cmd/moai/main | 측정 regex 아티팩트 — 결함 아님 | overview.md:161 `cmd/moai/main() → cli.Execute()` 함수 호출 표기. 실체 cmd/moai/main.go 실재. t432 §1 정리 규칙(`.go` 복원)이 해소 |

## 식별자 축 (data-flow.md)

- `ListActive` 3곳: data-flow.md:197 (mermaid 노드), :214 (PreToolUse 흐름), :356 (인터페이스 블록 `ListActive(spec string) ([]Session, error)`)
- 실제 API (이번 런 grep 실측): `QueryActiveWork(optSpecID string) ([]Entry, error)` — internal/session/registry.go:261 패키지 함수. Registry 메서드 Register/Heartbeat/Deregister는 registry.go:169/215/241 (receiver 메서드 — t432 §3.1 #19-22 HIT와 일치). 진짜 미적중은 ListActive 1건.

## F1 MINOR (핸드오프 흡수 — 기록만, t432 트리 쓰기 없음)

t432 증거 파일 `codemaps-accuracy-verification.md`: §3.1 표제 "26항목"(194행) vs 표 27행(행 번호 1-27, 251행 Gaps에도 "27항목"). 표제가 오기. 파일은 t432 워크트리 local-only(미추적) + 트리 frozen이라 본 레인이 수정하지 않음 — 본 기록과 리드 보고로 흡수.

## Gaps

- 측정 regex는 백틱/코드펜스/인용을 구분하지 않는다(의도된 전수 포함 — t432 §1 Residual-risk와 동일).
- `Query` 메서드(registry.go:266, t432 보고 인용)는 이번 런에서 미직접 관측 — 내 grep 패턴(`func Query`)이 receiver 메서드를 놓치는 t432 §3.1 #17/#23과 동일한 측정 결함 여지. plan-phase에서 manager-spec이 R3로 직접 확인.

## Residual-risk

- t432 병합이 착지하면 codemaps 본문이 재생성판으로 바뀐다 — 본 재측정의 줄번호는 트리 A 기준이며, run-phase는 origin/develop 흡수 후 재인벤토리로 재확정한다(카드 배차 지시 "t432 병합 착지 뒤 재측정" 준수).
- 부정 각주 판별은 본문 형태(blockquote ">")에 의존한다 — 재발 방지 축(D2)의 기계 판별 규칙은 SPEC이 정한다.
