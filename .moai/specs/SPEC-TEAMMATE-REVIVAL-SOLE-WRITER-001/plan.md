# plan — SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001

## §A Context

### A.1 배경

카드 t269(Tier S, state=picked, 브랜치 `WT-revival-solo-writer`) — t232 sync-audit verdict(2026-08-25, PASS 0.92)의 process Finding F4("소유권 위반·오-revert·감사 중 동시 작성 … Required fix: 별도 후속 카드") 후속. 산출물은 **교리 계층**뿐: (1) 정지 티메이트 SendMessage 부활 방지 교리, (2) 활성 감사 중 워크트리 단독-작성자 규율. 메커니즘 계층은 동반 카드 t267 경계 밖(spec.md §7).

근거 경로(dev 전용, 템플릿 반영 금지):

- `.moai/specs/SPEC-ZONE-REGISTRY-RESYNC-001/progress.md` — §F:30-31(M1 자발 실행 + TaskStop 2회·부활 SendMessage 1건 전망), §E.2:381(M2 저술자 = 부활 티메이트, 리드 재정지·영구 정지)
- `.moai/reports/t232/sync-audit-verdict-2026-08-25.md` — :4(측정 트리 `ef93a9d1e`→`11df9587a`), §프로세스 판정 (c):49, F4:78, 권고 3:98, 잔여위험:92
- auto-memory `feedback_sendmessage_revives_stopped_teammate.md`(MoAI 프로젝트 메모리 루트; 오케스트레이터 존재 확인 2026-08-26)
- `.moai/research/cc-changelog-snapshot-2.1.233.md` :3236-3237(SendMessage auto-resume, 2.1.77)
- plan-phase 종합: `.moai/reports/t269/research.md`(4-렌즈 팬아웃 + 교정 C1-C4)

### A.2 대상 위치 (내용 기반 재좌표, 2026-08-26 t269 트리 실측)

| 위치 | 현재 상태 |
|------|-----------|
| `cross-session-messaging.md` Rules 절 | 정지·부활 언어 0 (`stopped teammate` 0, `TaskStop` 0) — greenfield |
| `cross-session-messaging.md` Anti-patterns | `Silent write race`가 최근접 기존 절 — 신규 엔트리는 별도 추가 |
| `agent-common-protocol.md` §Background Agent Execution | 동시성 스코프 단독-작성자 불변식 존재(감사창 스코프 없음, 세션 벡터 없음) — 문단 뒤 확장 |
| 트윈 패리티 | ACP↔ACPT `cmp` 무출력(동일); CSM↔CSMT 단일 hunk `113,114d112`(Origin 라인+공백, 의도적) |

### A.3 접근법

1. 두 절 블록을 **양 사본 동일·중립 영어 텍스트**로 저작(사고 근거 인용은 SPEC 산출물에만 — REQ-TRSW-005).
2. CSM: Rules 절에 `[ZONE:Evolvable] [HARD]` 불릿 1개 + Anti-patterns에 `Reviving a stopped teammate` 엔트리 1개. ACP: §Background Agent Execution 기존 문단 뒤 `[ZONE:Evolvable] [HARD]` 문단 1개(heading 무변경 — anchor 안전).
3. 토큰 앵커: CSM `stopped teammate`(≥2)/`owning orchestrator`(≥1), ACP `actively audited`(≥2)/`foreign commit`(≥1) — 전부 오늘 0 실측(§C).
4. 마일스톤 2개(M1 저작·패리티, M2 검증 배치·증거) — 의사결정 가역성 내림차순.

### A.4 PRESERVE 목록 (무편집)

- `cross-session-messaging.md` 로컬↔트윈 의도적 diff — Origin 라인 `> Origin: SPEC-CODEX-SESSION-MSG-001 (design.md §8 mapping).` + 공백(hunk `113,114d112`). 바이트 보존, "수리" 금지.
- `orchestration-mode-selection.md` 전체(자체 의도적 중립화 재작성 :142/144 포함) — 무편집.
- `zone-registry.md` 양 미러 + `internal/constitution/registry_sync_test.go` 핀(`wantRegistryEntries=101`, `wantTupleDigest=2edb5384…`) — 무편집. 신규 레지스트리 엔트리 없음(CSM 엔트리 0개; ACP 기존 13개 clause verbatim 무변경, 신규 텍스트의 등록 리터럴 재인용 금지).
- 에이전트 정의 전부(`.claude/agents/moai/*.md`) 및 `.codex` `.toml` 미러 — 무편집.
- `kanban-dispatch.md`(+`-detail.md`), `cross-session-messaging-detail.md`, `sync.md`, `worktree-integration.md`, `.claude/rules/local/*` — 무편집.
- Go 소스 0행. `lint.skip` 요청 없음.

