# codemaps 정확성 검증 증거 — SPEC-CODEMAPS-REFRESH-002 (카드 t475)

**측정 트리**: 워크트리 `.claude/worktrees/t475`, 브랜치 `WT-codemaps-stale`, HEAD `52f863f36`
**merge-base(HEAD, origin/develop)**: `52f863f3666c9ec754253a06b96ed1fe844f1590`
**스탬프 앵커(작업 전)**: `25a3212a93b4c811cbb22e3c0b34d43571fa65b4` — `tree_root` t476, `generated_at` `2026-09-03T18:18:34Z`

7개 섹션: ① 후보 산출 + 판별 ② 편입 ③ `docs-truth.md` 손 갱신 ④ 구간별 `diff -u` ⑤ 인용 경로 실존 ⑥ 패키지 대조 ⑦ 인용 식별자 hit/miss.

---

## §0. 기준선 재측정 (REQ-CM2-001)

```
$ ./bin/moai graph check ; echo EXIT=$?
codemaps  metric=described-source-diff value=64 threshold=40 verdict=stale
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=absent  (mx-index absent (untracked runtime artifact — fresh worktree state))
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=absent  (edges.jsonl absent (untracked derived artifact — fresh worktree state))
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
  measured from: 25a3212a9 (working-tree-differs-from-stamp)
EXIT=1
```

```
$ cat .moai/project/codemaps/provenance.json
{
  "schema_version": 1,
  "tree_root": "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t476",
  "commit_sha": "25a3212a93b4c811cbb22e3c0b34d43571fa65b4",
  "dirty": false,
  "described_roots": [ "internal", "cmd", "pkg" ],
  "generated_by": "codemaps-gen",
  "generated_at": "2026-09-03T18:18:34Z"
}
```

```
$ go list ./internal/... ./cmd/... ./pkg/... | wc -l
     136
$ git rev-parse --short HEAD
52f863f36
$ git branch --show-current
WT-codemaps-stale
$ git merge-base HEAD origin/develop
52f863f3666c9ec754253a06b96ed1fe844f1590
```

**spec.md §A 저작 시점 값과의 차이: 없다.** value 64 / threshold 40 / verdict stale, `go list` 136, 앵커 `25a3212a9` 모두 저작 시점 판독과 동일하다. 실행 시점의 `origin/develop` 원격 추적 ref 는 `9dddac8828dac771d6780a9acde85844c91a5f42` 이며 merge-base 는 여전히 HEAD 와 같다.

수출 파일 `described-roots-diff-since-anchor.txt` 는 이 트리에서 재현된다:

```
$ git diff --name-status 25a3212a93b4c811cbb22e3c0b34d43571fa65b4 HEAD -- internal cmd pkg > /tmp/D-recomputed.txt
$ diff -q /tmp/D-recomputed.txt .moai/reports/t475/described-roots-diff-since-anchor.txt
(무출력 — 동일)
```

---

## §① 후보 집합 산출 + 전수 판별 (REQ-CM2-002 / AC-CM2-002)

### ①-a 후보 집합은 규칙의 출력이다 — 손 열거가 아니다

spec.md §A.3(a) 의 A층·B층 명령을 그대로 실행했다(중간 계산 `/tmp`, 결과는 이 파일과 `candidates.txt` 로 수출).

```
$ cat .moai/project/codemaps/*.md > /tmp/cm.txt
$ D=.moai/reports/t475/described-roots-diff-since-anchor.txt
$ go list ./internal/... ./cmd/... ./pkg/... | sed 's|^[^/]*/[^/]*/[^/]*/||' \
    | while read -r p; do /usr/bin/grep -q -F "$p" /tmp/cm.txt || echo "$p"; done > /tmp/Zero.txt
$ while read -r p; do awk -v P="$p/" '$2 ~ "^"P {f=1} END{exit !f}' "$D" && echo "$p"; done < /tmp/Zero.txt > /tmp/A.txt
$ awk '$1=="A" && $2 ~ /\.go$/ && $2 !~ /_test\.go$/ && $2 !~ /testdata\//{print $2}' "$D" \
    | while read -r f; do
        /usr/bin/grep -q -F "$f" /tmp/cm.txt && continue
        /usr/bin/grep -qx -F "$(dirname "$f")" /tmp/A.txt && continue
        echo "$f"
      done > /tmp/B.txt
$ wc -l /tmp/Zero.txt /tmp/A.txt /tmp/B.txt
      48 /tmp/Zero.txt
       5 /tmp/A.txt
      14 /tmp/B.txt
      67 total
```

**A층 5** — `internal/core/git`, `internal/settings/yamlpatch`, `internal/stateanchor`, `internal/template/agentemit`, `internal/template/commandemit`.

**B층 14** —

```
internal/cli/codex_skills_disable.go     internal/kanban/prlink_landedref.go
internal/cli/codex_skills_prune.go       internal/kanban/settings_drift.go
internal/cli/doctor_hook_delivery.go     internal/statusline/state_anchor.go
internal/cli/integration_settings_drift.go internal/template/published_skills.go
internal/cli/skills.go                   internal/template/skill_mirror_repair.go
internal/cli/update_mirror_heal.go       internal/web/codexmirror.go
internal/hook/quality/step_git_env.go    internal/web/fieldsets_codex_templ.go
```

**C층 1 (운영자 지명 이월)** — `internal/chain`. 실측으로 두 조건을 확인했다: 6문서 연결 텍스트에 대한 `/usr/bin/grep -c -F 'internal/chain' /tmp/cm.txt` → `0`(exit 1), 앵커 이후 변경 집합에 대한 `/usr/bin/grep -c 'internal/chain' "$D"` → `0`(exit 1). 트리에는 실재한다(`/bin/ls -d internal/chain` → `internal/chain`).

**후보 = 5 + 14 + 1 = 20.** 저작 시점 실측(48 / 5 / 14 / +1)과 동일하며 차이가 없다. 전수 목록은 `.moai/reports/t475/candidates.txt`(한 줄에 하나, 20행).

### ①-b 판별 — §A.3(a1) 책임 질문 전수 적용

판별식(운영자 판정, spec.md §A.3(a1)): **부모 패키지의 기존 서술이 이 단위가 지는 책임을 실제로 담고 있는가?** 히트 수도 변경 파일 수도 근거가 아니다.

인용 좌표는 **재생성 전 사본** `.moai/reports/t475/pre-regen/<문서>.md` 의 행 번호에 고정한다(재생성 후에는 행이 이동하므로, 감사 시점에 다시 해석 가능한 좌표는 사본 쪽이다). 각 행의 좌표·인용문은 다음 형식으로 기계 재확인된다:

```
$ sed -n '<n>p' .moai/reports/t475/pre-regen/<문서>.md | /usr/bin/grep -qF "<인용문>"
```

