---
id: SPEC-CODEX-SKILL-LOADER-001
title: "코덱스 에이전트가 moai 스킬을 참조할 수 있게 한다 — 적재 뿌리를 먼저 재측정하고, 그 판정으로 방출 규칙을 연다"
version: "0.1.0"
status: completed
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/template
lifecycle: spec-anchored
tier: M
tags: "codex, skills, agentemit, skill-loader, measurement-first, dual-harness"
related_specs: [SPEC-CODEX-SKILLS-CANONICAL-001, SPEC-CODEX-SKILL-NEUTRAL-001, SPEC-CODEX-WIRING-001]
---

# SPEC-CODEX-SKILL-LOADER-001 — 코덱스 스킬 적재 축

## HISTORY

- 2026-09-03 (plan-phase, v0.1.0, scoped delta-audit iter-3 결함 5 건 + 형제 스윕) — 범위 한정 감사(네 번째 상태 `inconclusive` 의 의미론 축)가 **2 blocking / 1 major / 2 minor** 를 냈고, 그중 **X2 가 N2 를 실제로는 닫지 않은 상태로 만들고 있었다.** **REQ 13 / AC 13 불변**, 분기 구조·판정 정체성 불변.
  - **X2(blocking, N2 수리가 심었다)** — 규칙 4 가 "미발화 회차는 `inconclusive` 이지 분기 B 가 아니다"라고 적은 다섯 줄 아래에서, 분기 기록 지시문은 손대지 않은 채 "R1 이면 A, **아니면** B"로 남아 있었다. "아니면"은 무조건이라 미발화 회차가 그 갈래를 만족하고 **분기 B 로 기록된다** — N2 가 막으려던 거짓 확증 그대로다. 기록하는 순간 실제로 읽히는 문장은 규칙이 아니라 지시문이므로 **낡은 문장이 이긴다.** 지시문을 세 갈래로 다시 쓰고, 2 번의 조건절을 떼지 말라는 [HARD] 를 붙였다.
  - **X1(blocking, 선재 결함)** — AC-006/007 상호 배타 조항에 범위가 없어, 분기 B(004~009 = 해당 없음)와 `inconclusive` 양쪽에서 "둘 다 해당 없음 → FAIL"이 발효했다. **이 SPEC 이 더 유력하다고 본 분기가 판정할 것이 없던 한 쌍 때문에 강제 FAIL 로 끝나는** 구조였다. 조항을 분기 A 로 한정하고, 두 요구사항의 가드가 다른 경로에서 애초에 충족되지 않음을 명시했다.
  - **X3(major, 같은 수리가 심었다)** — 양성 대조를 R2 로 못박은 탓에 "R1 발화 + R2 미발화" 회차가 어느 조항에도 걸리지 않고, 평문으로 읽으면 **직접 관측된 분기 A 를 부정**했다. R1 의 발화 자체가 계측기 작동의 증거다. 대조를 "어느 뿌리에서든 발화하면 충족"으로 고치고 R2 는 전원 침묵 회차의 사전 기준으로 강등했다.
  - **X5(minor)** — `inconclusive` 회차의 나머지 12 개 판정에 아무 처분도 허가되지 않아(적용 조건 절이 분기 **안에서만** 해당 없음을 허가한다) "미기록 = FAIL"만 남았다. 완료 정의에 처분 한 줄과 산출물(blocker report — AC-CSL-003 의 (a)-(e) 는 분기 B 용이라 별개 문서)을 넣었다.
  - **X4(minor)** — 상태 집합을 13 개 전체에 열어 놓고 다음 줄에서 좁혀, 앞의 허용문이 뒤의 제한을 이겼다. 제한을 열거문 안으로 접었다.
  - **형제 스윕에서 2 건 추가 발견** — X3 가 대조를 any-root 로 바꾸면서 **재선택 소진 조건 두 자리**("3 회 뒤에도 **R2 가** 발화하지 않으면")가 낡았다. 소진 조건도 대조와 같은 기준이어야 하므로 둘 다 any-root 로 맞췄다. AC-CSL-003 의 적용 조건에도 `inconclusive` 를 명시해, 완료 정의 한 줄에만 기대지 않게 했다(X5 와 같은 형태의 재발 방지).
  - **두 blocking 이 같은 형태다** — 새 규칙을 넣고 그 규칙이 **낡게 만든 문장**(분기 기록 지시문)과 **새로 상호작용하게 된 문장**(AC-CSL-002 의 `Given`)을 훑지 않았다. `verification-completeness.md` §3 의 cross-layer sweep 이며, 이번엔 수리 **뒤에** 스윕을 돌려 위 2 건을 잡았다.