## §B Known Issues (관련 항목만)

- **워크트리 가드가 복합 shell 구조를 거부** — 변수 할당·루프·프로세스 치환·다중 `[[ ]]` 체인이 "too complex"로 거부됨(계획 단계 실측 3회). run-phase 검증 명령은 단일 단순 명령으로 발행한다.
- **`moai update`가 관리 뿌리를 통째 삭제**(CLAUDE.local §2.3) — 편집 대상 2경로 전부 template-managed다. 트윈 패리티(AC-003)가 로컬 사본의 재배포 복원 가능성을 보장한다. 기존 dev-only Origin 라인의 update 노출은 선존 상태이며 본 SPEC은 신규 분기를 만들지 않는다.
- **ACP는 등록된 파일**(레지스트리 엔트리 13개) — once 의미론: 신규 텍스트가 기존 등록 clause 리터럴과 우연히 일치하면 clause가 multi로 잡힌다. 저작 시 해당 리터럴 회피, M2에서 AC-005로 기계 확인.

## §C Pre-flight (착수 전 측정 — 2026-08-26 t269 트리, 전부 1인칭 실측)

```
$ grep -c 'stopped teammate' <CSM> <CSMT> <ACP> <ACPT>        → 0 / 0 / 0 / 0
$ grep -c 'owning orchestrator' <CSM> <CSMT>                  → 0 / 0
$ grep -c 'TaskStop' <CSM> <ACP>                              → 0 / 0
$ grep -c 'actively audited' <ACP> <ACPT>                     → 0 / 0
$ grep -c 'foreign commit' <ACP> <ACPT>                       → 0 / 0
$ cmp <ACP> <ACPT>                                            → (무출력 — 바이트 동일)
$ diff <CSM> <CSMT>                                           → 113,114d112 / < > Origin: … / <
$ grep -cE 'SPEC-[A-Z]|…' <CSMT> <ACPT>                       → 0 / 1 (ACPT 292행 --filter-spec=<SPEC-ID> 플레이스홀더 — 정밀 패턴 미부합)
$ wc -l <CSM> <ACP>                                           → 126 / 362
$ go test -count=1 ./internal/constitution/...                → ok  0.708s
$ echo 'SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001' | grep -cE '^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$' → 1
$ ls .moai/specs/ | grep -E 'TEAMMATE|REVIVAL|SOLE|ROGUE|STOP-RE' → 전-ID 충돌 0 (E2E-REVIVAL·STOP-EVIDENCE-WRITER 등과 전체 ID 상이)
```

RED 셀 근거(verification-completeness §2): AC-001/002는 "토큰 부재"로 RED — 이유: 해당 교리가 아직 어디에도 없음(research §4 grep 0-hit와 동일 결론). AC-003/005는 오늘 초록인 PRESERVE형 — 실패 모드는 회귀이며, 기준값을 위 SHA·행수로 고정한다. AC-004는 기준선 자체가 0/0(조악 패턴 유일 적중은 플레이스홀더).

## §D Constraints

### D.1 규율

- Template-First: 트윈 먼저 → `make build` → 로컬. spec.md §4 전항 준수.
- 커밋: 오케스트레이터 소관(이 문서의 독자는 run-phase 위임자). 커밋 메시지에 카드 id t269 명시. 이 저장소는 전 Tier Route B(PR) — `repo-local-pr-policy.md`.
- 시간 추정 금지 — Priority 라벨·순서로만.

### D.2 리스크 분석

| # | 리스크 | 완화 |
|---|--------|------|
| 1 | always-loaded 문맥 비용 증가(두 파일 전부 상시 로드) | AC-006 상한: CSM ≤16행·ACP ≤10행 순증(~26행 ≈ ~230 토큰 최악). 절 텍스트를 압축해 저작. |
| 2 | ACP 레지스트리 결합(13개 pinned clause + anchor 검사) | append-only 배치(신규 문단, heading 무변경), 등록 리터럴 비재인용, AC-005 기계 확인 |
| 3 | 템플릿 중립성 CI 가드 | 양 사본 동일·중립 저작(REQ-005); 인용은 SPEC 산출물 한정; AC-004 정밀 패턴 |
| 4 | 교리≠강제 — 규칙 위반에 기계적 게이트 없음 | 명시된 한계(spec.md §4·§7). t267이 메커니즘을 소관; 그 전까지 REQ-004 표면화 의무가 탐지층 |
| 5 | `moai update` wipe 노출 | 트윈 패리티(AC-003)가 재배포 복원 보장; 신규 분기 0 |
| 6 | 워크트리 가드의 복합 명령 거부 | 검증은 단일 단순 명령으로 발행(§B) |

