# MoAI-ADK 하네스 전수 조사와 개선 계획

**기준:** 2026-09-23, 공유 기본 체크아웃 `main` / `2213871af`  
**범위:** 템플릿 배포 Go 코드와 CLI 연결부, `.claude/skills`, `.claude/rules/moai`, `CLAUDE.md`, `.agents/skills` 연결, 배포용 `internal/template/templates` 미러  
**판정:** 확인된 충돌·동작 6건, 재현 또는 정책 결정이 더 필요한 후보 2건. Codex CLI 완전 지원은 **미입증**이며, 정적 실행 계약과 설치 바이너리/소스 사이에 불일치가 확인됐다. 이 문서는 개선 계획이며 소스 수정·배포 완료 보고가 아니다.

## 1. 조사 방법과 적용 범위

세 조사자가 템플릿 코드, 워크플로 규칙, 스킬·Codex 연결을 나눠 읽었다. 루트에서는 CLI와 배포 경로, 파일 목록, 미러 차이, 좁은 테스트를 교차 확인했다. `.claude/worktrees/**`와 과거 생성 보고서는 현재 동작 집계에서 제외했다. 바이트 차이는 자동으로 결함으로 분류하지 않았다. 현재 저장소에는 다른 작업의 수정 파일이 있으며, 조사 중 그 파일을 수정하지 않았다.

`git ls-files` 기준 `internal/template/**` 682개, `.claude/skills/**` 333개, `.claude/rules/**` 87개, `CLAUDE.md` 1개를 목록화했다. 현재 파일과 배포용 템플릿의 대응 쌍 392개 중 274개가 바이트 동일, 118개가 달랐다. 이 수치는 **목록·동일성 조사 범위**이지 1,103개 파일의 모든 런타임 분기를 실행했다는 뜻이 아니다. 스킬 트리는 3,261,737바이트, 규칙 트리는 1,158,081바이트, `CLAUDE.md`는 20,019바이트다. 트리 전체 크기를 세션 토큰 비용으로 환산하지 않았다. 스킬은 호출할 때 읽히는 부분이 있기 때문이다.

### 실제 관측 출력

```text
pwd && git branch --show-current && git rev-parse --short HEAD
/Users/goos/MoAI/moai-adk-go
main
2213871af

git ls-files ... | awk ...
template 682 claude_skills 333 rules 87 agents_skills 0 CLAUDE 1

python3 (대응 파일 바이트 비교)
mirror_pairs 392 byte_equal 274 different 118

go test ./internal/cli -run 'TestUpdateFlagMatrixCharacterization|TestUpdateCmd_TemplatesOnlyAndBinaryMutuallyExclusive' -count=1 -timeout 120s
ok github.com/modu-ai/moai-adk/internal/cli 1.006s

go test ./internal/template -run 'Test(Atomic|Validate|ApplyProfile|Deploy)' -count=1 -timeout 120s
ok github.com/modu-ai/moai-adk/internal/template 0.416s

python3 (규칙 frontmatter의 paths 필드와 파일 크기 집계)
rules 84 unscoped 14 bytes 185368 words 26852
kanban_dispatch_bytes 32800
```

위 출력은 이 체크아웃에서 이번 조사 중 관측한 값이다. `git status --porcelain=v1 -uno | awk 'END{print ...}'`는 추적 파일 수정 15줄을 반환했다. `origin/main` 최신성·원격 CI·실제 사용자 프로젝트에서의 업데이트는 확인하지 않았다.

## 2. 확인된 문제와 개선안

### F1 · 높음 · 공유 체크아웃 작업 절차가 서로 충돌

**근거:** `AGENTS.md:47-53,88-107`은 기본 체크아웃에서 브랜치 전환과 `git reset --hard`를 금지하고 launcher worktree를 요구한다. `.claude/rules/moai/workflow/spec-workflow.md:21-59`는 plan/run/sync 기본 경로를 main checkout에 놓고, PR 경로에서는 `git checkout main`, `git reset --hard origin/main`을 실행하도록 쓴다. 배포 미러 `internal/template/templates/AGENTS.md`와 `.../.claude/rules/moai/workflow/spec-workflow.md`에도 같은 충돌이 있다.