- 2026-09-03 (plan-phase, v0.1.0, plan-audit iter-2 부채 청산) — iter-2 **PASS-WITH-DEBT 0.86**(Tier M 임계 0.80)의 부채 3 건을 run-phase M0 **이전에** 닫는다. 셋 다 M0 앞에 있으므로 이월하면 run 이 깨진 계측기로 시작한다. **REQ 13 / AC 13 불변.**
  - **N1(critical, iter-2 가 심었다)** — 맨 `git diff` 는 워킹 트리와 인덱스를 비교하므로 run 커밋(REQ-CSL-008) 뒤 판정 시점에 **빈 출력**을 낸다. 그 빈 출력이 iter-2 가 넣은 "빈 집합 → 해당 없음" 경로로 흘러, 필드를 도입한 실행이 "도입 없음"으로 기록된다 — 빈 스윕 가드가 눈멂을 **자신 있는 오답**으로 바꾸는 형태이며, 그것이 대체한 impossible 방향보다 나쁘다. plan.md §C.0 에 `BASELINE_SHA` 동결(remedy **R2** — 작성 시점에 값을 알 수 없어 R1 은 불가)을 넣고 AC-CSL-005·011·012 세 판정을 전부 그 기준으로 결정하게 했다. 같은 편집에서 키 문자 집합을 `[a-z_]` → `[a-z0-9_]` 로 넓혔다(숫자를 담은 키를 놓쳤다).
  - **N2(major, D3 잔여)** — 음성 대조는 **거짓 양성**만 잡는다. 신호가 아예 발화하지 않으면 다섯 뿌리가 전부 미적재로 읽히고 무표식 대조는 예상대로 통과해, M0 이 깨진 계측기 위에서 **분기 B 를 기록한다** — 분기 B 는 이 SPEC 이 더 유력하다고 본 결과이므로 이 실패 양식은 **기대하던 답을 내놓는다**. 계측기의 붉음은 관측되고 초록은 한 번도 관측되지 않는다. R2(`$CODEX_HOME/skills`, 코덱스 설치기가 문서화한 뿌리)를 **양성 대조**로 세우고, 어느 뿌리도 발화하지 않은 회차를 `inconclusive`(분기 아님)로 못박았으며, 무한 재선택을 막기 위해 재선택을 **3 회**로 묶었다.
  - **N3(major, 분기 A 한정)** — `codex_measured_version` 은 "아래 필드 의미론을 잰 버전"인데(`agents-codex.yaml:13`) 그 값만 새 버전으로 덮으면 이 SPEC 이 재지 않은 선행 7 개까지 커버리지를 주장하게 되고, AC-CSL-005 가 그 7 개를 재판정 대상에서 제외한다고 적어 둔 것과 **내부 모순**이 된다. 버전을 잰 대상 옆(`skill-loader` rationale)에 기록하도록 범위를 좁혔다 — D5 와 같은 계열의 과대 주장이다.

