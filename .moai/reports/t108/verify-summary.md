# t108 — 오케스트레이션 모드 재명명 6→4 — 검증 요약

작성: 2026-08-17 · 워크트리 `.claude/worktrees/t108` [WT-t108] (base = release/v3.1.1 d169c4aec 병합)

## 새 체계 (단일 축 = 동시 스폰 수, 오름차순)

| 모드 | 동시 스폰 | 구(舊) 토큰 |
|---|---|---|
| `direct` | 0 | Mode 1 `trivial` |
| `serial` (기본 폴백) | 1 (마일스톤당 순차) | Mode 5 `sub-agent` |
| `fanout` | N (3-5 자문 밴드, 경계=런타임 캡 20) | Mode 4 `parallel` |
| `sweep` | 수십~수백 (16-동시/1000-총) | Mode 6 `workflow` |

- `background` → 모드→**실행 옵션** 강등 (CC v2.1.198+ 서브에이전트 기본 백그라운드)
- `agent-team` → §A 각주 + §C.1 (t92 프레임: 실험적 재허용, tombstone 아님)
- 핸드오프 열거형: `serial | fanout | agent-team | sweep` (omitted default = serial)
- 구토큰 수용: 읽기 측 **무기한** 매핑 수용 (solo-sequential→serial, parallel-subagents→fanout, dynamic-workflow→sweep) — 쓰기는 신규 토큰만. Go 파서 부재 실측(handoff CLI·주입기 본문 verbatim, `mode:` 줄은 모델 해석 프로토콜 텍스트) → 런타임 마이그레이션 불필요. 근거: 과거 텍스트 매핑 비용 0 + 구 메모리 항목 무기한 잔존 → sunset 불필요.

## Claim (주장)

1. 살아있는 교리 전 표면(harness `.claude/` + 템플릿 미러 + 루트 CLAUDE.md + docs-site 4로케일)에서
   `grep 'Mode [1-6]|Mode 7'` = **0건** — 카드 [HARD] 달성.
2. 구 열거형 토큰 잔여 = 0 (하위호환 문서의 의도적 언급 제외).
3. 치환 부작용(중복 토큰·잔여 숫자) = 0.
4. 템플릿 미러 byte-parity + 중립성(STRICT) + catalog 해시 전량 통과, make build 성공.
5. always-loaded 예산 통과 (증분 +2 줄 — 1,000B 진술 임계 미만).
6. docs-site hugo 빌드 경고 0.

## Evidence (증거)

- `grep-mode-refs.txt` — `grep -rn "Mode [1-6]\|Mode 7" .claude/ templates/ docs-site/ CLAUDE.md` → **0행**
- `grep-enum-residue.txt` — 구 열거형 토큰 (하위호환 문서 제외) → **0행**
- `grep-damage.txt` — 중복 토큰/잔여 숫자 패턴 → **0행**
- `green-template.txt` — `go test ./internal/template/ -count=1` → `ok 23.355s` (미러·중립성·해시 포함)
- `green-budget.txt` — `go test ./internal/config/ -run TestAlwaysLoadedTokenBudget -count=1` → `ok 0.384s`
- `rename-modes.pl` — 순차 규칙 일괄 치환 스크립트 (방법 증거; 구→신 규칙 순서 명시)
- hugo — `hugo --quiet` RC=0 (docs-site)

## Baseline-attribution (baseline 귀속)

WT-t108 (HEAD = release/v3.1.1 d169c4aec + base 병합 커밋), 위 커맨드는 본 세션 미커밋 작업 트리에서 실행.

## Gaps (미검증 / 명시적 예외)

1. **불변 역사 기록은 의도적으로 미수정**: `.moai/specs/*` (~800행)와 `CHANGELOG.md` (2개 항목)의
   구 토큰은 완결된 시점의 사실 기록 — 재작성 시 역사 왜곡. 살아있는 표면 전부 0 달성과 대비 기재.
2. `.claude/skills/moai-foundation-core/modules/execution-rules.md`의 "Mode 1/2/3"은 **git 전략 모드**
   (Manual/Personal/Team)라 축돌 제거 차원에서 헤더를 "Git strategy: …"로 명확화 (grep=0 달성 수단).
3. 카드가 영향 파일로 명시한 agent-common-protocol.md는 t114 다이어트본에 수치 참조 0건 — 실측
   결과 변환 대상 없었음 (보고용).
4. docs-site 4로케일 패리티: 동일 페이지 4로케일 동시 갱신 (머메디어 노드 라벨 1곳 포함) — 헤딩
   수 parity는 hugo 빌드로 간접 검증 (전수 비교 스크립트 미실행).
5. 실제 `/moai run` 세션에서 progress.md § Mode Selection 로그의 신규 토큰 기록 — 런타임 미실증
   (doctrine-only 변경, Go 코드 0행).

## Residual-risk (잔여 위험)

- 붙여넣기 재개 메시지의 `mode:` 줄은 프로토콜 토큰 — 구 토큰을 쓴 재개본이 과거 memory에서
  주입될 수 있으나 설계상 읽기 측 매핑이 무기한 수용 (Gap 아님, 설계).
- 일괄 치환이 문맥별 미세 어색함을 남길 수 있음 — 3종 damage grep으로 구조 손상은 0 확인했으나
  문체 수준 감수.
- `fanout`/`sweep` 토큰이 기존 다른 의미(예: FO-* 팬아웃 ID)와 시각적 유사 — 카탈로그 표가
  정본 §A에서 단일 원천을 유지하므로 실질 충돌 없음.