**영향:** 동일한 작업에 양립 불가능한 지시가 주어진다. 강제 초기화를 실행하면 공유 체크아웃의 미커밋 변경이 사라질 수 있다. 이번 조사에서 해당 명령을 실행하지 않았고 실제 손실은 관측하지 않았다.

**계획:** 하나의 정책 문서를 권위자로 정한다. 현재 공유 체크아웃 계약을 기준으로 SPEC의 plan/run/sync 표, 후기 브랜치 정리 명령, manager-git 안내, 배포 미러를 한 변경 묶음으로 개정한다. 금지 명령을 단순 삭제하는 데 그치지 않고, 브랜치 재진입·원격 병합 후 갱신 절차를 launcher worktree 기준으로 다시 적는다. 두 문서가 같은 작업 시나리오에 상충 명령을 내리지 않는 구조화 검사를 추가한다.

**완료 조건:** 공유 체크아웃 작업 시나리오의 모든 단계가 `AGENTS.md`의 금지 명령 없이 실행 가능한 문서로 정렬되고, 배포된 템플릿에서 같은 검사가 통과한다.

### F2 · 높음 · Codex 스킬 래퍼가 현재 도구 인터페이스와 맞지 않음

**근거:** `.agents/skills/moai-goal/SKILL.md:7`은 `Use Skill("moai")`를 유일한 실행 지시로 갖고, `.agents/skills/moai/SKILL.md:8,110,125-164`는 `Skill`, `Agent`, `AskUserQuestion`, `${CLAUDE_SKILL_DIR}`을 전제로 한다. `internal/template/skill_mirror.go:178-193,230-245`는 본문 변환 없이 링크 또는 복사한다. **이번 Codex 세션에 노출된** 도구 API에는 `Skill()`이 없다. 실제 이번 요청에서는 래퍼를 문자 그대로 실행하지 못하고 `moai goal arm --auto` CLI를 사용했다.

**범위:** 모든 Codex 버전에서 실행 불가능하다는 뜻은 아니다. 설치 바이너리는 `v3.2.0-rc.12` / `g34dea4ff0`, 조사 체크아웃은 `2213871af`로 서로 다른 빌드다. 이번 CLI의 `--auto`는 현재 스킬/워크플로 문서에 없고, 현재 소스 `internal/cli/goal.go:53-104`는 이전 arm/status/clear 표면이다. 바이너리 동작과 이 소스의 구현을 섞어 단정하지 않았다.

**계획:** Claude용 호출 문구와 Codex용 절차를 생성 단계에서 분리한다. Codex 래퍼에는 실제 사용 가능한 CLI 또는 네이티브 도구 호출, 환경별 경로 해석, 오류 반환을 적는다. 출시 바이너리와 동일 커밋에서 대표 명령 `goal`, `plan`, `run`, `sync`의 스모크 검사를 수행하고, 생성된 스킬에 사용 불가능한 필수 도구 이름이 남으면 실패시키는 검사를 추가한다.

**완료 조건:** 새로 초기화한 Codex 프로젝트에서 대표 래퍼가 수동 번역 없이 실행되고, 같은 빌드의 CLI 도움말·스킬·워크플로 문서가 일치한다.

### F3 · 높음 · 부모 심볼릭 링크를 통해 프로젝트 밖에 파일을 씀

**재현:** 임시 `root`와 `outside`를 만들고 `root/.claude -> outside` 링크를 둔 뒤, `fstest.MapFS`의 `.claude/probe.txt`를 `template.NewDeployer(...).Deploy(...)`로 배포했다. `GOCACHE=/tmp/moai-template-auditprobe/gocache go run -mod=mod .`은 종료 0으로 다음을 출력했다.

```text
deployErr=template deploy track ".claude/probe.txt": manifest track: manifest: file not found outsideReadErr=<nil> outsideContent="OUTSIDE"
```

