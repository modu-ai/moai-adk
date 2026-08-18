---
title: 업데이트
weight: 30
draft: false
---

MoAI-ADK를 최신 버전으로 유지하는 방법을 안내합니다. `moai update` 하나로 바이너리와 템플릿이 함께 갱신되며, 사용자가 만든 커스텀 자산은 자동으로 보존됩니다.

## 업데이트 명령

플래그 없이 실행하면 바이너리와 템플릿을 모두 갱신합니다 — 이것이 기본 동작입니다.

```bash
moai update
```

### 3단계 스마트 업데이트

```mermaid
flowchart TD
    A["moai update 실행"] --> B["Stage 1: 패키지 버전 확인"]
    B --> C{"최신 버전?"}
    C -->|"예"| D["Stage 2: 설정 버전 비교"]
    C -->|"아니오"| E["이미 최신 상태"]
    D --> F{"설정 형식 변경?"}
    F -->|"예"| G["설정 마이그레이션 (백업 후)"]
    F -->|"아니오"| H["설정 유지"]
    G --> I["Stage 3: 템플릿 동기화"]
    H --> I
    I --> J["완료 보고서"]
```

### Stage 1: 패키지 버전 확인

현재 설치된 버전과 GitHub Releases의 최신 버전을 비교합니다.

```bash
# 현재 버전 확인
moai --version

# 사용 가능한 업데이트만 확인 (실제 업데이트 안 함)
moai update --check
```

### 체크섬 의무 검증 (Mandatory Checksum Verification) {#checksum-verification}

`moai update` 의 바이너리 다운로드는 **checksum 검증을 우회할 수 없습니다**. 릴리스의 `checksums.txt` 다운로드가 실패하거나 파싱에 실패하면 업데이트 흐름을 **중단(abort)** 합니다 — 바이너리 다운로드를 시도하지 않습니다.

#### Retry 정책

`checksums.txt` 다운로드는 **3회 retry** 를 지수 백오프로 시도합니다:

| 시도 | 대기 시간 |
|------|-----------|
| 1차 (즉시) | 0s |
| 2차 retry | 2s 대기 |
| 3차 retry | 4s 대기 |
| 추가 retry 없음 | 합계 ~6s 대기 후 실패 |

모든 retry 가 실패하면 다음과 같은 메시지가 출력됩니다:

```
error: checksum unavailable: persistent retry failure after 3 attempts
```

**`--skip-checksum` 같은 우회 옵션은 존재하지 않습니다** (CWE-345 의도된 정책).

#### 실패 시 복구 절차

1. **네트워크 연결 확인**:
   ```bash
   curl -I https://github.com/modu-ai/moai-adk/releases/latest
   ```
2. **Proxy / firewall 확인** — GitHub release asset 도메인 (`github.com`, `objects.githubusercontent.com`) 허용 여부
3. **일시적 GitHub CDN 장애 가능성** — 잠시 후 재시도
4. **수동 바이너리 설치** (영구 차단 시):
   ```bash
   curl -fsSL https://adk.mo.ai.kr/install.sh | bash
   ```
   수동 설치 시 GitHub Release 의 `checksums.txt` 를 별도로 확인하는 것을 권장합니다.

