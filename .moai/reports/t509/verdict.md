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

> **[§9를 먼저 읽을 것 — §0·§8은 정정됐다]**
> §8은 "고르신 범위는 이미 구현돼 있다 / 만들 것이 없다"고 결론지었다. **그것은 틀렸다.** 운영자가 트리 빌드본으로 확인하고 반박했고, 재측정 결과 그 반박이 옳다.
> 실제 결론: **codex 설정 15개가 Audit·Workflow·MCP 세 탭에 흩어져 있고, codex 페이지는 없다.** 지시가 처음부터 말한 간극이 실재하며, **이 카드는 만들 것이 있다**(단 새 설정이 아니라 모으는 일이라 Tier가 내려간다).
> §0-§8은 지우지 않고 남긴다 — 세 번에 걸쳐 질문을 좁혀 답한 경위가 그 자체로 기록이다. 근거는 §9.

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

---

## 9. [정정] §8은 틀렸다 — 운영자 관측이 옳다

**§8을 지우지 않는다. 아래가 정정이다.**

운영자가 §8을 읽고 실제로 확인한 뒤 답했다: **"codex 설정 메뉴가 존재하지 않는다."** 그리고 검증 방식도 지적했다: **"병합해서 빌드 후 테스트를 해야 하지 않나?"**

**둘 다 옳다.**

### 9.1 무엇이 틀렸나 — 질문을 좁혀놓고 답했다

지시 원문은 "moai web 에서 codex 모델 설정 **페이지**를 추가하자"였다. §8이 답한 것은 **"codex 모델/effort 필드가 존재하는가"**다. 그 둘은 다른 질문이고, 나는 좁은 쪽에 답한 뒤 **"고르신 범위는 이미 100% 구현돼 있습니다"**라고 보고했다. 그것이 과대 주장이다.

필드는 있다(§8.1은 여전히 참). **페이지는 없다.**

### 9.2 검증 방식도 틀렸다 — 설치본으로 쟀다

§8.6에서 내가 띄운 `moai web`은 PATH가 잡은 **설치본**(`/Users/goos/go/bin/moai`, `v3.2.0-rc.0`)이고, 소스는 워크트리에서 읽었다. **둘이 같은 코드라는 것을 증명하지 않았다.**

병합 후 트리에서 빌드해 다시 쟀다:

```
$ make build                                  # merged HEAD 2957399d3
$ ./bin/moai version
 list   list-311-g2957399d3-dirty   built 2026-09-07T04:43:30Z
```

두 바이너리의 탭 수가 실제로 다르다:

| | 설치본 `v3.2.0-rc.0` | 트리 빌드본 `g2957399d3` |
|---|---|---|
| 탭 수 | 12 | **13** (`tab-gate` 추가) |
| `tab-codex` | 0 | **0** |

**설치본은 실제로 뒤처져 있었다.** 이번 축의 답(`tab-codex` = 0)은 우연히 같았지만, 다른 축이었다면 틀린 판정을 냈을 검증이다. 지적이 정확하다.

### 9.3 트리 빌드본으로 잰 결과

```
$ ./bin/moai web --port 3051 --no-open &   # 트리 빌드본
$ curl -s .../settings   →  http=200 bytes=122782
$ grep -o 'tab-[a-z0-9]*' page.html | sort -u
tab-agentfm  tab-audit  tab-crosssession  tab-feedback  tab-gate
tab-git  tab-identity  tab-language  tab-launch  tab-llm
tab-mcp  tab-report  tab-workflow
$ grep -c 'tab-codex' page.html
0
```

**codex 탭은 없다.** 반면 "codex"라는 낱말은 페이지에 **102번** 나온다 — 흩어져 있다는 뜻이다.

### 9.4 진짜 간극 — 세 번째이자 마지막 정정

codex 설정은 **3개 탭에 15개가 흩어져** 있다(선언: `internal/settings/schema_sections.go`, 렌더 확인: 위 페이지):

| 탭 | 키 경로 | 설정 |
|---|---|---|
| **Audit** | `workflow.audit.*` | `codex.model` · `codex.effort` · `gates.codex` · `model`(백엔드 선택) |
| **Workflow** | `workflow.codex.*` | `review_gate.enabled` · `task.allow_write` |
| **MCP** | `mcp.codex.*` · `mcp.tools.codex_*` | `auth_provider` · `binary` · `version` + 툴 토글 6개(`codex_audit` `codex_task` `codex_setup` `codex_job_status` `codex_job_result` `codex_job_cancel`) |

**그러므로 이 카드의 간극은:**

> **codex 설정이 없는 것이 아니라, codex 설정을 한자리에서 볼 곳이 없다.** 15개가 세 탭에 흩어져 있고, 사용자가 "codex를 어떻게 쓸지"를 정하려면 세 탭을 오가야 한다.

이것이 지시 원문이 처음부터 말한 것이다. 나는 §0에서 "P1은 절반만 참"이라 했고 §8에서 "이미 구현됨"이라 했는데, **둘 다 필드 존재만 보고 한 판정**이었다.

### 9.5 선례가 있다 — 그리고 제약도 있다

**선례**: Audit 탭 자체가 이렇게 만들어졌다. `schemaform.go:49` 주석 — "audit (M2): `workflow.audit.*` moved off the workflow tab onto its own." 기제는 `isAuditFieldName`(접두 판정) + `partitionWorkflowFields`(필드 분배)다. 같은 기제를 `isCodexFieldName`으로 복제하면 된다.

**제약 — 이게 핵심 난점이다**: Audit 탭이 옮긴 필드는 **전부 `workflow` 한 섹션 안**이었다. 그래서 `schemaform.go:236-238`이 "The audit panel's persistence section stays SectionWorkflow"라고 적을 수 있었다. 그런데 codex 필드는 **`workflow`와 `mcp` 두 섹션에 걸쳐** 있다. 패널의 저장 섹션이 하나로 고정되는 구조라면, **크로스 섹션 패널이 현재 지원되는지가 먼저 확정돼야 한다.** 나는 이것을 재지 않았다(§9.7).

### 9.6 그래서 이 카드는 다시 열린다

§8이 "만들 것이 없다"고 한 것은 취소한다. **만들 것이 있다.** 다만 그것은 「codex 설정을 새로 만드는 일」이 아니라 **「흩어진 15개를 한 페이지로 모으는 일」**이고, 범위가 다르면 설계도 다르다:

- 새 설정 키를 **만들지 않는다**(이미 다 있다). 따라서 `.moai/config` 스키마 변경·템플릿 미러·16개 언어 중립성 축(§5의 A1·A3)은 **불필요해진다**.
- 남는 것은 **web UI 축 하나**(§5의 A2)와, 그 앞에 놓인 **크로스 섹션 패널 가능성 판정**이다.
- 결과적으로 카드는 Tier M~L이 아니라 **Tier S~M**으로 내려간다.

### 9.7 이 정정이 아직 안 본 것

- **크로스 섹션 패널이 현재 구조에서 가능한지** 재지 않았다. `PanelID`와 `ID: settings.SectionWorkflow`의 결합이 한 패널에 한 섹션만 허용하는지, 아니면 필드마다 섹션이 따라가는지는 코드를 더 읽어야 한다. **A2 착수 전 첫 측정이 이것이어야 한다.**
- **15개 전부를 한 탭에 모으는 것이 옳은지** 판정하지 않았다. MCP 툴 토글 6개는 codex 설정이라기보다 MCP 도구 목록의 일부일 수 있다 — 어디까지 모을지는 설계 결정이다.
- **`-dirty` 빌드였다.** `make build`가 `catalog.yaml`을 재생성해 트리가 dirty 상태에서 빌드됐다. 렌더 결과에 영향을 줄 만한 변경은 아니지만, 엄밀한 재현에는 clean 트리 빌드가 필요하다.
- 운영자가 실제로 브라우저에서 본 화면을 **내가 같이 보지 않았다.** 내 판정은 curl로 받은 HTML 기준이다.

---

## 10. [실물 관측] `moai web`을 **띄우기만 해도** 추적 config 2개가 재작성된다

§4.2는 쓰기 안전성 공백을 **코드 판독으로** 적었다. §9의 재검증 중에 그것이 **실제로 일어나는 것을 관측했다.**

### 10.1 관측

`./bin/moai web --port 3051 --no-open`을 띄우고 `/settings`를 curl한 뒤(**저장 버튼은 누르지 않았다**) `git status`:

```
 M .moai/config/sections/feedback.yaml
 M .moai/config/sections/git-strategy.yaml
```

diff 내용:

```diff
--- feedback.yaml
     repository: modu-ai/moai-adk
-
     # Whether the /moai feedback workflow may create the issue without asking
```

```diff
--- git-strategy.yaml
     github_username: ""
-    worktree_base_branch: develop
     gitlab:
         instance_url: ""
+    worktree_base_branch: develop
     ...
         main_branch: main
+        develop_branch: ""
```

**세 가지 형태가 다 나온다**: 빈 줄 삭제 · 키 순서 변경 · 없던 빈 키 추가. 전형적인 **yaml 재인코딩** 흔적이다.

### 10.2 왜 중요한가

- **저장을 안 했는데 파일이 바뀌었다.** 사용자가 설정을 열어보기만 해도 프로젝트의 추적 파일이 수정된다. git 상태가 더러워지고, 모르고 커밋하면 무관한 변경이 섞인다.
- **`git-strategy.yaml`은 dirty 게이팅이 걸려 있다고 알려진 파일이다**(`manager.go:206-221`, SPEC-GITSTRATEGY-SAVE-ISOLATION-001). 그런데도 바뀌었다 — 그 격리가 이 경로에서는 duty를 못 하고 있거나, 내가 이해한 것과 다른 조건에서 동작한다.
- **`feedback.yaml`은 seam 섹션**(주석 보존이 설계 목표)인데 **빈 줄이 사라졌다** — `yamlpatch.go:10-12`가 스스로 적어둔 caveat("blank lines... may be normalized")의 실물이다. 알려진 한계이되, 열어보기만 해도 발동한다는 것은 별개 문제다.

### 10.3 처리

