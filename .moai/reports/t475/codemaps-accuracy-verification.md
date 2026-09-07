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