| 단위 | 판정 | 근거 (책임 명명 + 검사한 부모 산문 위치) |
|---|---|---|
| `internal/core/git` | fold | 책임: 저장소·브랜치·체크아웃·충돌·워크트리 조작의 저수준 리포지터리 프리미티브 계층. 부모 산문이 이 책임을 담는다 — `overview.md:L93` "``core/git``은 인프라", 그리고 `modules.md:L102` "``core/git`` 위에 얹은 상위 유틸리티"(= `internal/git` 이 이 단위 위에 얹혔다고 명시). 히트 0 은 경로 표기 차이(`core/git` vs `internal/core/git`)의 산물이지 서술 공백이 아니다. |
| `internal/settings/yamlpatch` | omission | 책임: 주석·미모델링 키·키 순서를 보존하는 YAML 노드 수술 기반 부분 패치 seam이며, `ConfigManager.Save()` 경로가 없는 8개 섹션의 유일한 쓰기 경로. 검사 범위: `internal/settings` 적중 3행 전수(`modules.md:L86`, `overview.md:L74`, `dependencies.md:L47`). L86 은 `yamlpatch` 를 하위 패키지 **열거 칸**에 이름으로만 싣고 부모 책임을 "두 표면이 공유하는 설정 스키마"로 기술한다 — 스키마 공유는 보존 쓰기 경로 책임을 담지 않는다. |
| `internal/stateanchor` | omission | 책임: `.moai/state/` 를 읽고 쓰는 모든 표면이 공유하는 단일 state anchor(프로젝트 루트) 해석 seam — stdin `workspace.project_dir` → `worktree.original_cwd` → 리포지터리 공통 디렉터리 순의 고정 우선순위. 검사 범위: 부모가 `internal`(적중 236행)뿐이라 `modules.md` 의 패키지 표 5개 전수(L15–L131)와 `overview.md:L70–L99` / `dependencies.md:L15–L51` 의 패키지 목록 전수를 읽었다 — `internal/stateanchor` 행이 없고 "state anchor" 개념 자체가 6문서에 서술되지 않는다. |
| `internal/template/agentemit` | omission | 책임: 보존된 에이전트 정의(`.md`)와 임베드 매니페스트(`agents-codex.yaml`)의 쌍을 중립 원본으로 삼아 `.codex/agents/` TOML 을 결정적으로 이중 발행하는 fail-closed 방출기. 검사 범위: `internal/template` 적중 19행 전수. `modules.md:L56` 은 `agentemit` 을 하위 패키지 **열거 칸**에만 싣고 부모 책임을 "배포기, 렌더러, settings 생성, 스킬 미러, 카탈로그 트리 해시, 모델 정책·프로파일 매트릭스"로 기술한다 — 어느 항목도 codex 측 이중 발행을 서술하지 않으며, `data-flow.md:L71–L110` 의 빌드타임·런타임 템플릿 경로도 이 방출기를 거치지 않는다. |
| `internal/template/commandemit` | omission | 책임: `/moai` 명령 소스를 codex 스킬 아티팩트(`.agents/skills/moai-<command>/SKILL.md`)로 발행 — 본문은 바이트 동일 verbatim, 이름 충돌·설명 누락에 fail-closed. 스탬프 앵커 시점에 존재하지 않던 패키지다. 검사 범위: `internal/template` 적중 19행 전수 — 어느 행도 이 발행 책임을 서술하지 않는다(운영자가 omission 으로 명명한 사례이며 판정이 그와 일치한다). |
| `internal/chain` | omission | 책임: 워크트리 세션 origin-trail 체인 — `.moai/state/chain/events.jsonl` 에 spawn 경계·`session_id` 백필·완료 엣지를 append-only JSONL 로 적는 계보 원장. 검사 범위: 부모가 `internal`(236행)뿐이라 `modules.md:L15–L131` 패키지 표 전수를 읽었다 — `internal/chain` 행이 없다. 6문서의 `chain` 적중 2건은 전부 cobra **명령 이름** 열거(`entry-points.md:L50`, `docs-truth.md:L80`)이며 패키지 책임 서술이 아니다. |
| `internal/cli/codex_skills_disable.go` | omission | 책임: 사용자 계층 `~/.codex/config.toml` 에 `enabled = false` 를 실은 `[[skills.config]]` 항목을 발행해 Codex 쪽에서만 스킬 하나를 끄는 verb. 검사 범위: `internal/cli` 적중 64행 전수, 특히 파일명 접두어가 귀속되는 `codex*` 클러스터 행 `modules.md:L38` 과 명령 등록 열거 `entry-points.md:L46–L55`. L38 은 그 클러스터의 책임을 "외부 에이전트 백엔드 런처, 잡 제어, 준비 상태 점검, 리뷰 게이트" 넷으로 기술하며, 사용자 HOME 설정에 대한 스킬 노출 제어는 그중 어느 것도 아니다. |
| `internal/cli/codex_skills_prune.go` | omission | 책임: 가리키는 파일이 사라진 유령 `[[skills.config]]` 등록을 `~/.codex/config.toml` 에서 제거 — 부재를 증명할 수 있는 것만 지우고 7개 부류는 절대 지우지 않는 allowlist 형 판정, 기본 dry-run. 검사 범위: `internal/cli` 적중 64행 전수, 특히 `modules.md:L38`(codex* 클러스터 4책임)과 `entry-points.md:L46–L55` — 유령 등록 정리는 어디에도 서술되지 않는다. |
| `internal/cli/doctor_hook_delivery.go` | fold | 책임: 배포 템플릿의 훅 엔트리를 프로젝트 `.claude/settings.json` 과 대조해 누락 엔트리를 배치·처방과 함께 보고하는 읽기 전용 doctor 점검. 부모 산문이 담는다 — `modules.md:L35` "진단 — config, disk, harness, hook wiring, mcp version, permission, sandbox, skills, worktree base, agentemit embed, codex" 의 **hook wiring** 이 정확히 이 점검이다. |
| `internal/cli/integration_settings_drift.go` | omission | 책임: `moai integration acquire` 의 선행 조건이자 독립 verb `moai integration preflight` 로서, 병합 창 진입 전 tracked `.claude/settings.json` 워킹 사본의 드리프트를 단정하고 보존·원장 기록을 남긴다. 검사 범위: `internal/cli` 적중 64행 전수 — `modules.md:L31–L44` 클러스터 표에 `integration*` 클러스터가 없고 잔여 버킷 `modules.md:L44` 가 열거하는 이름에도 없으며, `entry-points.md:L46–L55` 의 두 등록 열거에도 없다. |
| `internal/cli/skills.go` | omission | 책임: `moai skills` 명령 트리 — 스킬 노출을 **계층별** 관심사로 두고 계층을 verb 가 아니라 플래그로 명명하며, `--codex` 를 필수로 만들어 사용자 HOME 쓰기를 호출 시점 opt-in 으로 고정한다. 검사 범위: `internal/cli` 적중 64행 전수 — `modules.md:L31–L44` 클러스터 표에 `skills*` 클러스터가 없고, `entry-points.md:L46–L55` 의 root.go 26개 명시 등록과 자기 등록 파일 예시 어디에도 `skills` 가 없다. |
| `internal/cli/update_mirror_heal.go` | omission | 책임: 버전 일치 `moai update` 가 Deploy 앞에서 조기 반환하는 경로 **옆에서** `.agents/skills` 미러를 복구하는 패스 — 재배포에 의존하지 않는 것이 요건이며 존재 게이트는 프로젝트의 기록된 배포 버전이다. 검사 범위: `internal/cli` 적중 64행 전수, 특히 `update*` 클러스터 행 `modules.md:L33`("템플릿 재배포 — 계획/분류/네임스페이스 보호, 3-way 머지, 백업·롤백, 클린 인스톨, dry-run")과 `data-flow.md:L100–L110` 의 update 경로. 둘 다 재배포 파이프라인을 기술할 뿐, 재배포가 일어나지 않는 경로의 복구는 서술하지 않는다. |
| `internal/hook/quality/step_git_env.go` | fold | 책임: gate 스텝의 자식 프로세스에서 리포지터리 **위치**를 정하는 환경변수(`GIT_DIR`·`GIT_INDEX_FILE` 등)만 제거해 실행 대상 리포지터리를 고정한다 — 신원·동작 변수는 의도적으로 남긴다. 부모 산문이 담는다 — `data-flow.md:L58` "린터·포매터·게이트 요약". 판단: 게이트 스텝을 **실행한다**는 부모 책임 안에 그 실행의 환경 범위를 정하는 일이 포함된다고 읽었다. 이 파일은 새 역량을 지지 않고 부모가 기술한 실행의 정확성 세부를 이룬다. |
| `internal/kanban/prlink_landedref.go` | fold | 책임: landed ref 의 3단 해석 사슬(설정 키 → `origin/HEAD` symref 판독 → 컴파일 상수)과 어느 단계가 답했는지의 보고 — 착지/PR 링크 질문이 어느 브랜치를 향하는지를 정한다. 부모 산문이 담는다 — `modules.md:L54` "백로그 큐의 상태 레코드·컬럼·역할 모델, SQLite 저장 엔진, 보드 락, PR 링크, 정합성 조정" 의 **PR 링크**. |
| `internal/kanban/settings_drift.go` | omission | 책임: 병합 전 tracked `.claude/settings.json` 드리프트 검출의 도메인 절반 — `--no-optional-locks` 강제(인덱스 쓰기 락 회피), 사본 보존, 원장 기록. 검사 범위: `internal/kanban` 적중 10행 전수. `modules.md:L54` 가 세는 다섯 책임 중 어느 것도 이 설정 파일 드리프트 단정을 담지 않는다 — "정합성 조정"은 카드·보드의 reconcile 이지 워킹 트리 설정 파일이 아니다. |
| `internal/statusline/state_anchor.go` | omission | 책임: statusline 렌더가 수행하는 모든 state 읽기·쓰기의 앵커를 공유 seam(`internal/stateanchor`)에 위임하는 어댑터 — 세션의 현재 디렉터리는 앵커가 아니라 walk-up 입력일 뿐이다. 검사 범위: `internal/statusline` 적중 7행 전수, 특히 `modules.md:L23` "Claude Code statusLine 렌더러. git·github·model·backlog·goal·usage 세그먼트 조립". 부모는 렌더 **세그먼트 조립**을 기술할 뿐 렌더가 어느 프로젝트 루트에 상태를 쓰는가라는 앵커 책임을 서술하지 않는다. |
| `internal/template/published_skills.go` | omission | 책임: 발행된 `/moai` 명령 스킬(`.agents/skills/moai-<command>/SKILL.md`) 경로에 대해 update 모드에서도 init 모드의 provenance 보호를 유지해 사용자 소유 파일이 살아남게 하고, 건너뜀을 침묵이 아니라 보고로 남긴다. 검사 범위: `internal/template` 적중 19행 전수 — `modules.md:L56` 의 "스킬 미러"는 심볼릭 링크 우회 미러 **생산자**(`data-flow.md:L93`)를 가리키며, 발행 스킬 경로의 배포측 **보호**와는 다른 책임이다. |
| `internal/template/skill_mirror_repair.go` | omission | 책임: Deploy 를 거치지 않고 `.agents/skills` 의 두 생산자 결과를 복구하는 패키지 수준 패스 — DeployerOption 이 아닌 형상을 의도적으로 골라 배포 경로를 건드리지 않는다. 검사 범위: `internal/template` 적중 19행 전수 + `data-flow.md:L93`("`internal/template/skill_mirror.go` 심볼릭 링크 우회 미러링"). 부모는 미러 **생성**을 기술하고 배포 없는 **복구**는 서술하지 않는다. |
| `internal/web/codexmirror.go` | omission | 책임: codex 탭의 행 모델 — Audit·MCP 탭에 사는 codex 설정의 읽기 전용 미러이며, 행은 `settings.AllFields()` 와 공유 MCP 도구 카탈로그에 대한 술어로 파생된다(손 열거 금지). 검사 범위: `internal/web` 적중 5행 전수, 특히 `modules.md:L22` "루프백 전용 브라우저 콘솔. `a-h/templ` 컴파일 뷰(`*_templ.go`) + htmx + SSE(fsnotify)로 프로파일·설정·todo 큐를 편집". 부모는 **편집** 표면을 기술하며 읽기 전용 미러 탭이라는 표면은 서술하지 않는다. |
| `internal/web/fieldsets_codex_templ.go` | fold | **생성 파일**이다(머리 표식 `// Code generated by templ - DO NOT EDIT.`, 짝 소스 `internal/web/fieldsets_codex.templ` 실재). §A.3(a1) 생성 파일 조항에 따라 질문은 "생성기의 서술이 이 산물의 존재 이유를 담는가"이며 담는다 — `dependencies.md:L139` "`.templ` → `_templ.go` 생성이"(L140 "빌드 전제이며 생성물이 트리에 커밋돼 있습니다"), 그리고 `modules.md:L22` "`a-h/templ` 컴파일 뷰(`*_templ.go`)". 편입 대상이 될 것은 생성 산물 목록이 아니라 **생성 관계**이며, 그 관계가 이미 서술돼 있다. |