- 2026-09-03 (plan-phase, v0.1.0, plan-audit iter-1 결함 배치 반영) — iter-1 **FAIL 0.79**(Tier M 임계 0.80, Testability 0.60 이 손실 전부)의 결함 7 건을 기존 번호 안에서 닫았다. **REQ 13 / AC 13 불변**, 재번호 없음(Traceability 1.00 유지).
  - **D1(critical)** — AC-CSL-005 가 방출된 TOML 의 **모든** 키를 쓸면서 AC-CSL-004 가 판정한 것만 허용해, 선행 SPEC 이 확인한 7 개 키 때문에 도착 시점부터 붉고 어떤 올바른 작업으로도 초록이 되지 않았다. 스윕을 **이 SPEC 의 diff 가 새로 도입한 키 집합**으로 좁히고, 빈 집합은 통과가 아니라 해당 없음으로 못박았다. REQ-CSL-005 도 같은 범위로 문면을 맞췄다(요구사항은 원래 새 필드만 구속한다).
  - **D2(major)** — 완료 정의의 분기 B 행이 AC-CSL-012·013 을 PASS 로 기록하게 했으나, 분기 B 는 템플릿 트리를 바꾸지 않고 Go 코드를 더하지 않아 두 판정 모두 **스윕 대상이 빈다**. 두 AC 본문에 스윕 수 요건과 대조군을 넣고(AC-CSL-011 의 기존 대조군 조항을 본떴다), 완료 정의의 분기 B 행을 해당 없음으로 고쳤다.
  - **D3(major)** — M0 프로브가 뿌리를 **판별하지 못했다**. 후보를 5 개로 늘리고(`.claude/skills/` 신설 — 미러 링크의 대상이라 이것 없이는 R1 의 양성을 귀속시킬 수 없다), 뿌리당 고유 표식 이름 + 한 실행에 한 뿌리 + 신호 선고정 + 음성 대조를 규칙 1~4 로 못박았다. plan-phase 픽스처는 두 뿌리가 같은 파일·같은 이름을 공유해 이 축에서 **결함**이므로 재사용을 금지했다.
  - **D4** — M0 증거에 실행 귀속이 없었다. AC-CSL-001 의 기록 요소를 다섯(명령·출력 전문·exit code·`codex --version`·트리 SHA)으로 확장하고, 분기 B 의 유일한 버전 고정점이 되도록 AC-CSL-003 의 blocker report 필수 항목에 codex 버전을 넣었다.
  - **D5** — §A.1 의 부재 주장이 셀렉터보다 넓었다. 세 뿌리(`internal`·`pkg`·`cmd`) × 세 표기로 넓힌 명령을 인용하고 결론이 유지됨을 확인했다.
  - **D6** — 리터럴 `11` 을 도출식으로 바꾸고, "재생성된 11 개 각각에" 가 선점하던 **균일 적용** 설계를 "방출 규칙이 선택한 에이전트"로 되돌렸다(M0/M2 가 아직 정하지 않은 사항이다).
  - **D7** — 수정 불요. Tier 비대칭(분기 B 는 파일 축이 접힌다)을 plan §B 에 기록만 했다.

- 2026-09-03 (plan-phase, v0.1.0, 프로브 2 회차 반영) — 레인 2 회차 프로브(§A.5 P1·P3·P4)가 분기 B 를 **더 유력한 쪽으로 올렸다.** 요구사항 개수는 불변(REQ 13 / AC 13)이며, 바뀐 것은 §A.5 의 측정 표와 분기 서술의 무게다: 분기 B 를 예외적 꼬리가 아니라 값이 매겨진 결과로 적고, 분기 B 에서 §A.4 방출 행만으로는 카드가 닫히지 않는다는 의존 관계를 REQ-CSL-003 에 명시했다. `codex doctor` 가 이 질문의 측정 표면이 **아니라는** 관측도 함께 기록해 run-phase 가 그것을 시도하지 않게 한다.

- 2026-09-03 (plan-phase, v0.1.0) — 카드 t452 로부터 착수. **카드 제목의 전제는 부분적으로 반증됐고**, 이 SPEC 은 카드 문면이 아니라 §A 의 실측 위에 선다. 반증 내역과 그것이 만든 범위 차이는 §A.1~§A.5 에 남긴다 — 나중에 읽는 사람이 "카드는 49종이라 했는데 왜 SPEC 은 다른 것을 하는가"를 문서 안에서 답할 수 있어야 한다.

---

## §A 측정된 현재 상태 (착수 트리 `d592b0551`, 워크트리 `.claude/worktrees/t452`)

이 절의 값은 전부 이 트리에서 이 회차에 실행한 명령의 출력이다. 추론인 것은 추론이라고 적는다.

### §A.1 카드가 말한 "49종"은 배포 대상이 아니라 유령 등록이다

`grep -c '^\[\[skills.config\]\]' ~/.codex/config.toml` → `49`. 49 개 전부 `/Users/goos/.codex/skills/moai-*/SKILL.md` 를 가리키고, `ls ~/.codex/skills/` 는 `.system` 과 `hatch-pet` 만 보여 준다 — **가리키는 경로가 하나도 존재하지 않는다.** 이름에는 현재 카탈로그에 없는 은퇴 스킬(`moai-design-tools`, `moai-docs-generation`, `moai-domain-uiux`, `moai-foundation-claude`)이 섞여 있다.

