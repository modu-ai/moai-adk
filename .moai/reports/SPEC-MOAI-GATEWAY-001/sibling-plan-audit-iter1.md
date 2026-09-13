# CG-RETIRE·TEAMMATE 독립 계획 감사
Iteration: 1
Verdict: FAIL
Overall Score: 0.81 (CG 로컬 구현 준비)
TEAMMATE Verdict: PASS
TEAMMATE Score: 1.00 (후보 조사·preflight 준비에 한정)

## Claim

CG-RETIRE 0.1.0은 로컬 이전 기능 구현 전에 SP-B1을 보완해야 한다. 기존 경로 조사·실패 fixture 준비는 진행할 수 있다.
TEAMMATE 0.1.0은 지원 seam 조사와 격리 fixture 준비를 진행할 수 있다. native Claude teammate 지원 구현·활성화·제품 완료를 승인한 판정은 아니다.

Reasoning context ignored per M1 Context Isolation. 작성자의 결론 대신 각 SPEC의 6종 문서와 실제 코드·man을 읽었다.
확인 대상으로 번호·형식·필수 필드·추적 누락, 이전 의미 축약, 오래된 CG 설정의 실행 경로 유입, 불명확한 이전 명령/저장 계약,
다른 provider fallback, 공유 tmux 인증, bootstrap 재사용, 부모 종료·PID 재사용, 같은 세션 경합, 지원 seam을 가정한 완료 판정을 두었다.
코어 전체나 작성 중 CLI 구현의 코드 감사는 수행하지 않았다.

## Must-Pass Results

두 SPEC에 공통으로 적용했다. CR/GT는 각각 해당 SPEC 디렉터리의 파일을 가리킨다.

- PASS MP-1: CR spec.md:30–44의 REQ-CR-001…008, GT spec.md:30–46의 REQ-GT-001…009는 연속·중복 없음. 실제 추출 결과는 아래와 같다.
- PASS MP-2: 요구사항 층만 판정했다. 두 spec.md의 모든 REQ는 When / Where / The system shall 형식이다. AC의 Given/When/Then은 올바른 검증층이다.
- PASS MP-3: 두 spec.md:2–14의 canonical 12필드, quoted version, draft, P1, 날짜, 문자열 tags를 확인했고 지정 바이너리 린트도 각각 통과했다.
- N/A MP-4: 단일 Go 저장소의 실행·설정 계약이다. CR의 4locale은 문서 언어 범위이며 프로그래밍 언어 도구 지원 목록이 아니다.
- PASS MP-5: 추출된 정식 SPEC ID는 모두 존재하고 draft다. retired/superseded/archived 참조가 없다. CR spec.md:44의 TEAMMATE 약칭 대상도 실제 입력 문서로 읽었다.
- PASS MP-6: 두 spec.md에 syscall 문자열 없음. 플랫폼 지원의 runtime 판정은 별도 Gate다.
- PASS MP-7: 두 plan.md/research.md에 [NEEDS CLARIFICATION] 표식 없음. 표식이 없다는 관측은 SP-B1 같은 본문 미결을 해소하지 않는다.
- N/A MP-8: 두 acceptance.md에 release-blocking으로 분류된 AC/RED-now cell이 없다. 향후 실제 완료 게이트 문장을 이미 실행된 RED cell로 해석하지 않았다.

## Category Scores

| 대상 | Clarity | Completeness | Testability | Traceability | 근거 |
|---|---:|---:|---:|---:|---|
| CG 로컬 구현 준비 | 0.75 | 0.75 | 0.75 | 1.00 | design.md:9–13, acceptance.md:5–7의 이전 입력/저장 결과 미확정. 8 REQ/AC는 모두 연결됨 |
| TEAMMATE 후보/preflight | 1.00 | 1.00 | 1.00 | 1.00 | design.md:11–19에서 지원 seam·대안을 미채택으로 구분. acceptance.md:3–19가 인증·재사용·종료·실제 연결의 관측 결과를 명시 |

Tier L 기준은 .claude/rules/moai/workflow/spec-workflow.md:142의 0.85다. CG는 평균 0.8125이며 blocking finding도 남아 있다.
TEAMMATE 점수는 후보/preflight 계획의 점수다. 살아 있는 native pane의 보안·기능 점수가 아니다.

## Defects Found

SP-B1 — CG design.md:9–13; acceptance.md:5–7 — 이전의 실행 가능한 입력과 저장 상태 계약이 아직 없다. Severity: major. Class: blocking. Confidence: high.