### D.3 mx_plan 노트

본 SPEC은 docs-only(변경 파일 전부 markdown 규칙 문서, Go/코드 파일 0) — `@MX:` 태그 부착 대상이 없다. sync의 MX 스캔은 경량 라티오널로 처리한다: "변경 파일 전부 markdown, @MX 부착 대상 0"을 스캔 기록에 한 줄로 남긴다. 코드 어노테이션 계약(`mx-tag-protocol.md`)은 이 SPEC에서 발동하지 않는다.

## §E Self-Verification (run-phase 보고 형식)

위임받은 manager-develop는 완료 보고에 VCI 5-섹션(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)으로 다음을 남긴다 — 각 항목 (a) 명령 (b) 관측 출력 verbatim (c) 측정 트리 SHA:

- E1. AC 6종 바이너리 매트릭스(AC-TRSW-001..006, PASS/FAIL)
- E2. `make build` rc + `git status --porcelain` 무출력(임베드 no-op)
- E3. AC-005 `go test -count=1 ./internal/constitution/...` 재실행 출력
- E4. AC-003 패리티: `cmp` 무출력 + `diff` 필터 잔여 0
- E5. AC-004 중립성 3-grep `0/0`
- E6. 신규 커밋 SHA 목록 + push 결과(오케스트레이터 조정 하)
- E7. 블로커 보고(있으면 — AskUserQuestion 금지, 구조화 보고)

커버리지(E3 축)는 해당 없음 — Go 코드 0행. 문서 전용 SPEC의 검증 축은 위 E1-E6이다.

## §F Milestones (의사결정 가역성 내림차순 — 검토 민감도 높은 것부터)

### M1 — 교리 절 저작, template-first (Priority High)

1. 두 절 블록 최종 저작(영어·중립; CSM 불릿+안티패턴, ACP 문단) — **가장 가역성 높은 결정: 문면·배치. 검토 집중 여기.**
2. 템플릿 트윈 먼저 편집 → 로컬 미러 동일 텍스트 편집.
3. AC-001..004 토큰 전환·패리티 자기검증 + AC-006 예산 측정.

### M2 — 검증 배치 + 증거 (Priority Medium)

1. `make build` → porcelain clean 확인(임베드 재생성 결정성).
2. AC-005 registry 가드 재실행.
3. AC 6종 전 행 명령+출력 verbatim을 progress.md §E.2에 기록(§E 보고 형식).

의존성: M2는 M1에 의존(직렬). 병렬화 이득 없음 — coding-heavy 순차 작업.

## §G Anti-Patterns

- 로컬 사본 먼저 편집 후 트윈 "나중에" — Template-First 위반, 패리티·중립성 가드 양쪽에 걸림.
- Origin 라인 의도적 diff를 "정리" — 중립성 규율의 산물 제거(연구 교정 C2가 쌍둥이 저작 선례로 확정).
- 사고 SHA·t232 토큰을 절 텍스트에 삽입 — 트윈 중립성 위반(REQ-005). 인용은 SPEC 산출물에만.
- ACP에서 기존 등록 clause 문장을 인용·변형해 재사용 — once 의미론 파손(§B 3항).
- 대량 `git add` — 명시 pathspec만(공유 체크아웃 규율).
- 검증 명령을 `&&`·루프로 묶기 — 워크트리 가드 거부(§B 1항). 단일 단순 명령씩 발행한다.

## §H Cross-References

- spec.md 본 SPEC — REQ/AC 계약의 단일 원천
- `.moai/reports/t269/research.md` — 표면 지도·교정 C1-C4
- `.moai/reports/t232/sync-audit-verdict-2026-08-25.md` — F4 위임 원문
- 카드 t267(backlog) — 메커니즘 계층 경계
- `internal/template/internal_content_leak_test.go`, `internal/template/template_neutrality_audit_test.go` — 중립성 가드
- `internal/constitution/registry_sync_test.go` — 레지스트리 핀(무손상 대상)
- `CLAUDE.local.md` §2(Template-First·update 위험), `repo-local-pr-policy.md`(전 Tier PR)