`grep -rn 'skills\.config\|SkillConfig\|skills_config' internal pkg cmd | grep -v _test` 는 **쓰기 주체를 찾지 못한다** — 읽기 전용 파서(`internal/codexwiring/skills.go`, 카드 t451 착지), 권고성 진단(`internal/cli/doctor_codex.go`), 그리고 매니페스트의 유예 rationale 문면(`internal/template/agentemit/agents-codex.yaml`)뿐이다. **셀렉터는 이 결론의 범위와 같아야 하므로 세 뿌리(`internal`·`pkg`·`cmd`)와 세 표기(헤더 리터럴·구조체 이름·snake_case)를 함께 쓴다** — `internal` 한 뿌리에 헤더 리터럴 하나만 쓰면 `pkg`/`cmd` 의 쓰기 주체나 구조체를 marshal 하는 쓰기를 보지 못하므로, 좁은 셀렉터로 "코드베이스 전체가 쓴 적 없다"를 주장하는 것은 근거보다 넓은 단언이 된다. 즉 이 49 개는 Go 이전 moai 버전이 남긴 죽은 장부이며, t451 의 doctor 가 이미 보고한다. 이 SPEC 의 대상이 아니다(§D).

### §A.2 스킬 노출 축은 이미 착지했다

`.moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/spec.md` frontmatter: `status: completed`, `version: "0.7.0"`. 그 SPEC 이 배포 시점 미러(`internal/template/skill_mirror.go`, 172 줄)를 넣었다 — `.agents/skills/<name>` 을 `../../.claude/skills/<name>` 로 향하는 상대 심볼릭 링크로 만들고, 링크가 불가하면 실 디렉터리 복사로 물러난다.

`go test ./internal/template/ -run 'Mirror' -count=1` → `ok  github.com/modu-ai/moai-adk/internal/template  1.517s`.

### §A.3 이 개발 저장소에 `.agents/` 가 없는 것은 정상이다

미러는 배포 시점에 만들어진다. 이 저장소는 템플릿의 **원본**이지 `moai init` 의 대상이 아니다. 여기서 `.agents/` 부재를 결함으로 읽으면 없는 결함을 만들게 된다.

### §A.4 착지하지 않은 실제 축 — 방출기의 `deferred-m1` 행

`internal/template/agentemit/agents-codex.yaml` 이 `skill-loader` 클래스에 대해 스스로 유예를 적어 두었다:

```yaml
  - class: skill-loader
    disposition: deferred-m1
    rationale: >-
      skills.config value set is unmeasured and M1 owns skills
      canonicalization to .agents/skills; M5 emits no skills field. When M1
      lands this row switches from deferred to an emission rule without any
      M5 artifact change.
```

M1(= §A.2 의 SPEC)은 **착지했는데 이 행은 아직 `deferred-m1` 이다.** 실측으로 확인한 결과:

```
$ grep -h '^[a-z_]* *=' internal/template/templates/.codex/agents/moai/*.toml | sed 's/ *=.*//' | sort -u
args
command
description
developer_instructions
model_reasoning_effort
name
sandbox_mode
$ grep -n '^skills' internal/template/templates/.codex/agents/moai/*.toml
(출력 없음)
```

방출된 11 개 TOML 어디에도 스킬 필드가 없다. 미러가 존재하는 배포판에서도 **코덱스 에이전트가 moai 스킬을 참조할 수단이 없다.** 이것이 이 SPEC 의 핵심 대상이다.

### §A.5 물려받은 전제 하나가 미검증이다 — 그리고 이 SPEC 은 그것을 측정으로 푼다

SPEC-CODEX-SKILLS-CANONICAL-001 §A(85 행)은 설계 전체를 "코덱스는 `.agents/skills/` 아래의 심볼릭 링크를 따라간다"에 기대며, 근거를 배포 사용자 프로젝트에서의 M0 런타임 측정으로 귀속한다. **이 회차의 독립 프로브는 그것을 적재 경로로 확인하지 못했다.**

이 회차에 실행한 프로브(codex-cli **0.152.1**, `/Users/goos/.local/bin/codex`):