### ①-c 판정 행 수 = 후보 수 (AC-CM2-002 조건 2)

```
$ wc -l < .moai/reports/t475/candidates.txt
      20
$ /usr/bin/grep -cE '^\| `[^`]+` \| (fold|omission) \|' .moai/reports/t475/codemaps-accuracy-verification.md
20
```

**분류 요약: fold 5 / omission 15.**

- fold 5 — `internal/core/git`, `internal/cli/doctor_hook_delivery.go`, `internal/hook/quality/step_git_env.go`, `internal/kanban/prlink_landedref.go`, `internal/web/fieldsets_codex_templ.go`
- omission 15 — 나머지 전부.

omission 15 는 §② 에서 편입 대상이 된다. fold 5 는 codemaps 산문을 그 단위에 대해 **바꾸지 않는다**.

### ①-d 인용 좌표 해석 가능성 — 15개 전수 재확인

판정 행이 인용한 모든 `<문서>.md:L<n>` + 인용문 쌍에 대해, 그 좌표에 그 문장이 실제로 있는지를 기계 대조했다(plan-audit iter-2 O-1 의 권고 형식). 전부 exit 0.

```
$ sed -n '<n>p' .moai/reports/t475/pre-regen/<문서>.md | /usr/bin/grep -qF '<인용문>'
overview.md:L93       exit=0    `core/git`은 인프라
modules.md:L102       exit=0    `core/git` 위에 얹은 상위 유틸리티
modules.md:L35        exit=0    hook wiring
data-flow.md:L58      exit=0    린터·포매터·게이트 요약
modules.md:L54        exit=0    PR 링크
dependencies.md:L139  exit=0    `_templ.go` 생성이
modules.md:L22        exit=0    `a-h/templ` 컴파일 뷰(`*_templ.go`)
modules.md:L86        exit=0    두 표면이 공유하는 설정 스키마
modules.md:L23        exit=0    세그먼트 조립
modules.md:L33        exit=0    템플릿 재배포
modules.md:L38        exit=0    외부 에이전트 백엔드 런처
modules.md:L56        exit=0    스킬 미러
data-flow.md:L93      exit=0    심볼릭 링크 우회 미러링
entry-points.md:L50   exit=0    migration, chain, harness-router
docs-truth.md:L80     exit=0    `chain`
```

집합 대조도 함께 돌렸다 — `candidates.txt` 와 판정 표의 단위 집합이 양방향으로 동일하다(`comm -23`/`comm -13` 모두 무출력).

---

## §② 재생성 + 편입 (REQ-CM2-003 / REQ-CM2-004 / AC-CM2-003 / AC-CM2-004)

### ②-a 사본 선행 (M2.0) — 재생성의 선행 조건

사본은 M1 단계에서 이미 떴다(판정 인용 좌표를 재생성 후에도 해석 가능한 파일에 고정하기 위해). 재생성 직전 재확인:

```
$ /bin/ls .moai/reports/t475/pre-regen/ | wc -l
       6
$ /bin/ls .moai/reports/t475/pre-regen/
data-flow.md
dependencies.md
docs-truth.md
entry-points.md
modules.md
overview.md
```

**측정 주의**: 이 셸의 프로필이 `ls` 를 `ls -la` 로 alias 하므로 aliased `ls | wc -l` 은 `total` 행과 `.`/`..` 을 함께 세어 `9` 를 낸다. 판정에 쓴 것은 unaliased `/bin/ls` 다.

### ②-b 생성기 5문서 재생성 (M2.1)

`/moai codemaps --force` 의 실행면(`.claude/skills/moai/workflows/codemaps.md` Phase 3)이 선언하는 산출은 **5개**다 — `overview.md` · `modules.md` · `dependencies.md` · `entry-points.md` · `data-flow.md`. 다섯 전부를 현재 트리(HEAD `52f863f36`, described_roots `[internal, cmd, pkg]`)에서 재측정해 다시 썼다.

```
$ /bin/ls .moai/project/codemaps/
data-flow.md
dependencies.md
docs-truth.md
entry-points.md
modules.md
overview.md
provenance.json
```

7항목(문서 6 + `provenance.json`). `docs-truth.md` 는 이 재생성의 산출물이 **아니며** §③ 이 별도로 담당한다.

재측정된 주요 값(전부 이 트리에서 실행한 명령의 출력):

| 값 | 직전 판(앵커 `25a3212a9`) | 이 판(`52f863f36`) |
|---|---|---|
| 비테스트 Go 파일 | 1096 | **1114** |
| 테스트 Go 파일 | 1771 | **1812** |
| Go 패키지 총수 | 137 | **139** |
| 임베드 템플릿 파일 | 564 | **581** |
| 최상위 집계 import 엣지 | 205 | **208** |
| `internal/cli` fan-out | 58 | **59** |
| `internal/core` fan-in | 6 | **7** |
| `internal/statusline` fan-out | 7 | **8** |
| `rootCmd.AddCommand` (root.go `init()`) | 26 | **29** |
| `AddCommand` 호출 파일 수 | 64 | **65** |
| `mcp_server.go` 의 `add(...)` | 28 | **29** |
| go.mod direct require | 30 | **29** |

**새 발견 — 직전 판의 두 진술이 이 트리에서 거짓이다.** 둘 다 재생성 문서 본문에 정정으로 실었다.

1. **패키지 단위 엣지 1638 은 재현되지 않는다.** 이 판의 명령(`go list -f '{{.ImportPath}} {{join .Imports " "}}' ./...` + 모듈 경로 필터)은 **345** 를 낸다. 직전 판이 인용한 명령 문자열이 생략형(`'{{range .Imports}}...'`)이라 무엇을 셌는지 복원할 수 없으므로, 이 판은 명령과 그 출력을 함께 싣고 차이를 명시했다(`overview.md` · `dependencies.md`).
2. **"`internal/hook/session_start.go`(67KB)가 트리 최대 비테스트 Go 파일"은 더 이상 사실이 아니다.** 실측 상위는 `internal/web/fieldsets_templ.go` 168KB · `screens_templ.go` 121KB(둘 다 `templ` 생성 산물)이고, `session_start.go` 는 61KB 로 5위권이다. `modules.md` 에 생성/손저작을 갈라 세는 절을 새로 넣었다.

### ②-c 편입 확인 (M2.3) — omission 15개 전수

```
$ while read -r u; do n=$(/usr/bin/grep -rl -F "$u" .moai/project/codemaps/ | wc -l); printf '%-45s %s\n' "$u" "$n"; done < omissions
internal/settings/yamlpatch                   3
internal/stateanchor                          4
internal/template/agentemit                   4
internal/template/commandemit                 4
internal/chain                                4
internal/cli/codex_skills_disable.go          1
internal/cli/codex_skills_prune.go            1
internal/cli/integration_settings_drift.go    3
internal/cli/skills.go                        1
internal/cli/update_mirror_heal.go            2
internal/kanban/settings_drift.go             2
internal/statusline/state_anchor.go           2
internal/template/published_skills.go         2
internal/template/skill_mirror_repair.go      2
internal/web/codexmirror.go                   2
```

**15/15 이 1개 이상. 0 은 없다.**

**직접 보정한 단위 3개**: 첫 통과에서 `internal/cli/codex_skills_disable.go` · `internal/cli/codex_skills_prune.go` · `internal/cli/skills.go` 가 **0** 이었다 — `modules.md` 의 클러스터 표에 파일명만(`codex_skills_disable.go` · `skills.go`) 적어 전체 경로가 문자열로 존재하지 않았기 때문이다. 세 자리를 전체 경로 표기로 고쳐 편입을 성립시켰다. 나머지 12개는 재작성 과정에서 편입됐다.

편입 위치 요약:

| 단위 | 편입된 곳 |
|---|---|
| `internal/chain` | `modules.md` data/persistence 표 신규 행 + `data-flow.md` §F 신규 경로 |
| `internal/stateanchor` · `internal/statusline/state_anchor.go` | `modules.md` cross-cutting 표 + 전용 절 + `data-flow.md` §E 신규 경로 + `overview.md` 레이어 표 |
| `internal/settings/yamlpatch` | `modules.md` 신규 절 "보존 쓰기 경로" + `data-flow.md` §G |
| `internal/template/agentemit` · `commandemit` · `published_skills.go` · `skill_mirror_repair.go` | `modules.md` 신규 절 "템플릿 방출·미러 계열" + `data-flow.md` §C(빌드타임 + 재배포 없는 경로) + `entry-points.md` |
| `internal/cli/update_mirror_heal.go` | `modules.md` `update*` 클러스터 행 + `data-flow.md` §C |
| `internal/cli/skills.go` · `codex_skills_disable.go` · `codex_skills_prune.go` | `modules.md` `skills*` · `codex*` 클러스터 행 + `entry-points.md` root 등록 목록 |
| `internal/cli/integration_settings_drift.go` · `internal/kanban/settings_drift.go` | `modules.md` `integration*` 클러스터 + kanban 행 + 중복 능력 절 + `entry-points.md` + `data-flow.md` §G |
| `internal/web/codexmirror.go` | `modules.md` 신규 절 "codex 미러 탭" + `entry-points.md` 웹 콘솔 표면 |

---

## §③ `docs-truth.md` 손 갱신 (REQ-CM2-014 / AC-CM2-003a)

**이 섹션은 §② 와 분리돼 있다.** `docs-truth.md` 는 생성기 산출물이 아니므로 §②-b 의 재생성이 이 파일을 건드리지 않으며, 합쳐 적으면 "재생성했으니 이것도 됐다"는 오독이 그대로 통과한다.

근거(실측, 이 트리):

```
$ /usr/bin/grep -n "docs-truth" .claude/skills/moai/workflows/codemaps.md
(무출력, exit 1)
```

즉 실행면 문서 전체에 `docs-truth` 가 0회 등장하며, `/moai codemaps --force` 는 이 파일을 만들지 않는다.

### ③-a 갱신 사실

`docs-truth.md` 를 손으로 갱신했다. 갱신 범위:

- 머리말에 **[HARD] 이 파일은 생성기 산출물이 아니다** 블록과 마지막 손 갱신 좌표(2026-09-08, `52f863f36`)를 신설.
- §1 에이전트 카탈로그 — 전수 대조 표 신설(아래 ③-b).
- §2 상태 enum — 재검증 날짜·HEAD 갱신(값 8 불변).
- §3 프론트매터 12필드 — 재검증 + 실제 슬라이스 좌표 `internal/spec/lint.go:982-998` 로 갱신(직전 판의 `956-971` 은 낡음).
- §4.1 CLI verb 표 — **렌더의 실제 그룹으로 교체**(아래 ③-c).
- §4.2 `/moai` 명령 16개 — 재검증 + `commandemit` 발행 관계 한 줄 추가.
- §5 GLM 티어 — 재검증 + 상수별 실제 행 번호로 갱신.

### ③-b §1 에이전트 카탈로그 전수 대조 (표본 추출 없음)

```
$ find .claude/agents/moai -maxdepth 1 -name '*.md' | wc -l
      11
$ find .claude/agents -maxdepth 2 -name '*.md' | sort
.claude/agents/harness/cli-template-specialist.md
.claude/agents/harness/hns-github-specialist.md
.claude/agents/harness/hns-oss-docs-content-author-specialist.md
.claude/agents/harness/hns-oss-docs-locale-translator-specialist.md
.claude/agents/harness/hns-oss-docs-structure-curator-specialist.md
.claude/agents/harness/hns-release-specialist.md
.claude/agents/harness/hns-release-update-specialist.md
.claude/agents/harness/hook-ci-specialist.md
.claude/agents/harness/quality-specialist.md
.claude/agents/harness/workflow-specialist.md
.claude/agents/moai/builder-harness.md
.claude/agents/moai/e2e-tester.md
.claude/agents/moai/manager-design.md
.claude/agents/moai/manager-develop.md
.claude/agents/moai/manager-docs.md
.claude/agents/moai/manager-git.md
.claude/agents/moai/manager-lead.md
.claude/agents/moai/manager-spec.md
.claude/agents/moai/plan-auditor.md
.claude/agents/moai/super-advisor.md
.claude/agents/moai/sync-auditor.md
```

**대조 결과: 11/11 일치, 양방향 잉여 0.** `.claude/agents/moai/` 의 11개 파일 각각이 §1 표의 행 1-11 과 이름으로 일대일 대응하고(대조 표를 `docs-truth.md` §1 에 실었다), 표에만 있는 행은 12행 `Explore` 하나인데 이것은 파일이 없는 Anthropic built-in 이므로 트리 부재가 정상이다. 같은 스캔이 보여주는 `.claude/agents/harness/` 10개는 user-owned harness specialist 이며 retained catalog 의 원소가 아니다 — 이 사실도 §1 에 적었다.

### ③-c §4.1 정정 — 직전 판의 그룹 분류는 렌더에 존재하지 않았다

`./bin/moai --help` 의 실제 출력은 **COMMANDS / LAUNCH COMMANDS / PROJECT COMMANDS / TOOLS** 4개 그룹이다. 직전 판이 실은 5분류(Project / Launchers / Autonomous-Dev / Governance / Tools-Infra)는 렌더에 없는 손 분류였고, 그 표는 **`skills` 를 빠뜨렸다** — `skills` 는 `internal/cli/root.go:177` 의 `newSkillsCmd()` 로 등록된 라이브 루트 명령이다. 이 판은 렌더의 4개 그룹을 그대로 옮기고 정정 사실을 문서 안에 남겼다.

부수 census 갱신(전부 이 트리 실측): `.AddCommand(` 비테스트 호출 202 → **207**, `rootCmd.AddCommand(` 60 → **62**(그중 29가 `root.go` 의 `init()` 안), `internal/cli` 비테스트 파일 264 → **279**.

---

## §④ 구간별 전후 대조 — `diff -u` (REQ-CM2-005 / AC-CM2-005)

구간 목록은 손으로 열거하지 않고 §A.3(b) 컷오프 명령의 출력을 채택했다:

```
$ awk '{print $2}' .moai/reports/t475/described-roots-diff-since-anchor.txt \
    | xargs -n1 dirname | sort | uniq -c | sort -rn \
    | awk '$1>=6 && $2 !~ /^internal\/template\/templates\//'
  59 internal/cli
  13 internal/web
  12 internal/statusline
  11 internal/kanban
  10 internal/template
   6 internal/settings
   6 internal/codexwiring
```

**7구간 — 저작 시점과 동일**(`internal/settings` 6 포함).

문서 전체의 전후 규모(`diff -u .moai/reports/t475/pre-regen/<doc>.md .moai/project/codemaps/<doc>.md | wc -l`):

```
overview.md      115
modules.md       282
dependencies.md  169
entry-points.md  135
data-flow.md     198
docs-truth.md    120
```

아래는 구간마다, 6문서 전체의 `diff -u` 에서 그 구간 경로를 담은 **변경 행만** 뽑은 것이다:

```
$ for d in overview modules dependencies entry-points data-flow docs-truth; do
    diff -u ".moai/reports/t475/pre-regen/$d.md" ".moai/project/codemaps/$d.md" \
      | /usr/bin/grep -E '^[-+][^-+]' | /usr/bin/grep -F "<구간>"
  done
```

########## AREA internal/cli
--- overview.md
+  `internal/cli/hook.go`가 61KB, `internal/hook/quality/gate.go`가 53KB입니다. 이들은
--- modules.md
-| `internal/cli` | 273 | 아래 클러스터 표 참조 | `update`(+`plan`/`deploy`/`merge`/`backup`/`report`), `harness`, `worktree`, `agentlint`, `preference`, `wizard`, `uikit`, `printer`, `specid`, `taskledger`, `pr` |
+| `internal/cli` | 279 | 아래 클러스터 표 참조 | `update`(+`plan`/`deploy`/`merge`/`backup`/`report`), `harness`, `worktree`, `agentlint`, `preference`, `wizard`, `uikit`, `printer`, `specid`, `taskledger`, `pr` |
+| `update*` | 23 | 템플릿 재배포 — 계획/분류/네임스페이스 보호, 3-way 머지, 백업·롤백, 클린 인스톨, dry-run. 단계 로직은 `cli/update/{plan,deploy,merge,backup,report}` 하위로 분해돼 있다. **재배포가 일어나지 않는 경로에도 복구 하나가 붙는다** — `internal/cli/update_mirror_heal.go`는 버전 일치 update가 Deploy 앞에서 조기 반환하는 자리 옆에서 `.agents/skills` 미러를 복구하며, 존재 게이트는 프로젝트의 기록된 배포 버전이다 |
+| `doctor*` | 15 | 진단 — config, disk, harness, hook wiring, mcp version, permission, sandbox, skills, worktree base, agentemit embed, codex. hook wiring 점검의 실체는 `internal/cli/doctor_hook_delivery.go`로, 배포 템플릿의 훅 엔트리를 프로젝트 `.claude/settings.json`과 대조해 누락분을 배치·처방과 함께 보고하며 **사용자 파일을 쓰지 않는다** |
+| `codex*` | 10 | 외부 에이전트 백엔드 런처, 잡 제어, 준비 상태 점검, 리뷰 게이트. **여기에 사용자 HOME 계층에 대한 스킬 노출 제어 두 개가 함께 산다** — `internal/cli/codex_skills_disable.go`는 `~/.codex/config.toml`에 `enabled = false`를 실은 `[[skills.config]]` 항목을 발행하고, `internal/cli/codex_skills_prune.go`는 가리키는 파일이 사라진 유령 등록을 제거한다(부재를 증명할 수 있는 것만 지우는 allowlist 형 판정, 기본 dry-run) |
+| `skills*` | 1 | `moai skills` 명령 트리(`internal/cli/skills.go`). 스킬 노출을 **계층별** 관심사로 두고 계층을 verb 가 아니라 플래그로 명명하며, `--codex`를 필수로 만들어 사용자 HOME 쓰기를 호출 시점 opt-in으로 고정한다 |
+| **`internal/git`** (루트 패키지) | **이 판에서 새로 잡혔다.** `core/git` 위의 상위 유틸리티 8 파일인데 루트 패키지를 import 하는 비테스트 코드가 0이다. 실제로 import 되는 것은 하위 `internal/git/convention` 하나뿐이며(`internal/cli` → `internal/git/convention`), 최상위 집계 fan-in 1은 그것이다. 소비자가 `internal/core/git`로 직접 내려가면서 중간 계층만 남은 모양으로 읽힌다 — 확인이 필요한 관찰이며, 이 문서가 답을 주지는 않는다 |
+| `internal/cli/taskledger` · `internal/lsp/aggregator` · `internal/hook/testutil` · `internal/timing` · `internal/tui/golden` | 테스트 전용 소비자만 갖는 leaf. 앞의 셋은 의도로 보이고, `timing`은 이름이 그것을 말한다 |
-  헬퍼, 후자는 `Write` + guard입니다. `internal/cli`의 6개 파일이 후자를, 나머지 트리가 전자를
+  `internal/cli/integration_settings_drift.go`(CLI 절반). 이것은 중복이 아니라 의도된 분할이며,
+ 89KB internal/cli/mcp_codex.go            (손 저작 — CLI 최대)
--- dependencies.md
+| `internal/stateanchor` | 2 | 상태 앵커 seam. 소비자는 `internal/statusline`과 `internal/cli` |
-| 1 | `internal/cli` | **58** |
+| 1 | `internal/cli` | **59** |
-`internal/cli`가 최상위 68개 중 **58개**를 import 합니다 — 사실상 전 트리에 닿습니다.
-합성 루트(`internal/cli/deps.go`)가 여기 있으므로 일부는 의도된 것이지만, 58 중 상당수는
+`internal/cli`가 최상위 68개 중 **59개**를 import 합니다 — 사실상 전 트리에 닿습니다.
+합성 루트(`internal/cli/deps.go`)가 여기 있으므로 일부는 의도된 것이지만, 59 중 상당수는
-{{end}}' ./internal/cli/... \
-  | awk -F/ '{print $1"/"$2}' | sort -u | grep -v '^internal/cli$' | wc -l
-| `github.com/charmbracelet/huh` v1.0.0 | **v2와 병존하는 v1 폼** | `internal/cli` 5개 파일 |
-| `github.com/charmbracelet/lipgloss` v1.1.1-… | **v2와 병존하는 v1 스타일링** | `internal/statusline` 3개 파일, `internal/cli` 2개 |
+| `github.com/charmbracelet/huh` v1.0.0 | **v2와 병존하는 v1 폼** | `internal/cli` |
+| `github.com/charmbracelet/lipgloss` v1.1.1-… | **v2와 병존하는 v1 스타일링** | `internal/statusline`, `internal/cli` |
--- entry-points.md
+`internal/cli/integration_settings_drift.go`가 tracked `.claude/settings.json`의 워킹 사본
+`moai doctor`의 hook delivery 점검(`internal/cli/doctor_hook_delivery.go`)이 프로젝트가 이미
--- data-flow.md
+internal/cli/update_mirror_heal.go      그 조기 반환 자리 옆에서 실행
-internal/cli/mcp_server.go              add(name, mcp.NewTool(...), handler) × 28
+internal/cli/mcp_server.go              add(name, mcp.NewTool(...), handler) × 29
+internal/cli/statusline.go              렌더 진입
+  B4  CLI 설정 캐시 사슬          internal/cli
+internal/cli (moai chain)              조회 표면
+  └ internal/cli/integration_settings_drift.go   CLI 절반 — 두 표면의 배선
--- docs-truth.md
-**Source:** `moai --help` rendered output (2026-09-02, built from this tree) + `grep -rn '\.AddCommand(' internal/cli/ --include='*.go' | grep -v _test` (202 non-test calls) + `grep -rn 'rootCmd\.AddCommand(' internal/cli --include='*.go' | grep -v _test | wc -l` (60 root registrations, 33 files) + `find internal/cli -name '*.go' ! -name '*_test.go' | wc -l` (264).
+**Source:** `./bin/moai --help` rendered output (2026-09-08, built from this tree at HEAD `52f863f36`) + `grep -rn '\.AddCommand(' internal/cli/ --include='*.go' | grep -v _test | wc -l` (**207** non-test calls) + `grep -rn 'rootCmd\.AddCommand(' internal/cli --include='*.go' | grep -v _test | wc -l` (**62** root registrations across the package; **29**의 `rootCmd.AddCommand`가 `internal/cli/root.go:143-269`의 `init()` 안에 있다) + `find internal/cli -name '*.go' ! -name '*_test.go' | wc -l` (**279**).
########## END internal/cli
########## AREA internal/web
--- modules.md
-| `internal/web` | 29 | 루프백 전용 브라우저 콘솔. `a-h/templ` 컴파일 뷰(`*_templ.go`) + htmx + SSE(fsnotify)로 프로파일·설정·todo 큐를 편집 | `assets` |
+| `internal/web` | 31 | 루프백 전용 브라우저 콘솔. `a-h/templ` 컴파일 뷰(`*_templ.go`) + htmx + SSE(fsnotify)로 프로파일·설정·todo 큐를 편집하고, **codex 탭 하나는 편집이 아니라 읽기 전용 미러**다(§ codex 미러 탭) | `assets` |
+### codex 미러 탭 — `internal/web`의 편집하지 않는 표면
+`internal/web/codexmirror.go`는 codex 탭의 **행 모델**이며, Audit·MCP 탭에 사는 codex 설정의
+렌더 쪽 `internal/web/fieldsets_codex_templ.go`는 `a-h/templ`이
+`internal/web/fieldsets_codex.templ`에서 생성한 산물입니다(`// Code generated by templ - DO NOT EDIT.`).
+168KB internal/web/fieldsets_templ.go      (생성)
+121KB internal/web/screens_templ.go        (생성)
--- dependencies.md
+   생성물이 트리에 커밋돼 있습니다 — `internal/web/fieldsets_codex_templ.go`(codex 미러 패널)와
--- entry-points.md
+codex 탭(`internal/web/codexmirror.go` 행 모델 +
+`internal/web/fieldsets_codex_templ.go` 렌더)은 Audit·MCP 탭에 사는 codex 설정의 읽기 전용
########## END internal/web
########## AREA internal/statusline
--- modules.md
-| `internal/statusline` | 20 | Claude Code statusLine 렌더러. git·github·model·backlog·goal·usage 세그먼트 조립 | — |
+| `internal/statusline` | 21 | Claude Code statusLine 렌더러. git·github·model·backlog·goal·usage 세그먼트 조립. 렌더가 읽고 쓰는 상태의 **앵커는 세션의 현재 디렉터리가 아니라** `internal/stateanchor` seam이 정한 프로젝트 루트이며, 그 어댑터가 `internal/statusline/state_anchor.go`다 | — |
+`internal/statusline/state_anchor.go`입니다.
--- dependencies.md
-다만 7위 `internal/hook`(6)과 10위 `internal/statusline`(5)은 **presentation인데 피의존
+다만 8위 `internal/hook`(6)과 10위 `internal/statusline`(5)은 **presentation인데 피의존
+| `internal/stateanchor` | 2 | 상태 앵커 seam. 소비자는 `internal/statusline`과 `internal/cli` |
-| 5 | `internal/statusline` | 7 |
+| 5 | `internal/statusline` | 8 |
+`internal/statusline`이 7 → 8로 오른 것도 같은 seam 때문입니다 — 렌더의 상태 앵커가
-| `github.com/charmbracelet/lipgloss` v1.1.1-… | **v2와 병존하는 v1 스타일링** | `internal/statusline` 3개 파일, `internal/cli` 2개 |
+| `github.com/charmbracelet/lipgloss` v1.1.1-… | **v2와 병존하는 v1 스타일링** | `internal/statusline`, `internal/cli` |
--- data-flow.md
+internal/statusline/state_anchor.go     resolveStateAnchor(...)  ← statusline 쪽 어댑터
########## END internal/statusline
########## AREA internal/kanban
--- modules.md
-| `internal/kanban` | 33 | 백로그 큐의 상태 레코드·컬럼·역할 모델, SQLite 저장 엔진, 보드 락, PR 링크, 정합성 조정 | — |
+| `internal/kanban` | 35 | 백로그 큐의 상태 레코드·컬럼·역할 모델, SQLite 저장 엔진, 보드 락, PR 링크(착지 ref 3단 해석 사슬 `prlink_landedref.go` 포함), 정합성 조정. **여기에 워킹 트리 검사 하나가 더 있다** — `settings_drift.go`가 병합 전 tracked `.claude/settings.json`의 워킹 사본 드리프트를 단정하고 사본을 보존하며 원장에 남긴다(`--no-optional-locks` 강제 — 평범한 status가 인덱스 쓰기 락을 잡아 병합 직전 경합을 스스로 만들기 때문) | — |
--- dependencies.md
-   `internal/kanban/backlog_sqlite.go`가 드라이버와 상수를 직접 씁니다 — `go mod tidy`가 아직
+   `internal/kanban/backlog_sqlite.go`가 드라이버(`_ "modernc.org/sqlite"`)와 상수
--- data-flow.md
+      └ internal/kanban/settings_drift.go        도메인 절반 — 검출 · 보존 · 원장
########## END internal/kanban
########## AREA internal/template
--- overview.md
-| 임베드 템플릿 파일 | 564 | `find internal/template/templates -type f \| wc -l` |
+| 임베드 템플릿 파일 | 581 | `find internal/template/templates -type f \| wc -l` |
--- modules.md
-| `internal/template` | 25 | `//go:embed all:templates` + `catalog.yaml`. 배포기, 렌더러, settings 생성, 스킬 미러, 카탈로그 트리 해시, 모델 정책·프로파일 매트릭스 | `agentemit`, `scripts` |
+| `internal/template` | 30 | `//go:embed all:templates` + `catalog.yaml`. 배포기, 렌더러, settings 생성, 스킬 미러, 카탈로그 트리 해시, 모델 정책·프로파일 매트릭스. **배포 뒤편에 두 개의 기계 방출기와 두 개의 미러 보호·복구 seam이 붙어 있다**(§ 템플릿 방출·미러 계열) | `agentemit`, `commandemit`, `scripts` |
+`internal/template`의 책임 칸 한 줄로는 담기지 않는 네 단위가 하위에 있습니다. 넷 다
+| `internal/template/agentemit` | 6 | 보존된 에이전트 정의(`.md`)와 임베드 매니페스트(`agents-codex.yaml`)의 쌍을 **중립 원본**으로 삼아 `.codex/agents/` TOML을 결정적으로 이중 발행한다. `.md`의 발행은 항등(identity)이라 재렌더·재정렬이 없고, Codex 쪽은 (`.md` × 매니페스트)의 결정적 변환이다. **fail-closed** — 알 수 없는 tool 토큰·미매핑 effort·유효하지 않은 sandbox 값이면 어느 파일의 어느 토큰인지 지목하며 실패하고 부분 산출물을 남기지 않는다(codex-cli가 알 수 없는 설정을 조용히 무시하므로 생성기 쪽이 자기 출력을 검증해야 한다) |
+| `internal/template/commandemit` | 3 | `/moai` 명령 소스를 codex 스킬 아티팩트(`.agents/skills/moai-<command>/SKILL.md`)로 발행한다. 명령 소스는 읽기 전용으로 소비하며 **본문은 바이트 동일 verbatim** — 본문에 남은 Claude 전용 도구 참조는 여기서 고치지 않고 경계 플래그로만 기록한다(그 수리는 명령 본문 계층 소관). fail-closed: 프론트매터 구분자 누락, 설명 누락, 무조건 분기 없는 로케일 조건부 설명, 기존 정본 스킬 디렉터리와 충돌하는 파생 이름 |
+| `internal/template/published_skills.go` | (파일) | 위 발행 스킬 경로에 대한 **배포측 보호**. 발행 스킬은 보통의 템플릿 파일처럼 배포되지만 경로 네임스페이스가 스킬 미러가 쓰는 `.agents/skills` 루트와 겹치고, update 모드(forceUpdate)는 다른 곳에서 provenance 검사를 건너뛴다. 이 검사가 그 경로들에 한해 init 모드의 provenance 동작을 살려 사용자 소유 파일이 update를 살아남게 하고, 건너뜀을 침묵이 아니라 보고로 남긴다 |
+| `internal/template/skill_mirror_repair.go` | (파일) | `.agents/skills`의 두 생산자(심볼릭 링크 미러, 발행 SKILL.md) 결과를 **Deploy 없이** 복구하는 패키지 수준 패스. DeployerOption이 아닌 형상을 의도적으로 골랐다 — 옵션이었다면 배포 경로에서도 살아나 수리 기능의 부작용으로 배포 동작이 바뀐다. 항목별 의미는 미러 생산자의 것을 재사용하므로 생산자와 갈라질 수 없다 |
+두 방출기는 **비테스트 코드에서 아무도 import 하지 않습니다**(`internal/template/agentemit`,
+`internal/template/commandemit` 둘 다 패키지 단위 fan-in 0). 소비자는 빌드 타깃
+| `internal/template/agentemit` · `internal/template/commandemit` | **고아가 아니다.** 소비자가 `make agents-emit` / `make commands-emit` 빌드 타깃과 골든 테스트다. 방출기는 빌드타임 도구이므로 런타임 fan-in 0이 정상 상태다 |
--- dependencies.md
-| 7 | `internal/template` | 6 | domain |
+| 8 | `internal/template` | 6 | domain |
+두 방출기(`internal/template/agentemit`, `internal/template/commandemit`)는 이 표에 **나타나지
--- entry-points.md
+통해 실행됩니다 — `internal/template/agentemit`(`make agents-emit`, `.md` × 매니페스트 →
+`.codex/agents/*.toml`)와 `internal/template/commandemit`(`make commands-emit`,
--- data-flow.md
-internal/template/embed.go                            //go:embed all:templates   (564개 파일)
+internal/template/embed.go                            //go:embed all:templates   (581개 파일)
+internal/template/agentemit                           make agents-emit
+internal/template/commandemit                         make commands-emit
-internal/template/skill_mirror.go       심볼릭 링크 우회 미러링
+internal/template/published_skills.go   발행 스킬 경로(.agents/skills/moai-<command>/SKILL.md)에
+internal/template/skill_mirror.go       심볼릭 링크 우회 미러링 (Deploy 마지막 단계)
+  └ internal/template/skill_mirror_repair.go
--- docs-truth.md
+이 16개 소스는 codex 쪽으로도 발행됩니다 — `internal/template/commandemit`이 각각을
########## END internal/template
########## AREA internal/settings
--- overview.md
-| data/persistence | 디스크상 named artifact 하나의 스키마와 읽기·쓰기 계약을 소유한다 | `internal/config`, `internal/session`, `internal/settings`, `internal/manifest` … |
+| data/persistence | 디스크상 named artifact 하나의 스키마와 읽기·쓰기 계약을 소유한다 | `internal/config`, `internal/session`, `internal/settings`, `internal/manifest`, `internal/chain` … |
--- modules.md
+### `internal/settings/yamlpatch` — 보존 쓰기 경로
+`internal/settings`의 책임은 "두 표면이 공유하는 **스키마**"지만, `yamlpatch`가 지는 것은
--- dependencies.md
-| 5 | `internal/settings` | 7 |
+| 6 | `internal/settings` | 7 |
+| `gopkg.in/yaml.v3` v3.0.1 | 설정·카탈로그·프론트매터 파싱 + **노드 트리 수술**(`internal/settings/yamlpatch`) | 트리 전역 |
+   `internal/settings/yamlpatch`는 같은 라이브러리의 **노드 트리**를 직접 수술해 주석과
--- data-flow.md
+internal/settings/*                     두 표면(moai web 콘솔 / moai profile setup TUI)이
+  └ Save() 경로가 없는 8개 섹션          internal/settings/yamlpatch
########## END internal/settings
########## AREA internal/codexwiring
########## END internal/codexwiring

### ④-a `internal/codexwiring` — 변경 없음, 빈 diff 를 증거로 첨부

위 블록의 `########## AREA internal/codexwiring` 과 `########## END internal/codexwiring` 사이는 **비어 있다**. 그 빈 출력이 이 행의 증거다 — 산문만 적힌 "변경 없음"은 반증 불가능하므로 쓰지 않는다.

```
$ for d in overview modules dependencies entry-points data-flow docs-truth; do
    diff -u ".moai/reports/t475/pre-regen/$d.md" ".moai/project/codemaps/$d.md" \
      | /usr/bin/grep -E '^[-+][^-+]' | /usr/bin/grep -F "internal/codexwiring"
  done
(무출력)
```

**왜 변경이 없는가**: 앵커 이후 그 디렉터리의 파일 6개가 움직였지만, 비테스트 파일 수(5)도 패키지가 지는 책임("Codex 측 배선 파일 생성·갱신")도 재측정에서 그대로였다. 구간이 컷오프에 걸린 것은 변경 **양** 때문이고, 서술이 바뀌지 않은 것은 그 변경이 서술 수준의 사실을 옮기지 않았기 때문이다. 컷오프는 **무엇을 들여다볼지**를 정할 뿐 서술 변경을 강제하지 않는다.

구간별 변경 행 수:

| 구간 | 변경 파일 수(앵커 대비) | 6문서에서 그 구간을 담은 변경 행 |
|---|---|---|
| `internal/cli` | 59 | 42 |
| `internal/web` | 13 | 14 |
| `internal/statusline` | 12 | 15 |
| `internal/kanban` | 11 | 8 |
| `internal/template` | 10 | 32 |
| `internal/settings` | 6 | 14 |
| `internal/codexwiring` | 6 | **0** (빈 diff 첨부, 위) |

---

## §④-b 사후 정정 — fold 판정 단위의 산문을 되돌렸다

§⑥(패키지 대조)을 처음 돌렸을 때 **후보 20개 중 히트-0 으로 남은 것이 0개**였다. 그것이 결함의 신호였다 — §A.3(a1) 은 fold 판정 단위에 대해 **"codemaps 산문을 바꾸지 않는다"** 고 규정하고, AC-CM2-007 은 잔여 히트-0 목록에 fold 단위가 판정과 함께 **남아 있을 것**을 전제한다. 첫 통과의 재생성은 fold 5개 전부에 대해 산문을 새로 썼고, 그것은 판정을 편입 허가로 오용한 형태다.

되돌린 내용(모두 fold 단위에 **대한** 서술만 제거하고, 같은 문단의 omission 단위 서술은 유지):

| fold 단위 | 첫 통과에서 넣었던 것 | 처분 |
|---|---|---|
| `internal/core/git` | `modules.md` infrastructure 표 신규 행 + `overview.md` 레이어 표 등재 + 전체 경로 표기 6곳 | 신규 행·등재 삭제, 표기를 직전 판의 `core/git` 형태로 환원 |
| `internal/cli/doctor_hook_delivery.go` | `modules.md` `doctor*` 클러스터 행 확장 + `entry-points.md` 훅 절 문단 | 둘 다 삭제(클러스터 행은 직전 판의 열거로 복귀) |
| `internal/hook/quality/step_git_env.go` | `data-flow.md` §B 하위 3행 + `entry-points.md` "게이트 스텝의 실행 환경" 문단 | 둘 다 삭제 |
| `internal/kanban/prlink_landedref.go` | `modules.md` kanban 행 괄호 삽입구 | 삭제 |
| `internal/web/fieldsets_codex_templ.go` | `modules.md` · `entry-points.md` · `dependencies.md` 3곳의 파일명 인용 | 파일명 인용만 삭제, 생성 관계 서술은 유지(직전 판에도 있던 사실) |

되돌린 뒤 재확인:

```
$ /usr/bin/grep -rc -F "<fold 단위>" .moai/project/codemaps/ | grep -v ':0$'
(다섯 단위 모두 무출력 — 히트 0)
```

omission 15개는 되돌림의 영향을 받지 않았다(§⑥ 표 참조, 전부 ≥1 유지).

**이 항목을 지우지 않고 남기는 이유**: 되돌리기 전 상태에서도 12개 AC 는 전부 통과했을 것이다 — 어떤 AC 도 "fold 단위가 여전히 히트 0 인가"를 직접 묻지 않는다. AC-CM2-007 의 전제로만 간접적으로 걸리며, 그 간접성이 이 결함을 조용히 통과시킬 수 있는 경로다. 리드가 판단할 수 있도록 기록한다.

---

## §⑤ 인용 경로 실존 (REQ-CM2-006 / AC-CM2-006, accuracy a)

### ⑤-a 추출 규약 — `internal/graph/check_citations.go` 의 정본 3요소

| # | 요소 | 좌표 | 값 |
|---|---|---|---|
| 1 | 정규식 | `check_citations.go:23` | `\b(?:internal\|pkg\|cmd)/[A-Za-z0-9_/.-]*` |
| 2 | 후행 구두점 절삭 | `check_citations.go:35` | `citedPathTrailingPunct = ".,;:)]}\"'"` |
| 3 | blockquote 면제 | `positiveCitedPaths` (`:118-134`) | `strings.TrimSpace(line)` 이 `>` 로 시작하는 줄은 스캔에서 제외. **코드펜스·mermaid 는 면제가 아니다** |

정본을 직독하며 규약이 **3요소보다 넓다**는 것을 확인했고, 그 나머지도 적용했다 — `normalizeCitedPath`(`:136-155`)의 후행 슬래시 절삭, `cmdMainPathMap`(`cmd/moai/main` → `cmd/moai/main.go`), 그리고 `.go` 접미 복원(`…checkgo` → `…check.go`). 이 셋을 빠뜨리면 표가 게이트보다 많은 absent 를 보고해 교차 대조가 어긋난다.

```
$ for f in .moai/project/codemaps/*.md; do
    /usr/bin/grep -v '^[[:space:]]*>' "$f" \
      | /usr/bin/grep -oE '(internal|pkg|cmd)/[A-Za-z0-9_/.-]*'
  done > /tmp/cited_raw.txt        # 정규식 + blockquote 면제
$ (후행 구두점 절삭 → 후행 슬래시 절삭 → cmd/moai/main 매핑 → .go 접미 복원 → 유니크)
```

### ⑤-b 결과

```
raw tokens        373
unique raw        180
unique normalized 175
absent              0
```

**전수 표는 유니크 정규화 경로 175개이며 absent 는 0건이다.** 목록 전문은 `.moai/reports/t475/cited-paths-table.txt` 로 수출했다(경로 → exists/absent, 175행).

### ⑤-c 게이트 계층과의 교차 대조

```
$ ./bin/moai graph check --json | grep -A4 '"layer": "citations"'
      "layer": "citations",
      "metric": "positive-cited-path-absence",
      "value": 0,
      "threshold": 0,
      "verdict": "fresh"
```

**표의 absent 0 = 계층의 value 0.** 두 수가 일치하므로 추출이 규약을 벗어나지 않았다.

### ⑤-d 이 검증이 잡아낸 실제 회귀 1건

첫 통과에서 표는 **absent 1건**을 냈다 — `cmd/templ`. 내가 `dependencies.md` 에 `tool github.com/a-h/templ/cmd/templ` 를 그대로 적었고, 정본 정규식이 그 모듈 경로 안의 `cmd/templ` 를 인용으로 집었기 때문이다. 트리에 `cmd/templ` 는 없으므로 이대로 두었으면 **citations 계층이 fresh → stale 로 뒤집혔을 것**이다.

수리는 blockquote 면제로 숨기지 않고 문장을 고쳤다 — `go.mod:106` 을 좌표로 남기고 모듈 경로 리터럴을 걷어냈다. 면제는 부존재를 **일부러** 인용한 줄을 위한 것이지, 실수로 만든 팬텀을 감추는 장치가 아니다.

---

## §⑥ 패키지 구조 대조 (REQ-CM2-007 / AC-CM2-007, accuracy b)

재생성 후 §A.3(a) 히트-0 패키지 명령을 다시 돌렸다.

```
$ cat .moai/project/codemaps/*.md > /tmp/cm3.txt
$ go list ./internal/... ./cmd/... ./pkg/... | sed 's|^[^/]*/[^/]*/[^/]*/||' \
    | while read -r p; do /usr/bin/grep -q -F "$p" /tmp/cm3.txt || echo "$p"; done | wc -l
      39
```

**48 → 39.** 편입으로 9개가 히트를 얻었다(`internal/chain`, `internal/stateanchor`, `internal/settings/yamlpatch`, `internal/template/agentemit`, `internal/template/commandemit` 및 파일 편입에 딸려 부모 경로 문자열이 생긴 것들).

잔여 39개 전수와 M1 판정:

| 패키지 | M1 판정 |
|---|---|
| `internal/core/git` | **fold** (후보였고 fold 로 판정 — 산문 무변경이 규정이므로 히트 0 유지가 정상) |
| `internal/cli/agentlint` · `internal/cli/harness` · `internal/cli/printer` · `internal/cli/uikit` · `internal/cli/worktree` | 후보 아님(앵커 이후 무변경) — 판정 대상 밖 |
| `internal/config/toolpolicy` · `internal/core/project` · `internal/graph/symbol` · `internal/runtime/gobin` · `internal/settings/agentfm` · `internal/tui/internal` | 후보 아님 — 판정 대상 밖 |
| `internal/harness/{capture,cluster,curator,delegationmap,proposalgen,router,routing,safety,seeds,throttle,tier,v4manifest}` (12) | 후보 아님 — 판정 대상 밖 |
| `internal/hook/{handoff,memo,memo/taxonomy,perf,trace}` (5) | 후보 아님 — 판정 대상 밖 |
| `internal/lsp/{cache,core,gopls,hook,subprocess}` (5) | 후보 아님 — 판정 대상 밖 |
| `internal/navigator/{detect,fix,route,sync,tiers}` (5) | 후보 아님 — 판정 대상 밖 |

**omission 판정 단위로서 여전히 히트 0 인 것은 0개다.** 즉 REQ-CM2-004 미이행이 없다. 기록 전용 처분이 적용된 것은 fold 판정 단위 `internal/core/git` 하나이며, 나머지 38개는 애초에 후보 집합에 들어오지 않은 단위(앵커 이후 무변경)로 REQ-CM2-013 ③ 의 이관 목록 소관이다.

**개수 자체는 합격 조건이 아니다**(§B.1 — 트리와 함께 움직이는 값). 판정은 위 표의 **분류가 존재하는가**로 이분한다.

---

## §⑦ 인용 식별자 hit/miss (REQ-CM2-008 / AC-CM2-008, accuracy c)

"식별자"의 정의는 REQ-CM2-008 이 명명한 **명령의 출력**이다:

```
$ for f in entry-points data-flow; do
    /usr/bin/grep -o '`[A-Za-z0-9_.]*`' ".moai/project/codemaps/$f.md" \
      | tr -d '`' | /usr/bin/grep -E '^([a-z][A-Za-z0-9_]*\.)?[A-Z][A-Za-z0-9_]*$'
  done | sort -u
```

**출력 10행 — 0행이 아니므로 빈 집합 위의 공허한 통과가 아니다**(재생성 전 `52f863f36` 에서도 10행이었다).

| 식별자 | 명명 위치(문서가 가리키는 곳) | 해석 결과 | hit/miss |
|---|---|---|---|
| `AddCommand` | `internal/cli` — root.go `init()` 29회 + 자기 등록 65 파일 | `internal/cli/migrate_restore_skill.go:110` 외 다수 | **hit** |
| `BacklogPathForRoot` | `internal/kanban` — 큐 경로 해석 | `internal/kanban/state_dir.go:129` `func BacklogPathForRoot(root string) string` | **hit** |
| `ExitCoder` | `cmd/moai/main.go` 종료 코드 매핑 seam | `internal/cli/exitcode.go:13` `type ExitCoder interface`(주석 `:11` — 원래 `cmd/moai/main.go` 에 있던 인터페이스가 매칭 규칙과 함께 이리로 옮겨졌다) | **hit** |
| `PreToolUse` | 훅 이벤트 이름 / `hook.EventPreToolUse` | `internal/hook/types.go:25` `EventPreToolUse EventType = "PreToolUse"` | **hit** |
| `RunE` | cobra 명령 필드 (`internal/cli`) | `internal/cli/migrate_restore_skill.go:106` 외 다수 | **hit** |
| `Shutdown` | `internal/hook/registry.go` 비동기 trace writer 플러시 배리어 | `internal/hook/registry.go:464` `func (r *registry) Shutdown()` | **hit** |
| `cli.ResolveExitCode` | `internal/cli` | `internal/cli/exitcode.go:35` `func ResolveExitCode(err error) (int, bool)` | **hit** |
| `hook.EventType` | `internal/hook` | `internal/hook/types.go:18` `type EventType string` | **hit** |
| `mcp.NewTool` | `internal/cli/mcp_server.go` 의 `add(...)` 첫 인자 계약 | `internal/cli/mcp_server.go:158` `add("session_list", mcp.NewTool(` | **hit** (서드파티 `mark3labs/mcp-go` 심볼, 명명 위치에서 실사용 확인) |
| `syscall.Exec` | `internal/cli` 런처(`cc`/`cg`/`glm`) 프로세스 교체 | `internal/cli/update.go:808` `return syscall.Exec(exe, os.Args, os.Environ())` | **hit** (stdlib 심볼; 런처 파일에도 존재하나 첫 비테스트 적중을 인용) |

**10 hit / 0 miss.** miss 는 기록만 하고 인용 본문을 지우지 않는 처분이나, 이번 실행에서는 miss 가 없다.
