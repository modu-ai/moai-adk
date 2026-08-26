# plan — SPEC-ZONE-REGISTRY-HARDEN-001

> Tier M · Route B(PR — repo-local `repo-local-pr-policy.md`: 본 저장소는 전 Tier PR 강제, Route A 비활성) · 개발 방식 `tdd`
> 모든 행 좌표는 워크트리 t268 트리 `db1362739`(= origin/main) 기준 **내용 재좌표** 값. 감사 보고서의 좌표(t232 트리 `11df9587a`)와 다르면 본 문서 값을 따른다.

## §A Context

### A.1 배경

- 선행 SPEC: SPEC-ZONE-REGISTRY-RESYNC-001(completed, PR #1646 머지 `39c677f47`). 종결 sync-audit `.moai/reports/t232/sync-audit-verdict-2026-08-25.md` PASS 0.92, Findings F1/F2/F3(optional)의 기술적 후속.
- 가드 구조(`internal/constitution/registry_sync_test.go`): mirror 2개(local/template) 각각 ① production validator ② LoadRegistry 데이터 검사(anchor 해석 + literal clause + 자기참조 금지) ③ 평가 수 단정 ④ bucket 단정 ⑤ SKIP-env 양방향 서브테스트. CI 차단 경로는 `go test ./...` Test 잡(`constitution-check` 잡은 continue-on-error 2차 신호).
- 파서 필드명(`internal/constitution/loader.go:62-66`): `zone_class` → `ZoneClass string`, `canary_gate` → `CanaryGate bool` (Entry 변환시 동일명 유지, `loader.go:122-126`).

### A.2 대상 위치 (재좌표 완료)

| Finding | 파일:행 | 내용 앵커 |
|---------|---------|-----------|
| F1-레지스트리 | `.claude/rules/moai/core/zone-registry.md:763-768` | `- id: CONST-V3R5-010` / `clause: "Semantic failures (data race, deadlock, panic, test assertion failure) MUST"` |
| F1-레지스트리(미러) | `internal/template/templates/.claude/rules/moai/core/zone-registry.md:763-768` | 좌측과 바이트 동일(`cmp` 실측 확인) |
| F1-원본(트윈) | `.claude/rules/moai/workflow/ci-autofix-protocol.md:106-108` + 템플릿 원본 `internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md`(바이트 동일) | `[ZONE:Frozen] [HARD] Semantic failures … MUST` / `NOT be automatically patched. The orchestrator MUST immediately escalate via` / `AskUserQuestion with the diagnosis report.` |
| F2 | `internal/constitution/registry_sync_test.go:45`(`wantRegistryEntries = 101`), `:135-137`(개수 단정) | 개수 단정 뒤에 digest 단정을 추가 |
| F3 | `.moai/specs/SPEC-ZONE-REGISTRY-RESYNC-001/plan.md:108` | `strings.Count(rawFileContent, clause)` 문단 (같은 문서의 `:145` `grep -F -c` 언급은 공격 시나리오 표 안 — 이미 정합하므로 무편집) |

### A.3 접근법 (finding별)

**F1 — 서식 전용 rewrap 후 clause 재선택** (감사 예시 문장 채택)

1. 템플릿 원본 ci-autofix-protocol.md의 106-108행 단락을 다음과 같이 rewrap(단어 변화 0, 개행 위치만 이동):

   ```
   [ZONE:Frozen] [HARD] Semantic failures (data race, deadlock, panic, test assertion failure) MUST
   NOT be automatically patched.
   The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report.
   ```

   → 감사 예시 문장 `The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report.`이 **단일 완결 행**이 된다. 배포판 트윈에 동일 편집. 이 문장 선택의 근거: 같은 [ZONE:Frozen] [HARD] 단락의 두 번째 의무문으로 MUST-NOT 금지와 함께 escaltation 계약을 온전히 전달하며, 감사관이 F1 Required fix에 예시로 명시한 문장이다.
2. zone-registry.md 양쪽 미러 `:768` clause 값을 `"The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report."` 로 교체(나머지 필드 — `file:`/`anchor: #semantic-failure-no-auto-patch`(레지스트리 :767 실측값 — 원본 파일의 `<!-- anchor: #semantic-failure-handling -->` 주석이 아니라)`/`zone`/`zone_class`/`canary_gate: true` — 무변경).
3. 검증: 핀된 파일에서 신규 문장 `grep -c -F` → 정확히 1. validator는 공백 정규화 containment를 쓰므로 rewrap 후에도 통과(정규화 텍스트 불변).

**F2 — 정렬 튜플 digest pinning**

1. 직렬화: 엔트리마다 `fmt.Sprintf("%s|%s|%s|%t", e.ID, e.Zone, e.ZoneClass, e.CanaryGate)` 행 생성 → `sort.Strings`(id 유일하므로 사실상 id 정렬) → `strings.Join(lines, "\n")` → `sha256.Sum256` → `hex.EncodeToString`.
2. 상수 블록(`:42-53`)에 `wantTupleDigest` 추가. 각 mirror 서브테스트의 개수 단정(`:135-137`) 직후에 digest 단정 추가 — 실패 메시지는 **계산된 digest를 출력**하고 "의도한 레지스트리 변경이면 `wantRegistryEntries`와 `wantTupleDigest`를 같은 커밋에서 함께 갱신" 절차를 안내(REQ-ZRH-006).
3. helper는 `func registryTupleDigest(entries []constitution.Entry) string`로 분리 — 변이 테스트가 재사용한다.
4. 변이 서브테스트 `TestRegistryTupleDigestRejectsSubstitution`(신규 독립 테스트 함수): `writeMutatedRegistryCopy` 패턴 계승 — 실제 레지스트리를 TempDir에 복사하되 `- id: CONST-V3R2-004`를 런타임에 읽어 개수 보존 ID 치환(예: 끝에 `X` 삽입) 적용 → `LoadRegistry` → digest 산출 → `wantTupleDigest`와 불일치 단정. **RED 먼저**: digest 단정 구현 전 이 변이가 현재 가드를 통과해버림을 관측(REQ-ZRH-005의 RED 근거).
5. zone/zone_class/canary_gate 치환 변이도 동일 helper로 부가 서브케이스(테이블 드리븐 — 3종).

**F3 — 폐쇄 SPEC plan.md 의미론 정정 (erratum 방식)**

1. `RESYNC-001/plan.md:108` 문단의 `strings.Count(rawFileContent, clause)` 서술을 구현이 재는 의미론으로 교체: "정규화 없이 엔트리별 **라인 수 의미론**(`grep -F -c` 등가 — clause를 리터럴 부분문자열로 포함하는 **행**의 개수. 같은 행 2회 적중도 1)으로 세어 …".
2. 정정 출처 주석 부기: `[정정 — SPEC-ZONE-REGISTRY-HARDEN-001 (2026-08-25): 원문은 발생 의미론(strings.Count)으로 기술했으나 구현·acceptance는 라인 수 의미론. 현 데이터는 양쪽 모두 once=97(t232 감사 독립 측정)이라 판정 불변.]` — 역사 무단 재작성이 아니라 erratum 각주로 남긴다.
3. 그 외 plan.md 내용(공격 시나리오 표의 `:145` 포함) 무편집.

### A.4 PRESERVE 목록 (무편집)

- `internal/constitution/validator.go` — 매처 불변(RESYNC D1 계약; `git diff 1ae6e5c36..HEAD -- internal/constitution/validator.go` = 0 유지)
- `internal/constitution/loader.go`, `rule.go`, 그 외 `internal/constitution/**` 비테스트 파일
- `.claude/rules/local/ci-autofix-protocol.md` (dev 원본 — 레지스트리가 pin하지 않음)
- CONST-V3R5-010 외 전 레지스트리 엔트리(은퇴 4건 포함)
- `.github/workflows/ci.yml` 및 `.github/workflows/**`
- `.moai/reports/t232/**` (감사 기록 불변)
- CHANGELOG.md — sync-phase(manager-docs) 소관

## §B Known Issues (관련 항목만)

- **B5 CI 3-tier**: spec-lint / golangci-lint / Test(OS별)는 각자 실패한다 — 신규 digest 코드는 `GOOS=windows` 컴파일 포함(테스트 파일이므로 `GOOS=windows go vet ./internal/constitution/`).
- **B8 워킹트리 위생**: `git add` 명시 pathspec만. 병렬 세션 파일 혼입 금지(카드 worktree에서 작업).
- **B10 스코프**: §A.4 PRESERVE 외 무편집 — 특히 "보다시" 인접 clause 개선 유혹 금지(§G AP-2).
- **Template-First**(CLAUDE.local §2): 템플릿 원본 선행 편집 → `make build` → 배포판 동기 → 트윈 `cmp`. 템플릿 neutrality CI(`template-neutrality-check.yaml`)가 `internal/template/templates/**` 경로 변경에 트리거되나 서식 전용 변경이라 금지 클래스(C3 날짜·C7 SHA·SPEC-ID)를 새로 삽입하지 않는 한 통과.
- **정렬 결정성**: digest는 행 정렬 후 산출 — YAML 원문 순서와 무관하게 안정. `CanaryGate`는 `%t`(true/false)로 직렬화.

## §C Pre-flight (착수 전 측정)

```bash
git rev-parse --short HEAD && git branch --show-current        # db1362739 / WT-zone-registry-f13 확인
go test ./internal/constitution/                               # 변경 전 baseline 초록 관측
go test -cover ./internal/constitution/                        # coverage baseline (직전 측정 85.8%)
cmp .claude/rules/moai/workflow/ci-autofix-protocol.md internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md   # 트윈 동일 baseline
cmp .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md                        # 미러 동일 baseline
golangci-lint run --timeout=2m ./internal/constitution/... 2>&1 | tail -3
grep -c -F 'The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report.' .claude/rules/moai/workflow/ci-autofix-protocol.md   # → 0 (RED 근거)
```

## §D Constraints

- 금지: `--no-verify` / force-push / validator.go 편집 / 은퇴 엔트리 clause 편집 / 미러 한쪽만 편집 후 커밋 / `moai update` 실행(§2.3 로컬 전용 파일 삭제 위험)
- 커밋: Conventional Commits — `fix(SPEC-ZONE-REGISTRY-HARDEN-001): M1 …` 형식, 카드 id t268 병기, `🗿 MoAI` 트레일러
- F1의 ci-autofix-protocol.md diff는 **공백 정규화 후 base 대비 바이트 동일**이어야 함(AC-ZRH-002가 `git show db1362739:<path>` 대비 기계 검증)
- F2의 직렬화 포맷(`id|zone|zone_class|%t` + sort + "\n" join + sha256 hex)은 AC가 digest 값이 아닌 **동작**(치환 탈출 차단·정상 통과)을 검증하는 기준이므로, 포맷 자체는 구현 재량이나 문서화 없이 바꾸지 않는다

## §E Self-Verification (run-phase 보고 형식)

각 항목 Claim / Evidence(명령+출력 원문) / Baseline-attribution(트리 SHA) / Gaps / Residual-risk 5분할 보고. 최소 축:
1. AC-ZRH-001..009 이진 PASS/FAIL 매트릭스(acceptance.md §D)
2. `go test ./internal/constitution/` + `-run TestRegistrySyncGuard -v` bucket 라인 원문 인용
3. `go test -cover ./internal/constitution/` (≥85%)
4. `golangci-lint run ./internal/constitution/...` 0 issues + `gofmt -l internal/constitution` 빈 출력
5. `GOOS=windows GOARCH=amd64 go vet ./internal/constitution/` rc=0
6. `make build` rc=0 후 `git status --porcelain` 트래킹파일 무변경(임베드 no-op)
7. 신규 커밋 SHA 목록 + push 결과

## §F Milestones (의사결정 가역성 내림차순 — 검토 민감도 높은 것부터)

### M1 — F1: rewrap + clause 재선택 (Priority High)

1. 템플릿 원본 ci-autofix-protocol.md rewrap → 배포판 트윈 동기 → `cmp` 동일 확인
2. zone-registry.md 양쪽 미러 clause 교체 → `cmp` 동일 확인
3. `make build` → `go test ./internal/constitution/` 초록(validator+guard 동시 통과 확인 — 공백 정규화 containment와 단일 행 유일 적중이 모두 성립)
4. AC-ZRH-001/002/003 측정
5. 커밋 `fix(SPEC-ZONE-REGISTRY-HARDEN-001): M1 re-select CONST-V3R5-010 clause to a complete single-line sentence (t268)`

### M2 — F2: 튜플 digest pinning (Priority High)

1. **RED**: `TestRegistryTupleDigestRejectsSubstitution` 먼저 작성 — digest 단정 부재 상태에서 개수 보존 ID 치환이 가드를 통과함을 관측(실패 출력 원문 보존)
2. `registryTupleDigest` helper + `wantTupleDigest` 상수 + mirror별 단정 구현(실패 메시지에 계산 digest + 갱신 절차)
3. 변이 테이블 확장: ID 치환 / zone 치환 / zone_class 치환 / canary_gate 반전 4종
4. `go test ./internal/constitution/` 초록 + lint/vet/coverage
5. AC-ZRH-004/005/008(가드 축) 측정
6. 커밋 `feat(SPEC-ZONE-REGISTRY-HARDEN-001): M2 pin sorted (id,zone,zone_class,canary_gate) tuple digest in registry sync guard (t268)`

### M3 — F3: plan.md 의미론 정정 (Priority Medium)

1. `RESYNC-001/plan.md:108` 문단 교체 + erratum 출처 주석(§A.3 F3 안)
2. AC-ZRH-007 측정(grep 3종)
3. 커밋 `docs(SPEC-ZONE-REGISTRY-HARDEN-001): M3 align RESYNC-001 plan.md literal-check semantics with implementation (t268)`

> **소유권(D3)**: 1·3단계(RESYNC-001 plan.md 본문 편집과 그 커밋)는 manager-develop 이 아니라 **오케스트레이터 재위임을 받은 manager-spec** 이 실행한다 — 완료된 타 SPEC 의 plan.md 본문은 spec-frontmatter-schema.md § Forbidden ownership crossings 로 보호된다. manager-develop 은 2단계(AC-ZRH-007 측정)만 담당한다.

완료 후: push → CI 판독 → `/moai sync`(manager-docs) — 본 저장소는 Route B(PR)로 마감.

## §G Anti-Patterns

- **AP-1**: 가드를 발생 의미론으로 전환하며 "F3도 같이 해결" — 범위 밖(문서 정렬안 채택). 구현·acceptance는 이미 정합하다.
- **AP-2**: F1을 하다가 다른 엔트리의 clause도 "더 좋은 문장"으로 교체 — 감사는 V3R5-010 1건만 지목. 리뷰 노이즈+신규 위험.
- **AP-3**: dev 원본(`.claude/rules/local/ci-autofix-protocol.md`)까지 rewrap — 계약상 불필요(레지스트리 `file:` 전수 확인). 인접 파일 무편집.
- **AP-4**: rewrap에 단어 교정("escalate"를 다른 동사로) 섞기 — AC-ZRH-002의 공백 정규화 동일 검사가 붉게 잡는다. 서식만.
- **AP-5**: digest 상수만 갱신하고 개수 상수를 안 옮기는(혹은 그 반대) 레지스트리 증감 — 본 SPEC엔 증감이 없으므로 둘 다 불변. 향후 증감 시 같은 커밋 갱신이 REQ-ZRH-006 계약.
- **AP-6**: 폐쇄된 RESYNC-001 plan.md를 조용히 재작성 — erratum 각주 방식 유지(감사 기록과의 정합성).
- **AP-7**: 미러 한쪽만 고치고 "배포는 나중에" — parity 테스트가 붉게 잡지만 커밋 단위에서도 한쪽만 커밋하지 않는다.

## §H Cross-References

- 요구 원천: `.moai/reports/t232/sync-audit-verdict-2026-08-25.md` §Findings F1-F3, §Recommendations 2
- 선행 SPEC: `.moai/specs/SPEC-ZONE-REGISTRY-RESYNC-001/{spec,plan,acceptance,progress}.md` (가드 계약·옵션C·bucket 규격의 정본)
- 가드 구현: `internal/constitution/registry_sync_test.go` / 파서: `internal/constitution/loader.go` / 매처(불변): `internal/constitution/validator.go`
- frontmatter 스키마: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- 템플릿 규율: CLAUDE.local.md §2(Template-First)·§2.3(update 삭제 위험)·`.moai/docs/template-internal-isolation-doctrine.md` §25.1(neutrality 클래스)
- 저장소 PR 정책: `.claude/rules/local/repo-local-pr-policy.md` (전 Tier Route B)