| # | 프로브 | 관측 |
|---|---|---|
| P1 | `strings <bin> \| grep -cE 'agents/skills'` → `1` | 바이너리 전체에서 **단 1 회**. 그 1 회는 `.mcp.json` · `.claude.json` · `.codex/hooks` 와 나란히 묶인 문자열 덩어리 안, `external-agent-migration/src/detect/mod.rs` 문맥이다 — 저장소가 **다른 에이전트용으로 설정되어 있는지 감지하는 표식**이지 스킬 적재 뿌리가 아니다 |
| P2 | `strings <bin> \| grep -oE '[./A-Za-z_-]*skills?[/A-Za-z_.-]*'` | `CODEX_HOME/skills`, `/.codex/skills`, `./skills/` 가 경로 형태로 나타난다 |
| P3 | `sed -n '46,60p' $CODEX_HOME/skills/.system/skill-installer/SKILL.md` | 코덱스 자신의 설치기가 다른 뿌리를 지목한다 — "Installs into `$CODEX_HOME/skills/<skill-name>` (defaults to `~/.codex/skills`)", "Installed annotations come from `$CODEX_HOME/skills`". 저장소 로컬 뿌리는 이 파일 어디에도 없다 |
| P4 | `codex doctor --json` (격리 `CODEX_HOME`, exit 1) | 스킬 **뿌리를 하나도 열거하지 않는다.** 스킬 형태의 필드는 기능 플래그 이름 `skill_mcp_dependency_install`·`skill_search` 둘뿐 — **`doctor` 는 이 질문의 측정 표면이 아니다**(run-phase 가 시도하지 않도록 여기 적어 둔다) |
| P5 | 같은 덤프 | 설정 구조체 이름 `SkillsConfig{bundled, include_instructions, max_context_tokens, config}`, `SkillConfig{path, name, enabled}` |
| P6 | 같은 덤프 | 에이전트 설정 필드로 보이는 연접: `route_class · developer_instructions · model · model_reasoning_effort · model_reasoning_summary · model_verbosity · personality · service_tier · **skills**` |

[HARD] **P5·P6 은 추론이다.** 바이너리 문자열 덤프의 연접은 구조체 필드의 강한 신호이지만, (a) 에이전트 TOML 에 `skills` 키가 실제로 받아들여지는지, (b) 그 값이 스킬 **이름**인지 **경로**인지, (c) 오값이 `sandbox_mode` 처럼 에이전트 파일 전체를 버리게 하는지는 **한 번도 실행해 본 적이 없다.** 도달 가능성은 정당화가 아니다(`verification-claim-integrity.md` §1).

[HARD] **P1~P4 도 분기 B 를 세우지는 못한다 — 확률을 올릴 뿐이다.** `strings` 는 **런타임에 조립되는 경로를 볼 수 없다.** `.agents` 는 같은 덩어리 안에 **독립 토큰으로도** 존재하므로, 런타임의 `join(".agents", "skills")` 는 이 프로브에 보이지 않는다. 결정적 측정은 여전히 실제 코덱스 세션이며, REQ-CSL-001 이 그것을 run-phase 의무로 못박는다.

그래서 이 SPEC 은 이 갈림을 단언으로 닫지 않고 **측정이 닫게** 설계한다. **P1~P4 를 반영하면 분기 B 가 더 유력한 쪽이다** — 아래 순서는 확률 순이 아니라 서술 순이며, 분기 B 를 예외적 꼬리로 취급하지 않는다:

- **분기 A — `.agents/skills` 가 적재된다**: 남은 간극은 §A.4 의 방출 행 하나뿐이다. 이 SPEC 의 예산 안에서 닫힌다.
- **분기 B — 적재되지 않는다 (현재 더 유력)**: §A.2 의 착지한 미러가 codex-cli 0.152.1 에서 **무력(inert)** 하다는 뜻이다. **그 경우 §A.4 의 방출 행 하나만으로는 이 카드가 닫히지 않는다** — 방출된 `skills` 값이 가리킬 적재 뿌리 자체가 없기 때문이다. 실제 적재 기구(`$CODEX_HOME/skills` 사용자 계층, 또는 프로젝트 계층 `.codex/skills`)를 지목하는 일이 새로 생기며, 그 일은 배포 기구 재설계 + 사용자 계층 동의 문제라 이 SPEC 의 예산을 넘는다. REQ-CSL-003 이 그 경우를 **기록된 정지**로 못박는다 — 놀라움이 아니라 값이 매겨진 결과다.