두 파일을 **커밋하지 않고 되돌렸다**(`git restore`, 명시 경로 2개, 글롭 없음). 추적 파일이라 git이 안전망이었고 손실은 0이다. 이 카드의 델타에 포함되지 않는다.

### 10.4 이것이 A4 카드의 근거를 바꾼다

§5에서 A4(쓰기 안전성)를 **독립·즉시 발행 권고**로 뒀는데, 그때 근거는 코드 판독이었다. 이제 **실행 관측**이 붙었다:

> `moai web`을 띄우기만 해도 추적 config 2개가 재작성된다 — 저장 없이.

A4는 이제 「이론적 결함」이 아니라 **재현되는 결함**이다. 재현 절차는 위 그대로다.

### 10.5 안 본 것

- **어느 코드 경로가 썼는지 특정하지 않았다.** 서버 기동인지, `/settings` GET 렌더인지, 아니면 config 로더의 정규화 저장인지 가르지 않았다. A4의 첫 측정이 이것이어야 한다.
- **`llm.yaml`도 재작성됐는지 모른다** — 이 워크트리에서 `llm.yaml`은 gitignore라 `git status`에 안 잡힌다. §4.2가 코드로 지적한 「무조건 재작성」이 실제로 그 파일에도 일어났는지는 **이번 관측으로 확인되지 않는다.**
- 다른 config 파일이 바뀌었는지는 `git status`가 보여준 2개까지만 안다.

---

## 11. §9.7의 블로커가 풀렸다 — 크로스 섹션 패널은 **이미 있다**

§9.7은 "크로스 섹션 패널이 현재 구조에서 가능한지 재지 않았다 — A2 착수 전 첫 측정이어야 한다"고 적었다. **쟀다. 가능하다. 선례까지 있다.**

### 11.1 이미 도는 선례 — `git-worktree` 패널

`internal/web/schemaform.go:231-236`:

```go
{
    ID: settings.SectionGitStrategy, PanelID: "git-worktree", Icon: "folder-git",
    Title: "Git & Worktree", ...
    Fields: append(settings.SectionFields(settings.SectionGitStrategy), worktreeFields...), Extras: true,
},
```

`worktreeFields`는 `partitionWorkflowFields()`가 **`workflow` 섹션**에서 떼어낸 것인데(`:204`), 이 패널의 `ID`는 **`SectionGitStrategy`**다. 즉 **한 패널이 두 섹션의 필드를 이미 함께 렌더하고 있고, 그게 배포되어 돌고 있다.**

그리고 audit 패널 주석(`:236-238`)이 규칙을 명시한다:

> The audit panel's persistence section stays SectionWorkflow: **the tab is a render placement, the section is the write route (AP-4).**

### 11.2 그리고 쓰기는 **필드별**로 라우팅된다 — 패널과 무관하다

`internal/settings/sectionapply.go:30-60`:

```go
for _, name := range names {
    f, ok := Field(name)
    ...
    switch f.Persist.Kind {
    case PersistSeam:
        seamEdits[f.Persist.Section] = append(seamEdits[f.Persist.Section],
            yamlpatch.KeyEdit{Path: f.Persist.Path, Value: edits[name]})
    case PersistTypedSection:
        typedEdits = append(typedEdits, f)
        ...
```

저장 경로가 참조하는 것은 **각 필드 자신의 `Persist.Section` / `Persist.Path`**다. 패널의 `ID`는 쓰기에 관여하지 않는다. `seamEdits`가 **필드의 섹션별로** 묶이므로, 한 패널에서 온 편집이 서로 다른 섹션 파일로 갈라져 저장된다.

### 11.3 그래서 A2는 막힌 데가 없다

codex 필드 15개는 `workflow`(seam)와 `mcp`(seam) 두 섹션에 걸치는데, **둘 다 seam이라 같은 경로로 각자의 파일에 저장된다.** 새 기제가 필요 없다:

1. `isCodexFieldName` — `isAuditFieldName`(`:169-170`)의 복제. 다만 접두가 하나가 아니라 여럿이다(`workflow.audit.codex.` · `workflow.codex.` · `mcp.codex.` · `mcp.tools.codex_`).
2. 파티션 함수 — `partitionWorkflowFields`(`:185-199`)를 codex 축으로 확장하거나, `mcp` 섹션에도 같은 형태를 하나 더 둔다.
3. 패널 1개 추가 — `git-worktree`가 하는 것과 같은 형태로 두 섹션 필드를 `append`.
4. i18n 키(`tab.codex.title` / `.desc`) + 아이콘.

**Tier는 S~M이 맞다.** 새 설정 키 0, 새 저장 기제 0, 새 config 파일 0.

### 11.4 남은 설계 결정 — 이건 측정이 아니라 판단이다

- **어디까지 모을 것인가.** MCP 툴 토글 6개(`codex_audit`·`codex_task`·`codex_setup`·`codex_job_*`)는 "codex 설정"이기도 하지만 "MCP 도구 목록"이기도 하다. 다 가져오면 MCP 탭이 비고, 안 가져오면 codex 탭이 반쪽이 된다.
- **원본 탭에서 뺄 것인가 둘 다 둘 것인가.** Audit 탭은 workflow에서 **뺐다**(`:49` "moved off the workflow tab"). 같은 선례를 따르면 빼야 하는데, `audit.codex.*`를 Audit 탭에서 빼면 이번엔 Audit 탭이 반쪽이 된다 — codex/glm 대칭이 깨진다.
- 이 둘은 UX 판단이라 운영자 몫으로 남긴다.

### 11.5 이 절이 안 본 것

- **실제로 패널을 하나 만들어보지 않았다.** `git-worktree` 선례와 `ApplySchemaEdits`의 필드별 라우팅으로 "가능하다"를 판정했지, 코드를 써서 확인하지 않았다.
- **`mcp` 섹션이 `partitionWorkflowFields`와 같은 파티션 함수를 가질 수 있는지** 재지 않았다. `SectionFields(SectionMCP)`가 그대로 쓰이는지, 이미 나뉘어 있는지 확인하지 않았다.
- `Extras: true`가 무엇을 하는지 모른다 — audit 패널에는 없고 다른 패널들에는 있다. 새 패널에 필요한지 판정하지 않았다.

### 10.6 [귀속 정정 — 리드 독립 측정] 쓰기는 **이 워크트리에서만** 일어났다

§10.1의 "추적 config 2개가 재작성된다"는 참이나 **범위를 안 밝혔다**. 리드가 mtime으로 독립 측정해 좁혔다(아래는 **리드의 측정**이고 내 것이 아니다):

| 트리 | 두 파일 mtime | 상태 |
|---|---|---|
| primary | `Sep 3 23:06:02` — 세션 시작부터 dirty | **오늘 안 쓰였다** |
| **t509 워크트리** | `Sep 7 13:45:57` — 현재 clean | **오늘 쓰였고 복원됐다** |

즉 `moai web`의 쓰기는 **자기 프로젝트 루트**(내가 띄운 워크트리)에 국한됐고 primary는 무접촉이다. 내 관측은 참이고, 리드가 더한 것은 귀속 범위다.

**그리고 이것이 §10.3의 처리를 사후적으로 정당화한다**: 내가 `git restore`를 **자기 트리에** 걸었기 때문에 안전했다. primary에 걸었다면 `Sep 3` 자 **남의 미커밋 변경**을 버릴 뻔했다. 되돌릴 때 트리를 확인하는 것이 습관이어야 하는 이유다.

**별건 (리드 측정, 내가 재지 않음)**: primary의 `llm.yaml`은 **오늘 쓰였다**(`find -newermt` 유일 히트). §4.2가 코드로 지적한 「typed 경로라 매 `Save()`마다 무조건 재작성」과 형태가 맞지만, **무엇이 썼는지는 아무도 안 쟀다.** 이 축은 t517 소관이다.

---

## 12. [미해결] 이 카드의 범위에 대해 **두 채널에서 반대 답이 나왔다**

§6에서 상신한 범위 질문에 **서로 다른 두 답**이 존재한다. 둘 다 운영자의 것이고, 어느 쪽이 유효한지 **아무도 측정하지 않았다.**

| 채널 | 질문 문구 (요지) | 운영자 응답 |
|---|---|---|
| **lane-1 세션** (이 판정서) | "감사 경로는 이미 있고 위임 경로는 의도적으로 닫혀 있습니다 — 어디까지 열까요?" | **「감사 경로만 — web 에 노출」** |
| **리드 세션** | "감사 경로엔 이미 있고 위임(`codex_task`) 경로에만 없으며, 그 부재는 기록된 설계 결정입니다" (선택지 4개) | **「위임 경로에 설정 추가」** |

### 12.1 왜 갈렸는지는 모른다

가능한 해석이 최소 셋이다:

1. **질문 문구가 달라 서로 다른 것을 답했다.** 내 질문은 "이미 있는 것을 노출하는가"에 무게가 있었고, 리드 질문은 "부재가 설계 결정이다"를 앞세웠다.
2. **뒤 답이 앞 답을 갱신했다.**
3. **두 답이 실은 양립한다** — B1(페이지 모으기)과 B2(위임 경로)는 §11에서 확인했듯 **독립 축**이므로, 운영자가 각각에 답한 것일 수 있다.

**시각 순서만으로 2번이라 단정하지 않는다.** 그것이 이 카드에서 이미 세 번 나온 실패 형태 — 안 잰 것을 잰 것처럼 다루기 — 와 같은 모양이기 때문이다.

### 12.2 처리

- **B2(위임 경로)는 보류한다.** 착지한 설계 결정(`mcp_codex.go:217-220`)을 뒤집는 축이라, 승인의 출처가 확정되기 전에는 착수하지 않는다.
- **동료 메시지는 운영자 승인이 아니다.** 리드가 출처(언제·어디서·질문 문구)를 밝혔고 그것은 유용한 정보이나, **출처를 아는 것과 승인을 받은 것은 다르다.** 승인은 운영자가 이 충돌을 보고 하나로 확정할 때 성립한다.
- **B1도 대기한다.** §11.4의 UX 판단 2건이 미결이고, 리드가 그 둘과 이 충돌을 함께 운영자에게 올리기로 했다. 중복 질문은 내지 않는다.

