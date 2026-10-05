# plan.md — SPEC-INIT-HARNESS-001

## §A Context

- **카드**: t585 (조사 `.moai/reports/init-tui-audit-20260909.md` 최종 카드 분할 C3 — "하네스 3-way 배포")
- **작업 위치**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t585` — 브랜치 `WT-init-harness-q`, plan 착지 시 HEAD `a404132e7`
- **SPEC 경로**: `.moai/specs/SPEC-INIT-HARNESS-001/{spec,plan,acceptance,design,research,progress}.md` (Tier L, 6파일 — progress 포함)
- **선행**: SPEC-INIT-QUIET-WIZARD-001 `status: completed` (depends_on 성립 — AC-IH 이전에 위저드 4문항 프레임이 이미 착지)
- **PRESERVE(기존 인프라)**: `resolveAgentWiringWithWizard` 단일 해석 지점, `codexwiring` 3종 배선과 신뢰 사이드카, `mirrorOneSkill` skip+report, R-011 게시 스킬 보호, slimFS, `moai tool enable codex`, reconfigure 12문항
- **EXTEND(이번에 늘리는 것)**: harnessFS 래퍼, `llm.harness` 키, doctor 하니스 조건, AGENTS.md 결속표 행, 위저드 옵션 서술

## §B Known Issues (이 SPEC과 관련된 것만)

- **B2 계열 — cross-SPEC 계약**: SPEC-CODEX-WIRING-001의 AC-CW-001..005(`internal/cli/init_agent_flag_test.go`)는 `--llm codex`의 현재 효과(배선+MCP 생략)를 핀다. 의미 변경(D2)은 기존 단정을 유지하고 부정 단정을 추가하는 방향으로만 건드린다 — 기존 테스트 **본문 무변경**이 원칙, 변경이 불가피해지는 순간 blocker 보고.
- **B2 계열 — t583 유보**: t583 Exclusions가 "하네스 3-way 배포(t585)"를 명시 유보 — 충돌 아님, 인수 기록은 본 SPEC HISTORY.
- **B4 — frontmatter**: `created:`/`updated:`/`tags:` 캐노닉 명칭(스네이크 케이스 금지) — 이미 준수, run 단계에서 status 전이 시에도 동일.
- **B6 — spec-lint 제목 관례**: Out of Scope는 H3(`### Out of Scope —`) 필수 — spec.md §5 준수.
- **B8/B10 — 작업 나무 위생**: runtime-managed(`.moai/state/`·`.moai/cache/`·`.moai/logs/`)와 타 카드 SPEC 디렉터 불touch. 템플릿 변경은 Template-First(`internal/template/templates/` 우저작) + `make build`.
- **B12 — sync CHANGELOG**: `--llm codex` 의미 변경은 Breaking 라벨로 기재 — manager-docs가 중복 검사 규약 준용.
- **F1/F4는 이 카드 범위 밖**(t753) — plan 착지 후에도 만지지 않는다.

## §C Pre-flight (run 진입 전 실행)

```bash
git branch --show-current        # WT-init-harness-q
git rev-parse --short HEAD       # a404132e7 (변했다면 흡수 후 재측정)
go build ./...                   # green
GOOS=windows GOARCH=amd64 go build ./...   # green
go test ./internal/cli/... ./internal/template/... ./internal/core/project/... -count=1   # baseline green
golangci-lint run --timeout=2m 2>&1 | tail -5
ls internal/template/templates/.claude/skills | wc -l    # 헤더 제외 34 (재매핑 원천 — 수치 변화 시 design.md D4 재판정)
```

baseline의 기존 실패가 있으면 신규 결함과 분리 기록(§E.5 — NEW vs baseline).

## §D Constraints