**run-phase 가 쓸 픽스처는 이미 디스크에 있다**: `/tmp/t452-probe/proj/` 가 `.claude/skills/moai-probe-x/SKILL.md` 와 그것을 가리키는 상대 심볼릭 링크 `.agents/skills/moai-probe-x -> ../../.claude/skills/moai-probe-x` 를 담고 있고, 링크를 통한 파일 판독이 확인됐다. **링크 통과가 확인된 것은 파일 시스템 층의 사실이지 코덱스가 그 경로를 적재한다는 뜻이 아니다** — 그 판정이 REQ-CSL-001 의 대상이다.

---

## §B 결정 기록

**D1 — 카드 제목이 아니라 실측을 SPEC 의 기준으로 삼는다.** 카드는 "49종 전부 부재"라고 적었으나 §A.1 이 그것을 유령 등록으로 반증했다. 카드 문면을 따라가면 사용자 계층 설정 쓰기라는 별개의 동의 문제로 들어간다. 이 SPEC 은 §A.4 의 방출 간극을 대상으로 한다.

**D2 — 측정을 마일스톤 0 으로 앞세운다.** §A.5 의 두 미검증 전제(적재 뿌리, 에이전트 `skills` 키)가 설계의 나머지를 지배한다. 측정 전에 방출 규칙을 쓰면 그 규칙은 추론 위에 서게 되고, 코덱스는 모르는 값을 조용히 무시하므로(선행 SPEC 의 t91 관측) **틀린 방출이 초록으로 보인다.**

**D3 — 분기 B 는 실패가 아니라 산출물이다.** 분기 B 판정은 "이 SPEC 이 실패했다"가 아니라 "물려받은 전제가 이 버전에서 거짓임을 측정으로 확정했다"는 결과이며, 그 자체가 후속 카드의 근거가 된다. blocker report 로 돌려보내는 것이 재계획 게이트다.

**D4 — 유예 행은 어느 분기에서도 `deferred-m1` 로 남지 않는다.** 유예의 조건("M1 이 착지하면")은 이미 충족됐다. 충족된 조건 아래 유예를 유지하는 것은 그 자체로 드리프트다. 방출 규칙(REQ-CSL-006)이 되거나 측정 근거를 단 `documented-drop`(REQ-CSL-007)이 되거나 둘 중 하나다.

**D5 — 쓰기 자세를 명시한다.** 카드 t451 은 신설 소스 2 개가 쓰기 계열 호출을 0 개 담는다는 것을 보장으로 삼았다. 이 카드는 성질상 쓰기가 필요하다. 그래서 어디에·언제·누구의 동의로 쓰는지를 요구사항으로 못박는다: **쓰기는 이 저장소의 템플릿·방출 트리 안에서만 일어나고**(프로젝트 자체 산출물이므로 별도 동의가 필요하지 않다), **사용자 계층(`$CODEX_HOME/config.toml`, `$CODEX_HOME/skills/`)에는 쓰지 않는다**(REQ-CSL-010·011).

---

## §C 요구사항 (GEARS)

### 측정 축

