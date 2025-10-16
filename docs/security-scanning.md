# @DOC:SECURITY-001 | 보안 스캔 가이드

MoAI-ADK 프로젝트의 보안 스캔 도구 사용 가이드입니다.

## 개요

MoAI-ADK는 두 가지 보안 스캔 도구를 사용합니다:

1. **pip-audit**: 의존성 패키지의 알려진 취약점 검사
2. **bandit**: Python 소스 코드의 보안 이슈 검사

## 로컬 보안 스캔 실행

**보안 도구 설치**:
```bash
pip install pip-audit bandit
```

**pip-audit 실행**:
```bash
pip-audit
```

**bandit 실행**:
```bash
bandit -r src/ -ll
```

옵션 설명:
- `-r src/`: src 디렉토리를 재귀적으로 스캔
- `-ll`: Low severity 이슈는 무시 (Medium/High만 보고)

## CI/CD 통합

GitHub Actions 워크플로우가 자동으로 다음 상황에서 보안 스캔을 실행합니다:

- `main`, `develop`, `feature/**` 브랜치에 push할 때
- Pull Request 생성/업데이트 시

워크플로우 파일: `.github/workflows/security.yml`

## 발견된 취약점 해결

### pip-audit 취약점

pip-audit가 취약점을 발견하면, 다음과 같이 출력됩니다:

```
Name      Version ID                  Fix Versions
--------- ------- ------------------- ------------
starlette 0.46.1  GHSA-2c2j-9gv5-cj73 0.47.2
```

**해결 방법**:
1. `Fix Versions` 열에 표시된 버전으로 업데이트
2. `pyproject.toml`에서 해당 패키지 버전 수정
3. `pip install -e ".[security]"` 실행
4. 다시 `pip-audit` 실행하여 확인

### bandit 보안 이슈

bandit이 보안 이슈를 발견하면:

```
>> Issue: [B605:start_process_with_a_shell] Starting a process with a shell
   Severity: High   Confidence: High
   Location: src/example.py:42
```

**해결 방법**:
1. 파일 위치로 이동 (예: `src/example.py:42`)
2. bandit이 제시하는 보안 권장사항 확인
3. 코드 수정 또는 정당한 사유가 있으면 `# nosec` 주석 추가

**nosec 사용 예시**:
```python
# 정당한 사유로 보안 검사 제외
subprocess.run(cmd, shell=True)  # nosec B605
```

## 심각도 레벨

### pip-audit
- 모든 취약점이 동등하게 중요하게 취급됩니다
- `--strict` 모드에서는 어떤 취약점도 허용하지 않음

### bandit
- **High**: 즉시 수정 필요
- **Medium**: 가능한 빨리 수정
- **Low**: 무시 (기본 설정: `-ll`)

## 보안 의존성 설치

프로젝트에 보안 도구를 포함하려면:

```bash
pip install -e ".[security]"
```

또는 개발 환경 전체 설치:

```bash
pip install -e ".[dev,security]"
```

## 참고 자료

- [pip-audit 공식 문서](https://github.com/pypa/pip-audit)
- [bandit 공식 문서](https://bandit.readthedocs.io/)
- [Python Security Best Practices](https://python.readthedocs.io/en/stable/library/security_warnings.html)

---

**관련 TAG**: @CODE:SECURITY-001
**최종 업데이트**: 2025-10-15
