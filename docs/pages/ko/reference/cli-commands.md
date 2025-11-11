---
title: CLI 명령어
description: MoAI-ADK의 5개 핵심 CLI 명령어와 사용법
---

# CLI 명령어

MoAI-ADK는 개발자가 효율적으로 프로젝트를 관리할 수 있도록 5개의 핵심 CLI 명령어를 제공합니다.

**참고**: CLI 명령어(`moai-adk`)와 Alfred 명령어(`/alfred:`)는 다른 명령어 체계입니다. CLI 명령어는 시스템 관리 및 프로젝트 초기화에 사용되며, Alfred 명령어는 개발 워크플로우에 사용됩니다.

## 명령어 목록

### 1. `init` - 프로젝트 초기화
새로운 MoAI-ADK 프로젝트를 생성합니다.

```bash
moai-adk init [OPTIONS] [PROJECT_NAME]
```

#### 옵션
- `--template`: 프로젝트 템플릿 지정 (기본값: standard)
- `--language`: 프로그래밍 언어 선택 (python, javascript, typescript)
- `--with-mcp`: MCP 서버 자동 설치
- `--mcp-auto`: 추천 MCP 서버 모두 설치

#### 사용 예시
```bash
# 기본 프로젝트 생성
moai-adk init my-project

# TypeScript 프로젝트 생성
moai-adk init --language typescript --template react my-web-app

# MCP 통합 프로젝트
moai-adk init --with-mcp context7 --with-mcp playwright
```

#### 생성되는 파일 구조
```
my-project/
├── .claude/
│   ├── agents/
│   ├── commands/
│   └── skills/
├── .moai/
│   ├── config.json
│   ├── specs/
│   └── docs/
├── src/
├── tests/
├── CLAUDE.md
├── README.md
└── pyproject.toml
```

### 2. `doctor` - 시스템 진단
프로젝트와 환경의 상태를 진단합니다.

```bash
moai-adk doctor [OPTIONS]
```

#### 옵션
- `--verbose`: 상세한 진단 정보 표시
- `--fix`: 자동 수정 가능한 문제 자동 해결
- `--check-updates`: 업데이트 확인

#### 진단 항목
- **프로젝트 구조**: 필수 디렉토리와 파일 확인
- **의존성**: 패키지 의존성 충돌 검사
- **TAG 시스템**: @TAG 체인 완전성 검증
- **Git 상태**: Git 설정과 브랜치 상태 확인
- **환경 변수**: 필수 환경 변수 설정 확인

#### 진단 결과 예시
```
🔍 MoAI-ADK 시스템 진단

✅ 프로젝트 구조: 정상
✅ 의존성: 최신 상태
⚠️  TAG 체인: 3개의 누락된 TAG 발견
✅ Git 설정: 정상
❌ 환경 변수: DATABASE_URL 설정 필요

📊 진단 요약:
- 정상: 3항목
- 경고: 1항목
- 오류: 1항목
```

### 3. `status` - 상태 확인
프로젝트의 현재 상태를 표시합니다.

```bash
moai-adk status [OPTIONS]
```

#### 옵션
- `--format`: 출력 형식 (table, json, markdown)
- `--detailed`: 상세 정보 표시
- `--tag-only`: TAG 상태만 표시

#### 표시 정보
- **프로젝트 정보**: 버전, 생성일, 언어
- **TAG 통계**: 전체 TAG, 체인 완전도
- **파일 상태**: 생성된 파일 수, 최근 변경
- **Git 정보**: 현재 브랜치, 커밋 상태
- **품질 지표**: 테스트 커버리지, 코드 품질

### 4. `backup` - 백업 관리
프로젝트 데이터를 백업하고 복원합니다.

```bash
moai-adk backup [SUBCOMMAND] [OPTIONS]
```

#### 서브명령어
- `create`: 백업 생성
- `list`: 백업 목록 표시
- `restore`: 백업 복원
- `cleanup`: 오래된 백업 정리

#### 옵션
- `--path`: 백업 저장 경로 지정
- `--compress`: 압축하여 백업
- `--include`: 포함할 파일/디렉토리
- `--exclude`: 제외할 파일/디렉토리

#### 사용 예시
```bash
# 백업 생성
moai-adk backup create --compress

# 백업 목록 확인
moai-adk backup list

# 특정 백업 복원
moai-adk backup restore backup-2024-01-15.tar.gz

# 7일 이상 된 백업 정리
moai-adk backup cleanup --days 7
```

### 5. `update` - 업데이트 관리
MoAI-ADK와 관련 패키지를 업데이트합니다.

```bash
moai-adk update [OPTIONS] [PACKAGE_NAME]
```

#### 옵션
- `--check-only`: 업데이트 가능 여부만 확인
- `--dry-run`: 업데이트 시뮬레이션
- `--force`: 강제 업데이트
- `--all`: 모든 패키지 업데이트

#### 업데이트 대상
- **MoAI-ADK 코어**: 핵심 프레임워크
- **에이전트 템플릿**: 최신 에이전트 버전
- **스킬 라이브러리**: 새로운 스킬과 기능
- **CLI 도구**: 명령어 기능 개선

#### 사용 예시
```bash
# MoAI-ADK 코어 업데이트
moai-adk update moai-adk

# 전체 패키지 업데이트
moai-adk update --all

# 업데이트 가능성 확인
moai-adk update --check-only
```

## 추가 CLI 명령어

다음 명령어들은 핵심 명령어는 아니지만 `__main__.py`에서 직접 추가되어 사용할 수 있습니다.

### `validate-links` - 링크 검증
문서의 온라인 링크 유효성을 검증합니다.

```bash
moai-adk validate-links [OPTIONS]
```

