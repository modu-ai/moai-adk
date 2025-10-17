# @DOC:CMD-CLI-001 | Chain: @SPEC:DOCS-003 -> @DOC:CMD-001

# CLI Reference

MoAI-ADK CLI 명령어 전체 레퍼런스입니다.

## moai-adk init

새 프로젝트 초기화:

```bash
moai-adk init <project-name> [options]
```

### Options

- `--language, -l`: 프로젝트 언어 (python, typescript, java, go)
- `--template, -t`: 템플릿 선택 (basic, full, minimal)
- `--mode, -m`: 모드 (personal, team)

### Examples

```bash
# Python 프로젝트 생성
moai-adk init my-project --language python

# Team 모드로 Full 템플릿 프로젝트 생성
moai-adk init team-project --mode team --template full
```

---

## moai-adk version

버전 정보 확인:

```bash
moai-adk --version
```

---

## moai-adk config

설정 확인 및 수정:

```bash
moai-adk config [command]
```

### Subcommands

- `show`: 현재 설정 표시
- `set <key> <value>`: 설정 변경
- `reset`: 기본값으로 리셋

### Examples

```bash
# 현재 설정 확인
moai-adk config show

# 모드 변경
moai-adk config set mode team

# 설정 리셋
moai-adk config reset
```

---

**다음**: [Alfred Commands →](alfred-commands.md)