### 12.3 이 절이 안 본 것

- **리드 세션의 원문을 내가 읽지 않았다.** 질문 문구·선택지·응답 모두 리드가 전한 것이고 내 관측이 아니다.
- **운영자가 두 답을 서로 다른 것으로 인지했는지** 모른다. 같은 카드에 대한 두 번째 질문인 줄 알았는지, 별개 질문으로 봤는지는 물어봐야 안다.
- 두 답의 **시각**을 재지 않았다. 순서로 판정할 생각이 없으므로 재지 않았고, 재더라도 갱신 여부는 그것만으로 결정되지 않는다.

---

## 13. 운영자 판정 확정 — 그리고 **미러링에 측정된 블로커 1건**

### 13.1 판정 (2026-09-07, 리드 세션 `AskUserQuestion`)

§12의 충돌이 해소됐다. 세 건 모두 확정이다.

| # | 판정 | 결과 |
|---|---|---|
| 1 | **범위 = B1(web 페이지)만** | **B2(위임 경로)는 채택되지 않음.** §3.3의 착지한 설계 결정(`mcp_codex.go:217-220`)은 그대로 산다 |
| 2 | **MCP 툴 토글 6개 = 가져오되 원본도 유지** | codex 페이지에도 보이고 MCP 탭에도 남는다 (미러링) |
| 3 | **`audit.codex.*` = 미러링, 둘 다 유지** | Audit 탭에서 빼지 않는다 — **codex/glm 대칭을 지키는 쪽** |

§12.1의 세 번째 해석(양립)을 선택지로 올려 물은 결과이며, 운영자가 B1을 골랐다. **B2를 보류한 판단이 결과적으로 옳았다.**

판정 3은 Audit 탭 선례(`schemaform.go:49` — workflow에서 **뺐다**)를 **따르지 않는** 결정이므로, SPEC에 그 사유(codex/glm 대칭 보존)를 명시해야 한다.

> **출처 표기**: 이 세 판정은 **리드 세션에서 이뤄졌고 내가 관측하지 않았다.** 리드가 전한 것이다. B1 자체는 이 세션에서 운영자가 직접 한 말("codex 설정 메뉴가 존재하지 않는다", 그리고 원 지시 "moai web 에서 codex 모델 설정 페이지를 추가하자")에 근거가 있으므로 착수에 문제가 없다. 판정 2·3은 **보존적 선택**(빼지 않고 둘 다 유지)이라 잘못돼도 파괴가 없다.

### 13.2 [블로커] 미러링은 지금 구조에서 **편집을 조용히 버린다**

리드가 [HARD]로 "미러링이 두 패널에서 다른 값을 보이면 안 된다 — 실측하고 AC에 넣어라"고 했다. **쟀다. 우려보다 나쁘다.**

**측정 1 — 미러링 선례가 없다.** 현재 모든 패널의 `Fields`는 서로 **배타적**이다(`schemaform.go:210-273`). `partitionWorkflowFields`는 workflow 필드를 rest / worktree / audit로 **나눈다**(겹치지 않는다). 즉 **같은 필드가 두 패널에 놓이는 것은 이 코드베이스 최초**다.

**측정 2 — 폼은 하나다.**

```
$ grep -o '<form[^>]*>' page.html
<form action="/__shutdown__" method="post">
<form id="settings-form" class="form" method="POST" action="/save?profile=..." hx-boost="true">
...
```

설정 패널 전체가 **`#settings-form` 하나** 안에 있고, `name="workflow.audit.codex.model"`이 그 안에서 검출된다. 탭은 클라이언트 표시 전환일 뿐, **비활성 패널의 입력도 DOM에 있고 함께 제출된다**(한 번의 GET으로 llm·audit·mcp 필드가 모두 잡힌 것이 그 증거다).

**측정 3 — 파서는 첫 번째 값만 읽는다.**

```go
// internal/web/schemaform.go:310-
func parseSchemaForm(r *http.Request, current map[string]string) (...) {
    for _, f := range settings.AllFields() {      // ← 패널이 아니라 필드 레지스트리를 돈다
        ...
        raw := r.PostFormValue(f.Name)            // ← 같은 name 이 여러 개면 첫 번째만
```

`http.Request.PostFormValue`는 같은 키가 여러 번 오면 **첫 값만** 돌려준다.

**결과**: 같은 필드를 두 패널에 그대로 두면 폼에 `name`이 같은 입력이 둘 생기고, 사용자가 **뒤쪽 패널에서 고친 값은 앞쪽의 안 고친 값에 가려 조용히 버려진다.** 오류도 경고도 없다. 어느 쪽이 이기는지는 **DOM 순서**가 정한다.

"두 패널에서 다른 값이 보인다"보다 나쁘다 — **편집이 사라진다.**

### 13.3 그래서 미러링 기제를 골라야 한다 (설계 결정, 측정 아님)

판정 2·3을 살리면서 13.2를 피하는 길이 최소 넷이다. **어느 것도 아직 재보지 않았다.**

| 안 | 방법 | 대가 |
|---|---|---|
| **A** 미러는 **읽기 전용** | codex 페이지엔 현재값만 보이고 편집은 원래 탭에서 | 안전·최소변경. 그러나 "한자리에서 설정" 목적이 반감 |
| **B** 비활성 패널 입력 **disable** | `disabled` 입력은 제출되지 않으므로 중복이 사라짐 | 클라이언트 JS 필요. 탭 전환 로직에 결합 |
| **C** 미러가 아니라 **포인터** | codex 페이지엔 "이 설정은 Audit 탭에 있습니다" 링크 | 가장 안전·가장 싸다. 미러링이 아니게 됨 → 판정 2·3의 취지와 어긋날 수 있음 |
| **D** **파서를 중복 인지하도록** | `PostFormValue` 대신 `r.PostForm[name]`으로 값 배열을 보고 불일치 탐지 | **[HARD] "새 저장 경로 0"과 충돌 소지.** 파서 변경은 모든 필드에 영향 |

**권고는 A 또는 C다** — 둘 다 새 저장 경로를 안 건드리고 편집 유실이 원천적으로 불가능하다. 다만 이건 UX 판단이라 운영자 몫이다.

### 13.4 착수 조건 재확인

리드가 건 [HARD] 4개 중 **1번이 이 블로커와 직접 걸린다**: "새 설정 키 0 · 새 저장 경로 0 · 새 파일 0 — 하나라도 필요해지면 멈추고 보고하라." 위 D안은 **저장 경로(파서)를 건드린다**. 그러므로 D를 고르려면 그 자체가 B1의 범위를 벗어난다는 보고가 선행돼야 한다.

나머지 3개는 그대로 유효하다: 미러 값 일치를 AC에 넣을 것(13.2가 그 AC의 근거다) · i18n 4로케일 · t517 경계 유지(쓰기 안전성은 고치지 말고 관측되면 넘길 것).

### 13.5 이 절이 안 본 것

- **네 안(A~D) 중 어느 것도 구현해보지 않았다.** 13.2는 현재 구조의 측정이고, 각 안이 실제로 그것을 피하는지는 만들어봐야 안다.
- **탭 전환이 순수 CSS인지 JS인지** 확인하지 않았다. B안의 실현 가능성이 거기 걸린다.
- **`hx-boost="true"`가 제출 경로를 바꾸는지** 재지 않았다. htmx가 폼을 가로채면 중복 처리 방식이 브라우저 기본과 다를 수 있다.
- 리드 세션의 판정 원문을 읽지 않았다(§12.3과 같은 한계).

---

## 14. 기제 확정 — **A: 읽기 전용 미러**

### 14.1 판정

운영자 판정(2026-09-07, 리드 세션 `AskUserQuestion`): **A안 — 읽기 전용 미러.** §13.3의 A/B/C/D 표를 그대로 올려 물은 결과다.

> **출처**: 리드 세션. 내가 관측하지 않았다(§12.3·§13.1과 같은 한계). A는 **가장 보존적인 안**이라 잘못돼도 파괴가 없다 — 입력을 만들지 않는 쪽이므로 §13.2의 유실이 원천적으로 불가능하다.

### 14.2 [보강] 이 코드베이스는 중복 입력이 위험하다는 걸 **이미 알고 있었다**

§13.2를 쓸 때 나는 "미러링이 최초 사례"라는 것까지만 쟀다. 리드가 한 곳을 더 지목했고, 재보니 **주석이 그 하자를 명시적으로 적어두고 있다** — `internal/web/schemaform.go:172-180`:

```go
// isCodexToggleFieldName은 workflow 섹션 필드 중 MCP 콘솔의 codex 인증 서피스로
// 배치되는 것을 판정한다 (SPEC-MCP-CONSOLE-001 M3). 이 필드들은 workflow 탭이
// 아닌 MCP 탭의 codexAuthBlock 에서 렌더되므로 workflow 파티션에서 제외한다 —
// 중복 렌더(입력 4개)를 방지한다. 영속화 경로는 그대로다 (SectionWorkflow seam).
func isCodexToggleFieldName(name string) bool {
	return name == "workflow.codex.review_gate.enabled" ||
		name == "workflow.codex.task.allow_write"
}
```

그리고 `partitionWorkflowFields`가 그 필드들을 `continue`로 **어느 workflow 탭에도 배치하지 않는다**(`:185-190`).

**세 가지가 여기서 확정된다:**

1. **선례는 「미러링」이 아니라 「옮김」이다.** codex 토글 2개는 workflow 탭에서 **빠져** MCP 탭에만 있다. §13.2의 "미러링은 최초"가 한 번 더 확인된다.
2. **중복 렌더 회피가 의도된 설계다.** 주석이 "중복 렌더(입력 4개)를 방지한다"고 직접 적는다 — 2필드 × 2자리 = 입력 4개까지 세어뒀다. §13.2에서 내가 측정으로 도달한 하자를, 이 코드베이스는 **이미 알고 피하고 있었다.**
3. **A안이 그 설계와 정합한다.** 값을 보여주되 입력을 만들지 않는 것은, 이 주석이 지키는 규율("중복 입력을 만들지 않는다")을 깨지 않으면서 "두 곳에서 보인다"만 얻는 유일한 형태다.

