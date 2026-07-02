---
id: SPEC-TOOLPOLICY-DEPLOY-REVIEW-001
title: "tool-policy.yaml 사용자 프로젝트 배포 게이팅 리뷰 및 결정"
version: "1.0.0"
status: completed
created: 2026-07-02
updated: 2026-07-02
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/template/templates/.moai/config/sections"
lifecycle: spec-anchored
tags: "template, deployment, tool-policy, distribution-neutrality, cleanup"
tier: S
era: V3R6
---

## HISTORY

- 2026-07-02 (v0.1.0): 초안 작성. `tool-policy.yaml`(48KB / 181 entries)가 모든 사용자 프로젝트에 배포되지만 어떤 사용자 런타임 경로에서도 소비되지 않음을 검증. 조사 결과 기반 Direction A(dev-only 재배치) 권고. — manager-spec

---

## §1 Context / 문제 정의

`internal/template/templates/.moai/config/sections/tool-policy.yaml`는 **48,315 bytes / 181 entries** 파일로, `moai init` / `moai update`가 **모든** 사용자 프로젝트에 배포한다. 그러나 이 파일은 `moai tool-policy build|list` codegen CLI에서만 소비되며(`internal/cli/tool_policy.go` + `internal/config/toolpolicy/{loader,codegen}.go`), `moai init`·템플릿 deployer·hook·기타 런타임 경로 어디에서도 읽지 않는다.

핵심 사실: 런타임 권한 강제(permission enforcement)는 `tool-policy.yaml`을 직접 읽지 않고, **생성된 `.claude/settings.json`의 `permissions` 블록**을 사용한다. 그리고 그 `settings.json`은 init 시점에 `settings.json.tmpl`로부터 **독립적으로** 렌더링된다 — `tool-policy.yaml`과 무관하다.

이 SPEC은 "사용자 프로젝트가 `tool-policy.yaml`을 필요로 하는가?"를 판정하고, 판정에 따라 배포를 게이팅할지(Direction A) 유지+문서화할지(Direction B) 결정하는 **리뷰 SPEC**이다.

### §1.1 조사 결과 (검증된 증거)

| # | 질문 | 관측 증거 | 판정 |
|---|------|-----------|------|
| 1 | `moai init`이 `tool-policy.yaml`에서 `settings.json`을 생성하는가? | `init.go` `runInit`는 `template.NewDeployerWithRenderer`/`NewSlimDeployerWithRenderer` → `deployer.Deploy()`만 호출. `internal/config/toolpolicy` 패키지를 import/호출하지 않음. `settings.json.tmpl`(L387에 permissions 블록 baked-in)이 `.claude/settings.json`으로 **독립 렌더링**됨. | **아니오** — settings.json은 .tmpl에서 독립 생성 |
| 2 | `tool-policy.yaml`을 소비하는 Go 코드는? | `grep -rln "config/toolpolicy" --include="*.go" \| grep -v _test.go` → **오직** `internal/cli/tool_policy.go` 1개 | codegen CLI 전용 |
| 3 | `moai tool-policy build`가 사용자 자동 플로우에 포함되는가? | init.go/deployer/hook 어디서도 호출 안 함. 수동 CLI 명령. | 수동 전용 |
| 4 | slim(기본)/full 배포에서 `tool-policy.yaml`이 배포되는가? | `slim_fs.go`는 **catalog non-core 항목만** 숨김. `tool-policy.yaml`은 catalog 항목이 아니므로 slim·full 모두 통과 배포. `catalog.yaml`에 `tool-policy` 항목 0건. | 예 — 항상 배포됨(전제 성립) |
| 5 | 사용자 대상 문서가 `moai tool-policy`를 안내하는가? | `grep -rn "moai tool-policy" docs-site/ README.md README.ko.md` → **0건** | 미문서화 |
| 6 | 템플릿 `tool-policy.yaml`이 dev 빌드의 SSOT인가? | dev SSOT는 **repo-root** `.moai/config/sections/tool-policy.yaml`(48,420 bytes). 템플릿 복사본(48,315 bytes)과 **byte-different**. `moai tool-policy build --template-only`는 repo-root를 읽어 `.tmpl`을 씀 — 템플릿 `tool-policy.yaml` 자체는 절대 읽지 않음. | 템플릿 복사본은 orphan(고아) SSOT |
| 7 | 템플릿 `tool-policy.yaml` 배포를 단언하는 테스트가 있는가? | 참조 3건(`tool_policy_test.go`, `toolpolicy/coverage_test.go`, `audit_loader_completeness_test.go`) 모두 **repo-root** fixture(`../../.moai/config/sections/...`) 또는 dedicated-loader allowlist 참조. 템플릿 복사본 배포를 단언하는 테스트 **0건**. `TestAuditLoaderCompleteness`는 uncovered-only 검사(역방향 stale-allowlist 검사 없음)라 파일 제거를 허용. | 없음 — 제거 안전 |

