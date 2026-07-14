---
title: 프로필 관리
weight: 80
draft: false
---

MoAI-ADK의 프로필 시스템으로 여러 Claude Code 설정을 격리하여 관리합니다. 업무용과 개인용, 고품질 세션과 비용 절감 세션을 프로필 하나씩으로 분리해 두면 모델·언어·표시 설정을 매번 바꿀 필요가 없습니다.

## 프로필이란?

프로필은 **격리된 Claude Code 설정 디렉토리**(`CLAUDE_CONFIG_DIR`)입니다. 프로필별로 독립적인 설정, 모델 선택, 언어 환경을 유지할 수 있습니다.

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

인터랙티브 설정 위자드를 실행합니다.

```bash
moai profile setup          # 기본 프로필 설정
moai profile setup work     # "work" 프로필 설정
```

**위자드 설정 항목:**
- **Identity**: 사용자 이름, 역할
- **Languages**: 대화 언어, 코드 주석 언어
- **Model Settings**: 기본 모델, 1M 컨텍스트 모델 선택
- **Display**: 출력 스타일, 상태 표시줄 설정

### moai profile current

현재 활성 프로필 이름을 표시합니다.

```bash
moai profile current
```

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
프로필 미지정 시 기본 프로필이 사용됩니다. 첫 실행 시 자동으로 설정 위자드가 시작됩니다.
{{< /callout >}}

## 1M 컨텍스트 모델 선택

프로필 설정 시 1M 컨텍스트 윈도우를 지원하는 모델을 선택할 수 있습니다. `[1m]` 접미사는 별도 모델이 아니라 Claude Code의 네이티브 컨텍스트 윈도우 수정자입니다.

**선택 가능한 모델 별칭:**
- `opus` / `opus[1m]`
- `sonnet` / `sonnet[1m]`
- `fable` / `fable[1m]`

설정 위자드의 "Model Settings" 단계에서 선택하거나, 프로필 설정 파일을 직접 수정합니다. 1M 컨텍스트 모델은 대용량 코드베이스 분석이나 긴 문서 작업에 적합합니다.

## 프로필 전환 시 동작

| 전환 | 동작 |
|------|------|
| `moai cc` → `moai glm` | GLM 환경 변수 자동 주입 |
| `moai glm` → `moai cc` | GLM 환경 변수 자동 제거 |
| `moai cc` → `moai cg` | GLM env를 tmux 세션에만 주입, Leader는 Claude 유지 |

## 관련 문서

- [CLI 레퍼런스](/getting-started/cli) - 전체 CLI 명령어
- [빠른 시작](/getting-started/quickstart) - 처음 시작하기
- [초기 설정](/getting-started/init-wizard) - 프로젝트 초기화