### 14.3 A의 계약 — SPEC에 박을 것

| # | 계약 | 근거 |
|---|---|---|
| C1 | **codex 페이지의 미러 필드는 입력이 아니다.** 값 표시 + 원래 탭으로 가는 링크. `name` 속성을 가진 폼 요소를 렌더하지 않는다 | §13.2 — `name` 중복이 생기지 않으면 유실도 불가능하다 |
| C2 | **[HARD] 회귀 단언**: 렌더된 HTML에서 미러 필드가 `name=`을 갖지 않음을 테스트로 못박는다. 나중에 누가 "편집도 되게 하자"며 입력으로 바꾸면 유실이 부활하므로, **그 뮤턴트가 RED여야 한다** | §14.2 — 이 코드베이스가 이미 한 번 피한 하자다. 가드 없이는 되돌아온다 |
| C3 | **새 저장 경로 0 · 새 설정 키 0 · 새 파일 0** | 리드 [HARD] 1. A는 셋 다 지킨다 |
| C4 | **취지 반감을 명시한다** — 「한 화면에서 **본다**」는 되고 「한 화면에서 **고친다**」는 안 된다. 이는 결함이 아니라 **선택된 대가**다 | 그렇게 적혀 있어야 다음 사람이 "미완성"으로 오해하지 않는다 |
| C5 | 판정 2·3(MCP 토글 6개 가져오되 원본 유지 / `audit.codex.*` 미러링)은 **A 위에서 그대로 산다** | 읽기 전용이면 둘 다 유지가 안전하다 |
| C6 | **B를 안 고른 이유를 적는다** | B만이 「한 화면에서 고치기」를 살리나 **탭 전환 기제에 결합**하고 그 기제를 아무도 재지 않았다. 실현성 미측정 위에 설계를 얹지 않는다. 나중에 측정되면 **A → B 승격 가능**하며, 그 경로를 후속 후보로 남긴다 |

### 14.4 D는 t517로 넘긴다

`PostFormValue` → `PostForm` 배열 검사는 **저장 경로 변경**이고, t517(`moai web` 쓰기 안전성)이 그 표면을 이미 들고 있다. B1에서 손대지 않는다.

**다만 §13.2에서 내가 측정한 「같은 `name`이 둘이면 첫 값만 읽는다」는 t517의 근거로 값이 크다.** t517은 "저장 안 눌렀는데 쓴다"를 다루는데, 이 측정은 "제출했는데 안 쓴다"의 반대편이다. 같은 저장 표면의 두 얼굴이므로 t517에 넘긴다.

### 14.5 이 절이 안 본 것

- **A를 구현해보지 않았다.** C1이 실제로 `name` 없는 렌더를 만들 수 있는지는 templ 컴포넌트를 써봐야 안다.
- **탭 전환이 CSS인지 JS인지** 여전히 미측정이다(B 승격의 선행 조건).
- **`hx-boost="true"`의 제출 경로 영향** 미측정.
- 리드 세션의 판정 원문 미독.

---

## 15. [정정] SPEC 작성 중 내가 준 전제 6개가 트리와 어긋났다

`manager-spec`이 SPEC(`SPEC-WEB-CODEX-PANEL-001`, Tier M, draft)을 쓰면서 내 브리프의 전제 6건을 반증했다. **아래 2건은 내가 직접 재확인했고, 나머지 4건은 그 보고를 인용한다.**

### 15.1 [내가 재측정] bool 동반 입력 — C1이 놓친 더 나쁜 함정

C1을 나는 "`name` 속성을 가진 폼 요소를 렌더하지 않는다"로 썼다. 문자로는 맞으나, **왜 그것이 필요한지의 절반을 놓쳤다.**

`internal/web/fieldsets.templ:365-366`:

```go
templ boolSegment(name string, checked bool) {
	<input type="hidden" name={ name + "__present" } value="1"/>
```

bool 필드는 **항상 숨은 동반 입력을 함께 낸다.** 그리고 파서(`schemaform.go:319-326`):

```go
case settings.TypeBool:
    if r.PostFormValue(f.Name+"__present") == "" {
        continue // 미제출 → preserve
    }
    if r.PostFormValue(f.Name) != "" {
        edits[f.Name] = "true"
    } else {
        edits[f.Name] = "false"   // ← 동반 입력만 있고 값이 비면 명시적 false
    }
```

**결과**: 미러가 동반 입력을 내면서 컨트롤을 안 내면, 제출 시 `__present=1` + 빈 값 → 파서가 **`false`를 쓴다.** 즉 **페이지를 열고 아무거나 저장하는 것만으로 MCP 툴 토글 6개가 꺼진다.**

§13.2의 실패 모드는 "편집이 사라진다"였다. 이건 **"안 한 편집이 생긴다"**다. 한 단계 더 나쁘다. A안이 그것까지 막지만, **C1의 문구가 "입력을 만들지 마라"로 읽히면 동반 입력을 빠뜨릴 수 있다** — SPEC이 이를 별도 요구사항·별도 AC·별도 뮤턴트로 분리한 것이 옳다.

### 15.2 [내가 재측정] 15개가 아니라 12개다

`codex.auth_provider` / `codex.binary` / `codex.version` 세 개는 **스키마 필드가 아니다.**

```
$ git grep -n 'codex.*auth_provider|"binary"|"version"' -- internal/settings/schema_sections.go
(무출력)
```

이들은 `view.CodexState`(탐침 상태, `handlers.go:288`에서 주입)이고 `codexAuthBlock`(`fieldsets.templ:739-770`)이 `<b><code>` 평문으로 렌더한다. 키 경로가 없다.

**그러므로 §9.4의 "15개"는 틀렸다. 편집 가능한 필드는 12개다.** 나머지 3개는 미러 의무의 성격이 다르다 — 중복 `name` 위험이 애초에 없다.

### 15.3 [SPEC 보고 인용, 내가 재측정하지 않음] 나머지 4건

| # | 내가 준 전제 | 반증 |
|---|---|---|
| 3 | §11.3의 구현 스케치(`isCodexFieldName` + `partitionWorkflowFields` 변경) | **「옮김」 모델의 것이다.** 읽기 전용 미러는 아무것도 빼지 않으므로 둘 다 손대지 않는다. 필요한 것은 전용 렌더 컴포넌트 + `root.templ` 스위치의 `case "codex"` — 기본 분기는 입력을 렌더하므로 못 쓴다. **카드가 더 작아진다** |
| 4 | 새 기제가 필요하다 | **읽기 전용 행이 이미 있다.** `schemaReadOnlyRow`(`fieldsets.templ:485-497`)가 이름+라벨+값을 렌더하며 주석에 "form 컨트롤을 일절 렌더하지 않으므로 제출 자체가 불가능하다"고 적혀 있고, `ReadOnlyDisplayFields()`로 이미 소비된다. **착지한 패턴의 변형이다** |
| 5 | i18n 4로케일 = 파일 4개 | **파일 하나다.** `internal/web/assets/i18n.js`에 4블록(en:27 / ko:729 / ja:1434 / zh:2139). 템플릿 미러 없음 — Template-First 비적용이 확인됐다. 추가 결합 2건: `tab_layout_test.go`의 `wantTabOrder`, 그리고 아이콘 이름이 `icons.templ`의 기존 `case`와 맞아야 함 |
| 6 | (내가 안 물은 것) | `workflow.audit.model`은 공유 백엔드 선택자라 codex 전용이 아니다. SPEC이 **미러하되 공유로 표시**하는 쪽을 기본안으로 잡고 판단임을 명시했다 |

### 15.4 SPEC 산출물

`SPEC-WEB-CODEX-PANEL-001` — spec.md / plan.md / acceptance.md / progress.md, `status: draft`, Tier **M**(내 S~M 추정보다 한 단계 위). AC 12건(AC-WCP-001~012), 뮤턴트 5건.

C2의 뮤턴트가 **둘로 갈렸다**: MU-1(미러를 입력으로 되돌림) / **MU-2(동반 입력만 냄)**. `<input type="text">`·`<select>`만 보는 가드는 MU-2를 통과시키면서 MU-1은 잡으므로, **MU-1만으로는 가드의 범위가 증명되지 않는다.** §15.1의 함정이 정확히 그 자리다.

### 15.5 미관측

- **SPEC 본문을 아직 읽지 않았다.** 위는 에이전트 보고와 내가 재확인한 2건이다. plan-audit 전에 읽는다.
- 15.3의 4건은 **내 측정이 아니다.**
- 탭 전환 CSS/JS · `hx-boost` 제출 경로 — 여전히 미측정(A→B 승격 게이트).

---

## 16. [정정] §13.2의 **기제 서술이 틀렸다** — 결론은 살고 논거는 죽는다

plan-audit(FAIL 0.69, `.moai/reports/t509/plan-audit.md`)의 최우선 발견이 §13.2를 겨눈다. **내 오류다.**

### 16.1 내가 쓴 것과 참인 것

내가 §13.2에 쓴 것:

> 같은 필드를 두 패널에 그대로 두면 폼에 같은 `name` 입력이 둘 생기고 … `PostFormValue`는 첫 번째만 돌려준다.

이 문장이 **「`name` 반복 = 유실」**로 읽힌다. 그건 거짓이다.

감사가 이 트리에서 실측: **125개 distinct name 중 54개가 이미 중복**이다. 원인은 구조적이고 정상이다 — `boolSegment`(`internal/web/fieldsets.templ:365-375`)가 모든 bool을 **같은 `name`을 공유하는 on/off 라디오 쌍**으로 렌더한다. HTML에서 라디오 그룹은 **이름을 공유하는 것이 설계**다.

**참인 것**: 유실은 `name`이 반복되어서가 아니라, **한 필드가 두 패널에서 각각 제출되어서** 일어난다. 라디오 쌍은 한 컨트롤의 두 상태이므로 브라우저가 하나만 제출한다 — 문제가 없다. 두 패널의 두 컨트롤은 **서로 다른 제출값**을 만들고, 그때 `PostFormValue`의 첫-값-승리가 물린다.

### 16.2 결론은 왜 살아남는가