- **REQ-CSL-001** — **While** codex-cli 가 어느 뿌리에서 스킬을 적재하는지 이 저장소에서 관측되지 않은 상태다, run-phase 는 어떤 방출 규칙도 편집하기 **전에** 격리 프로브로 적재 뿌리를 관측하고, 실행마다 명령·출력 전문·**exit code**·**같은 회차에 관측한 `codex --version`**·**트리 SHA** 를 증거 경로에 남겨야 한다. 관측은 뿌리를 **판별**해야 한다 — 한 실행에 한 뿌리만 채우고, 뿌리마다 다른 표식 이름을 쓰며, 신호를 실행 전에 고르고, 표식 없는 음성 대조로 그 신호를 검증한다.
- **REQ-CSL-002** — **When** REQ-CSL-001 의 프로브가 `.agents/skills/<name>/SKILL.md` 의 적재 여부를 판정하면, run-phase 는 그 판정을 분기 A 또는 분기 B 로 기록하고, 기록된 분기만을 근거로 후속 마일스톤에 진입해야 한다. **When** 어느 후보 뿌리에서도 신호가 발화하지 않으면, run-phase 는 그 회차를 `inconclusive` 로 기록하고 분기를 적어서는 **안 된다** — 신호 미발화는 미적재가 아니라 미측정이며, 그것을 분기 B 로 읽으면 더 유력하다고 본 결과를 깨진 계측기로 확증하게 된다.
- **REQ-CSL-003** — **When** 프로브가 분기 B(미적재)를 판정하면, run-phase 는 `agents-codex.yaml` 을 편집하지 않고 blocker report 를 반환해야 한다. 보고서는 (a) 관측된 실제 적재 뿌리, (b) `.agents/skills` 미적재의 관측 근거, (c) 착지한 미러가 이 코덱스 버전에서 무력하므로 **§A.4 의 방출 행만으로는 카드가 닫히지 않는다**는 의존 관계, (d) 그 뿌리를 쓰는 일이 왜 이 SPEC 의 예산 밖인지를 담는다.
- **REQ-CSL-004** — **While** 에이전트 TOML 의 `skills` 키의 존재·값 형태·오값 거동이 관측되지 않은 상태다, run-phase 는 그 키를 방출하기 전에 세 가지를 각각 관측하고 출력을 증거 경로에 남겨야 한다.
- **REQ-CSL-005** — 방출기는 값 집합이 프로브로 확인되지 않은 필드를 **이 SPEC 의 변경으로 새로 방출해서는 안 된다**. (선행 매니페스트의 ship-omitted 규칙을 이 축에도 그대로 적용한다.) **구속 범위는 이 SPEC 이 도입하는 필드다** — 이미 방출되고 있는 7 개 키는 선행 SPEC 의 측정으로 확인된 것이고 그 근거는 `agents-codex.yaml` 의 `fields:`·`classes:` rationale 에 있으므로, 이 요구사항이 그것들을 다시 판정하지 않는다.

### 방출 축

- **REQ-CSL-006** — **Where** 분기 A 가 확정되고 **When** REQ-CSL-004 의 관측이 사용 가능한 `skills` 키를 확인하면, `agents-codex.yaml` 의 `skill-loader` 행은 `deferred-m1` 에서 방출 규칙으로 전환되고, 방출된 값은 관측된 값 형태를 따라야 한다.
- **REQ-CSL-007** — **When** REQ-CSL-004 의 관측이 사용 가능한 키를 확인하지 못하면, 같은 행은 `documented-drop` 으로 전환되고 rationale 은 관측 명령·출력 요지·codex 버전을 담아야 한다.
- **REQ-CSL-008** — **When** `agents-codex.yaml` 이 변경되면, run-phase 는 `make agents-emit` 으로 `.codex/agents/moai/*.toml` **전체**를 재생성하고 `make build` 를 수행해야 한다. 재생성 없이 커밋하면 소스 축과 임베드 축이 갈린다. **개수는 리터럴로 적지 않는다** — 그 회차의 값은 `ls internal/template/templates/.codex/agents/moai/*.toml | wc -l` 로 세며, 에이전트가 추가되는 순간 낡는 숫자를 요구사항에 박지 않는다.
- **REQ-CSL-009** — 매니페스트는 이번 회차 프로브가 사용한 codex-cli 버전을 **이 SPEC 이 실제로 잰 것의 범위 안에서** 기록해야 한다. 매니페스트 최상단 `codex_measured_version` 은 "아래 필드 의미론을 잰 버전"을 뜻하므로(`agents-codex.yaml:13`), 그 값만 새 버전으로 덮어쓰면 이 SPEC 이 **재지 않은** 선행 7 개 필드까지 새 버전에서 확인된 것처럼 주장하게 된다(AC-CSL-005 가 그 7 개를 재판정 대상에서 제외한다고 명시하고 있으므로 SPEC 내부 모순이기도 하다). 따라서 버전은 이번에 잰 대상 옆에 — `skill-loader` 행의 처분 rationale 에 — 기록하고, 최상단 값을 넓히려면 그 넓힘이 덮는 필드를 이 회차에 실제로 재측정했을 때만 한다.
- **REQ-CSL-010** — run-phase 는 이 개발 저장소가 쓰는 사용자 계층 codex 상태(`$CODEX_HOME/config.toml`, `$CODEX_HOME/skills/`)를 읽기 외의 방식으로 건드려서는 **안 된다**. 모든 codex 실행은 `/tmp` 프로젝트 + 격리 `CODEX_HOME` 안에서 수행한다.
- **REQ-CSL-011** — 이 SPEC 이 만드는 코드는 사용자 계층 `[[skills.config]]` 항목을 쓰거나 지워서는 **안 된다**.