### §1.2 결론

- 사용자는 `tool-policy.yaml` **없이도** 작동하는 `settings.json`(권한 블록 포함)을 init 시점에 `.tmpl`에서 받는다.
- 사용자가 `tool-policy.yaml`을 필요로 하는 유일한 경우는 **직접** `moai tool-policy build`를 수동 실행해 권한 블록을 재생성/커스터마이징하려는 경우인데, (a) 이를 안내하는 문서가 없고, (b) 배포되는 파일은 moai-adk 자체의 181-entry 내부 정책이라 범용 사용자 커스터마이징 템플릿으로 적합하지 않으며, (c) 사용자는 `settings.json` permissions를 직접 편집하는 표준 Claude Code 방식으로 `tool-policy.yaml` 없이 커스터마이징할 수 있다.

따라서 배포되는 `tool-policy.yaml`은 사용자에게 **실현 가치 ≈ 0**, **비용 48KB/프로젝트**. **Direction A(dev-only 게이팅)** 를 권고한다. 상세 판단·잔여위험은 §5.

---

## §2 Requirements (GEARS)

- **REQ-TDR-001** (Unwanted behavior): The `moai init` template deployment **shall not** write `.moai/config/sections/tool-policy.yaml` into a newly initialized user project.
- **REQ-TDR-002** (Ubiquitous — preservation): The `.claude/settings.json` permissions block **shall** remain generated from `settings.json.tmpl` at init time, byte-unchanged and independent of `tool-policy.yaml`.
- **REQ-TDR-003** (Where — dev SSOT preservation): **Where** the repository is the moai-adk-go template source, `moai tool-policy build` and `moai tool-policy list` **shall** continue to operate against the repo-root `.moai/config/sections/tool-policy.yaml` dev SSOT.
- **REQ-TDR-004** (When — test integrity): **When** the template copy of `tool-policy.yaml` is removed, the full test suite (`go test ./...`) **shall** continue to pass, and the `TestAuditLoaderCompleteness` `acknowledgedDedicatedLoaders` comment **shall** be updated to reflect the dev-only relocation.
- **REQ-TDR-005** (Where — user customization path): **Where** a user needs to customize tool/permission decisions, the recommended path **shall** be documented in one durable surface (edit `.claude/settings.json` permissions directly; or author a `tool-policy.yaml` and run `moai tool-policy build`).

---

## §3 Acceptance Criteria (Tier S — inline)

### AC-TDR-001a — 신규 init에 tool-policy.yaml 미배포
- **Given** 빈 디렉터리에서 `moai init`을 실행하고
- **When** init이 성공적으로 완료되면
- **Then** `<project>/.moai/config/sections/tool-policy.yaml`이 **존재하지 않는다**.

### AC-TDR-001b — settings.json 권한 블록 불변(회귀 방지)
- **Given** 동일한 신규 init 프로젝트에서
- **When** `.claude/settings.json`을 확인하면
- **Then** `permissions` 블록(allow/deny/ask 리스트)이 존재하고 비어있지 않으며, 본 SPEC 변경 전과 **byte-identical** 하다.

### AC-TDR-002 — 전체 테스트 통과
- **Given** 템플릿에서 `tool-policy.yaml`이 제거된 트리에서
- **When** `go test ./...`를 실행하면
- **Then** 모든 테스트가 통과한다(`TestAuditLoaderCompleteness` 포함).

### AC-TDR-003 — dev SSOT 온전성
- **Given** moai-adk-go 리포지토리 루트에서
- **When** `moai tool-policy list` (repo-root 기본 경로) 를 실행하면
- **Then** 181개 정책 엔트리가 정상 반환된다(dev codegen 능력 보존).

### AC-TDR-004 — 소비 경로 불변 증명
- **Given** 변경 후 트리에서
- **When** `grep -rln "config/toolpolicy" --include="*.go" . | grep -v _test.go` 를 실행하면
- **Then** 결과는 여전히 `internal/cli/tool_policy.go` 1건뿐이다(신규 런타임 소비자 미추가).

