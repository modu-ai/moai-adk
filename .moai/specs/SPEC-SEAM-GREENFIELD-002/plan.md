---
id: SPEC-SEAM-GREENFIELD-002
title: "plan — greenfield 씨앗의 flow 스타일 고착 수리 (t1050)"
version: "0.1.0"
created: 2026-09-20
updated: 2026-09-20
author: GOOS
module: "internal/settings/yamlpatch"
tier: S
---

> 본 산출물은 `status:` 필드를 두지 않는다(SPEC 디렉터리의 상태축은 `spec.md` 단일 소관 — frontmatter 스키마 SSOT § Artifact Statelessness).

## §A Context

- **카드**: t1050 (Class B — 결함, 원인 특정 완료). 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1050`, 브랜치 `WT-greenfield-style`, plan-phase 기준 HEAD `159dd30df`.
- **Tier**: S. 산출물은 `spec.md` + `plan.md` + `acceptance.md`(디스패치 지시로 AC를 분리 저작) + `progress.md`.
- **결함 요약**: `PatchFile`의 greenfield 씨앗 `data = []byte("{}\n")`이 flow 스타일 문서로 파싱되고, yaml.v3 인코더가 루트 노드의 스타일을 보존해 생성 파일이 한 줄 flow YAML이 된다. 그 형상은 이후 저장에도 고착된다. 근거 표: `spec.md` §1.3(E1..E12), 판정 방향: §1.5.
- **상류**: 카드 t1043 판정서 `.claude/worktrees/t1043/.moai/reports/t1043/verdict.md` §C9 — "한 줄로 고칠 수 있으나 이 카드에서 고치지 않았다"로 남긴 신규 결함.
- **run-phase 방법론**: RED-GREEN-REFACTOR (tdd) — defect-fix with explicit reproduction. RED는 AC-SGF2-001·002가 강제한다.

## §B Known Issues (본 SPEC 도메인 관련만)

- **B-a (과잉 수리의 함정)**: 스타일 해제를 인코딩 직전에 **무조건** 걸면 다섯 셀이 전부 green으로 보인다 — 그리고 사용자가 일부러 flow로 적은 기존 파일을 조용히 재포맷한다. 판별 증거는 `acceptance.md` **AC-SGF2-003b**(deliberate-flow 셀)와 AC-SGF2-005의 (b) 방향 뮤턴트뿐이고, 후자는 전자를 자기 검출기로 지목하므로 **실질적으로 셀 하나에 수렴한다**. [HARD] 그 셀이 판별력을 갖는 것은 edit이 **upsert**이고 단정이 **원본 대비 바이트 비교**일 때뿐이다 — 스칼라 교체 edit은 `lineSplice` 빠른 경로에서 끝나 스타일이 결정되는 재직렬화 경로에 닿지 않고, 개행·들여쓰기 휴리스틱은 루트만 de-flow되는 (b) 뮤턴트를 놓친다. 측정 근거와 두 형태의 verbatim 출력은 `acceptance.md` AC-SGF2-003b에 있다.
- **B-b (단일 셀 공허 green)**: 섹션 하나만 재면 씨앗이 공유 지점이라는 사실이 가려진다. AC-SGF2-001은 다섯 루트 키 전부를 요구한다.
- **B-c (생성 ≠ 고착)**: 생성만 고치고 두 번째 저장을 재지 않으면 고착이 남아도 AC-SGF2-001은 통과한다 — AC-SGF2-002가 그 구멍이다.
- **B-d (기대값 전환 유혹)**: 통제군 6건 중 하나가 수리와 충돌하면 기대를 고치고 싶어진다. 그것은 수리가 아니라 통제군 파괴다 — 충돌은 수리 형태가 틀렸다는 신호이며, 고치기 전에 `progress.md`에 적는다(AC-SGF2-004).
- **B-e (배차 전제 3건 정정됨)**: `sectionRootKeys`는 11이 아니라 12이고, `WriteSectionViaSeam`으로 실제 도달하는 섹션은 6개다. `@MX:ANCHOR`의 "initializer_expansion ×3"은 0건이다. 상세: `spec.md` §1.4. run 단계는 이 정정된 값을 다시 잰 뒤 주석에 적는다(AC-SGF2-006).
- **B-f (셸 grep)**: 이 트리의 셸 `grep`은 ugrep 래퍼라 조용히 건너뛴다 — 부재 주장 증거는 `/usr/bin/grep`으로 채득하고 양성 대조를 나란히 인용한다.
- **B-g (줄 앵커 노후)**: `spec.md` E4/E10/E11의 file:line은 `159dd30df` 실측값이다. run 착수 시 content-token(`func <name>`, `@MX:ANCHOR`, 씨앗 리터럴)으로 재검증한다.
- **B-h (CI 3-tier)**: spec-lint / golangci-lint / per-OS test는 각자 실패할 수 있다 — 기존 baseline과 NEW 결함을 구분해 보고한다.

## §C Pre-flight (run-phase 착수 시)

```bash
# 1. 브랜치 + HEAD 재확인 (디스패치 값과 다르면 정지·보고)
git rev-parse --short HEAD && git branch --show-current