자세한 위협 모델은 [보안 노트 — CWE-345](/ko/advanced/security-notes/#cwe-345) 를 참조하세요.

### Stage 2: 설정 버전 비교

설정 파일의 형식과 호환성을 검사합니다. 형식이 바뀌었으면 자동으로 백업한 뒤 마이그레이션합니다.

**검사 파일:**

- `.moai/config/sections/` 하위 YAML 파일들

{{< callout type="info" >}}
설정 마이그레이션 전에 항상 `.moai/config/` 디렉터리가 백업됩니다.
{{< /callout >}}

### Stage 3: 템플릿 동기화

프로젝트 템플릿과 기본 파일을 최신 버전으로 동기화합니다. 사용자가 손댄 파일은 그대로 보존하고, 새 버전과 충돌하면 백업한 뒤 병합합니다.

```mermaid
graph TD
    A["템플릿 동기화"] --> B["SKILL.md 템플릿"]
    A --> C["에이전트 템플릿"]
    A --> D["규칙 파일"]
    A --> E["설정 기본값"]

    B --> F{"사용자 변경?"}
    C --> F
    D --> F
    E --> F

    F -->|"아니오"| G["자동 업데이트"]
    F -->|"예"| H["백업 후 3-way 병합"]

    G --> I["동기화 완료"]
    H --> I
```

### 삭제 전 백업 (pre-clean backup)

템플릿 재배포 전에 MoAI가 관리하는 뿌리(`.claude/`·`.moai/` 아래 템플릿 관리 경로)를 정리할 때, 그 안에 있던 **템플릿이 배포하지 않는 파일은 먼저 백업한 뒤**에만 정리됩니다. 백업은 실행별 타임스탬프 디렉터리 `.moai-backups/<timestamp>/pre-clean/<root>/...` 아래에 원래 상대 경로 그대로 들어가고, **백업에 실패하면 정리 자체가 중단**됩니다 — 백업 없이 지우는 경로는 없습니다.

이 안전장치는 `moai update`가 관리 뿌리 안에 둔 로컬 전용 파일(개인 규칙, 실험 스킬 등)을 잃지 않게 합니다. 무엇이 관리 뿌리인지와 로컬 전용 파일을 어디에 둬야 하는지는 프로젝트의 로컬 개발 안내를 따르세요. 백업이 만들어졌다면 요약 출력에 그 경로가 함께 안내됩니다.

## 플래그 레퍼런스

| 플래그 | 설명 |
|--------|------|
| `--check` | 새 버전이 있는지만 확인 (업데이트 안 함) |
| `-c, --config` | 설정 마법사 다시 실행 (템플릿 동기화 안 함) |
| `--force` | 강제 업데이트 (버전 일치 스킵, 백업+병합 강제) |
| `--yes` | 모든 확인 자동 승인 (CI/CD 모드) |
| `--templates-only` | 바이너리 업데이트 건너뛰고 템플릿만 동기화 |
| `--binary` | 템플릿 동기화 건너뛰고 바이너리만 업데이트 |
| `--version <tag>` | 최신 대신 특정 릴리스 태그(stable / rc / 이전 버전) 설치 |
| `--dry-run` | 파일시스템 변경 없이 계획된 작업만 표시 |
| `--no-hooks` | Git 훅 설치 건너뛰기 |
| `--verbose` | 모든 경고 표시 (진단 모드) |
| `--shell-env` | Claude Code 용 셸 환경변수 구성 |
| `--profile <high\|medium\|low>` | 모델+effort 프로필 덮어쓰기 (`llm.yaml` 의 `profile` 에 저장) |

### 동작 방식

| 명령어 | 바이너리 업데이트 | 템플릿 동기화 |
|--------|-------------------|---------------|
| `moai update` | {{< icon check ok >}} | {{< icon check ok >}} |
| `moai update --binary` | {{< icon check ok >}} | {{< icon x >}} |
| `moai update --templates-only` | {{< icon x >}} | {{< icon check ok >}} |
| `moai update --check` | {{< icon x >}} | {{< icon x >}} (버전 확인만) |

### 바이너리 전용 업데이트

바이너리만 업데이트하고 템플릿은 동기화하지 않습니다:

```bash
moai update --binary
```

### 특정 버전 설치 (`--version`)

`moai update --version <tag>`는 특정 GitHub 릴리스 태그(stable, rc, 이전 버전)를 기본 업데이트와 동일한 체크섬 검증 다운로드 경로로 설치합니다. 한 플래그로 세 가지 용도를 모두 처리합니다: 검증된 stable 버전 고정, 테스트용 rc 전환, 회귀 후 이전 버전으로 롤백.

```bash
# stable 릴리스 고정
moai update --version v3.0.0

# 앞의 "v"는 생략 가능
moai update --version 3.0.0

# rc 시도
moai update --version v3.1.0-rc1

# 이전 버전으로 롤백
moai update --version v2.14.0
```

{{< callout type="info" >}}
이 플래그는 `api.github.com` 호스트의 `https` 만 사용하며, 다운로드한 바이너리를 릴리스의 공개 체크섬으로 검증합니다 — `--skip-checksum` / `--insecure` 우회는 없습니다. 플랫폼에 맞는 바이너리 에셋이 없거나 체크섬이 일치하지 않으면 0이 아닌 종료 코드로 끝나며 파일시스템은 그대로 남습니다.
{{< /callout >}}

#### 플래그 상호작용 매트릭스

`--version`은 일부 플래그와 배타적이며, 나머지와는 함께 쓸 수 있습니다:

| 함께 쓰는 플래그 | `--version` | 동작 |
|--------|--------|------|
| `--check` | {{< icon x >}} | 배타적 (네트워크 호출 전 사용법 에러) |
| `--templates-only` | {{< icon x >}} | 배타적 |
| `--restore` | {{< icon x >}} | 배타적 |
| `--dry-run` | {{< icon x >}} | 배타적 |
| `--binary` | {{< icon check ok >}} | 요청한 태그의 바이너리만 설치, 템플릿 동기화 건너뜀 |
| `--force` | {{< icon check ok >}} | 실행 중 버전이 이미 일치해도 강제 재설치 |
| `--yes` | {{< icon check ok >}} | 다운그레이드 확인 프롬프트 건너뜀 (CI/CD 모드) |

#### 다운그레이드 확인

요청한 태그가 실행 중 버전보다 오래된 경우, 인터랙티브 터미널에서 확인 프롬프트가 나타납니다. `--yes`를 주거나(CI 등) 비-TTY stdin이면 프롬프트 없이 진행합니다.

#### stable vs. rc 동작

기본 `moai update`(--version 없음)는 GitHub의 `/releases/latest`를 가져오며, 이는 사전 릴리스를 자동으로 제외합니다 — 따라서 rc 및 사전 릴리스 태그는 기본 흐름에서 **절대** 노출되지 않습니다. `--version <tag>`만이 rc나 특정 이전 태그를 명시적으로 설치하는 유일한 경로입니다.

### 템플릿 전용 동기화

템플릿만 동기화하고 바이너리는 업데이트하지 않습니다:

```bash
moai update --templates-only
```

### 설정 마법사 재실행

설정 마법사를 다시 띄워 프로젝트 구성을 바꿉니다(템플릿 동기화는 하지 않습니다):

```bash
moai update -c
# 또는
moai update --config
```

### Dry Run

실제 변경 없이 계획된 아카이브와 설치 작업을 미리 확인합니다:

```bash
moai update --dry-run
```

### CI/CD 모드

모든 확인을 자동 승인합니다:

```bash
moai update --yes
```

## 업데이트 후 절차

### 1단계: 버전 확인

```bash
moai --version
```

### 2단계: 설정 검증

```bash
moai doctor
```

### 3단계: 새로운 기능 확인

```bash
moai --help
```

## 개인 설정 관리

MoAI-ADK 업데이트 시 **CLAUDE.md**와 `settings.json`은 새 버전으로 동기화됩니다. 개인적인 수정 사항은 별도 파일에 보관하세요.

| 파일 | 위치 | 업데이트 영향 |
|------|------|--------------|
| `CLAUDE.md` | 프로젝트 루트 | {{< icon warning warn >}} 업데이트 시 변경됨 (MoAI-ADK 관리) |
| `settings.json` | `.claude/` | {{< icon warning warn >}} 업데이트 시 변경됨 (MoAI-ADK 관리) |
| `CLAUDE.local.md` | 프로젝트 루트 | {{< icon check ok >}} 영향 없음 (개인 설정) |
| `.claude/settings.local.json` | 프로젝트 | {{< icon check ok >}} 영향 없음 (개인 설정) |

{{< callout type="info" >}}
**설정 우선순위:** Local > Project > User > Enterprise<br />
`settings.local.json` 이 프로젝트 설정을 오버라이드합니다.
{{< /callout >}}

### moai 폴더 구조

MoAI-ADK는 다음 폴더에서만 파일을 관리합니다:

```
.claude/
├── agents/
│   ├── moai/                # MoAI-ADK 에이전트 (업데이트 대상)
│   └── harness/             # 사용자 하네스 에이전트 (업데이트 제외, 보존)
│
├── hooks/
│   └── moai/                # MoAI-ADK 훅 스크립트 (업데이트 대상)
│
├── skills/
│   ├── moai-*               # MoAI-ADK 스킬 (moai- 접두사, 업데이트 대상)
│   └── hns-*                # 사용자 생성 스킬 (업데이트 제외, 보존)
│
└── rules/
    └── moai/                # 규칙 파일 (moai 관리)
```

| 유형 | 위치 | 업데이트 영향 |
|------|------|--------------|
| **에이전트** | `agents/moai/` | {{< icon warning warn >}} 업데이트 시 변경됨 |
| **훅** | `hooks/moai/` | {{< icon warning warn >}} 업데이트 시 변경됨 |
| **스킬** | `skills/moai-*` | {{< icon warning warn >}} 업데이트 시 변경됨 |
| **규칙** | `rules/moai/` | {{< icon warning warn >}} 업데이트 시 변경됨 |
| **사용자 에이전트** | `agents/harness/` | {{< icon check ok >}} 업데이트 영향 없음 (보존) |
| **사용자 스킬** | `skills/hns-*` (레거시 `harness-*`, `my-*` 포함) | {{< icon check ok >}} 업데이트 영향 없음 (보존) |

{{< callout type="warning" >}}
**중요:** <code>moai-*</code> 접두사 스킬은 MoAI-ADK가 관리하며 업데이트 시 덮어 쓰입니다. 직접 만든 스킬은 <code>hns-*</code> 접두사(사용자 소유 네임스페이스)를, 에이전트는 <code>.claude/agents/harness/</code> 디렉터리를 사용하세요.
{{< /callout >}}

## 롤백

업데이트 후 문제가 발생하면 이전 버전으로 롤백할 수 있습니다:

```bash
# 인프로세스로 특정 버전 롤백 (권장)
moai update --version <릴리스-태그>

# 부트스트랩 경로 (moai 설치 전): 설치 스크립트 사용
curl -fsSL https://adk.mo.ai.kr/install.sh | bash -s -- --version <릴리스-태그>

# 백업에서 설정 복원
cp -r .moai/config.bak .moai/config
```

{{< callout type="warning" >}}
롤백 전에 현재 작업을 커밋하세요.
{{< /callout >}}

## 문제 해결

### 업데이트 실패

```bash
# 네트워크 확인
curl -I https://github.com/modu-ai/moai-adk/releases/latest

# 수동 재설치
curl -fsSL https://adk.mo.ai.kr/install.sh | bash
```

### 설정 마이그레이션 오류

```bash
# 백업에서 복원
cp -r .moai/config.bak .moai/config

# 설정 검증
moai doctor
```

### 템플릿 충돌

사용자가 수정한 템플릿 파일은 자동으로 백업한 뒤 3-way 병합합니다. 충돌이 나면 `--verbose` 로 상세 경고를 확인하세요:

```bash
moai update --verbose
```

강제로 덮어쓰려면 `--force` 를 사용합니다 (기존 사용자 변경 사항은 `.moai/archive/` 에 백업됩니다):

```bash
moai update --force
```

## 다음 단계

1. **[변경 로그 확인](https://github.com/modu-ai/moai-adk/releases)** — 새 기능 살펴보기
2. **[핵심 개념](/ko/core-concepts/what-is-moai-adk)** — 새 에이전트와 기능 익히기
3. **[빠른 시작](/ko/getting-started/quickstart)** — 새 기능을 프로젝트에 적용하기
