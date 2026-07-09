# SPEC-TEMPLATE-RULES-CLEANUP-001 — CI 가드 설계 (design.md)

가드 4종의 탐지 패턴 / 스코프 / allowlist / 오탐 완화 / 테스트 배치. 모든 패턴은 Go `regexp`(RE2) 호환 — lookahead/lookbehind 사용 불가를 전제로 설계한다.

## §1 공통 설계 원칙

1. **스코프**: 신규/확장 가드는 `internal/template/templates/.claude/rules/**` 한정. skills 트리는 기존 SBN 가드 소유 (스코프 침범 금지 — 기존 `skillBodyScoped`/`skillMoaiScoped` 플래그 선례를 따라 `rulesScoped` 성격의 스코프 지정을 도입).
2. **allowlist 메커니즘**: 기존 `pedagogicalAllowlistEntry{File, LineStart, LineEnd, SpecID, Rationale}` (literal-substring 매칭, LineStart/LineEnd는 diagnostic-only) 구조를 재사용/일반화. 신규 entry는 `Rationale` 필수 — 사유 없는 allowlist 등재 금지.
3. **실패 출력**: `file:line: matched-token` + 가드별 greppable sentinel. 제안 sentinel: `RULE_REQ_AC_TOKEN_LEAK`, `RULE_PROVENANCE_LEAK`, `RULE_DATE_PROVENANCE_LEAK`, `RULE_GOVERNANCE_TOKEN_LEAK` (기존 `RULE_TEMPLATE_MIRROR_DRIFT` 명명 관례 준용).
4. **recurrence backstop**: 각 가드는 `TestSkillBodyLeakClassRecurrenceBackstop` 패턴의 셀프테스트를 동반 — synthetic re-leak probe가 fire하고 clean replacement가 pass함을 결정론적으로 문서화 (REQ-TRC-066).
5. **RED→GREEN 프로토콜**: M1에서 가드가 현재 트리의 기지 위반을 실명 검출(RED, progress.md §E.2 기록) → M2-M6 정리 → M7 GREEN. push는 GREEN 후에만.

## §2 가드 (a) — 일반화 REQ/AC 토큰 (REQ-TRC-060)

- **현황**: `internal_content_leak_test.go`의 `S3-req-ac-token-any-prefix` class가 정확히 목표 패턴 `\b(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+\b`을 보유하나 `skillBodyScoped: true` — 주석 명시 "REQ/AC tokens in agents/rules are owned elsewhere (EXCL-SBN-002)".
- **설계**: rules-스코프 **sibling class 추가** (기존 class의 스코프 플래그 변경보다 안전 — 기존 EXCL-SBN-002 계약 문서와 skills 측 allowlist를 건드리지 않음). 동일 regex, 스코프 술어는 `strings.HasPrefix(rel, ".claude/rules/")`.
- **prefix allowlist 대체**: 감사가 지적한 ATR/WO/COORD/UNP/LNC/TII prefix-allowlist 방식(허용 prefix 외만 검출 → 신규 prefix가 자동 통과하는 구조적 사각)을 폐기하고, **전 패턴 검출 + 파일단위 pedagogical allowlist**로 전환. 이행기: 기존 prefix-allowlist 테스트가 별도 파일에 존재하면(후보 4파일, M1에서 확정) 신규 가드로 대체 또는 병존시키되 이중 유지비용을 피하기 위해 대체를 우선 검토.
- **pedagogical allowlist 초기 entry**: (i) `development/manager-develop-prompt-template.md`의 `AC-XXX-001`/`REQ-XXX-*` placeholder — 또는 더 견고하게 regex 자체에서 도메인 세그먼트 `XXX` 리터럴 제외: 매치 후 `-XXX-` 포함 토큰 skip (allowlist entry보다 규칙이 자명). 두 방식 중 run-phase에서 택1, 결정 근거를 테스트 주석에 기록.
- **오탐 완화**: cleanup(M3) 완료 후 잔여 정당 토큰은 placeholder뿐이어야 함 — GREEN이 곧 오탐 0의 증명.

## §3 가드 (b) — lessons #N / 사건-W# provenance (REQ-TRC-061)

- **패턴 (RE2 호환, 협패턴 우선)**:
  - `\blessons? #[0-9]+` — lessons 참조 (한국어 문장 내 "lessons #21 W0 fix 패턴" 포함 매치).
  - `\bW[0-9] (meta-analysis|meta|fix)\b` + `\bW[0-9]/W[0-9]\b` + `\bW[0-9] 케이스` — 사건-W# 관용형 3종 (감사 실측 인스턴스가 전부 이 3형에 속함).
- **의도적 협패턴 선택 근거**: 광패턴 `\bW[0-9]\b`는 `SPEC-W3-*`류 미래 식별자·표 헤더 등에서 오탐 여지. RE2에 lookahead가 없어 문맥 배제 불가하므로, 관용형 열거 + allowlist가 유지비용이 낮다. 협패턴이 놓치는 신형 W# 표현은 recurrence 시 관용형 추가로 대응 (테스트 주석에 확장 절차 명기).
- **allowlist**: 초기 0건 목표 (M3 정리로 전량 소거). `W3C`류는 협패턴상 애초에 비매치.
- **배치**: 가드 (a)와 같은 rules-스코프 class 군 또는 신규 `rule_provenance_audit_test.go` — M1에서 파일 배치 확정 (사용자 결정 3의 "+2~3 test file changes" 범위 내).