즉 배포는 나중에 오류를 반환했지만, 이미 외부 파일에 `OUTSIDE`를 기록했다. `internal/template/deployer.go:386-412`는 문자열 경로를 검사하고, `:252-266`의 쓰기는 기존 부모 링크를 따른다. 프로젝트의 링크를 누가 배치할 수 있는지, 링크 허용 정책은 별도 판단이 필요하다. 이번 재현은 원본 프로젝트를 수정하지 않았다.

**계획:** 쓰기 전에 부모 구성요소의 심볼릭 링크 정책을 정하고 실효 경로를 루트에 묶는다. 검사와 열기 사이의 경로 교체까지 고려한 안전한 열기 방식을 사용한다. 루트 밖 파일이 만들어지지 않는 TempDir 회귀 테스트와 정상 배포 테스트를 함께 둔다.

**완료 조건:** 외부를 향하는 부모 링크를 둔 재현에서 프로젝트 밖 쓰기 자체가 발생하지 않고 실패가 반환된다.

### F4 · 중간 · 템플릿 목록 읽기 실패가 빈 목록으로 바뀜

**근거:** `internal/template/deployer.go:307-326`의 `ListTemplates()`는 `fs.WalkDir`의 callback 오류를 `nil`로 바꾸고 반환 오류를 `_ =`로 버린다. 읽기 실패를 내는 임시 `fs.FS`를 주입한 재현 명령 `GOCACHE=/tmp/moai-template-auditprobe/gocache go run -mod=mod .`의 출력은 `ListTemplates on broken FS: len=0 value=[]`였다. 비테스트 호출자는 `internal/cli/update_template_sync.go:154`, `internal/cli/update_dryrun_preview.go:49`, `internal/cli/update/merge/merge.go:307`에 있다.

**영향:** 템플릿 목록 조회 실패와 실제 빈 목록을 호출자가 구분하지 못한다. 특정 업데이트가 파일을 누락한 실제 사고는 재현하지 않았다.

**계획:** 목록 API를 `([]string, error)`로 바꾸거나 기존 API 호환이 필요하면 strict 목록 API를 도입한다. 세 호출자가 오류를 표시하고 후속 변경을 멈추게 한다. 실패 주입 `fs.FS` 회귀 테스트에 정상·중간 실패·빈 목록을 구분해 넣는다.

**완료 조건:** 실패 주입 시 업데이트·미리보기·병합 경로가 성공형 빈 목록을 출력하지 않고 실패 원인을 전파한다.

### F5 · 중간 · goal 문서와 현재 소스의 명령 계약이 어긋남

**근거:** `.claude/skills/moai/SKILL.md:159-164`는 `goal resume`을 명령으로 열거하지만 `.claude/skills/moai/workflows/goal.md:79-89`는 미출시라고 적고 `internal/cli/goal.go:53-56`도 등록하지 않는다. 현재 **별도 빌드** 바이너리 도움말에는 `resume`, `--auto`가 관측됐다. 따라서 문제는 특정 verb의 절대 부재가 아니라 문서·소스·배포 바이너리의 버전 귀속이 불명확한 점이다.

**계획:** 빌드된 CLI `--help`에서 명령 표를 생성하거나 검사한다. 워크플로 문서에 지원 버전과 `--auto` 생명주기(`draft` 이후 단계 포함)를 연결한다. 과거 소스 브랜치와 신규 바이너리를 비교 보고할 때 커밋 식별자를 의무화한다.

**완료 조건:** 동일 빌드에서 CLI 도움말과 배포 스킬의 verb/flag 집합이 일치한다.

### F6 · 중간 · 상시 로드 규칙에 Kanban 전용 긴 본문이 포함됨

**근거:** `.claude/rules/moai/development/rule-authoring.md:12-19`는 `paths:` 없는 규칙이 매 턴 들어간다고 설명한다. 현재 규칙 84개 중 14개가 이 범위이며 합계 185,368바이트·공백 기준 26,852단어다. `.claude/rules/moai/workflow/kanban-dispatch.md`는 32,800바이트이고 11행에서 lead 이외 세션에서는 효력이 없다고 밝힌다. `CLAUDE.md` 20,019바이트는 별도 상시 슬롯이다.

