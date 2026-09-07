# t509 — codex 모델별 설정: 착수 전 판정

| 항목 | 값 |
|---|---|
| 카드 | t509 · Class C · Tier M~L · **착수 전 판정 단계** (구현 0) |
| 트리 | `.claude/worktrees/t509` · 브랜치 `WT-codex-model-config` · HEAD `ace1c5440` (`origin/develop`과 델타 `0 0`) |
| 지시 원문 | "각 모델별 설정 할 수 있도록 .moai/config 설정과 moai web 에서 codex 모델 설정 페이지를 추가하자" (운영자, 2026-09-07) |
| 산출 | 이 판정서. **코드·템플릿 수정 없음** — 카드가 [HARD]로 축 분할 판단을 구현보다 앞에 두라고 요구한다 |

---

## 0. 결론 먼저 — 카드의 전제 3개 중 2개가 측정과 어긋난다

카드는 세 가지를 전제한다. 재보니 이렇다.

| # | 카드의 전제 | 측정 결과 |
|---|---|---|
| P1 | `llm.yaml`에 codex 키가 없다 = **누락** | **절반만 참.** `llm.yaml`에 없는 것은 맞다. 그러나 codex 모델/effort 설정은 **이미 존재한다** — `workflow.audit.codex.{model,effort}` |
| P2 | codex 블록은 `claude_models`·`glm`과 **같은 모양**이어야 한다 | **전제가 성립하지 않는다.** 그 둘은 이미 서로 **다른 모양**이다. 맞출 「하나의 기존 모양」이 없다 |
| P3 | 새 블록을 만들면 **로더에 분기가 셋 생긴다** | **거짓.** 로더는 블록별 분기가 없다 — 섹션 전체를 구조체 하나로 한 번에 unmarshal한다 |

그리고 이 카드가 지시하는 방향은 **이미 착지한 판정과 정면으로 부딪힌다**(§3). 그러므로 이 판정서의 결론은 "어떻게 만들 것인가"가 아니라 **"무엇이 진짜 간극이고, 그것을 만드는 것이 옳은가"**다.

> **[§8 먼저 읽을 것]** 운영자가 §6의 결정에서 **「감사 경로만 — web에 노출」**을 골랐다. 그 범위를 재보니 **이미 전부 구현돼 있다** — 전용 Audit 탭, 네 필드, 회귀 테스트, i18n까지. **이 카드는 그 범위에서 만들 것이 없다.** 근거는 §8.

---

## 1. 측정 — 이 트리에서 잰 것

### 1.1 `llm.yaml`은 추적되지 않는다

```
$ git check-ignore -v .moai/config/sections/llm.yaml
.gitignore:200:.moai/config/sections/llm.yaml    .moai/config/sections/llm.yaml
$ git ls-files .moai/config | wc -l
35
```