## §4 가드 (c) — 내부 날짜 provenance 휴리스틱 (REQ-TRC-062)

- **[HARD 제약] 배치**: default-tier `leakClasses` slice에 추가 **금지**. `TestLeakClassNoDateShaInDefaultTier`(:857)가 날짜 probe에 매치되는 default-tier class를 즉시 FAIL시키는 tier-ownership 계약을 보유 ("date detection is owned by the strict tier"). → **독립 테스트 함수** `TestRuleInternalDateProvenance` (신규 파일 `rule_date_provenance_audit_test.go` 권장)로 구현, `leakClasses`와 자료구조 비공유.
- **패턴**: `\b20[0-9]{2}-[0-9]{2}-[0-9]{2}\b` (ISO 날짜).
- **오탐 완화 (라인-컨텍스트 제외 + allowlist 2단)**:
  1. 라인-컨텍스트 제외: 매치 라인이 다음 접두/포함 형태면 skip — `Last Updated:`, `Version:`, `Status:`, `Relocated:` (footer 메타데이터 관례). 단 **본 SPEC의 M3가 footer 내부 작업 날짜(design/constitution.md:423-424)도 제거**하므로, 제외 규칙은 "미래의 정당한 문서 메타데이터"를 위한 보험이지 현재 위반의 면죄부가 아님 — RED 시점에는 제외 규칙을 임시 비활성화한 카운트도 함께 기록해 판정 일관성을 확인.
  2. 파일단위 allowlist: 정당 만료일(BC window 등)이 발견되면 `{File, Substring, Rationale}` entry로 등재. 초기 목표 0건.
- **리스크 잔존 인지**: 날짜 휴리스틱은 본질적으로 의미 판별 불가 (내부 작업 날짜 vs 정책 유효일). 가드의 역할은 "새 날짜 유입 시 사람의 명시 판정(allowlist 등재 or 제거)을 강제"하는 ratchet이지 자동 판별기가 아님 — 이 성격을 테스트 doc comment에 명기.

## §5 가드 (d) — CONST-V3R* / SPEC-V3R* rules 스코프 확장 (REQ-TRC-063)

- **현황**: leak test line ~199 pattern `\bSPEC-V3R[0-9]-[A-Z0-9-]+\b|\bCONST-V3R[0-9]-[0-9]+\b|...` — skill-body 스코프.
- **설계**: 가드 (a)와 동일하게 rules-스코프 sibling class 추가 (동일 regex 재사용 — regex 중복 선언 대신 패키지 상수로 추출 검토). zone-registry.md가 M3에서 제거되므로 121건 최대 오염원이 소멸; 잔여는 `manager-develop-prompt-template.md` 4줄 + `worktree-integration.md` 2곳 → M3 정리 후 GREEN.
- **MIG-003류 standalone 마이그레이션 ID**: 별도 regex `\bMIG-[0-9]{3}\b` 추가 여부는 run-phase 판단 — 현재 실측 1건(settings-management.md:174)뿐이라 정리 후 grep-AC로 충분할 수 있음. 가드화하면 오탐 검토 필요 (일반 영단어 충돌 없음, 추가 비용 낮음 → 포함 권장).

## §6 테스트 배치 요약

| 가드 | 배치 | 신규/수정 |
|------|------|-----------|
| (a) REQ/AC rules-스코프 | `internal_content_leak_test.go` sibling class 또는 `rule_provenance_audit_test.go` | 수정 or 신규 |
| (b) lessons/W# | 가드 (a)와 동일 파일 (provenance 계열 묶음) | 동일 파일 |
| (c) 날짜 | **신규** `rule_date_provenance_audit_test.go` (leakClasses 비공유 — tier 계약) | 신규 |
| (d) CONST/SPEC-V3R | `internal_content_leak_test.go` sibling class | 수정 |

합계 변경 파일: 2~3개 (사용자 결정 3의 예상 범위와 일치). 기존 계약 테스트 3종(`TestLeakClassNoDateShaInDefaultTier`, `TestSanitizedPairParity`, `TestRuleTemplateMirrorDrift`)은 무접촉 green 유지가 설계 불변식.

## §7 CI 연동

- 로컬: `go test ./internal/template/ -count=1` 에 자동 포함 (신규 테스트 함수는 별도 등록 불요 — go test 자동 발견).
- CI: `.github/workflows/template-neutrality-check.yaml` (path-trigger) — 신규 테스트가 `internal/template/` 패키지에 있으므로 기존 워크플로 트리거로 자동 커버. 워크플로 파일 수정 불요 (run-phase에서 트리거 경로에 `internal/template/templates/.claude/rules/**` 포함 여부 확인, 미포함 시 경로 추가).