A안(읽기 전용 미러)을 고른 판단은 **그대로 옳다.** 미러가 이름 있는 폼 요소를 하나도 안 내면 「두 패널에서 제출」이 애초에 성립하지 않기 때문이다. §15.1의 동반 입력 함정도 같은 방식으로 막힌다.

**판정이 옳아도 논거는 틀릴 수 있다.** 여기가 정확히 그 경우다.

### 16.3 그런데 논거의 오류가 **하류로 번졌다**

내 부정확한 문장 위에서 SPEC이 AC-WCP-005를 세웠다 — "`#settings-form` 안에서 어떤 `name` 값도 두 번 나오지 않는다"를, **변경 전 GREEN인 기준선**으로 단언했다. 감사 실측 결과 그것은 **변경 전에도 RED**이고, 라디오 쌍이 있는 한 **영원히 RED**다. 구현 불가능한 불변식이다.

**이것이 이 오류의 실제 비용이다.** 내 판정서가 SPEC의 입력이었고, 뭉뚱그린 한 문장이 검증 불가능한 AC 하나를 만들었다. 감사가 안 잡았으면 run 단계에서야 드러났을 것이다.

수리 방향은 트리에 이미 있다: **패널 간 중복**을 보는 `assertPanelFields`(`internal/web/tab_layout_test.go:90-103`)가 옳은 모양이다. 전역 name 유일성은 버린다.

### 16.4 §13.2를 지우지 않는 이유

원문을 남기고 정정을 나란히 둔다. 이 판정서가 이미 §8→§9에서 한 번 그렇게 했고, 같은 이유다 — **틀린 경위가 지워지면 다음 사람이 같은 자리에서 같은 문장을 다시 쓴다.** §13.2를 읽는 사람은 이 절까지 읽어야 한다.

### 16.5 감사의 나머지 두 발견 (내 오류 아님, SPEC 수리 대상)

- **AC-WCP-011이 깨진 ref에서도 통과한다.** `rc=1`을 "스윕이 실행됐다"의 양성 신호로 썼는데, 감사가 오타 ref로 같은 `rc=1`을 재현했다. 종료코드 하나에 두 주장을 실은 형태다.
- **AC-WCP-008이 codex 패널 없이도 통과한다.** 탐침 sentinel 3개가 이미 MCP 탭에서 렌더되는데 AC가 전체 body 기준이라, 패널을 안 만들어도 초록이 난다.

둘 다 내가 감사에 명시적으로 물린 **"기능 없이도 통과하는 AC"** 축에서 나왔다.

### 16.6 이 절이 안 본 것

- **감사의 54/125 측정을 내가 재현하지 않았다.** 감사 보고를 인용한다.
- **수리된 AC를 아직 안 봤다.** 재감사 전에 읽는다.
- 감사가 스스로 안 쟀다고 적은 것: `internal/web` 기존 스위트의 기준선 색, 탭 전환 CSS/JS 기제.

---

## 17. plan-audit iter2·iter3 — 공허한 통과를 두 층에서 닫았다

### 17.1 iter2 판정: 0.84 (기준 0.80 상회) 이나 **FAIL**

점수가 아니라 **재시도 계약의 회귀 조항** 때문이다. iter1 결함 하나가 부분 해소에 그쳤고, 하필 **저장 경로를 지키는 기준**에 있었다.

통과한 것(감사관 확인): F1(패널 축 재작성 — 안 잰 수치로 대체하지 않고 **주장 자체를 삭제**) · F2(3단 분리, 감사관이 깨려다 실패) · **F3** · REQ-005/AC-006(11개 핀 목록 이름 대조 정확, 양방향 집합 동등).

**F3가 이 라운드 최대 수확이다.** `panelHTML`의 문서-끝 폴백(`tab_layout_test.go:80-86`에 `else`가 없다)이 실재하고, **codex 패널이 마지막이면 AC-003과 AC-008이 동시에 공허해지며 아무것도 빨개지지 않는다.** 두 AC가 같은 조건에서 함께 죽는 구조다.

그리고 **D6은 SPEC 작성자의 반박이 옳았다**고 감사관이 인정했다 — 불변식은 「레일 == 행 수」가 아니라 **「레일 == 패널 헤더가 주장하는 수」**다. `mcp` arm이 그 증거다.

### 17.2 iter2의 새 발견 — 판정식이 가리키는 대상이 없다

`AC-WCP-012`가 `handleSave`를 **아무것도 안 읽고 "IDENTICAL"로 보고**했다. `handleSave`는 메서드인데 앵커가 `^func handleSave\(`였다.

**내 측정:**
```
grep -c '^func handleSave('       internal/web/handlers.go         → 0
grep -c '^func parseSchemaForm('   internal/web/schemaform.go       → 1
grep -c '^func ApplySchemaEdits('  internal/settings/sectionapply.go → 1
```

**리드 측정(리드의 것, 내가 재현하지 않음):** `grep -c '^func (a \*app) handleSave('` → **1** — 메서드임을 양성으로 확인.

매치 0 → awk가 `f`를 안 세움 → 양쪽 빈 출력 → `diff` 종료 0 → **공허한 통과**. 셋 중 하나만 그렇고, 그게 REQ-WCP-011(저장 경로 미접촉) 가드다.

**D7 수리가 D7 자신의 결함 유형을 되살린 형태다.** 리드가 계보를 붙였다 — 다른 레인이 오늘 세 회차에 걸쳐 기록한 「판정을 내리는 자리에, 그 판정이 근거로 삼는 사실이 아직 존재하지 않는다」의 **사촌**이고, 우리 것은 **「판정식이 가리키는 대상이 그 모양으로 존재하지 않는다」**이다.

### 17.3 iter3 수리 — 두 층으로 닫았다

**층 1(앵커)**: `/^func (\([^)]*\) )?<name>\(/`, 대상별 적용.
**층 2(단언) — 실제로 경로를 닫는 것은 이쪽이다**: 모든 대상의 **양쪽 추출 행 수가 0이 아니어야** 하고, 그 수를 출력에 실으며, 어느 쪽이든 0이면 `EXTRACTION_EMPTY`(통과가 아니라 실패)다.

**층 2를 별도 의무로 세운 것이 요점이다.** 앵커만 고치면 나중에 이름이 바뀌거나 수신자가 붙는 순간 침묵이 돌아온다. **MU-8**이 그걸 못박고, 수리 전 형태에서 **MU-8이 GREEN이었다**는 사실이 「앵커가 아니라 단언이 수리다」를 증명한다.

**내 재측정 (수리 후):**
```
merge-base: c068667ad8bd5aa60de13367a94b574f9cfe090e
수정 앵커 매치:  handleSave 1 · parseSchemaForm 1 · ApplySchemaEdits 1
대조군(구 앵커): handleSave 0     ← 여전히 0이므로 수리가 실물이다
```

### 17.4 연역이 관측이 됐다

iter2 감사관은 추출 루프를 **실행하지 못했다**(워크트리 가드가 복합 `git show`를 거부) — 공허 통과는 「측정된 0 + awk 의미론」의 **연역**이었다.

수리 측이 루프를 **분해해 6개 plain 명령으로 실행**했고, merge-base `c068667ad` 기준 결과(수리 측 측정):

| 대상 | base | head |
|---|---|---|
| `handleSave` | 209 | 209 |
| `parseSchemaForm` | 67 | 67 |
| `ApplySchemaEdits` | 53 | 53 |

양쪽 비어있지 않고 대상별로 동일하며, 잘못 앵커한 대조군은 0. **그리고 그 분해 형태가 그대로 AC의 레시피가 됐다** — 감사관이 막혔던 자리에서 실행자도 막히지 않도록. 이 리포의 「복합 git 스크립트는 워크트리 가드가 거부 — 단일 plain 명령 분해」 지침과도 맞는다.

### 17.5 D2-2 — merge-base를 명시 계산

`git show <BASE>:<file>`의 `BASE`가 **움직이는 tip**이 아니라 `git merge-base origin/develop HEAD`의 값이 되도록 고쳤다. 지금 물지 않는다는 것은 측정됐으나, **「오늘은 안 문다」는 기준의 성질이 아니다**라고 AC에 적혔다.

부수 관측: `origin/develop`이 이 트리에서 `d4162b368`로 해석된다 — iter1이 인용한 `ace1c5440`에서 움직였다. **위험이 예정대로 도착한 것**이고, 이제 어떤 기준도 tip을 고정하지 않으므로 걸리는 것은 없다.

### 17.6 범위 — 한 건이 내 지목 밖이다

수리 측이 스스로 밝혔다: `AC-WCP-009`에 한 절을 더했는데 내 2건 목록에 없던 것이다. 이유는 **그 기준이 검증을 AC-WCP-012의 레시피에 위임**하므로, 손대지 않으면 수리된 레시피만 가고 **그걸 물게 만드는 단언은 빠진 채** 간다는 것이다.

**D2-1의 폭발 반경 안으로 판단해 수용한다.** 되돌리면 수리가 반쪽이 된다. 다만 **내가 지목하지 않은 편집**이므로 여기 적어 재감사가 범위 이탈 여부를 따로 볼 수 있게 한다.

mtime으로 확인된 이번 라운드 편집: `acceptance.md`·`progress.md`만. `spec.md`·`plan.md`는 iter1 그대로. AC/REQ id 집합 불변(14/12).

### 17.7 이 절이 안 본 것

- **iter3 재감사를 아직 안 돌렸다.** 위 수리는 **수리 측이 자기 산출을 잰 것**이고, 실행자가 자기 판정을 내리는 형태다. 독립 확인이 남았다.
- 수리 측의 209/67/53 추출 수를 **내가 재현하지 않았다.** 앵커 매치 1/1/1과 대조군 0은 재현했다.
- Go 테스트·빌드는 이 라운드에도 돌리지 않았다.

### 17.8 [내 재측정] 공허 통과를 관측으로 확정

수리 측이 「연역이 관측이 됐다」고 보고했다. **그 전환이 이 라운드의 핵심이므로 내가 다시 쟀다.**

**양방향 분류** — 각 대상을 두 형태 모두로 세어, 측정 자체가 공허하지 않음을 먼저 고정했다:

| 대상 | method 형태 | plain 형태 | 합 |
|---|---|---|---|
| `handleSave` (`handlers.go`) | **1** | 0 | 1 |
| `parseSchemaForm` (`schemaform.go`) | 0 | **1** | 1 |
| `ApplySchemaEdits` (`sectionapply.go`) | 0 | **1** | 1 |

각 행이 정확히 1이다 — 어느 대상도 두 형태 모두이거나 어느 쪽도 아니지 않다. **하나는 메서드, 둘은 평범한 함수**다.

**공허 통과 자체의 관측**:
```
git show c068667ad:internal/web/handlers.go > /tmp/base  (666 lines)
순진한 앵커 '^func handleSave(' — base: 0 · head: 0
대조군(수신자 허용 앵커) — base: 1
```

base 사본이 666줄로 비어 있지 않은데도 순진한 앵커가 **양쪽 0**이고, 수신자 허용 앵커는 base에서 **1**을 잡는다. 즉 0은 파일이 없어서가 아니라 **앵커가 그 모양을 못 맞춰서**다 — `diff`가 빈 것끼리 비교해 종료 0을 내고 `IDENTICAL`을 찍는 경로가 이렇게 확정된다.

**연역이 관측이 됐다.** iter2 감사관이 워크트리 가드에 막혀 남긴 자리를, 수리 측이 뚫고 나도 재현했다.

### 17.9 워크트리 가드가 루프를 **두 번, 다른 이유로** 거부한다

수리 측이 기록한 것 중 재사용 가치가 큰 관측이다.

1. 첫 거부 — 루프가 `git`을 가드가 정적으로 검증 못 하는 복합 형태로 부른다.
2. `git show` 셋을 밖으로 빼내 루프 본체에 **`awk`만** 남겼는데 **또 거부**됐다 — 이번엔 가드가 **비-리터럴 `awk` 프로그램을 읽지 못하기** 때문이다.

두 번째 이유는 예상되는 쪽이 아니다. 이 리포의 「복합 git 스크립트는 워크트리 가드가 거부 — 단일 plain 명령 분해」 지침이 **git에 국한되지 않는다**는 뜻이다.

그래서 AC는 **대상×양쪽 = 6개 plain 명령**으로 쓰였고, **한 번 뚫은 방법이 그대로 계약에 실렸다** — 실행자가 감사관이 막힌 자리에서 안 막힌다.

### 17.10 수리 측이 남긴 질문 — 답한다

수리 측이 flag 했다: "`git status`가 `spec.md`·`plan.md`를 clean으로 보이는데, 한 시간 전엔 modified였다. 당신이 흡수한 게 아니라면 병합 전에 봐야 한다."

**나다.** iter3 수리 직전 커밋 `ba7d93411`이 iter1 상태의 네 파일을 함께 담았다. 흡수가 아니고 develop은 아직 안 들어왔다(§17.5의 `d4162b368`은 여전히 미흡수).

**침묵하지 않고 물은 것이 옳다.** 트리 상태가 자기 예상과 다를 때 「누가 썼나」를 묻는 것은 이 저장소가 병렬 작성자 사고에서 배운 형태다 — 이 경우엔 나였고 설명이 되지만, 안 물었으면 확인되지 않았을 것이다.

---

## 18. plan-audit PASS 0.89 · 흡수 · 좌표 재확인

### 18.1 판정

**PASS 0.89** (Tier M 기준 0.80). 궤적 **0.69 → 0.84 → 0.89**, STOP 없음. 판정서 `.moai/reports/t509/plan-audit-iter3.md`.

다섯 델타 항목 전부 통과했고, **감사관이 모든 수치를 인용하지 않고 자기 손으로 재현**했다 — 209/209 · 67/67 · 53/53, 대조군 0/0 + `diff` rc=0, 앵커 비교, method/plain 행 합. 전부 일치.

`MU-8`은 양쪽을 실제로 돌렸다: 수리 전 **GREEN**(빈 스트림 둘 → `diff` 0 → `IDENTICAL`) → 수리 후 **RED**.

`AC-WCP-009`는 **폭발 반경**으로 판정됐다 — iter1 권고 #6과 iter2 D7 행이 이미 지목하고 있었다. **내 2건 목록이 결함 자신의 반경보다 좁았다.** 수리를 좁게 이름 붙인다고 수리가 닿는 범위가 좁아지지 않는다.

### 18.2 [내 위반] 감사 창 안에서 내가 두 번째 작성자가 됐다

감사관이 잡았다: "감사 창 안에서 두 번째 행위자가 스테이징했다. 감사받는 워크트리에는 작성자가 하나여야 한다."

**나다.** §17.x를 쓰면서 이 카드 기록에 그 규칙을 적어놓고, 감사가 도는 중에 스테이징하고 커밋을 시도했다. `index.lock`이 거부해 `HEAD`는 안 움직였고 감사는 옮겨간 기준 위에서 이뤄지지 않았다 — **그러나 그건 락이 한 일이지 내가 한 일이 아니다.** 락이 쓰기를 거절한 것과 작성자가 쓰기를 삼간 것은 다른 일이다.

### 18.3 잔여 3건 종결 (재감사 없이 내가 확인)

| # | 내용 | 내 검증 |
|---|---|---|
| D3-5 | `version: 0.3.0` + HISTORY에 0.2.1·0.3.0 행 추가 | 버전 1행 · HISTORY 2행 확인 |
| D3-3 | AC-WCP-012가 대상별 3중 리터럴로 재작성 — `git show`·`awk`·`wc -l` 쌍·`diff` | 리터럴 `diff` **3** · `wc -l` **6**(=3쌍) |
| D3-2 | MU-8이 대상별 기제를 명시 | 아래 재현 |

**D3-3의 값**: 판정 순서가 명시됐다 — **행 수를 먼저 읽고 그다음 `diff`**. 빈 것끼리의 `diff`는 조용하므로, 순서를 뒤집으면 **그 공허를 막으려 만든 기준 안에서 공허한 답**을 받는다.

**D3-2 내 재현** — 수신자 제거 기제가 셋 중 둘에 무효다:

| 기제 | `handleSave` | `parseSchemaForm` | `ApplySchemaEdits` |
|---|---|---|---|
| 수신자 제거 | **0** | 1 | 1 |
| 이름 오타 | **0** | **0** | **0** |

id 집합 불변 확인: AC **14** · REQ **12** · MU **8**.

### 18.4 흡수와 좌표 재확인

흡수 `f9bac43af`, 충돌 **0**. 로컬 develop `098a631b3`(흡수 시점) · `origin/develop` `d4162b368`.

**인용 좌표 6곳이 병합 트리에서 그대로 대상을 가리킨다:**

| 좌표 | 확인 |
|---|---|
| `fieldsets.templ:365-375` | `__present`/`radio` 3히트 (boolSegment) |
| `fieldsets.templ:739-762` | `CodexState` 6히트 (codexAuthBlock) |
| `tab_layout_test.go:76-88` | `panelHTML` 1 |
| `tab_layout_test.go:90-103` | `assertPanelFields` 2 |
| `settings_shell.go:112-129` | `settingsTabFieldCount` 2 |
| `mcp_codex_surface_test.go:24` | 비어있지 않음 (50바이트) |

**AC-012 앵커도 병합 후 1/1/1**, 순진한 앵커는 여전히 0.

### 18.5 D2-2 수리가 **실증됐다** — merge-base가 실제로 움직였다

흡수로 merge-base가 `c068667ad` → **`d4162b368`**으로 바뀌었다. D2-2가 겨눈 바로 그 상황이다.

```
$ git diff --name-only c068667ad d4162b368 -- <저장 경로 3파일>
(무출력)                                    ← 대상은 안 움직였다

새 merge-base 추출 행 수:  209 · 67 · 53    ← 옛 base 와 동일
```

**AC가 SHA를 고정하지 않고 `git merge-base`를 동적으로 계산하기 때문에 base 이동이 무해했다.** 만약 tip을 읽는 종전 형태였다면 지금 다른 값을 읽었을 것이다 — 이번엔 대상이 안 움직여 결과가 같았겠지만, 그건 **운이지 기준의 성질이 아니다.** 수리가 그 운에 기대지 않게 만들었다.

### 18.6 아직 아무도 실행하지 않은 것

**Go 테스트·빌드·vet·templ-generate가 plan 단계 전체에서 한 번도 안 돌았다.** 감사관도 안 돌렸고 나도 안 돌렸다. `AC-WCP-012`의 green-build arm은 **run 단계가 첫 실행**이다. `progress.md`에 알려진 공백으로 기록됐다 — 놀람이 아니라 예고로 도착하도록.

그리고 수리 델타는 **git으로 분리 불가**하다(iter2가 미커밋 트리를 감사했고 수리 커밋이 v0.2.0과 수리를 함께 담았다). 이미 통과한 v0.2.0 표면 중 `AC-WCP-012`/`AC-WCP-009`/`§D.2` 밖은 재검토되지 않았다.

---

## 19. run 진입 — 기준선 (구현 착수 **전에** 잰 것)

Implementation Kickoff Approval 획득(운영자, 2026-09-07, 리드 세션). 병합 순서도 운영자가 확정: **run·sync 후 창 한 번.**

### 19.1 왜 기준선을 먼저 잡는가

`AC-WCP-012`의 green-build arm은 **run이 첫 실행**이다(§18.6). 지금 초록인지 모르는 상태에서 구현하면, **내 변경이 깨뜨린 것과 원래 빨갛던 것을 구분할 수 없다** — 그 순간 run 판정 전체가 오염된다. 리드가 이 계획을 [HARD]로 승격했다.

### 19.2 기준선 — 세 축 전부 초록

흡수 후 트리(`725071d91`, develop `0b1e27877` 포함) 기준:

| 축 | 명령 | 결과 |
|---|---|---|
| 빌드 | `go build ./...` | **rc=0** |
| 테스트 | `go test ./internal/web/ ./internal/settings/ -count=1` | `ok internal/web 3.960s` · `ok internal/settings 0.624s` |
| templ 드리프트 | `templ generate` (from `internal/web/`) + `git status` | **드리프트 0** |

**그러므로 이후 나오는 빨강은 내 것이다.**