**한계:** 바이트와 공백 단어만 측정했다. 실제 토크나이저 비용과 런타임의 캐시 할인은 측정하지 않았다. `session-handoff.md`처럼 상시 로드가 필요한 다른 규칙까지 일괄 이동하자는 뜻은 아니다.

**계획:** Kanban lead가 항상 알아야 할 진입·안전 규칙만 작은 상시 stub에 두고 상세 절차는 lead 시작 시에만 명시적으로 로드한다. 변경 전후 동일 작업 묶음에서 실제 입력 토큰·캐시 적중·지시 준수율을 측정한다. 절감량이 없거나 lead 성공률이 낮아지면 되돌린다.

**완료 조건:** 일반 세션에서 불필요한 본문 로드가 제거되고 lead 세션의 절차·안전 테스트가 유지된다.

## 3. 추가 검증이 필요한 후보

| ID | 관측 사실 | 아직 확인할 것 | 검증 후 가능한 조치 |
|---|---|---|---|
| H1 | `internal/template/deployer.go:19-29`는 항상 `<dest>.moai-tmp`를 `os.WriteFile`로 쓰고 같은 경로를 지운다. | 동시 CLI 실행 차단 여부, 임시 파일 선점·동시 배포 재현 | 겹침이 가능하면 같은 디렉터리 `CreateTemp`와 배타 생성으로 바꾸고 원자 rename 검증 |
| H2 | `internal/template/profile_matrix.go:81-110`과 `model_policy.go:215-235`가 같은 `llm.yaml`을 별도 읽기·쓰기하고 CLI가 연속 호출한다. | 첫 쓰기 뒤 두 번째 실패 주입, 동시 실행 경로 | 한 번 읽고 두 변경을 합쳐 원자 저장 |

`internal/template/validator.go:114-137`의 “기대 파일이 디렉터리면 warning과 `Valid=true`”는 처음엔 결함처럼 보였지만 `validator_test.go:183-207`의 명시 테스트와 `TestValidatorValidateDeployment` 통과를 확인했다. 현재 계약으로 분류하며 결함 목록에서 제외했다. `renderCache`도 같은 인스턴스 재사용 위험 가설은 있으나 현재 생성 범위와 `TestDeployerSingleRender` 결과만으로 생산 결함이라 주장하지 않는다.

## 4. 미러와 근거 없는 수치 문구

배포 원본은 `internal/template/embed.go:28-43`의 `templates`이며, 설치된 `.claude/skills` 파일과는 별개다. 배포 스킬 309개 대응 파일 중 106개, 그중 `SKILL.md` 15개가 현재 설치본과 바이트가 다르다. 예를 들어 `moai-foundation-cc/SKILL.md:66,115,254-280`에는 중첩 agent/허용 도구 설명 차이가 있다. 단순 차이를 drift 결함으로 판정하지 않았다. 먼저 의도된 사용자 수정·구버전 설치본·동기화 누락을 분류하고, 배포 원본과 실제 호출 파일을 각각 검사해야 한다. 완전히 같은 본문 파일은 `.gitkeep` 5개뿐이라 스킬 본문 복제가 주요 낭비라고 볼 근거도 없다.

`.claude/skills/moai-foundation-core/SKILL.md:129,181-185`에는 “성공률 40%”, “이해도 80%를 시간 5%에”, “에스컬레이션 70% 감소”와 같은 수치가 출처·기준선 없이 단정돼 있다. 이번 조사에서 수치의 진위를 검증하지 않았다. 증명 자료가 없다면 수치 표현을 제거하고, 남기려면 실험 조건과 재현 결과를 명시한다.

## 5. 실행 순서와 검증 기준