# 2. content-token 앵커 재검증 (줄 인용은 낡는다)
/usr/bin/grep -n 'data = \[\]byte("{}' internal/settings/yamlpatch/yamlpatch.go
/usr/bin/grep -n '@MX:ANCHOR' internal/settings/yamlpatch/yamlpatch.go
/usr/bin/grep -rn 'yamlpatch\.PatchFile(' --include='*.go' internal/ cmd/ pkg/

# 3. 부재 주장 + 양성 대조 (AC-SGF2-006용)
/usr/bin/grep -rn 'yamlpatch' internal/core/project/ ; echo "exit=$?"
/usr/bin/grep -c 'func' internal/core/project/initializer_expansion.go   # 양성 대조: 비어 있지 않아야

# 4. 기존 baseline 측정 (수정 전 상태 — NEW vs 기존 구분용)
go test ./internal/settings/... ./internal/cli/... -count=1
```

## §D Constraints

### PRESERVE (수정 금지)

- 기존 섹션 파일에 대한 라인-스플라이스 바이트 보존 불변(빈 줄·주석·키 순서·unknown key·인용 스타일·typed 스칼라) — REQ-SGF2-004.
- 사용자가 일부러 flow로 적은 **기존** 문서의 형상 — 재포맷 금지(B-a).
- `atomicWrite`의 temp+rename 원자성과 absent-허용 분기(SPEC-SEAM-GREENFIELD-001의 결과물).
- 통제군 6건(`acceptance.md` AC-SGF2-004)의 기대값 — 무수정 GREEN.
- `internal/settings/sectionapply.go`의 C6 무변경-스킵 게이트, `sectionroute.go`의 라우팅 판정 — 본 SPEC 무접촉.
- `internal/web/` 렌더러 — 무접촉.

### 금지 사항

- `go test ./...` 로컬 전체 실행 금지 — 범위는 `./internal/settings/... ./internal/cli/...`.
- 기존 flow 파일 소급 재포맷 금지(`spec.md` §5).
- `sectionRootKeys` 12 엔트리 / `RouteSeam` 6 섹션의 불일치 정리 금지 — 관측만, 별도 카드 소관.
- `sectionwrite.go:57`의 "8개 섹션" 독스트링 수정 금지 — 본 SPEC이 고치는 함수 위가 아니다.
- 새 의존성 추가 금지. `--no-verify`·force-push 금지. 커밋은 Conventional Commits + 카드 id(t1050) + `🗿 MoAI` 트레일러.

## §E Self-Verification (run-phase 완료 보고 항목)

E1 AC PASS/FAIL 매트릭스(`acceptance.md` AC-SGF2-001..006, 커맨드 + 실출력 + 트리 SHA 귀속) · E2 크로스플랫폼 빌드(`GOOS=windows GOARCH=amd64 go build ./...`) · E3 범위 패키지 커버리지 · E5 lint(NEW vs baseline 구분) · E6 브랜치 HEAD + 미푸시 커밋 수 · E8 RED verbatim 채득(AC-SGF2-001·002 수리 전 실패 출력 — 반증 가능성의 근거) · E9 뮤턴트 2방향 채득 + 못 잡은 뮤턴트 기록.

## §F Milestones

결정 가변성 순 — 가장 변할 가능성이 큰 결정(수리 형태와 그 경계)을 앞에 두고, 기계적 재검증을 뒤로 보낸다.

### M1 — 수리 형태의 경계 확정 (가장 되돌리기 비싼 결정)

1. 스타일 해제를 **greenfield 경로에만** 거는 조건 표현을 정한다(씨앗을 취했음을 기록하는 방식). `spec.md` §4의 두 대안 — 씨앗 플래그 방식 vs 씨앗 리터럴 자체 변경 — 중 어느 쪽을 취하는지와 그 근거를 `progress.md`에 적는다.
2. 조건이 `lineSplice` 빠른 경로와 재직렬화 폴백 **양쪽**에 올바로 걸리는지 직접 측정한다 — greenfield 문서가 항상 폴백으로 간다는 가정을 측정 없이 전제하지 않는다(`acceptance.md` §D).
3. `detectIndent(data)`가 `{}\n`에 대해 무엇을 반환하는지 측정해 기록한다. 형제 파일 관례와 어긋나면 **고치지 말고 기록**하고 별도 카드로 올린다(범위 규율).

### M2 — RED 가드 (생성 + 고착 + 보존)

1. greenfield block-스타일 가드 신설: 다섯 루트 키(`mcp`/`report`/`crosssession`/`gate`/`cacheStrategy`)에 대해 생성 출력이 block임을 단정 (AC-SGF2-001). 수리 전 트리에서 다섯 셀 RED 채득 — 실패 출력에 실제 flow 형상이 보여야 한다.
2. 고착 가드 신설: 생성된 파일에 두 번째 실변경 edit → 여전히 block + 첫 키 잔존 (AC-SGF2-002). 수리 전 RED 채득.
3. 바이트 보존 가드 신설: 기존 block 파일의 미편집 바이트 불변 (AC-SGF2-003) + **deliberate-flow 파일은 flow 유지** (AC-SGF2-003b — edit은 **upsert**, 단정은 **원본 대비 바이트 비교**). 이 두 셀은 수리 전에도 PASS한다. 그것을 "정상"으로 귀속하지 않는다 — 수리가 아직 없어 재포맷할 주체가 없기 때문이고, 셀의 역할은 RED 채득이 아니라 M4 뮤턴트 (b) 아래에서 FAIL하는 것이다. 스칼라 교체 edit으로 쓰면 (b) 아래에서도 PASS해 판별력이 0이 된다(B-a).
4. **[HARD] 순서**: RED 채득 산출물은 수리 커밋보다 **앞선 커밋**에 들어간다 — 같은 커밋에 담으면 순서 주장이 영구히 검증 불가가 된다(커밋 그래프만이 순서를 증언한다).

### M3 — 수리 + 스테일 주석 정정

1. M1에서 확정한 형태로 수리를 적용한다 (REQ-SGF2-001..004).
2. M2의 가드가 PASS로 뒤집히는지 확인한다.
3. `@MX:ANCHOR` 주석(`yamlpatch.go:52-53`)을 **run 시점에 직접 잰** 호출점 분포로 정정한다 (AC-SGF2-006). 본 SPEC E5/E11의 수치를 그대로 옮겨 적지 않는다 — 재측정 커맨드·출력·양성 대조를 채득한다.

### M4 — 뮤턴트 + 통제 재확인

1. 뮤턴트 2방향(`acceptance.md` AC-SGF2-005): (a) 스타일 해제 제거 → 신규 가드 FAIL, (b) 조건 없이 무조건 해제 → **AC-SGF2-003b** deliberate-flow 셀 FAIL(돌리기 전에 그 셀이 upsert + 바이트 비교 형태인지 확인한다 — 아니면 검출기가 없는 것이지 뮤턴트가 잡힌 것이 아니다). 각각 verbatim 채득 후 복원, `git diff --stat` 무출력 증명.
2. 못 잡은 뮤턴트가 있으면 내용과 함께 `progress.md`에 기록 (REQ-SGF2-008).
3. 통제군 6건 무수정 GREEN 확인 (AC-SGF2-004).

### M5 — 범위 한정 판정 + 선행 SPEC 패치

1. `go test ./internal/settings/... ./internal/cli/... -count=1` + `GOOS=windows GOARCH=amd64 go build ./...` + `golangci-lint run`.
2. `SPEC-SEAM-GREENFIELD-001/spec.md`에 **정확히 두 가지만** 덧붙인다 — HISTORY 한 행(greenfield 출력 스타일의 소유가 002임) + `related_specs`에 `SPEC-SEAM-GREENFIELD-002` 추가. `status`·본문 섹션·`progress.md` 무접촉. (이 패치는 plan 단계에서 이미 적용돼 있으면 재적용하지 않는다 — 현재 상태부터 읽는다.)
3. §E 자가검증 항목 채움 → 완료 보고(카드 id + 브랜치/HEAD + 미푸시 수 + 증거 경로 + 재측정 범위).

## §G Anti-Patterns

- 섹션 하나만 재고 "greenfield가 block이 됐다"고 보고하기 (B-b).
- 생성만 고치고 두 번째 저장을 재지 않기 (B-c).
- 스타일 해제를 인코딩 직전 무조건 걸어 다섯 셀을 통과시키기 — 사용자의 deliberate-flow 파일을 조용히 재포맷한다 (B-a).
- AC-SGF2-003b의 edit을 **스칼라 교체**로 쓰기 — `lineSplice` 빠른 경로에서 끝나 스타일 결정 지점에 닿지 않으므로, 올바른 수리와 과잉 적용 뮤턴트가 바이트 동일 출력을 내고 셀의 판별력이 0이 된다(측정됨: `spec.md` E13).
- AC-SGF2-003b를 "개행이 생겼는가 / 들여쓰기가 있는가"로 단정하기 — 뮤턴트(b)는 **루트만** de-flow하므로 그 휴리스틱을 통과한다. 단정은 원본 대비 바이트 비교여야 한다.
- 통제군 6건의 기대값을 수리에 맞춰 고치기 (B-d).
- 상류 t1043 §C9의 수치를 이 트리의 측정으로 인용하기 — 그것은 다른 트리(`0bf27ea69`)의 값이다.
- 본 SPEC E5/E11의 호출점 수치를 재측정 없이 주석에 옮겨 적기 (AC-SGF2-006).
- 범위 밖 정리(등록부 12 vs 라우팅 6, `sectionwrite.go:57` 독스트링)를 "지나가는 김에" 같이 고치기.

## §H Cross-References

- `spec.md`: `.moai/specs/SPEC-SEAM-GREENFIELD-002/spec.md` (REQ-SGF2-001..009, §1.3 증거 사슬 E1..E12, §1.4 배차 전제 정정, §4 수리 방향 소견, §5 Out of Scope)
- `acceptance.md`: 같은 디렉터리 — AC-SGF2-001..006 정본(Given-When-Then), §C 완료 게이트, §D 전방 점검
- 상류 증거: `.claude/worktrees/t1043/.moai/reports/t1043/verdict.md` §C9 (flow 고착 관측 + block-씨앗 양성 대조)
- 본 카드 측정 기록: `.moai/reports/t1050/plan-measurements.md` — E1(§C1)·E2(§C2)·E12(§C4) verbatim. §C3은 2026-09-20 정정 블록으로 기각된 해석이므로 인용하지 않는다(정정된 사실은 `spec.md` §1.4).
- plan-audit 판정서: `.moai/reports/t1050/plan-audit-verdict.md` (iter1 FAIL 0.74 — D1..D8; 수정 반영 기록은 `progress.md` §E.1)
- 선행 SPEC: `.moai/specs/SPEC-SEAM-GREENFIELD-001/{spec.md,plan.md}` (greenfield 생성 자체의 수리, 뮤턴트-오버레이 방법론, §4 암묵 기본값을 명시 선택으로 문서화한 선례)
- 방법론 원천: `SPEC-WEB-WRITE-SAFETY-001/progress.md` (렌더-오버레이 뮤턴트 절차), `SPEC-WEB-WRITE-SAFETY-001/spec.md` REQ-WWS-005 (라인-스플라이스 바이트 보존 불변)
- 규율: `.claude/rules/moai/core/verification-claim-integrity.md` §2(baseline 귀속)·§2.3(커밋 그래프만이 순서를 증언한다), `.claude/rules/moai/development/verification-completeness.md` §2(two-cell + 뮤턴트 프로브)