### 19.3 [내 오류] templ 드리프트 가드는 **cwd 민감**하다 — 가짜 빨강을 내가 만들었다

처음에 `templ generate`를 **리포 루트에서** 돌렸다. 결과: 생성 파일 7개, **386 insert / 386 delete** 드리프트. 하마터면 "기준선이 빨갛다"고 보고할 뻔했다.

차이의 성질을 보니 전부 한 종류였다:

```diff
-  return templ.Error{..., FileName: `icons.templ`, Line: 108, Col: 44}
+  return templ.Error{..., FileName: `internal/web/icons.templ`, Line: 108, Col: 44}
```

`templ generate`가 `FileName:`을 **실행 디렉터리 기준 상대 경로**로 적는다. 커밋된 파일은 `internal/web/`에서 생성됐고, 나는 루트에서 돌렸다. **트리가 틀린 게 아니라 내 호출 위치가 틀렸다.**

판별:
```
git restore internal/web/        → 0 파일
cd internal/web && templ generate → 드리프트 0
```

**교훈**: 드리프트 가드는 **생성 시점의 cwd를 재현해야** 한다. 재현하지 않으면 386줄짜리 오탐이 나오고, 그것은 "코드가 낡았다"가 아니라 "내가 다른 자리에서 쟀다"이다. §1.2에서 내가 이미 한 번 밟은 형태 — **잰 대상을 명시하지 않은 측정** — 의 변형이다. run 단계와 CI가 이 가드를 돌릴 때 같은 자리에 서야 한다.

### 19.4 [정정] §18.4의 흡수 값이 틀렸다

§18.4에 "로컬 develop `098a631b3`(흡수 시점)"이라 적었다. **틀렸다.** 실제로 들어온 것은 `0b1e27877`이다.

원인: tip을 읽은 시점(`098a631b3`, 15:27)과 `git merge develop`을 실행한 시점 사이에 develop이 한 번 더 전진했고(`0b1e27877`, 15:36), **머지는 내가 읽은 값이 아니라 그 순간의 develop을 가져갔다.**

```
git merge-base --is-ancestor 098a631b3 0b1e27877 → YES  (098 이 먼저)
git merge-base --is-ancestor 0b1e27877 HEAD      → YES  (내 HEAD 가 이미 포함)
git rev-list --count --left-right 0b1e27877...HEAD → 0  16
```

손해는 없다 — 더 많이 흡수했고 충돌 0이며 지금 develop과 완전히 동기다. **그러나 보고한 값이 실제 입력이 아니었다.** 이 리포가 [HARD]로 적은 「커밋·푸시 직전에 HEAD를 다시 읽어라, 턴 앞에서 읽은 값을 쓰지 마라」의 정확한 사례이고, 내가 그것을 밟았다.

### 19.5 이 절이 안 본 것

- `go vet`은 안 돌렸다. 기준선 세 축에 넣지 않았다.
- 테스트는 **두 패키지만** 돌렸다(`internal/web`, `internal/settings`). 전체 스위트는 로컬에서 돌리지 않는다는 이 리포의 규율에 따른 것이고, 전 패키지 판정은 CI 몫이다.
- `templ generate`의 **버전**을 고정해 확인하지 않았다. `go run github.com/a-h/templ/cmd/templ`이 go.mod에서 해석하는 판을 썼고, CI가 같은 판을 쓰는지는 재지 않았다.

---

## 20. run 완료 — AC 14/14 · 뮤턴트 8/8 RED · 그리고 **SPEC 반증 5건**

### 20.1 내 독립 검증

구현 커밋 `dc817ff65`·`d8f92092a`. 아래는 **내가 다시 돌린 것**이고 구현 측 보고의 인용이 아니다.

| 축 | 결과 |
|---|---|
| `go build ./...` | **rc=0** |
| `go test ./internal/web/ ./internal/settings/` | `ok 4.431s` · `ok 0.531s` |
| `go vet ./internal/web/ ./internal/settings/` | **rc=0** |
| codex 테스트 12개 | 전부 PASS, **실행 12개**(0이 아니므로 공허하지 않음) |

**금지선 유지 확인**: `schemaform.go`가 바뀌었으나 변경은 **탭 등록 10줄뿐**이고 `parseSchemaForm`은 무접촉(`git diff … | grep -c '(PostFormValue|func parseSchemaForm)'` → **0**). `.moai/config` 무접촉.

**진단 오탐 1건**: LSP가 `zz_t509dup_test.go` 등 뮤턴트 잔재와 미정의 심볼 5개를 보고했으나, `git status` 빈 출력 · `ls internal/web/zz_*` no matches · `go build` rc=0. **워크트리에서 LSP 진단이 낡는 알려진 형태**다. 도구가 아니라 실행으로 판정했다.

### 20.2 [최중요] SPEC이 명시하지 않은 결합이 **실제로 깨졌다**

`internal/web/mcp_console_test.go`의 `TestMCPConsoleWriteCapableTextDistinction`이 실패했다. 원인:

그 테스트는 `strings.Index(body, chip)`로 **페이지 전체 첫 등장** key chip에 창을 앵커한다. codex 미러는 같은 chip을 **탭 순서상 더 앞에서** 되풀이하므로, 창이 배지도 컨트롤도 없는 미러 행에 내려앉아 **"MCP 표면이 열화됐다"**고 보고했다 — MCP 표면은 전혀 안 바뀌었는데.

수리는 **범위 축소이지 약화가 아니다**:
```diff
-	body := renderConsolePage(t)
+	body := panelHTML(t, renderConsolePage(t), "mcp")
```
그 테스트가 원래 말하려던 패널로 창을 좁혔다. 프로덕션 코드·저장 경로 무접촉이고, 이유가 주석으로 그 자리에 남았다.

**일반화 — 후속 카드 후보다.**
> **not-last 규칙은 `panelHTML` 소비자만 보호한다.** 미러가 복제하는 것은 `name`이 아니라 **텍스트**이므로, **페이지 전체 첫-등장 앵커를 쓰는 기존 단언은 전부 같은 위험을 갖는다.**

SPEC의 커플링 목록에 이 파일이 없었고, 하필 `AC-WCP-009`가 이 테스트로 "소유 탭 불변"을 판정한다. 감사 3회가 못 잡은 것을 **첫 실행이 잡았다** — 그래서 「아무도 실행한 적 없다」가 §18.6의 [HARD] 고지였던 것이다.

### 20.3 SPEC 반증 나머지 4건 (구현 측 보고, 내가 재현하지 않음)

| # | 반증 |
|---|---|
| 2 | **`AC-WCP-013`의 레일 팔은 문자 그대로 만족 불가능하다.** `shell.templ:156`이 모든 탭에 무조건 `<span class="count">`를 찍으므로 레일은 **`0`을 보여주지 카운트 부재를 보여주지 못한다.** 괄호의 `(zero)`를 정본으로 읽고 진행. "카운트를 말하지 않는다"가 가능한 곳은 패널 헤더뿐이라 거기서만 생략. 레일까지 없애려면 전 탭에 영향 가는 변경이라 범위 밖 |
| 3 | **`AC-WCP-011`은 커밋 전에는 다른 질문에 답한다.** `git diff origin/develop...HEAD`는 커밋된 이력만 본다 — 구현 커밋 전에 돌리면 경로 9개(plan 산출물뿐)에 `filter_rc=1`로 **통과처럼 보이지만 구현에 대해 아무것도 재지 않는다.** §D.4가 구제하나 기준 자체가 "run-phase 커밋 이후"를 말했어야 한다. 커밋 후 재관측(경로 20개) |
| 4 | **`REQ-WCP-005`의 "and the shared MCP tool catalogue"는 하중을 받는 구절이다.** predicate를 `strings.Contains(name,"codex")` 한 줄로 썼다면 `AC-WCP-006`의 두 팔이 **바이트 동일한 계산**이 되어 레지스트리 스윕 팔이 독립 oracle이 아니게 되고, MU-3은 적용할 팔이 없어 **작성 불가**가 된다. SPEC이 이를 가정하지만 경고하지 않는다 |
| 5 | **MU-6을 잡는 것은 어떤 Go 테스트도 아니다.** count 팔은 초록이고 `grep -c 'case "codex"' settings_shell.go`만 적색이다. 즉 **CI는 이 뮤턴트를 못 잡고**, 가드는 누군가 그 grep을 손으로 돌릴 때만 산다. SPEC이 의도한 분리이나 **잔여 위험**으로 남는다 |

### 20.4 SPEC이 결정하지 않아 구현 측이 정한 것 1건

`REQ-WCP-003`은 "current value as read from disk"인데, MCP 도구 bool 6개는 **디스크 값이 보통 비어 있고 콘솔 의미론은 "비었으면 켜짐"**이다. 원시 빈 값을 그대로 두면 "꺼짐"으로 읽히고, 유효값을 계산하면 패널이 **두 번째 분류기**가 된다(`REQ-WCP-006`이 probe에 대해 금지하는 바로 그것).

절충: `(unset)` 플레이스홀더 + 그룹 머리글에 "비어 있으면 켜진 것으로 읽는다"는 사실 문구. **후속에서 명시할 값이 있다.**

### 20.5 t517 — 목격 없음, 그러나 부재 증거는 아니다

작업 전 구간에서 `git status --short`에 `.moai/config/**`가 한 번도 나타나지 않았다. **다만 실제 서버를 띄운 적이 없다**(모든 테스트가 자기 `t.TempDir()` 루트를 만든다). 이것은 **결함이 없다는 증거가 아니라 관측이 없다는 사실**이다 — §10의 관측은 서버를 띄웠을 때 나왔다.

### 20.6 안 본 것

- 구현 측의 AC별 출력을 **전부 재현하지는 않았다.** 빌드·테스트·vet·codex 테스트 12개·금지선·mcp 수리 성질을 재현했고, 나머지 AC 판정과 뮤턴트 8건은 구현 측 보고다.
- **`moai web`을 띄워 눈으로 보지 않았다.** 렌더는 테스트가 판정했다.
- `golangci-lint`를 내가 돌리지 않았다(구현 측이 `0 issues.` 보고).
- 전체 스위트 미실행 — CI 몫.