- **PRESERVE 목록**: §A의 PRESERVE 전부 + `init_agent_flag_test.go` 본문 + 기존 claude/both init 실행 테스트 본문 + `doctor_golden_test.go` 기존 시나리오(신규 시나리오는 추가).
- **금지**: 위저드 문항 수·순서 변경, reconfigure 문항 변경, `mirrorOneSkill` 본문 재작성(옵션으로 끔), deployer walk 본문에 하니스 분기(design.md D3 기각 사유), 실제 HOME에 쓰는 테스트, `--no-verify`, 타 카드 범위(F1·F4·t589·t586) 침범, 기존 사용자 프로젝트 마이그레이션 코드.
- **필수**: 템플릿 변경 후 `make build`, init 실행 테스트의 홈 지문 슬롯 규약(선언→전후 지문→기록 — t583 REQ-IQW-014/016 준용), Conventional Commits + `🗿 MoAI` 트레일러, 이 리포 규율상 카드 커밋에는 카드 id(t585) 포함. 커밋·push는 레인 오케스트레이터가 창을 줄 때 수행(레인 자율 push 금지).

## §E Self-Verification 매핑

- E1 = AC-IH-001..016 PASS/FAIL 매트릭스 (acceptance.md)
- E2 = `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`
- E3 = 영향 패키지 커버리지: `go test -cover ./internal/cli/... ./internal/template/... ./internal/core/project/...` (목표 85%+ — 신규 파일 기준)
- E4 = 서브에이전트 경계 grep: `grep -rn 'AskUserQuestion' internal/cli internal/template | grep -v _test.go | grep -v '// '` → 0
- E5 = lint (baseline과 NEW 분리)
- E6 = 커밋 SHA 목록 + push 상태(오케스트레이터 창 경유)
- E7 = blocker 보고(있으면)
- E8 = RED 관측 출력(각 마일스퀀스 첫 RED — verbatim)

## §F Milestones (결정 가역성 순 — 데이터·인터페이스 먼저, 기계적 마지막)

### M1 (Priority High) — 영속화와 해석 축 [데이터 모델 — 변경 가능성 최상위]

- `llm.harness` 키: 템플릿 `llm.yaml` 기본 키 추가, `defaults.go` 시딩 `"claude"`, init 3경로(대화형/비대화형×3값)에서 해석값 기록.
- `--llm` help 재작성(`init.go:132`), `agentWiring` 주석의 소비자 목록에 배포기·영속화 추가.
- Tests(RED 우선): init 실행 후 `llm.harness` 존재·값 단정(3값), 키 생존은 M4와 결합.
- [RESOLVED 2026-09-14 → §I NC-1] `--llm codex` 값 의미 변경 — 운영자 채택 확정(codex-only deployment). 원 마커 본문: 기존 사용자는 무영향(키 없음→claude 폴백, research.md §4)으로 분석했고 값의 뜻 변경은 운영자 확인 사항이었다. 불수용 대안(신규 값 추가·위저드 4옵션화)은 기각 확정.

### M2 (Priority High) — codex 단독 배포 축 [신규 타입·인터페이스]

- `harnessFS`(제안명, `internal/template/`): 은닉+재매핑 래퍼(design.md D3/D4) — 단위 테스트(은닉 목록, 재매핑 경로, slimFS와 중첩 조합).
- initializer 시접: opts에 하니스 전달, codex일 때 harnessFS + `WithSkillMirror(false)`로 deployer 구성.
- Tests: codex 단독 init 실행 테스트(홈 지문 슬롯) — 부정 단정(AC-IH-002), 긍정 단정(AC-IH-003), 재매핑 무결성(AC-IH-004). RED는 오늘 나무에서 `.claude/`가 심어지는 것부터 관측.
- claude·both 보존 핀: 기존 실행 테스트 무변경 green 확인(AC-IH-005/006).

### M3 (Priority High) — 공개 축 [사용자 대면]

- 위저드 `agent_wiring` Label·Desc·Title 갱신 + 번역 3블록 동기(design.md D1 — 값 동결).
- 템플릿 AGENTS.md 결속표 보강(동결 목록 미커버 5건 — design.md D8) + 바이트 상한 가드 확인.
- README init 절 + docs-site init 가이드 3-way 절(4-locale 동기 — oss-docs 규칙).
- Tests: 문항 3옵션·값 동정 단정(3로케일), AGENTS.md 공개 완전성 grep(AC-IH-007), 문서 존재(AC-IH-008).