### 규율 축

- **REQ-CSL-012** — **Where** 변경이 `internal/template/templates/` 아래에 닿으면, 산출물은 SPEC ID·내부 날짜·커밋 SHA·특정 운영체제 편향 경로를 담아서는 **안 된다**.
- **REQ-CSL-013** — 프로브 절차와 그것이 남기는 스크립트·코드는 홈 경로를 하드코딩하지 않고 `CODEX_HOME` 을 존중해야 하며, `os.Stat` 오류를 "경로 부재"로 접어서는 **안 된다**.

**요구사항 13 개 / Tier M 상한 16 — 예산 안.**

---

## §D 범위 밖

이 절은 무엇을 짓지 않는지를 못박는다. 아래 항목들은 각각 별개의 판단을 요구하며, 이 SPEC 안에서 조용히 처리되면 안 된다.

### Out of Scope — 사용자 계층 유령 등록 49 개의 청소

- §A.1 의 죽은 `[[skills.config]]` 항목 49 개를 지우거나 고쳐 쓰는 일은 이 SPEC 이 하지 않는다. 사용자 계층 파일에 대한 쓰기는 별도의 동의 문제이고, 이 SPEC 의 REQ-CSL-011 은 정반대 방향(쓰지 않는다)을 못박는다.
- 그 항목들의 **보고**는 이미 카드 t451 의 doctor 소관이며 착지해 있다. 이 SPEC 은 보고 표면도 건드리지 않는다.

### Out of Scope — 미러 기구 자체의 재설계

- `internal/template/skill_mirror.go` 의 링크·복사 전략, 갱신 경로, 청소 등록은 SPEC-CODEX-SKILLS-CANONICAL-001 과 그 승계 카드의 소관이다.
- 분기 B 가 확정되어 미러가 무력임이 드러나더라도, 이 SPEC 은 그 사실을 **기록하고 정지**할 뿐(REQ-CSL-003) 대체 기구를 짓지 않는다.

### Out of Scope — 코덱스 e2e 스윕과 파서 잔여 위험

- 카드 t462(codex e2e 스윕)는 **현재 상태를 측정**하는 카드이지 이 카드의 수리를 판정하는 주체가 아니다. 이 SPEC 은 t462 의 착지 순서에 의존하지 않는다.
- 카드 t468(codex 파서 잔여 위험 3 건)은 별개 카드다.

### Out of Scope — 에이전트 본문(`.md`) 문면 변경

- `.codex/agents/moai/*.toml` 은 `.claude/agents/moai/*.md` 로부터 **기계 방출**된다. 이 SPEC 은 매니페스트의 처분 행과 방출 규칙만 바꾸며, 에이전트 본문 산문을 편집하지 않는다.

---

## §E 추적성

| 요구사항 | 판정 |
|---|---|
| REQ-CSL-001 | AC-CSL-001 |
| REQ-CSL-002 | AC-CSL-002 |
| REQ-CSL-003 | AC-CSL-003 |
| REQ-CSL-004 | AC-CSL-004 |
| REQ-CSL-005 | AC-CSL-005 |
| REQ-CSL-006 | AC-CSL-006 |
| REQ-CSL-007 | AC-CSL-007 |
| REQ-CSL-008 | AC-CSL-008 |
| REQ-CSL-009 | AC-CSL-009 |
| REQ-CSL-010 | AC-CSL-010 |
| REQ-CSL-011 | AC-CSL-011 |
| REQ-CSL-012 | AC-CSL-012 |
| REQ-CSL-013 | AC-CSL-013 |

판정 전문은 `acceptance.md`.

---

## §F 참조

- `.moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/` — 착지한 미러(M1). §A.5 의 물려받은 전제의 출처.
- `.moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/` — 중립 지시 계층. `tool_classes` 어휘의 소관.
- `internal/template/agentemit/agents-codex.yaml` — 이 SPEC 이 바꾸는 유일한 매니페스트.
- `internal/codexwiring/skills.go` — 카드 t451 이 착지시킨 읽기 전용 `[[skills.config]]` 파서. 이 SPEC 은 읽기 자세를 이어받는다.
- `CLAUDE.local.md` §2.0 — 에이전트 정의 사본 3 벌과 `make agents-emit` 의무.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1 — §A.5 의 추론 표시와 REQ-CSL-001·004 의 근거.