---

## 21. sync 완료 — 그리고 **범위 확대 1건을 내가 수용한다**

### 21.1 내 독립 검증

sync 커밋 `be968543f`(11파일) · `c891d50ab`(SHA 백필). 아래는 내가 다시 잰 것이다.

| 축 | 결과 |
|---|---|
| `go build ./...` | **rc=0** |
| `go test ./internal/web/ ./internal/settings/` | `ok 5.462s` · `ok 0.487s` |
| SPEC 상태 | `version: 0.3.0` · **`status: completed`** · `updated: 2026-09-07` |
| `sync_commit_sha` | `be968543f` 기록됨 |
| 탭 수 | `consoleTabs()` **14** · `wantTabOrder` **14**, `codex`가 **8번째**(audit 직후, 마지막 아님) |

### 21.2 [수용] 문서의 탭 수가 **이 카드 이전부터 틀려 있었다**

sync 측이 밝혔다: 문서는 **11탭**이라 적고 있었는데 실제는 **14탭**이다 — `feedback`과 `gate`가 **이 카드와 무관하게** 빠져 있었다.

그래서 codex만 더하면 **"12"**가 나오는데, 그건 **알면서 거짓을 쓰는 것**이다. 두 누락을 함께 접어 넣어 14로 맞췄다.

**범위 확대가 맞고, 나는 수용한다.** 근거는 sync 측 논리 그대로다 — **수는 원자적이라 절반만 고칠 수 없다.** 대안은 "새 거짓을 쓰기"였고, 그건 낡은 드리프트를 남기는 것보다 나쁘다. `AC-WCP-009`(§17.6)와 같은 계열의 판단이고, 그때처럼 **밝히고 고친 것**이 옳다.

**내 검증**: `grep -c 'LabelKey:' internal/web/schemaform.go` → **14**. 주장이 참이다.

### 21.3 sync 측이 보고한 나머지 반증 4건

| # | 내용 | 처리 |
|---|---|---|
| 2 | `cli-reference/web.md`가 4로케일 전부 **"설정 9개 탭"** — 이 카드 이전부터 5개 틀렸다 | **손대지 않았다**(do-not-fix 경계). §E.4에 기록. 이제 형제 페이지와 모순되므로 **두 표면을 함께 훑는 문서 정확성 카드**가 필요하다 |
| 3 | `plan.md`·`acceptance.md`·`progress.md`에 **YAML frontmatter가 아예 없다** — heading-first 문서다. "4개 산출물 원자 전이"는 셋에 대해 **0의 작업**을 기술한다 | §E.4에 명시 기록(넷이 전이했다고 암시하지 않도록) |
| 4 | README 스크린샷 alt 텍스트가 탭 수를 담고 있었고, **이미지는 실제로 11탭을 보여준다** | 14로 고치면 **그림을 잘못 기술**하게 되므로 alt에서 **수를 제거**. 다만 **스크린샷 자체가 낡았다** — 재생성은 범위 밖이고 다른 데 기록이 없어 여기 남긴다 |
| 5 | **zsh는 `PIPESTATUS`를 안 채운다.** 첫 `go build`/`go test`를 `tail`에 파이프해 `tail`의 종료코드를 rc로 보고했고 하나는 빈 값이 찍혔다 | 파일 리다이렉트 후 `$?` 직독으로 **재측정**. 보고된 수치는 두 번째 것 |

**5번은 이 저장소가 이미 기록한 형태**다(「같은 지표를 양끝에서 재라」). 파이프 뒤의 `rc`는 파이프 마지막 명령의 것이지 내가 재려던 명령의 것이 아니다. sync 측이 스스로 잡고 다시 잰 것이 옳다.

### 21.4 반증 #2(`AC-WCP-011`) — **내가 커밋 후로 닫았다**

리드가 가장 위험하다고 짚은 건이다. 세 단계를 각각 관측했다:

```
merge-base                                    0b1e27877259fef70079188bf7714029cbaa7ded
step1  git rev-parse --verify -q <base>       ref_resolves_rc=0
step2  git diff --name-only <base>...HEAD     20 경로 (그중 internal/web/ 11)
step3  grep -E '\.moai/config/'               filter_rc=1
대조군 grep -cE 'internal/web/'                11
```

**핵심은 step2다.** 구현 커밋 **전에는 9경로(전부 plan 산출물)**라 `filter_rc=1`이 나면서도 구현을 하나도 안 쟀다. 지금은 20경로에 구현 파일 11개가 들어 있다.

대조군을 붙인 이유: `filter_rc=1`은 **「config 무접촉」과 「필터 고장」 둘 다에서** 나온다. `internal/web/` → 11이 필터가 살아 있음을 고정한다.

그리고 **가드가 변수 형태를 거부**해 리터럴로 나눠 돌렸다 — AC가 9개 리터럴 명령으로 쓰인 이유가 실행에서 그대로 재현됐다.

### 21.5 안 본 것

- sync 측의 `hugo build`·`docs-i18n-check.sh`·4로케일 패리티 수치를 **재현하지 않았다.** 빌드·테스트·SPEC 상태·탭 수는 재현했다.
- `moai web`을 띄우지 않았다. 렌더는 테스트가 판정했다.
- `golangci-lint`·전체 스위트·크로스플랫폼 미실행 — CI 몫.
- **t517**: 이 단계에서도 `.moai/config/**`가 `git status`에 미출현. 그러나 **서버를 안 띄웠으므로 관측 부재이지 결함 부재가 아니다**(§20.5와 동일).

---

## 22. Amendments 착지 — 그리고 **내가 증폭한 숫자의 단위가 어긋났다**

### 22.1 산출

`spec.md` v0.4.0(+159/−1, 유일한 삭제는 버전 줄 하나) — HISTORY 행 + **`## Amendments` 절 A1~A6**. `acceptance.md`에 `§D.6` 포인터 표(+25/−0). `status: completed` 불변.
검증: `moai spec lint` 무소견 · `spec_audit` `modern_era_clean: 1`, `drift_findings: []`.

원문은 **하나도 지우지 않았다**. 「무엇을 믿었고 실행이 무엇을 반증했는지」가 남는 것이 이 절의 목적이다.

### 22.2 [정정] §21.4의 「20경로」는 **고정할 수 없는 수**다

Amendment 측이 현재 HEAD(`c891d50ab`)에서 재측정했다:

```
ref_resolves_rc=0
git diff --name-only origin/develop...HEAD | wc -l              → 29   (내 관측 시점 20)
git diff --name-only origin/develop...HEAD | grep -cE 'internal/web/' → 11   (불변)
filter_rc=1
```

**총계가 20→29로 늘었다** — 내 관측 뒤에 sync 커밋과 기록 커밋들이 착지했기 때문이다.

**이건 내 수치의 결함이 아니라 기준의 성질이다**: 경로 총계는 **움직이는 값이고 고정할 수 없다**. 다음 독자가 비교해야 할 것은 **총계가 아니라 대조군(`internal/web/` 11)과 `filter_rc`**다. A5가 두 값을 각각 귀속해 적고 그 지시를 남겼다.

§21.4에서 내가 "20경로"를 성취처럼 적은 것은 **고정 불가한 수를 고정된 것처럼 인용한 것**이다. 대조군을 붙인 것은 옳았고, 총계에 무게를 실은 것이 틀렸다.

### 22.3 [내 오류] **「2 대 34」는 단위가 어긋난 비교다**

리드가 스윕 규모를 `strings.Index` 앵커 **34개소/15파일** 대 `panelHTML` 소비자 **2개**로 쟀고, 나는 그것을 **"보호받는 쪽이 소수라는 게 2:34로 확인됐다"**고 증폭해 옮겼다.

**내 재측정:**

| 대상 | 사이트 | 파일 |
|---|---|---|
| `strings.Index(` (원시 상위집합) | **60** | **19** |
| `panelHTML(` | **12** | **3** (정의 파일 `tab_layout_test.go` 포함 → 소비자는 2) |

**「2 대 34」는 파일 대 사이트다.** 같은 단위로 놓으면 `19 파일 대 3 파일`, 또는 `60 사이트 대 12 호출부`가 된다. 어느 쪽도 2:34가 아니다.

그리고 리드의 34/15는 **60/19의 필터링된 부분집합**이다 — 「페이지 전체 첫-등장 **앵커링**」으로 좁힌 것. 그 필터가 정확히 **t527의 [HARD] 1이 세우라고 요구한 판별식**이므로, **34/15는 스윕의 경계가 아니라 잠정치**다. t527이 그것을 경계로 읽으면 안 된다.

**이 저장소가 이미 기록한 형태다** — 「갈리면 단위부터」·「센 단위가 주장자의 단위와 맞아야 한다」. 나는 숫자가 내 일반화를 지지한다는 것에 이끌려 **단위를 확인하지 않고 증폭했다.** 일반화 자체(「보호받는 쪽이 소수」)는 어느 단위로 봐도 참이지만, **내가 인용한 비율은 참이 아니다.**

### 22.4 Amendment 측이 확인한 것 (내가 재현하지 않음)

- `shell.templ:156`이 count span을 **무조건** 찍는다 → A1의 전제 확인
- **MU-6은 정말 안 잡힌다** — `codex_panel_test.go:352-356`이 빈 names와 count 0을 단언하는데 **`default` 분기도 그 둘을 만족**하고, grep을 돌리는 테스트가 없다
- `(unset)` 플레이스홀더와 그룹 주석이 `fieldsets_codex.templ:61`·`codexmirror.go:183`에 실재
- **t527이 SQLite 백로그에 실재**한다 — 다만 `moai todo` 렌더 목록에는 **안 보인다**(queued 23장만 표시). §12에서 배운 그대로: **목록은 조회용이지 전수용이 아니다**

### 22.5 안 본 것

- Amendments 본문(A1~A6)을 **정독하지 않았다.** 구조(버전·HISTORY·삭제 1줄·lint 무소견)와 위 두 수치만 확인했다.
- `moai spec lint`·`spec_audit`를 **내가 돌리지 않았다.**
- t527의 34/15가 어떤 필터로 나왔는지 **모른다** — 리드에게 확인이 필요하다.