### M4 (Priority Medium) — 정합성 축

- update: `llm.harness` 읽기 → codex 단독 재배포(harnessFS 경유), 키 3-way 생존.
- doctor: 하니스 조건 분기 + codex 단독 golden 시나리오 추가(design.md D7 — claude 표면 INFO 강등).
- Tests: update 재배포 부활 방지(AC-IH-009), doctor golden(AC-IH-010), 병합 생존(AC-IH-015).

### M5 (Priority Medium) — 종합 [기계적]

- 비대화형 parity 테스트(AC-IH-011), `tool enable codex` 무변경 핀(AC-IH-014), AC-CW-004 보존 확인(AC-IH-013).
- 영향 패키지 스위트 + vet + lint + windows 크로스빌드(AC-IH-016), §E.2 슬롯 지문 총정리.

## §G Anti-Patterns

- "테스트 몇 개 고쳐서 green" — AC-CW-*·기존 실행 테스트 본문 수정은 보존 증거 소멸. 부정 단정 추가가 정답.
- deployer에 `if harness == codex` 산발 분기 — FS 층 하나로 (D3 기각 기록 준수).
- codex 단독에서 링크만 지우고 원본 `.claude/skills`를 남기는 절반 재매핑 — `.claude/` 0이 아니면 실패(AC-IH-002가 한 줄로 잡음).
- 위저드 값을 바꾼 "더 깔끔한" 개명 — 값 동결(D1).
- 실제 HOME에 init 실행 — 슬롯 규약 위반(t583 사고 재발).

## §H Cross-References

- spec.md §3 REQ-IH-001..014 ↔ acceptance.md AC-IH-001..016 (추적 표는 acceptance.md §D)
- design.md D1..D9 — 각 REQ의 설계 근거
- 선행·형제·후행: spec.md §6
- 조사 원문: `.moai/reports/init-tui-audit-20260909.md` (읽기전용 — primary 체크아웃)

## §I 운영자 결정 기록 (2026-09-14 확정)

두 NC 모두 **운영자 확인으로 해결**됐다(2026-09-14, 리드 경유 전달). 본문 마커는 `[RESOLVED …]`로 닫고 결정 본문은 이 절이 보관한다 — 조용히 삭제하지 않는다(§I 메커니즘: 감사 추적 보존).

### NC-1 — `--llm codex` 값 의미 변경 → 채택 (RESOLVED 2026-09-14)

- **결정**: 값 재정의를 "codex-only deployment"로 **수용**한다. 신규 값 추가 대안은 기각 확정 — 위저드는 3옵션을 유지한다.
- **반영**: design.md D2가 운영자 확정을 얻어 그대로 확정됐다. §F M1의 위저드 4옵션화 폐기 대안도 폐기 확정. REQ·AC 변경 없음.
- **원 마커**: NEEDS CLARIFICATION — "`--llm codex` 값 의미 변경의 breaking 수용" (§F M1에 있었음) → RESOLVED 2026-09-14, 본 절로 이관. 스캔 대상 괄호형은 해제 시 제거하고 서술형으로 보관한다(§I 메커니즘).

### NC-2 — claude 단독 배포 트리밍 → 현행 유지 (RESOLVED 2026-09-14)

- **결정**: (a) "현행 유지"로 **확정**. REQ 추가 없음.
- **반영**: REQ-IH-003(파일집·로직 보존)이 그대로 확정됐다 — codex 표면 제외 REQ는 만들지 않는다. 디스패치 문구("`.claude/` 표면만")와 조사 운영자 결정의 불일치는 조사 결정 쪽으로 종결(research.md §1 기록 유지).
- **원 마커**: NEEDS CLARIFICATION — "claude 단독 배포 트리밍" (이 절에 있었음) → RESOLVED 2026-09-14, 본 절로 이관.

두 마커 모두 plan-auditor clarification gate 대상에서 해제 — run 진입 준비 완료.
