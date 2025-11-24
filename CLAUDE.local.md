# MoAI-ADK 로컬 개발 가이드

## 빠른 시작

**작업 위치**: `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/` → 동기화 → `/Users/goos/MoAI/MoAI-ADK/`

**동기화 실행**:
```bash
bash .moai/scripts/sync-from-src.sh
```

## 파일 동기화 규칙

**동기화 대상**:
- `src/moai_adk/.claude/` ↔ `.claude/`
- `src/moai_adk/.moai/` ↔ `.moai/`

**절대 동기화 금지** (로컬 전용):
- `.claude/settings.local.json`
- `.claude/hooks/`
- `.moai/cache/`, `.moai/logs/`
- `CLAUDE.local.md` (이 파일)

## 코드 표준

- ✅ 모든 코드: 영문 작성
- ✅ 변수명: camelCase 또는 snake_case (언어별)
- ✅ 함수명: camelCase (JS/Python), PascalCase (C#/Java)
- ✅ 클래스명: PascalCase
- ✅ 주석: 영문

**금지**:
```python
# ❌ WRONG
def calculate():  # 점수 계산
    pass

# ✅ CORRECT
def calculate():  # Calculate final score
    pass
```

## 로컬 전용 파일

| 파일 | 위치 | Git 추적 |
|------|------|---------|
| `99-release.md` | `.claude/commands/moai/` | ✅ Yes |
| `CLAUDE.local.md` | 루트 | ✅ Yes |
| `settings.local.json` | `.claude/` | ❌ No |
| `config/config.json` | `.moai/` | ❌ No |
| `cache/`, `logs/` | `.moai/` | ❌ No |

## 자주 사용하는 명령어

```bash
# 동기화
bash .moai/scripts/sync-from-src.sh

# 검증
ruff check src/ && pytest tests/ -v --cov

# 문서 검증
python .moai/tools/validate-docs.py
```

## Git 작업 체크리스트

**커밋 전**:
- [ ] 코드가 영문으로 작성됨
- [ ] 로컬 전용 파일이 포함되지 않음
- [ ] 테스트 통과 (`pytest`)
- [ ] Linting 통과 (`ruff`)

**푸시 전**:
- [ ] 최신 개발 버전으로 rebase
- [ ] 커밋 메시지가 표준 포맷
- [ ] 문서 동기화됨

---

**Version**: 2.0.0 (최적화)
**Created**: 2025-11-24
**Target**: Local developers only
