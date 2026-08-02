---
title: 프로필 관리
weight: 40
draft: false
---

MoAI-ADK의 프로필 시스템은 여러 Claude Code 설정을 격리해 관리합니다. 업무용과 개인용, 고품질 세션과 비용 절감 세션을 프로필 하나씩으로 분리해 두면 모델·언어·표시 설정을 매번 바꿀 필요가 없습니다.

## 프로필이란?

프로필은 **격리된 Claude Code 설정 디렉터리**(`CLAUDE_CONFIG_DIR`)입니다. 프로필마다 설정, 모델 선택, 언어 환경을 따로 유지할 수 있습니다.

```
~/.moai/claude-profiles/
├── default/           # 기본 프로필
│   ├── settings.json
│   └── settings.local.json
├── work/              # 업무용 프로필
│   ├── settings.json
│   └── settings.local.json
└── personal/          # 개인용 프로필
    └── ...
```

## 명령어 레퍼런스

### moai profile list

사용 가능한 모든 프로필을 표시합니다.

```bash
moai profile list
```

### moai profile setup [name]

대화형 설정 마법사를 실행합니다.

```bash
moai profile setup          # 기본 프로필 설정
moai profile setup work     # "work" 프로필 설정
```

**마법사가 묻는 항목:**
- **Identity**: 사용자 이름, 역할
- **Languages**: 대화 언어, 코드 주석 언어
- **Model Settings**: 기본 모델, 1M 컨텍스트 모델 선택
- **Display**: 출력 스타일, 상태 표시줄 설정

### moai profile current

현재 활성 프로필 이름을 표시합니다.

```bash
moai profile current
```