### Edge cases
- **EC-1** (기존 프로젝트): `moai update`는 템플릿에 없어진 파일을 **삭제하지 않으므로**, 이미 `tool-policy.yaml`을 가진 기존 프로젝트는 파일을 그대로 보존한다(무해). 본 SPEC은 **신규 init**에만 영향(§5 잔여위험 R2).
- **EC-2** (`moai tool-policy list`를 신규 사용자 프로젝트에서 실행): `tool-policy.yaml` 부재 시 `load policy` 에러 발생. run-phase에서 graceful 처리(안내 메시지) 여부를 A1/A2 하위결정과 함께 확정(§5).

---

## §4 Direction 비교

| 축 | Direction A (배포 게이팅 — 권고) | Direction B (유지 + 문서화) |
|----|-------------------------------|----------------------------|
| 전제 성립 여부 | settings.json이 .tmpl에서 독립 렌더링됨 → **성립** | 사용자가 tool-policy.yaml에서 settings.json을 재생성함 → **불성립**(자동 경로 없음) |
| 사용자 가치 | 48KB 데드웨이트 제거, 배포 중립성 향상 | 미문서화 명령용 orphan SSOT 유지 |
| 회귀 위험 | `moai tool-policy list`가 신규 사용자 프로젝트에서 파일 부재 에러(EC-2, 완화 가능) | 없음 |
| §15/§25 정합 | moai-adk 내부 181-entry 정책 배포 제거 → isolation 정합 | 내부 정책 계속 배포 |

권고: **Direction A**. Direction B의 전제("사용자가 tool-policy.yaml에서 settings.json을 재생성")는 자동 플로우·문서 부재로 성립하지 않는다.

---

## §5 결정 및 잔여위험

- **결정**: Direction A — 템플릿 트리에서 `internal/template/templates/.moai/config/sections/tool-policy.yaml`를 제거해 dev-only(repo-root 전용)로 재배치.
- **run-phase 하위결정(오케스트레이터/사용자 확인 대상)**:
  - **A1 — 완전 제거**: 파일을 템플릿에서 삭제. 가장 깨끗하나 신규 사용자 프로젝트에서 `moai tool-policy list/build`가 파일 부재.
  - **A2 — 최소 stub로 치환**: 몇 개 예시 엔트리 + 주석만 담은 소형 `tool-policy.yaml`로 교체. `moai tool-policy` 명령 사용성 보존 + 48KB/내부콘텐츠 제거. (§25 내부 콘텐츠 격리에도 정합)
  - 본 SPEC은 **A(게이팅)** 를 확정하고 A1/A2는 plan.md §F M1에서 오케스트레이터 확인 후 선택.
- **잔여위험**:
  - **R1** (`moai tool-policy list` 회귀, EC-2): A2 선택 또는 명령의 graceful 부재-처리로 완화. A1 선택 시 명령 사용성은 dev-repo에 한정됨을 문서화(REQ-TDR-005).
  - **R2** (기존 프로젝트): `moai update`가 orphan 파일을 pruning하지 않으므로 기존 사용자는 파일을 보존(무해). 소급 정리는 본 SPEC 범위 밖.
  - **R3** (Chesterton's Fence): 원 SPEC-V3R6-TOOL-POLICY-SSOT-001의 `moai tool-policy build` consumer-project 인지(주석 "consumer project — skipping template target")는 "사용자가 로컬 tool-policy.yaml로 재생성 가능"을 **설계 의도**로 시사한다. 그러나 그 의도는 문서·자동 경로로 실현되지 않았다. A2(최소 stub)는 이 설계 의도를 최소 비용으로 존중하는 절충안이다.

---

## Exclusions

본 SPEC이 **다루지 않는** 범위. 각 항목은 별도 처리 경로를 가진다.

### Out of Scope — settings.json permissions 재설계
- `.claude/settings.json`의 `permissions` 블록 구조·정책 자체는 변경하지 않는다. 본 SPEC은 배포 대상 파일 목록만 바꾸며 권한 강제 동작은 불변(REQ-TDR-002).
- `settings.json.tmpl`의 baked-in permissions 블록 편집은 별도 `moai tool-policy build` 워크플로우 소관.

### Out of Scope — moai tool-policy CLI 제거/재설계
- `moai tool-policy build|list` 명령 자체는 dev codegen SSOT로 유지된다(REQ-TDR-003). 명령 제거·기능 변경은 본 SPEC 범위 밖.
- `internal/config/toolpolicy` 패키지 codegen 로직은 손대지 않는다.

### Out of Scope — 기존 사용자 프로젝트 소급 정리
- 이미 `tool-policy.yaml`을 가진 기존 프로젝트에서의 파일 삭제(`moai update` pruning)는 본 SPEC 범위 밖(EC-1 / R2). 신규 init 만 대상.
- orphan 파일 pruning 메커니즘 도입은 별도 SPEC 필요.