#### 옵션
- `--file`: 검증할 파일 경로 (기본값: README.ko.md)
- `--max-concurrent`: 동시 검색 수 (기본값: 3)
- `--timeout`: 요청 타임아웃 (기본값: 8초)
- `--output`: 결과 저장 파일 경로
- `--verbose`: 상세 진행 상황 표시

### `improve-ux` - 사용자 경험 개선
웹사이트의 사용자 경험을 분석하고 개선 제안을 제공합니다.

```bash
moai-adk improve-ux [OPTIONS]
```

#### 옵션
- `--url`: 분석할 URL (기본값: https://adk.mo.ai.kr)
- `--output`: 분석 결과 저장 경로
- `--format`: 출력 형식 (json, markdown, text)
- `--verbose`: 상세 진행 상황 표시
- `--max-workers`: 동시 작업 수 (기본값: 5)

## CLI 명령어 vs Alfred 명령어

MoAI-ADK에는 두 종류의 명령어 체계가 있습니다:

### CLI 명령어 (moai-adk)
- **목적**: 시스템 관리, 프로젝트 초기화, 유지보수
- **실행**: `moai-adk <command>` 또는 `uv run moai-adk <command>`
- **명령어**: init, doctor, status, backup, update 등
- **사용처**: 터미널에서 직접 실행

### Alfred 명령어 (/alfred:)
- **목적**: 개발 워크플로우, SPEC 작성, TDD 사이클
- **실행**: Claude 대화창에서 `/alfred:<command>` 형식으로 호출
- **명령어**: /alfred:0-project, /alfred:1-plan, /alfred:2-run, /alfred:3-sync
- **사용처**: Claude Code 대화 환경

#### 명령어 비교표

| 목적 | CLI 명령어 | Alfred 명령어 |
|------|------------|---------------|
| 프로젝트 생성 | `moai-adk init` | `/alfred:0-project` |
| 시스템 진단 | `moai-adk doctor` | 해당 없음 |
| 개발 계획 | 해당 없음 | `/alfred:1-plan` |
| 코드 구현 | 해당 없음 | `/alfred:2-run` |
| 상태 확인 | `moai-adk status` | `/alfred:3-sync` (일부) |
| 백업 관리 | `moai-adk backup` | 해당 없음 |

## 고급 기능

### 1. 명령어 체이닝
여러 명령어를 파이프라인으로 연결하여 실행할 수 있습니다.

```bash
# 프로젝트 진단 후 자동 수정
moai-adk doctor --fix | moai-adk status --format json

# 백업 후 상태 확인
moai-adk backup create && moai-adk status --detailed
```

### 2. 환경별 설정
`.moai/config.json` 파일에서 환경별 설정을 관리할 수 있습니다.

```json
{
  "cli": {
    "default_format": "table",
    "auto_backup": true,
    "backup_retention_days": 30
  },
  "environments": {
    "development": {
      "verbose": true,
      "auto_fix": true
    },
    "production": {
      "verbose": false,
      "auto_fix": false,
      "require_confirmation": true
    }
  }
}
```

### 3. 자동화 스크립트
CLI 명령어를 활용한 자동화 스크립트 예시입니다.

```bash
#!/bin/bash
# project-setup.sh

echo "🚀 MoAI-ADK 프로젝트 설정 시작"

# 프로젝트 초기화
moai-adk init $1 --template standard --language python

# 시스템 진단
moai-adk doctor --fix

# 초기 백업 생성
moai-adk backup create --compress

# 상태 확인
moai-adk status --detailed

echo "✅ 프로젝트 설정 완료!"
```

## 문제 해결

### 1. 일반적인 문제

#### 명령어를 찾을 수 없음
```bash
# PATH에 moai-adk 추가
export PATH="$PATH:/path/to/moai-adk"

# 또는 uv를 통해 실행
uv run moai-adk --help
```

#### 권한 오류
```bash
# 적절한 권한으로 실행
chmod +x $(which moai-adk)

# 또는 sudo로 실행 (권장하지 않음)
sudo moai-adk doctor
```

#### 의존성 충돌
```bash
# 가상 환경 생성
python -m venv venv
source venv/bin/activate

# 의존성 재설치
pip uninstall moai-adk
pip install moai-adk
```

### 2. 성능 최적화

#### 병렬 실행
```bash
# 여러 명령어 병렬 실행
moai-adk doctor --check-updates &
moai-adk backup create &
wait
```

#### 캐싱 활용
```bash
# 캐시된 결과 사용
moai-adk status --cached
moai-adk validate-links --use-cache  # 추가 CLI 명령어
```

## 모범 사례

### 1. 정기적인 유지보수
```bash
# 주간 유지보수 스크립트
moai-adk doctor --fix
moai-adk update --check-only
moai-adk backup create
moai-adk status
```

### 2. CI/CD 통합
```yaml
# .github/workflows/moai-check.yml
name: MoAI Quality Check
on: [push, pull_request]
jobs:
  quality-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup MoAI-ADK
        run: pip install moai-adk
      - name: Run quality checks
        run: |
          moai-adk doctor
          moai-adk validate-links  # 추가 CLI 명령어
```

### 3. 팀 협업
```bash
# 팀 표준 설정 파일
echo '{"cli": {"default_format": "json", "auto_backup": true}}' > .moai/cli-config.json

# 공유 스크립트
echo "#!/bin/bash\nmoai-adk doctor && moai-adk status" > team-check.sh
chmod +x team-check.sh
```

이러한 CLI 명령어들을 활용하여 MoAI-ADK 프로젝트를 효율적으로 관리할 수 있습니다. 개발 워크플로우는 Alfred 명령어(`/alfred:`)와 함께 사용하시기 바랍니다.