문서는 exact CLI 이전 명령을 “구현 전 계획에 반영”한다고 명시적으로 유보한다. 읽은 6종 문서에는 launcher/teammate 선택을 어느 명령으로 받고,
어떤 파일/키에 어떤 값으로 저장하며, 그 결과를 어느 launch 경계가 읽어 미이전 거절을 해제하는지 없다. AC-CR-003의 “선택한 필드”와
AC-CR-002의 “명시 정상 구성”에 입력/기대값을 고정할 수 없는 상태다. 이 빈칸은 TEAMMATE의 실제 지원 여부와 별개로 로컬 이전 구현에 영향을 준다.

Required fix: design/plan에 최소 한 표로 (1) preview/apply의 명령·명시 선택·무인 실행 조건, (2) 원본과 대상 경로/필드 및 leader/teammate 의미,
(3) 이전 완료/미완료 상태를 읽는 launch guard와 변경 전 거절 순서, (4) TEAMMATE 게이트 미충족 때 저장·완료 표시 정책을 확정한다.
AC-CR-002/003에 실제 이전 전/후 fixture 한 쌍, 기존 cg가 남은 상태의 cc/glm/gpt 거절 및 명시 정상 구성 대조군을 연결한다.
큰 새 기능을 추가할 필요는 없다. 해당 표와 fixture 계약만 보완한 변경분 감사로 재판정한다.

SP-A1 — CG acceptance.md:5,11 — 이전 거절이 기존 backend/모드 변경보다 먼저인지 관측하는 fixture를 구체화하면 좋다. Severity: minor. Class: optional. Confidence: high.

현재 소스의 template.IsGLMBackend는 team_mode=cg를 true로 읽고, legacy launcher는 설정 정리·모드 저장을 수행한다.
CR-002/004/005가 이미 원본 보존·새 provider 판정·자동 변환 금지를 요구하므로 별도 요구사항 누락으로 판단하지 않았다.
테스트에서 stale team_mode=cg와 llm.mode=glm, GLM 슬롯 및 알 수 없는 사용자 키를 함께 놓고 legacy guard 앞뒤의 설정/exec 계수를 관측한다.
작성 중 gateway launcher 변경을 실패 구현으로 판정한 것이 아니다.

TEAMMATE 후보/preflight 범위에서는 blocking finding을 발견하지 않았다.

## Evidence

아래 명령은 이 감사에서 지정 WT를 workdir로 직접 실행했다. 각 출력은 관측한 발췌이며 줄임말을 새 출력처럼 만들지 않았다.

```text
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
$ git rev-parse --short HEAD
81c1d58f9
$ /tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-CG-RETIRE-001
✓ No findings — all SPEC documents are valid
$ /tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-GATEWAY-TEAMMATE-001
✓ No findings — all SPEC documents are valid
```

각 exit 0. 부모가 이 HEAD에서 빌드한 바이너리이며 감사자가 다시 빌드했다고 주장하지 않는다.
Python pathlib/re를 사용해 REQ/AC 정의, 참조 차집합, syscall/clarification/release-blocking 문자열과 참조 SPEC의 실제 status를 추출했다. exit 0:

```text
SPEC-MOAI-CG-RETIRE-001 REQ 8 AC 8 uncovered [] orphan []
syscall= False clarification= False release-blocking-cells= 0
reference SPEC-MOAI-CG-RETIRE-001 draft
reference SPEC-MOAI-GATEWAY-001 draft
reference SPEC-MOAI-GATEWAY-PICKER-001 draft
reference SPEC-MOAI-GPT-AUTH-001 draft
SPEC-MOAI-GATEWAY-TEAMMATE-001 REQ 9 AC 9 uncovered [] orphan []
syscall= False clarification= False release-blocking-cells= 0
reference SPEC-MOAI-GATEWAY-001 draft
reference SPEC-MOAI-GATEWAY-PICKER-001 draft
reference SPEC-MOAI-GATEWAY-TEAMMATE-001 draft
reference SPEC-MOAI-GPT-AUTH-001 draft
```

`rg -n 'migrat|이전 명령|저장|schema|필드|team_mode|launcher|teammate' .moai/specs/SPEC-MOAI-CG-RETIRE-001`의 직접 출력 발췌:

```text
.moai/specs/SPEC-MOAI-CG-RETIRE-001/design.md:12:자동 사용자 구성 삭제·전역 tmux 환경 정리·프로필 교체는 하지 않는다. exact CLI 이전 명령은 현재 명령 관례 조사 후
```

`nl -ba`로 12개 SPEC 파일 전체와 `cat sibling-plan-design.md`를 읽어 위 문장 다음 줄의 구현 전 유보와 AC의 미정 기대값을 교차 확인했다.

실제 코드 관측 (`sed`/`rg`, exit 0):

- internal/cli/cg.go:57의 `rootCmd.AddCommand(cgCmd)`, runCG의 rejectFactoryOnCG→rejectKanbanOnCG 호출은 아직 존재한다.
- internal/cli/launcher.go:311의 applyCGMode, :385의 persistTeamMode(root, "cg")와 internal/cli/glm.go:624–668의 typed LLMConfig 읽기·저장/disableTeamMode를 읽었다.
- internal/template/glm_effort_overlay.go:310–315는 `case config.TeamModeGLM, config.TeamModeCG: return true`를 포함한다.
- internal/cli/spawn.go:87–88은 tmux new-window에 cwd와 command를 전달한다. internal/tmux/session.go:166–175는 split-window 후 sendKeys다.
- internal/tmux/session.go:206–210의 InjectEnv는 set-environment, :334–336의 InjectSensitiveEnv도 `set-environment %s %s` 내용을 만든다. pane 격리 기능으로 오인하지 않은 research와 일치한다.
- `sed -n '3104,3122p' /opt/homebrew/Cellar/tmux/3.6a/share/man/man1/tmux.1`은 new-window의 `-e environment` 및 shell-command/argument 문법을 보여 준다. Claude native 연결 증거가 아니다.
- `git diff -- internal/cli/cc.go internal/cli/launcher.go`를 읽어 runClaudeEntry 및 gateway binding 작업 중임을 확인했다. 이 미완성 구현의 품질 판정은 범위 밖이다.

## Baseline-attribution

HEAD 81c1d58f9의 dirty WT 문서를 판정했다. SPEC 자체는 0.1.0 draft다. 주요 입력 SHA-256:

```text
CG spec.md 02aadb976c0d6996325a9af55ea37364556fef44e7199617714aafeac9580b5d
CG design.md 8a13940b2ef5ea3db62856bf0fb264f44305d2dc4a0034bb2b7c59d1a5b293e2
CG acceptance.md fbf550b40de32988f81c4689db2d806cb1ecf5693ac3af890870ea0a833c234f
TEAMMATE spec.md 038d1dd333b071e7b6ff0e9eaf3e7d7cdef5244dae77085acd95051d6f1406a5
TEAMMATE design.md 06fa47c27534e1d50d61aeab90d388c5bdaecc0686849f46ff39e8b86ca3c244
TEAMMATE acceptance.md 41e9d058e1781df6fdf8c9fc083e56417acad3980334d1c3ece190e9e3bd1d65
```

## Gaps

Claude·tmux·provider·auth network 실행 없음. 4locale 전체 참조 분류·빌드·렌더, 이전 transaction runtime, 실제 pane 연결·동시성·강제 종료는 미검증이다.
TEAMMATE design.md:11–19의 seam 또는 대안은 채택된 제품 구현이 아니다. private bootstrap의 동일 OS 사용자 사이 소비 권한과 강제 종료 시 liveness 연결도 실제 설계/시험 게이트로 남는다.
코어의 wording-only/in-process 제한은 native pane 격리 PASS가 아니다. 외부 모델의 별도 감사 결과는 이 판정에 사용하지 않았다.

## Residual-risk

CR-008의 Claude leader/GLM teammate 동등성은 TEAMMATE 통합 전 완료로 기록할 수 없다.
TEAMMATE는 spec.md:42, design.md:13–19에 따라 지원 seam이 없으면 활성화를 거절해야 하며 experimental flag나 tmux PATH wrapper를 무음 채택할 수 없다.
실제 관측은 2026-09-11 19:00 Asia/Seoul 이후에만 가능하다. 지금의 구조/코드 판독 결과로 실제 routing·fallback·token 격리를 증명하지 않는다.

## Recommendation

CG는 SP-B1 표와 fixture 계약을 보완한 뒤 해당 변경분만 재감사한다. TEAMMATE는 후보 조사·실패 fixture 준비를 진행하되 AC-GT-007/009 실측 전 native pane 지원을 활성화하지 않는다.