이 값은 전역 기록을 기준으로 하므로, 프로젝트마다 기억하는 프로필이 다를 때는 [프로필 자동 선택](#프로필-자동-선택)의 제약을 함께 참고하세요.

### moai profile delete [name]

프로필을 삭제합니다.

```bash
moai profile delete old-profile
```

## 프로필로 Claude Code 실행

`-p` (또는 `--profile`) 플래그로 프로필을 지정합니다.

```bash
moai cc -p work          # work 프로필로 Claude 실행
moai glm -p cost-save    # cost-save 프로필로 GLM 실행
moai cg -p team          # team 프로필로 CG 모드 실행
```

{{< callout type="info" >}}
`-p` 로 지정한 프로필은 언제나 우선합니다. 지정하지 않았을 때 어떤 프로필이 쓰이는지는 아래 [프로필 자동 선택](#프로필-자동-선택)을 참고하세요. 프로필을 처음 사용할 때는 설정 마법사가 자동으로 시작됩니다.
{{< /callout >}}

## 프로필 자동 선택

`-p` 없이 `moai cc` 를 실행하면 `~/.moai/claude-profiles/launch.yaml` 기록을 참고해 프로필을 고릅니다. 이 기록은 `-p` 로 이름 있는 프로필을 실행할 때마다 갱신됩니다.

{{< callout type="note" >}}
아래 설명하는 프로젝트별 기억은 다음 릴리스에 포함됩니다. 지금 배포된 버전은 전역 기록 하나(`last_profile`)만 남기므로, 프로젝트 B 에서 `-p` 로 프로필을 지정하면 프로젝트 A 가 기억하던 값이 덮어써집니다.
{{< /callout >}}

`launch.yaml` 은 전역 기록과 함께, 프로젝트의 절대 경로를 키로 삼는 `projects:` 목록을 유지합니다. `-p` 없는 실행이 프로필을 정하는 순서는 다음과 같습니다.

1. 현재 프로젝트가 기억하는 프로필 (`projects:` 항목)
2. 전역 기록 (`last_profile`)
3. 기본 프로필

기록된 프로필이라도 디렉터리가 이미 지워졌다면 건너뛰고 다음 순서로 넘어갑니다. `-p` 로 지정한 이름은 이 순서 전체보다 앞서며, `-p default` 로 기본 프로필을 명시할 수도 있습니다.

두 조회를 모두 끄려면 환경 변수를 지정합니다.

```bash
MOAI_NO_PROFILE_FALLBACK=1 moai cc    # 기록을 무시하고 기본 프로필로 실행
```

프로젝트별 기록은 `-p` 로 실행할 때 남고, [웹 콘솔](/ko/cli-reference/web)에서 프로필을 전환할 때도 함께 갱신됩니다. 기본 프로필(`default`)은 기록 대상이 아닙니다.

**알아둘 제약**

- 프로젝트 디렉터리를 옮기거나 이름을 바꾸면 기존 항목은 어느 경로와도 맞지 않게 됩니다. 이 항목은 조용히 건너뛰므로 실행에 문제를 일으키지는 않습니다.
- `projects:` 목록은 프로젝트가 늘어날수록 함께 늘어나며, 정리해 주는 명령은 아직 없습니다.
- `moai profile current` 는 전역 기록을 그대로 보여줍니다. 따라서 기억된 프로필이 전역 기록과 다른 프로젝트에서는, `moai profile current` 가 알려주는 이름과 `-p` 없는 `moai cc` 가 실제로 띄우는 프로필이 서로 다를 수 있습니다.

## 새 프로필의 첫 실행

새로 만든 프로필 디렉터리에는 Claude Code 의 계정 상태를 담는 `.claude.json` 이 아직 없습니다. 계정 상태는 어느 플랫폼에서든 설정 디렉터리마다 따로 관리되므로, 쓰던 세션이 멀쩡하더라도 새 프로필로 처음 실행하면 로그인·온보딩 화면이 나타납니다.

{{< callout type="note" >}}
아래 안내 메시지는 다음 릴리스에 포함됩니다. 지금 배포된 버전은 아무 예고 없이 로그인 화면으로 넘어갑니다.
{{< /callout >}}

런처는 Claude Code 를 띄우기 전에 표준 오류로 다음을 알립니다.

```
Notice: profile "work" has no Claude Code configuration yet.
  Claude Code will show the login / onboarding screen on this launch.
  Account state is not inherited between profiles; sign in once and it
  persists for this profile.
```

자격 증명을 새 프로필로 복사하거나 옮기는 동작은 하지 않습니다. 계정 상태를 담는 곳이 플랫폼마다 달라, 한쪽에 맞춘 복사는 다른 쪽에서 어긋나기 때문입니다. 한 번 로그인하면 그 상태가 해당 프로필에 남으므로 다음 실행부터는 이 화면이 나오지 않습니다.

## 1M 컨텍스트 모델 선택

프로필 설정 시 1M 컨텍스트 윈도우를 지원하는 모델을 선택할 수 있습니다. `[1m]` 접미사는 별도 모델이 아니라 Claude Code의 네이티브 컨텍스트 윈도우 수정자입니다.

**선택 가능한 모델 별칭:**
- `opus` / `opus[1m]`
- `sonnet` / `sonnet[1m]`
- `fable` / `fable[1m]`

설정 마법사의 "Model Settings" 단계에서 고르거나, 프로필 설정 파일을 직접 고쳐도 됩니다. 1M 컨텍스트 모델은 코드베이스를 통째로 분석하거나 긴 문서를 다룰 때 유리합니다.

## 프로필 전환 시 동작

| 전환 | 동작 |
|------|------|
| `moai cc` → `moai glm` | GLM 환경 변수 자동 주입 |
| `moai glm` → `moai cc` | GLM 환경 변수 자동 제거 |
| `moai cc` → `moai cg` | GLM env를 tmux 세션에만 주입, Leader는 Claude 유지 |

## 관련 문서

- [moai web 웹 콘솔](/ko/cli-reference/web) - 브라우저에서 프로필 전환·편집
- [CLI 레퍼런스](/ko/getting-started/cli) - 전체 CLI 명령어
- [빠른 시작](/ko/getting-started/quickstart) - 처음 시작하기
- [초기 설정](/ko/getting-started/init-wizard) - 프로젝트 초기화