| 순서 | 우선순위 | 변경 묶음 | 소유 범위 | 검증 게이트 |
|---|---|---|---|---|
| A | 높음 | F1 정책 단일화와 배포 미러 정렬 | `AGENTS.md`, SPEC·git 규칙, 템플릿 | 충돌 시나리오 구조 검사, 금지 명령 검사, 배포 산출물 대조 |
| B | 높음 | F2 Codex 어댑터와 스킬 생성 계약 | 스킬 미러 생성, Codex 래퍼 | 같은 커밋의 CLI로 init 후 goal/plan/run/sync 스모크, 호출 가능한 도구만 사용 |
| C | 높음 | F3 경로 이탈 방지 | `internal/template/deployer.go`와 배포 호출자 | 부모 링크를 둔 TempDir 반례에서 외부 쓰기 0건 |
| D | 중간 | F4 목록 오류 전파와 F5 명령 목록 정합 | `internal/template`, 업데이트 호출자, goal 문서 | 실패 주입 FS, 좁은 Go 테스트, CLI 도움말 대조 |
| E | 중간 | F6 상시 로드 축소와 문구 정리 | 규칙·스킬 문서 | 동일 작업 입력 토큰/캐시/성공률 전후 측정, lead 회귀 시 되돌림 |
| F | 조건부 | H1·H2 재현 후 최소 수정 | 템플릿 쓰기·설정 저장 | 동시성/실패 주입 반례를 먼저 실패시킨 뒤 고침 |

각 묶음은 별도 카드와 worktree에서 진행하고, 공유 기본 체크아웃의 기존 변경을 보존한다. 수정 전에는 기준 브랜치·커밋·원격 차이를 다시 확인한다. 문서 또는 테스트가 녹색이라는 이유만으로 실제 배포·사용자 프로젝트 안전성을 주장하지 않는다. 모든 변경의 완료 판정에는 실제 명령과 출력을 남긴다.

## 6. `moai-adk for Codex` 목표 대비 전수 점검

