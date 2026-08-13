---
title: moai web 웹 콘솔
weight: 50
draft: false
---

`moai web` 은 브라우저 기반 설정 편집기인 **MoAI Web Console** 을 실행합니다. 터미널 프로필 마법사(`moai profile`)와 같은 검증·저장 로직을 그대로 쓰며, 프로필 선호도와 프로젝트의 user / language / statusline 섹션을 웹 UI에서 편집합니다.

프로필과 `settings.json` 항목이 늘어나면 터미널 마법사로 모든 값을 한 번에 들여다 보기 어렵기 때문에, 이 커맨드는 같은 저장 로직을 브라우저 폼으로 옮겨 놓습니다. 따라서 관리자 에이전트나 하네스(harness) 가 읽는 같은 YAML 을 더 빠르게 살펴 보고 고치는 보조 진입점 역할을 합니다.

## 개요

```bash
moai web [OPTIONS]
```

콘솔은 **루프백(127.0.0.1)에만 바인딩** 됩니다. 외부 데이터베이스, 인증, 네트워크 노출이 전혀 없습니다. 대상 포트를 이미 오래된 moai 인스턴스가 점유하고 있으면 해당 인스턴스를 종료하고 다시 바인딩합니다. 다만 moai 가 아닌 외부 프로세스는 절대 종료하지 않으며, 이때는 오류를 내고 `--port` 사용을 권합니다.

## 플래그

| 플래그 | 설명 |
|--------|------|
| `--port <N>` | 127.0.0.1 에 바인딩할 TCP 포트 (기본값: `3041`) |
| `--no-open` | 브라우저를 자동으로 열지 않음 |
| `--no-reuse` | 오래된 moai 인스턴스로부터 포트를 회수하지 않고, 포트 충돌 시 실패 |

## 예시

```bash
moai web                 # 127.0.0.1:3041 에 바인딩하고 브라우저를 엶
moai web --port 9000     # 다른 포트에 바인딩
moai web --no-open       # 브라우저를 열지 않고 시작
moai web --no-reuse      # 포트가 사용 중이면 회수 대신 실패
```

## 편집 대상

웹 콘솔은 다음을 편집합니다.

- **프로필 선호도** — 모델·언어·표시 설정 등 프로필별 설정
- **프로젝트 설정** — `.moai/config/sections/` 의 user / language / statusline 섹션

저장할 때 터미널 마법사와 똑같은 유효성 검사를 거치므로, 둘 중 어느 쪽으로 고쳐도 결과는 같습니다.

## 콘솔 화면

콘솔의 인터페이스 언어는 헤더 오른쪽 선택기에서 English · 한국어 · 日本語 · 中文 가운데 고릅니다. 아래 화면은 English 로 둔 모습이라, 이 절에서는 화면에 보이는 영어 표기를 괄호에 함께 적습니다.

헤더에는 프로젝트 이름과 현재 프로필, 주요 설정 요약(`lang · model · effort · dev`)이 나란히 놓입니다. 그 아래로 프로필 바가 오는데, 프로필 선택기 바로 옆에 추가·이름변경·삭제 컨트롤이 있어 프로필 생애주기 전체가 한 줄에 들어갑니다 (별도의 프로필 카드는 없습니다). 프로필 바 아래에 사용자 정보(Identity) · 언어(Language) · LLM · 서드파티 LLM(3rd Party LLM) · 워크플로우(Workflow) · Git·워크트리(Git & Worktree) · 감사(Audit) · 에이전트(Agents) · 리포트(Report) 아홉 개 탭이 이어집니다. 고친 값은 맨 아래 설정 저장(Save settings) 버튼으로 기록합니다.

![MoAI Web Console 첫 화면. 헤더의 프로젝트 이름과 프로필, 프로필 바, 아홉 개 탭, 사용자 정보(Identity) 탭의 표시 이름(Display name) 입력란, 설정 저장(Save settings) 버튼](/images/profile/web-console-overview.png)

프로필 바에서 프로필을 전환하고, 이름을 바꾸고, 삭제(Delete)하고, 새 프로필 이름(New profile name)을 적어 프로필 생성(Create profile)으로 새로 만들 수 있습니다. 다른 프로필을 고르면 헤더의 프로필 표시도 함께 바뀝니다. 아래는 `moai-cowork` 프로필로 전환한 뒤 언어(Language) 탭을 연 모습입니다.

![moai-cowork 프로필로 전환한 콘솔의 언어(Language) 탭. 대화 언어(Conversation language), 커밋 메시지 언어(Commit message language), 코드 주석 언어(Code comment language), 문서 언어(Documentation language) 네 항목](/images/profile/web-console-switch.png)

LLM 탭에서는 권한 모드(Permission mode)와 모델(Model), 추론 강도(Effort level)를 고칩니다. 터미널 마법사의 "Model Settings" 단계가 다루는 값과 같습니다.

![moai-adk 프로필의 LLM 탭. 권한 모드(Permission mode), 모델(Model), 추론 강도(Effort level) 세 항목](/images/profile/web-console-llm-tab.png)

## 프로필 기록의 범위

콘솔에서 프로필을 전환하면 그 선택이 `~/.moai/claude-profiles/launch.yaml` 에 현재 프로젝트의 기록으로 남습니다. 같은 프로젝트에서 `-p` 없이 `moai cc` 를 실행할 때 이 값이 쓰입니다.

{{< callout type="note" >}}
프로젝트 단위 기록은 다음 릴리스에 포함됩니다. 지금 배포된 버전은 프로젝트 구분 없이 전역 기록 하나만 다룹니다.
{{< /callout >}}

콘솔이 읽는 값과 쓰는 값은 모두 현재 프로젝트를 기준으로 하므로, 화면에 보이는 프로필과 실제로 기록되는 프로필은 언제나 같습니다. 다만 `moai cc -p X` 로 시작한 세션 안에서 콘솔을 열면 `CLAUDE_CONFIG_DIR` 이 이미 정해져 있어, 기록과 무관하게 `X` 를 그대로 표시합니다.

선택 순서와 제약은 [프로필 관리](/ko/cli-reference/profile#프로필-자동-선택)에서 자세히 다룹니다.

---

관련: [프로필 관리](/ko/cli-reference/profile) · [CLI 개요](/ko/getting-started/cli)