`.moai/config/` 전체 35파일은 **추적된다**. `llm.yaml` **하나만** 제외돼 있다 — API 키를 담아 모드 `0600`이기 때문이다(커밋 `54d748ddf`, #1499). 따라서 이 파일은 리포의 정본이 아니고, 사용자에게는 **템플릿을 통해서만** 간다.

### 1.2 내 선행 실측의 귀속을 정정한다

t509 카드에 실린 "실측 (1) `llm.yaml`에 codex 관련 키 grep 0히트"는 **내가 상신한 것**이고, 두 가지가 빠져 있었다.

- **잰 대상이 primary 체크아웃의 로컬 파일이었다** — §1.1에 따라 정본이 아니다. 정본인 템플릿에서 다시 쟀다:
  ```
  $ grep -ic 'codex' internal/template/templates/.moai/config/sections/llm.yaml
  0
  $ grep -ic 'glm' internal/template/templates/.moai/config/sections/llm.yaml    # 대조군
  28
  ```
  결론(“`llm.yaml`에 codex 없음”)은 **유지되지만** 귀속이 로컬 → 템플릿으로 바뀐다.
- **범위를 명시하지 않았다.** "codex 관련 키 0히트"는 `llm.yaml` **안에서** 참이지, 리포 전체에서 참이 아니다. 그 생략이 「누락」 판정의 근거가 됐다. §1.3이 그 대가다.

### 1.3 codex 모델/effort는 이미 존재한다 — 다른 자리에

`internal/template/templates/.moai/config/sections/workflow.yaml:85-91`, 템플릿에 이미 배포되고 있다:

```yaml
    audit:
        codex:
            model: ""
            effort: ""
        glm:
            model: ""
            effort: ""
```

- 파일은 **추적된다**(`llm.yaml`과 달리).
- Go 타입은 `config.AuditConfig`(`internal/config/audit_models.go:59-76`)이고 `Codex`/`GLM` 둘 다 재사용 가능한 페어 타입 `ModelEffort{Model, Effort}`(`internal/config/profile.go:73`)를 쓴다.
- 값은 **빈 문자열로 배포**된다(`internal/config/audit_models.go:69-70`: "Distributed default and Go default stay EMPTY (REQ-AMP-005 neutrality)").
- 출처: SPEC-V3R6-AUDIT-MODEL-PIN-001, 카드 t225 착지분.

---

## 2. 착수 전 판정 (1) — 기존 두 블록은 이미 같은 모양이 아니다

카드 [HARD]: "claude_models 와 glm 이 이미 있으므로 codex 블록은 그 둘과 같은 모양이어야 한다 — 세 번째 형태를 새로 발명하면 로더에 분기가 셋 생긴다."

**측정:** 그 둘은 이미 다르다.

| 블록 | 모양 | Go 타입 |
|---|---|---|
| `claude_models` | 평평한 3키 스칼라 `{high, medium, low}` | `ClaudeTierModels` (`internal/config/types.go:321`) |
| `glm` | 2단 중첩 — `base_url` + `models{7키}` + `effort{4키}` | `GLMSettings` (`internal/config/types.go:328`), `GLMModels`(`:362`), `GLMTierEffort`(`:353-358`) |

두 타입은 **서로 무관한 형제 타입**이고, 어느 쪽도 `ModelEffort` 페어를 쓰지 않는다. `glm`은 `models`와 `effort`를 **티어별 페어가 아니라 평행한 두 맵**으로 둔다.

**그러므로 "같은 모양"이라는 요구는 지시할 대상이 없다.** 맞출 기존 모양이 하나가 아니라 둘이고, 셋째를 만들지 않으려면 **어느 쪽에 맞출지를 먼저 골라야 한다** — 그 선택 자체가 이 카드의 설계 결정이다.

**그리고 P3(로더 분기 셋)은 거짓이다.** 로더는 블록을 구분하지 않는다:

```go
// internal/config/loader.go:194-206
func (l *Loader) loadLLMSection(dir string, cfg *Config) {
	wrapper := &llmFileWrapper{LLM: cfg.LLM}
	loaded, err := loadYAMLFile(dir, "llm.yaml", wrapper)
	...
}
```

`llm:` 루트 전체를 `LLMConfig` 구조체 하나로 한 번에 unmarshal한다. 블록을 추가해도 **로더 코드는 한 줄도 늘지 않는다** — 구조체 필드와 yaml 태그만 는다. 분기 걱정은 이 카드의 제약이 아니다.

---

## 3. 착수 전 판정 (2) — 충돌 2건. 이것이 이 카드의 본체다

### 3.1 착지한 판정이 `llm.yaml`을 이미 기각했다

`internal/cli/audit_pin.go:26-28`, 주석 원문:

> The pin lives in workflow.yaml — NOT llm.yaml — because llm.yaml is gitignored and wiped by `moai update`, so a pin there would be uncommitable and non-durable (plan-audit MF1 / lead ruling C).

SPEC-V3R6-AUDIT-MODEL-PIN-001이 **정확히 이 질문을 이미 다뤘고 `llm.yaml`을 기각했다.** 사유는 카드가 스스로 별도 [HARD]로 경고한 것과 같다 — "`moai update`는 `.moai/config`를 통째 삭제 후 재배포한다... 다음 update에 조용히 원복된다."

**즉 카드 안에서 두 [HARD]가 서로 당기고 있다**: 「llm.yaml의 기존 블록과 같은 모양으로 두라」와 「update가 원복하는 위험을 만들지 말라」. 앞의 것을 따르면 뒤의 것을 어긴다.

#### 3.1.1 [중요] 그런데 그 기각 사유의 **절반은 이미 낡았다**

`audit_pin.go`의 사유는 두 주장을 합친 것이다. 재보니 하나는 살아 있고 하나는 죽었다.

| 사유 절반 | 현재 상태 |
|---|---|
| "**gitignored** — uncommitable" | **여전히 참.** `.gitignore:200`. 프로젝트가 핀을 git으로 공유할 수 없다 |
| "**wiped by `moai update`** — non-durable" | **더 이상 참이 아닌 것으로 보인다.** 보존 기제가 프로덕션 경로에 있다 |

프로덕션 체인을 코드로 추적했다:

```
update_template_sync.go:251,403  backup.BackupMoaiConfig(projectRoot)   ← .moai/config 통째 백업
        ↓ (CleanMoaiManagedPaths 가 지우고 템플릿 재배포)
update_restore.go:53             backup.RestoreFromBackupDir(...)
        ↓
restore.go:160                   MergeYAML3WayRetained(newData, oldData, baseData)
        ↓
merge.go:39 → node_merge.go      deepMerge3WayTo   ← 3-way 병합
```

그리고 백업이 `llm.yaml`을 제외하지 **않는다**:

```go
// internal/cli/update/backup/backup.go:53
excludedDirs := []string{}
```

제외 목록이 비어 있다. 출처는 SPEC-UPDATE-YAML-PRESERVE-001이고, 카드 t239가 이 축을 다뤘다 — 이 트리에 `internal/cli/update_llm_preserve_test.go`와 `internal/cli/update_clean_install_config_preserve_test.go`가 존재하며, t239의 뮤턴트-RED 로그가 보존이 깨졌을 때 그 테스트가 **실패함**을 보인다(`llm.yaml: key "manager-develop" missing (path [llm agent_overrides manager-develop model])`).

**그러므로 `llm.yaml`에 둔 사용자 값은 `moai update`를 건너 보존된다** — 적어도 코드 경로와 회귀 테스트가 그렇게 말한다.

**이것이 A0/A1의 선택지를 넓힌다.** `llm.yaml`을 기각한 근거 두 개 중 하나가 무효화됐으므로, 그 기각을 지금 다시 적용할지 여부는 **재검토 대상**이다. 남은 사유("git으로 공유 불가")가 이 카드의 용도에 치명적인지는 별개 질문이다 — 사용자 개인 설정이라면 공유 불가가 오히려 정상이고, 프로젝트 공유 설정이라면 치명적이다.

**단, 나는 `moai update`를 실행하지 않았다.** 이 절은 코드 판독과 테스트 존재 확인이지 실행 관측이 아니다(§7 Residual-risk).

### 3.2 우선순위 규칙은 **이미 존재한다** — 부재가 아니라 존재가 발견이다

리드 [HARD]: "같은 키가 두 자리에 있으면 어느 쪽이 이기는지 아무도 모르게 된다."

**측정: 이미 두 자리에 있고, 이미 규칙이 있다.** 템플릿 주석이 직접 적는다(`workflow.yaml:76-77`):

> Precedence: this pin > the llm.yaml sync-auditor SSOT cell > empty (legacy behavior).

코드에서 실현된 순서(`internal/cli/mcp_codex.go:230-252`, `:197-210`):

```
호출자 명시 params["model"]
  > workflow.audit.codex 핀 (servable 할 때)
  > llm.yaml profiles.<active>.sync-auditor 셀   ← codexAuditAgentKey = "sync-auditor" (mcp_codex.go:153)
  > 빈 ModelEffort{}
```

GLM 쪽도 같은 모양이다(`internal/cli/mcp_glm.go:172-180`). **그러므로 규칙을 발명할 필요가 없다. 새 블록은 이 체인의 어디에 끼는지만 정하면 된다.**

### 3.3 task 경로에 설정이 없는 것은 **의도된 배제**다

이것이 가장 중요한 발견이다. `workflow.audit` 주석(`workflow.yaml:74-76`)이 범위를 스스로 적는다:

> Applies ONLY at the audit entry points (codex_audit / glm_audit / audit_multi) — **never to codex_task / glm_task delegation**.

그리고 코드가 이유를 적는다(`internal/cli/mcp_codex.go:217-220`): task 경로 리졸버 `resolveCodexModelEffort`는 핀을 **일부러 읽지 않는다** — "a config-file pin is persistent project state and must not leak into delegation tasks."

**그러므로 진짜 간극은 「codex 설정이 없다」가 아니다.** 정확히는:

> **감사(audit) 경로에는 codex 모델/effort 설정이 있고, 위임(task) 경로에는 없다. 그리고 그 부재는 기록된 설계 결정이다.**

운영자 지시("각 모델별 설정")가 겨냥한 것이 **감사 경로**라면 이미 충족돼 있고, **위임 경로**라면 그 간극은 실재하되 **착지한 설계 결정을 뒤집는 일**이 된다. 어느 쪽인지는 내가 고를 사안이 아니다(§6).

---

## 4. `moai web` 실측 — 붙일 자리는 있으나 바닥에 공백이 있다

### 4.1 이미 편집 가능하다

- `llm.yaml`은 **이미 web console에서 편집된다** — "GLM Settings" 탭(`internal/web/schemaform.go:39`), 키는 `glm.models.{high,medium,low,fable}`·`glm.effort.{...}`(`internal/settings/sectionapply.go:228-251`). 에이전트별 override는 "Agents" 탭(`agentfm`).
- `workflow.yaml`도 **이미 편집 가능하다** — 다만 다른 경로로: **seam 섹션**(`internal/settings/sectionroute.go:70-105`), 즉 `yamlpatch`로 노드만 수술해 **주석과 미모델링 키를 보존**한다.
- 라우트는 전부 HTML 폼 POST다. JSON REST API는 없다(`/events`만 비-HTML이고 값을 안 나르는 SSE 신호).

### 4.2 [HARD] "부분 저장 금지"는 현재 표면이 **이미 지키지 못한다**

카드 [HARD]: "moai web 은 사용자 설정을 쓰는 표면이다. 저장 실패 시 원본 보존, 부분 저장 금지."

측정 결과 현재 상태는 이렇다.

| 성질 | 실측 |
|---|---|
| 단일 파일 원자성 | **있다** — temp+rename (`internal/config/atomicfile/write.go:41-81`, `yamlpatch.go:189-218`) |
| 백업 | **없다** — 쓰기 경로에 `.bak`·스냅샷 없음 |
| 프로세스 간 락 | **없다** — 프로세스 내 `sync.RWMutex`만(`manager.go:172`) |
| 다중 파일 저장 원자성 | **없다.** `handleSave`(`internal/web/handlers.go:350-558`)는 8단계를 순차 쓰기하고, 뒤 단계가 실패해도 앞 단계는 **롤백되지 않는다**. 코드가 그 부분 상태를 오류 배너로 **명시해 알린다**(`:483-489` 등) — 즉 알려진·문서화된 동작이다 |
| 주석·미모델링 키 보존 | **typed 섹션은 파괴한다.** `llm.yaml`은 typed 경로라 매 저장마다 구조체에서 통째 재마샬된다 |
| `llm.yaml` dirty 게이팅 | **없다.** `manager.go:223-226`이 **모든 `Save()`마다 무조건** `llm.yaml`을 다시 쓴다 — `user`/`quality`만 고쳐도 `llm.yaml`이 재작성된다. (`git-strategy.yaml`은 dirty 플래그로 격리돼 있다 — `manager.go:206-221`, SPEC-GITSTRATEGY-SAVE-ISOLATION-001) |

**따라서 "부분 저장 금지"는 이 카드가 지킬 수 있는 조건이 아니다** — 이미 그렇지 않은 표면 위에 기능을 얹는 것이기 때문이다. 이 카드가 그 결함을 **만들지는** 않지만, 새 설정을 그 위에 놓으면 노출 면이 넓어진다. 별도 축으로 분리해야 한다.

---

## 5. 축 분할안 — 한 카드로 Tier L을 만들지 않는다

카드가 건 표면 4개를 재분류하면 이렇게 갈린다. **A0가 나머지 전부의 선행이다.**

| 축 | 내용 | Tier | 선행 | 성격 |
|---|---|---|---|---|
| **A0** | **범위 확정 — 감사 경로인가 위임 경로인가.** §3.3의 의도된 배제를 뒤집을지 말지 | — | 없음 | **운영자 결정.** 구현 0 |
| **A1** | 스키마 — codex 블록의 **자리와 모양** 확정 + Go 타입 + 템플릿 배포 | S~M | A0 | `workflow.yaml`(추적·seam·보존)이 유력. `ModelEffort` 페어 재사용 |
| **A2** | `moai web` 설정 페이지 | M | A1 | seam 섹션이면 기존 seam 패널 패턴 재사용 |
| **A3** | 템플릿 미러 + 16개 프로그래밍 언어 중립성 + CI 가드 | S | A1 | A1에 흡수 가능(같은 파일) |
| **A4** | **쓰기 안전성** — `llm.yaml` dirty 게이팅 부재, 다중 파일 부분 저장 | M | **없음 — 독립** | **이 카드와 분리.** 기존 결함이지 이 카드가 만드는 것이 아니다 |

**권고**: A0(결정) → A1+A3(한 카드) → A2. **A4는 별도 카드로 지금 발행**하는 것을 권한다 — 이 카드에 묶으면 Tier L이 되고, 묶지 않아도 결함은 이미 존재한다.

**A1의 자리 — 두 후보를 성질로 비교한다**(§3.1.1이 `llm.yaml` 기각 근거의 절반을 무효화했으므로, 한쪽을 미리 배제하지 않는다):

| 성질 | `workflow.yaml` | `llm.yaml` |
|---|---|---|
| git 추적 | **된다** — 프로젝트가 설정을 공유·리뷰 가능 | 안 된다(`.gitignore:200`) |
| `moai update` 보존 | 된다 | **된다** — 3-way 병합(§3.1.1) |
| web console 저장 방식 | **seam** — 주석·미모델링 키 보존 | typed — **주석 파괴**, 매 저장 무조건 재작성 |
| 기존 codex 설정과의 인접성 | **`audit.codex`가 이미 여기 있다** | 없다 |
| 우선순위 체인 | 이미 최상위 핀 | 이미 하위 SSOT 셀 |
| 비밀값 동거 | 없음 | **API 키와 같은 파일**(`0600`) |

측정만 놓고 보면 **`workflow.yaml`이 우세하다** — 특히 web console 저장 방식과 비밀값 동거 두 줄이 결정적이다. `llm.yaml`은 update 보존이 확인되면서 되살아났지만, 주석 파괴·무조건 재작성·비밀값 동거는 그대로 남는다.

**다만 이것은 A0의 답에 종속된 권고이지 판정이 아니다.** 그리고 "사용자 개인 설정이라 git에 안 올라가는 편이 맞다"는 요구가 있다면 `llm.yaml` 쪽이 옳아진다 — 그 요구가 있는지는 내가 모른다.

---

## 6. 운영자 결정이 필요한 지점

내가 고를 수 없는 것 하나다.

> **`codex_task` 위임 경로에 모델/effort 설정을 여는가?**

- **연다면**: 실재하는 간극을 메우지만, `mcp_codex.go:217-220`의 기록된 설계 결정("설정 파일 핀이 위임 task로 새어서는 안 된다")을 뒤집는 것이다. 그 결정을 내린 SPEC과 정합하게 무효화 근거를 남겨야 한다.
- **열지 않는다면**: 운영자 지시는 **이미 충족돼 있다** — `workflow.audit.codex.{model,effort}`가 그것이고, 남는 일은 그것을 `moai web`에서 편집 가능하게 만드는 것뿐이다(A2, Tier S~M). 카드가 훨씬 작아진다.

내 관측으로는 지시 원문("각 모델별 설정 할 수 있도록")이 **감사/위임을 구분하지 않으므로**, 지시만으로는 어느 쪽인지 결정되지 않는다.

---

## 7. 판정서

### Claim
1. `llm.yaml`은 `.moai/config/` 중 **유일하게** gitignore된 파일이다(35 추적 / 1 제외). 사용자에게는 템플릿으로만 간다.
2. codex 모델/effort 설정은 **이미 존재한다** — `workflow.audit.codex.{model,effort}`, 추적되는 `workflow.yaml`에, 빈 값으로 배포.
3. 우선순위 규칙도 **이미 존재한다** — 호출자 > `workflow.audit.codex` 핀 > `llm.yaml` `profiles.<active>.sync-auditor` > 빈 값.
4. `claude_models`와 `glm`은 **이미 서로 다른 모양**이다. "같은 모양으로"라는 요구는 지시 대상이 없다.
5. 로더에 블록별 분기는 **없다.** 블록 추가로 분기가 늘지 않는다.
6. 착지한 판정(SPEC-V3R6-AUDIT-MODEL-PIN-001, lead ruling C)이 **`llm.yaml`을 이미 기각했다**.
7. task 경로의 설정 부재는 **기록된 의도적 배제**이지 누락이 아니다.
8. `moai web`은 `llm.yaml`과 `workflow.yaml` 둘 다 이미 편집한다. 다만 다중 파일 저장이 **원자적이지 않고**(문서화된 부분 저장), `llm.yaml`은 dirty 게이팅이 없어 **무관한 저장에도 재작성**된다.

### Evidence
전 항목이 `file:line`으로 인용돼 있다(§1-§4). 부재 주장 2건은 대조군을 동반한다: 템플릿 codex 0히트 ↔ glm 28히트 · `.moai/config` 추적 35 ↔ `llm.yaml` 1건 제외(`git check-ignore -v` 출력).
서브에이전트 2건이 조사한 항목 중 **내가 직접 재확인한 것**: `audit_pin.go:26-28` 주석 원문 · `workflow.yaml:74-91` 템플릿 블록 · `codexAuditAgentKey = "sync-auditor"` · `.gitignore:200` · 템플릿/로컬 `llm.yaml` codex 히트 수. 재확인하지 않은 항목은 Gaps에 든다.

### Baseline-attribution
트리 `.claude/worktrees/t509`, HEAD `ace1c5440`, `origin/develop`과 델타 `0 0`. 모든 grep·파일 읽기는 이 트리에서 이번 런에 실행했다. 예외 1건: `.moai/config/sections/llm.yaml`은 이 트리에 **없어서**(gitignore) primary 체크아웃 사본을 읽었다 — §1.1에 그 사실과 이유를 적었고, 결론은 템플릿 쪽 재측정으로 뒷받침했다.

### Gaps — 관측하지 않은 것
1. **운영자 지시의 대상이 감사 경로인지 위임 경로인지** — 지시문이 구분하지 않는다. 판정하지 않았다(§6).
2. ~~`DeepMerge3Way`의 비테스트 호출부를 찾지 못했다~~ — **닫았다**(§3.1.1). 직접 호출은 테스트뿐이었고, 프로덕션 경로는 래퍼 `MergeYAML3WayRetained`를 거친다(`restore.go:160`). 첫 grep이 래퍼 이름을 안 찾아 공백처럼 보였던 것이다. 남은 미관측은 #3으로 옮긴다.
3. **`moai update`를 실제로 돌려보지 않았다.** §3.1.1의 보존 주장은 **코드 경로 추적 + 회귀 테스트 존재 확인**이지 실행 관측이 아니다. 특히 `llm.yaml`이 `0600`이라는 점이 백업·복원에서 어떻게 다뤄지는지(권한 보존, 읽기 실패 시 동작)는 재지 않았다.
4. **`moai web`을 띄워보지 않았다.** 4절은 전부 코드 판독이다. 부분 저장·dirty 게이팅 부재는 실행으로 재현하지 않았다.
5. **서브에이전트 보고 중 미재확인 항목**: 라우트 표 전체, 테스트 파일 목록, `GLMTierEffort` 정확한 행 범위, seam 섹션 목록의 stale 여부(`SeamSections()`가 2개만 반환한다는 관측).
6. **16개 프로그래밍 언어 중립성**을 실제로 검사하지 않았다. A3의 몫이다.
7. **t225·t239 판정서 본문**을 읽지 않았다. 카드 서술과 코드 주석으로만 확인했다.

### Residual-risk
- **§3.1.1이 착지한 주석 하나를 낡았다고 판정한다.** 그 판정은 코드 경로 추적에 기대고 실행 관측이 아니다(Gaps #3). `moai update`를 격리 랩에서 한 번 돌려 `llm.yaml`의 사용자 키가 살아남는지 보는 것이 A0 결정 전 가장 값싼 검증이며, **그것이 이 판정서에서 가장 뒤집히기 쉬운 문장이다.**
- 낡았다고 판정한 것은 주석의 **사유 절반**이지 그 결정 자체가 아니다. SPEC-V3R6-AUDIT-MODEL-PIN-001이 다른 근거(예: 비밀값 동거, 리뷰 가능성)를 함께 고려했다면 결정은 여전히 옳을 수 있다 — 나는 그 SPEC 본문을 읽지 않았다(Gaps #7).
- §5의 Tier 추정과 자리 비교표의 "우세" 판단은 **판단이지 측정이 아니다**.
- 이 판정서는 코드를 읽고 썼다. `moai update`도 `moai web`도 실행하지 않았다.

---

## 8. 운영자 결정 이후 — 고른 범위는 **이미 구현돼 있다**

§6의 상신에 운영자가 답했다: **「감사 경로만 — web에 노출」**(2026-09-07).

그 범위를 코드에서 재봤다. **만들 것이 없다. 전부 이미 있다.**

### 8.1 네 필드가 이미 web console에 선언돼 있다

`internal/settings/schema_sections.go:415-421`:

```go
s(SectionWorkflow, "workflow", TypeText, "workflow", "audit", "codex", "model"),
withEmptySubmits(withSelect(s(SectionWorkflow, "workflow", TypeSelect, "workflow", "audit", "codex", "effort"),
    "f.workflow.audit.codex.effort.opt.", v4EffortValues(), emptyLabelUnset, "opt.unset")),
withEmptySubmits(withSelect(s(SectionWorkflow, "workflow", TypeSelect, "workflow", "audit", "glm", "model"),
    "f.workflow.audit.glm.model.opt.", config.ValidGLMModels(), emptyLabelUnset, "opt.unset")),
withEmptySubmits(withSelect(s(SectionWorkflow, "workflow", TypeSelect, "workflow", "audit", "glm", "effort"),
    "f.workflow.audit.glm.effort.opt.", template.GLMReasoningStateNames(), emptyLabelUnset, "opt.unset")),
```

- `codex.model` — 자유 입력 텍스트(codex가 서빙 가능한 id, 예: `gpt-*`)
- `codex.effort` — 닫힌 선택(v4 effort 어휘), **비우면 핀 해제**가 저장된다(`withEmptySubmits`)
- `glm.model` / `glm.effort` — 각각 `ValidGLMModels()` / z.ai reasoning state 이름

출처는 그 위 주석(`:403-404`): **SPEC-V3R6-AUDIT-MODEL-PIN-001 M4 (REQ-AMP-009 / AC-AMP-008)**.

### 8.2 전용 Audit 탭이 이미 있고, 라우팅도 이미 있다

- 탭 등록: `internal/web/schemaform.go:52` — `{ID: "audit", LabelKey: "tab.audit.title", Baseline: "Audit"}`. 주석(`:49`)이 이유를 적는다 — "audit (M2): workflow.audit.* moved off the workflow tab onto its own."
- 패널: `schemaform.go:238-241` — `PanelID: "audit"`, `Icon: "check-circle"`, `Title: "Audit"`.
- 필드 배치: `isAuditFieldName`(`:169-170`)이 `workflow.audit.` 접두를 보고 그 탭으로 보낸다. `partitionWorkflowFields`(`:185-199`)가 workflow 필드를 rest / worktree / audit 셋으로 가른다.

### 8.3 죽은 코드가 아니다 — 대조군

이 경로를 검증하는 파일이 실재한다:

```
$ git grep -ln 'workflow.audit.codex' -- internal/
internal/cli/audit_pin_live_test.go
internal/cli/mcp_codex.go
internal/cli/mcp_codex_audit_pin_test.go
internal/cli/mcp_convergence.go
internal/config/testdata/shipped_key_inventory.yaml
internal/core/project/initializer_audit_test.go
internal/settings/audit_pin_fields_test.go
internal/settings/schema_sections.go
internal/web/assets/i18n.js
internal/web/widget_policy_test.go
```

`audit_pin_live_test.go`(라이브 테스트), `audit_pin_fields_test.go`(필드 선언), `widget_policy_test.go`(위젯 정책), 그리고 `i18n.js`에 번역 문자열까지 있다. 배포 키 인벤토리(`shipped_key_inventory.yaml`)에도 올라 있다.

### 8.4 저장 경로도 이미 옳다

주석(`:405-407`)이 적는다 — 이 필드들은 "persisted through the same workflow.yaml seam the audit resolvers read". 즉 **web이 쓰는 자리와 리졸버가 읽는 자리가 같다.** §4.1에서 확인한 대로 `workflow.yaml`은 seam 경로라 주석과 미모델링 키가 보존된다.

그리고 주석(`:413-414`)이 한 가지를 더 구분한다: "Unlike the llm tier effort map (stored-only, REQ-WCR-033), these efforts **ARE runtime-applied** — they ride the audit request builders." 즉 이 설정은 저장만 되는 장식이 아니라 실제로 적용된다.

### 8.5 그래서 이 카드의 결론

**운영자가 고른 범위에서 이 카드는 만들 것이 없다.** 요청은 이미 충족돼 있고, 남는 일은 **어디를 눌러야 하는지 알리는 것**뿐이다:

> `moai web` → **Audit 탭** → `codex.model` / `codex.effort` (그리고 `glm.model` / `glm.effort`)

**다만 한 가지를 되짚어야 한다.** 운영자는 §6에서 「감사 경로만」을 골랐을 때 **그것이 만들어야 할 일이라고 알고** 골랐다. 이미 있다는 사실을 알았다면 다른 답을 골랐을 수 있다 — 특히 원래 지시("각 모델별 설정")가 겨냥한 것이 §3.3의 **위임 경로**였다면, 그 간극은 여전히 열려 있고 이 카드는 아직 답하지 않았다. 그 재확인이 이 카드의 마지막 남은 행위다.

### 8.6 실행으로 확인한 것 — 선언에 그치지 않는다

§8.1-8.4는 코드 판독이다. 「선언이 있다」와 「동작한다」는 다르므로, 이 축을 지키는 테스트를 **실제로 돌렸다**:

```
$ go test ./internal/settings/ -run Audit -count=1 -v
--- PASS: TestAuditPinFields_ExistWithTypeAndPanel (0.00s)
--- PASS: TestAuditPinFields_SeamRoundTrip (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/settings	0.480s
rc=0
```

- `ExistWithTypeAndPanel` — 네 필드가 **올바른 타입과 패널로** 선언돼 있음을 지킨다.
- `SeamRoundTrip` — **seam 경로 왕복**, 즉 저장한 값이 다시 읽히는 것을 지킨다. §8.4의 "web이 쓰는 자리와 리졸버가 읽는 자리가 같다"를 실행으로 뒷받침한다.

**공허한 초록이 아니다**: `=== RUN` 개수가 **2**다(0이 아니다). `-run` 필터가 아무것도 안 잡았다면 테스트 0개로 `ok`가 났을 것이고, 그것이 이 패턴의 전형적 위장이다.

### 8.7 그럼에도 안 본 것

- **`moai web`을 띄워 눈으로 보지 않았다.** 렌더가 실제로 되는지는 여전히 실행 관측이 아니다 — 위 테스트는 필드 선언과 seam 왕복을 재지, HTML 렌더를 재지 않는다.
- **`internal/cli/audit_pin_live_test.go`는 돌리지 않았다.** 이름상 라이브 테스트라 `codex` 바이너리를 요구할 수 있고, `internal/cli` 는 이 리포에서 600초 하한이 걸린 패키지다. 존재만 확인했고 내용을 읽지 않았다.
- 가장 값싼 남은 확정: `moai web`을 띄워 Audit 탭에서 `codex.model`에 값을 넣고 저장한 뒤 `workflow.yaml`에 남는지 본다. 원하면 하겠다.