2026-08-17의 `.moai/reports/moai-adk-dual-harness-codex-20260817.md:104-120`은 성공 지표를 **스킬 노출, 신뢰된 훅 8종, AGENTS.md 32 KiB 이내, 양쪽 하네스에서 같은 SPEC plan→run→sync 완주**로 정했다. 2026-08-22의 `.moai/reports/codex-dual-harness/codex-dual-harness-plan-20260822.md:3-19`는 `moai init --agent claude|codex|both`, 스킬 링크, hooks.json, config.toml, agent TOML, MCP, doctor 경로를 제시했다. 두 문서는 과거 계획·측정 기록이므로 현재 동작 근거로 승격하지 않고 이번 소스와 설치 CLI를 따로 대조했다. [OpenAI 공식 문서의 Codex 스킬 구조](https://developers.openai.com/plugins/concepts/skills)와 [Codex의 MCP 설정 경로](https://developers.openai.com/learn/docs-mcp)도 현재 외부 계약 참고로 확인했다.

| 표면 | 현재 소스·설치본에서 관측한 것 | 판정과 경계 |
|---|---|---|
| 초기화 선택 | `internal/cli/init.go:128-154,393-400`은 `--agent claude|codex|both`를 등록한다. 설치된 `moai init --help`는 `--llm claude|gpt|both`를 보여주고 `--agent`는 없다. | **불일치.** 설치 명령으로 설계 문서의 진입 절차를 그대로 실행할 수 없다. 설치 바이너리 `g34dea4ff0`와 HEAD `2213871af`는 dirty 빌드 내용이 다르다. |
| 공통 지침 | 루트·배포 `AGENTS.md` 각각 14,229바이트, 두 파일 `cmp` 동일. | **정적 크기 통과.** 사용자 전역 AGENTS 합산과 Codex 실제 주입·잘림은 미측정. 루트 문서의 “중첩 AGENTS 없음” 문구는 `internal/template/templates/AGENTS.md` 존재와 불일치한다. |
| 스킬 노출 | `.agents/skills`에 정본을 향하는 링크 34개와 하위 명령 래퍼 16개, 깨진 링크 0개. 래퍼 16개 모두 `Use Skill("moai")` 형태. Codex 0.156.1의 로컬 `debug prompt-input`에서 이 저장소의 `moai` 카탈로그 행 50개와 루트 `AGENTS.md` 주입을 확인했다. | **발견 통과·실행 미입증.** SKILL 본문은 시작 입력에 즉시 들어가지 않았고, 실제 모델의 본문 읽기·지시 수행은 미측정. 이번 세션 도구에는 `Skill()`도 없다. |
| 에이전트 | `.codex/agents/moai/*.toml` 11개 생성. 배포 TOML 11개 모두 `Agent()`, `Skill()`, `/compact` 등 Claude식 호출 문구가 검색된다. | **게시 통과·실행 미입증.** `internal/template/agentemit/writer.go:99-109`가 본문을 그대로 실어 보내고, 기존 배포 검사는 바이트 동일성 중심이다. |
| 훅 | 현재 `.codex/hooks.json`은 JSON 파싱 가능하고 MoAI handler가 들어 있는 이벤트 8개가 관측됐다. 현재 소스 `internal/codexadapter/events.go:46-63`은 adapted 6개이고 `internal/codexwiring/hooks.go:95-113`은 그 목록으로 생성한다. | **버전 차이.** 현재 파일 8개와 현재 소스 생성 기준 6개를 같은 빌드의 결과로 볼 수 없다. 신뢰 등록과 실제 발화는 미측정. |
| MCP | `codex mcp get moai`는 enabled=true, stdio `moai mcp-server`, `default_tools_approval_mode=writes`를 표시했다. | **등록 확인.** 서버 기동·각 도구 호출·승인 흐름은 미측정. |
| 설정·상태줄 | `internal/codexwiring/wire.go`가 `.codex/config.toml`과 hooks를 생성·갱신하고 테스트가 통과했다. 현재 README:341은 Codex 상태줄에 MoAI 전용 값 표시가 안 된다고 적는다. | **일부 지원.** `tui.status_line` 기본 식별자와 MoAI 전용 상태를 구분해야 한다. |
| 업데이트·진단 | `internal/cli/update_codex_wiring.go`가 기존 배선의 갱신을 시도한다. 설치된 `codex doctor --summary --no-color --ascii`는 config/auth/mcp loaded `ok`, 전체 `19 ok/7 notes/5 warn/0 fail`을 보고했다. | **부분 관측.** 이 출력은 훅 신뢰나 workflow 성공 판정이 아니다. |
| 사용자 워크플로 | 동일 SPEC의 Codex plan→run→sync, 질문·위임·goal 지속 실행을 이번 조사에서 끝까지 실행하지 않았다. | **Gap.** “완벽 지원” PASS 불가. |
| 교차 모델 작업·감사 | `mcp_server.go:246-278`은 `codex_audit`, `codex_setup`, `codex_task`, `audit_multi`를 등록한다. `codex_task.go:173-286`은 쓰기 opt-in을 확인하고, `codex_review_gate.go:66-115`은 확정 FAIL만 차단한다. 관련 좁은 단위 테스트는 통과했다. | **소스 경로 확인.** Codex CLI 0.156.1에 대한 실제 RPC·인증·MCP 호출 성공은 미측정. 이는 Codex가 MoAI의 호스트가 되는 경로와 별개다. |

설치 CLI의 차이는 `moai init --agent codex --help` → 종료 코드 1, `Unknown flag: --agent.`로 직접 확인했다. 반면 소스의 Codex 배선 관련 테스트는 다음처럼 통과했다.

```text
go test ./internal/codexwiring -run 'Test(Wire|RefreshWiring|RenderHooks|EnsureMCPTable|StatusLine)' -count=1 -timeout 90s
ok github.com/modu-ai/moai-adk/internal/codexwiring 0.674s

go test ./internal/cli -run 'Test(Init.*Codex|Update.*Codex|MCP.*Codex|Codex.*Wiring)' -count=1 -timeout 90s
ok github.com/modu-ai/moai-adk/internal/cli 0.839s

go test ./internal/cli -run '^(TestCodexAudit_NativeDispatchesReviewStart|TestCodexAudit_AdversarialDispatchesTurnStart|TestReviewGate_CodexFailBlocks|TestCheckCodexWiring_HealthyProjectOK)$' -count=1 -v -timeout 60s
=== RUN   TestReviewGate_CodexFailBlocks
--- PASS: TestReviewGate_CodexFailBlocks (0.00s)
=== RUN   TestCheckCodexWiring_HealthyProjectOK
--- PASS: TestCheckCodexWiring_HealthyProjectOK (0.00s)
=== RUN   TestCodexAudit_NativeDispatchesReviewStart
--- PASS: TestCodexAudit_NativeDispatchesReviewStart (0.00s)
=== RUN   TestCodexAudit_AdversarialDispatchesTurnStart
--- PASS: TestCodexAudit_AdversarialDispatchesTurnStart (0.00s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 0.779s
```

이 테스트는 소스 로직과 모의 호출을 검증한다. 설치된 `codex-cli 0.156.1`의 실제 런타임 통합 검사는 아니다.

Codex 스킬 발견은 모델 호출 없이 `CODEX_HOME`을 분리해 `codex debug prompt-input`으로 확인했다. 현재 저장소 결과를 필요한 필드만 출력했더니 `moai_catalog_rows 50`, `moai_core True`, `moai_goal True`, `project_agents_contract_loaded True`, `claude_main_skill_body_eager_loaded False`였다. 별도 공개 `/tmp` fixture의 일반 스킬과 `.agents/skills` 상대 심볼릭 링크 스킬도 모두 카탈로그에 나타났다(`entries 4`, 두 스킬 발견 `true`). 따라서 이 버전에서 **상대 링크 발견은 확인**됐으며, 과거 설계의 링크 미발견 가능성을 현재 사실로 반복하지 않는다. `codex exec` 모델 호출은 분리된 `CODEX_HOME`에서도 사용자 전역 스킬 메타데이터가 초기 입력에 합쳐지는 것을 확인해, 공개 fixture만 외부로 보낸다는 조사 조건에 맞지 않아 실행하지 않았다.

같은 로컬 출력에서 MoAI 스킬 카탈로그 행의 UTF-8 길이를 세면 50행 **13,995바이트**였고, 정본 34개가 12,078바이트, 명령 래퍼 16개가 1,917바이트였다. 이는 시작 입력의 **문자 바이트** 측정이지 토큰·요금 측정이 아니다. 래퍼를 없애면 명령 진입성이 떨어질 수 있으므로, 먼저 설명문을 줄인 시안의 `prompt-input` 바이트와 실제 명령 선택 정확도를 함께 비교한다.

### Codex에서 우선 고칠 계약

**C1 · 높음:** F2를 Codex 전용 dispatcher와 에이전트 본문 변환으로 확장한다. `internal/template/skill_mirror.go:3-7,171-193`의 무변환 링크는 파일 게시에는 맞지만 실행 도구 호환성은 보장하지 않는다. `internal/template/templates/.codex/agents/moai/manager-spec.toml:218-224`와 `manager-lead.toml:155-157` 등에서 Claude 호출을 Codex에서 사용 가능한 절차로 생성한다. 11개 TOML과 16개 래퍼의 정적 검사는 필수 도구·환경변수·slash 명령을 기계적으로 검사해야 한다.

**C2 · 높음:** 출시 산출물의 옵션·파일·도움말을 한 커밋에서 생성·검증한다. 현재 `--agent`/`--llm`, goal `--auto`/옛 arm 문서는 빌드가 어긋나면 지원 주장을 망가뜨린다. 같은 `moai version` 빌드에서 `init --help`, 실제 임시 프로젝트 init, 생성 파일, Codex CLI 로드를 연속 검증한다. 빌드 메타데이터에 dirty 여부와 입력 트리를 기록한다.

**C3 · 높음:** 훅 6/8 차이를 의도된 지원 범위로 문서화하고, 생성된 훅 각각에 대해 신뢰·발화·출력 효과를 Codex CLI에서 확인한다. `codex doctor`의 일반 통과나 `.codex/hooks.json` 파일 존재는 실제 훅 동작의 대체 증거가 아니다. 2026-08-22 설계의 프로젝트 훅 미발화 블로커도 현재 버전에서 다시 재현·해소해야 한다.

**C4 · 중간:** `internal/codexwiring/wire.go:88-96`은 whitelist 거부 시 쓰지 않고 오류를 돌려준다. 그러나 `internal/cli/init.go:160-172,925`와 `internal/cli/update_codex_wiring.go:16-25`는 그 오류를 경고 후 계속 진행한다. 안전한 no-write는 좁은 테스트에서 확인됐지만, `--agent codex` 초기화가 배선 실패 후 성공처럼 보이는 것이 허용 정책인지 결정해야 한다. 성공의 정의에 훅·MCP 설치가 포함되면 CLI 실패로 반환하고 복구 지침을 제공한다.

**C5 · 중간:** `internal/codexwiring/configtoml.go:177-203`의 MCP 검사기는 TOML 전체 파싱 없이 줄을 센다. 중복 `command`가 있는 잘못된 입력에서 `Canonical:true`를 반환한 임시 주입 결과가 있다. 이는 **진단기의 잘못된 정규 판정 후보**이며 유효 설정을 잘못 처리한 증거는 아니다. 실제 Codex 파서와 대조하는 음성 테스트를 추가한다.

### Codex 지원 완료 게이트

1. **동일 빌드:** 릴리스 바이너리, 소스 커밋, 템플릿 해시, CLI 도움말이 하나의 manifest에 묶이고 dirty 빌드가 식별된다.
2. **깨끗한 임시 프로젝트:** `moai init`에서 Claude/Codex/both 각각 생성물·미생성물을 확인하고, 재실행과 `moai update` 후 사용자 파일 보존을 확인한다.
3. **Codex CLI 실제 세션:** 스킬 탐색·대표 명령 4종, 11개 중 대표 에이전트의 위임, 사용자 질문, 6개 또는 정책상 8개 훅의 신뢰와 발화, MCP 읽기·쓰기 승인 흐름을 각각 관측한다.
4. **동일 SPEC:** 한 개의 작은 fixture SPEC을 Claude와 Codex에서 각각 plan→run→sync로 완주하고 동일한 AC·evidence 계약을 만족하는지 독립 감사한다. 이 단계 전에는 “완벽 지원” 문구를 사용하지 않는다.
5. **버전 회귀:** 지원할 Codex CLI 버전별로 설정 스키마·훅·agent·skill smoke를 CI에서 실행하고, 최신 버전의 변화는 실패 또는 명시적 GAP으로 드러나게 한다.

## 7. 남은 범위와 잔여 위험

- 원격 `origin/main`과 CI 상태를 이번 보고서의 기준선으로 사용하지 않았다. 기준은 명시된 현재 체크아웃이다.
- 파일 목록은 전수화했지만 모든 Go 분기와 Claude/Codex 런타임을 실행하지 않았다. H1·H2는 후보로만 둔다.
- `go test`는 관련 일부 테스트만 실행했다. 전체 스위트, 실제 `moai init/update` 사용자 프로젝트, Claude Code 호출, 다른 Codex 버전은 실행하지 않았다.
- 기존 수정 파일 15개와 미추적 보고서가 있는 공유 트리다. 이 보고서 작성으로 소스 코드를 고치거나 기존 작업을 정리하지 않았다.

**결론:** F1의 파괴적 명령 충돌, F2의 호출 인터페이스, F3의 루트 밖 쓰기를 먼저 해결하고, 목록 오류 전파·명령 계약 정합·상시 로드 비용을 각각 독립 검증하자. H1·H2는 재현 결과에 따라 수정 여부를 정한다.
