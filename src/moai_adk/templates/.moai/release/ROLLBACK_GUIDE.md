# 🚨 릴리즈 롤백 가이드

> **로컬 전용 문서** - 패키지 배포에 포함되지 않음

긴급 상황에서 실패한 릴리스를 복구하기 위한 절차입니다.

---

## 📊 롤백 결정 기준

### 즉시 롤백 필요 (24시간 이내)

- 🔴 **심각한 버그**: 설치 후 기본 기능 작동 불가
- 🔴 **보안 취약점**: CVE 발표 또는 민감한 정보 노출
- 🔴 **데이터 손상**: 마이그레이션 오류로 사용자 데이터 영향
- 🔴 **호환성 깨짐**: 주요 의존성과 충돌하여 사용 불가능
- 🔴 **배포 오류**: 잘못된 파일 또는 설정으로 배포됨

### 모니터링 후 결정 (24-72시간)

- 🟡 **경미한 버그**: 일부 기능만 영향받음
- 🟡 **성능 저하**: 특정 시나리오에서만 발생
- 🟡 **문서 오류**: 기능은 정상이나 설명이 잘못됨

---

## 🚀 1단계: 현황 파악

### 1.1 PyPI 배포 상태 확인

```bash
# 현재 배포된 버전 확인
curl https://pypi.org/pypi/moai-adk/json | jq '.releases | keys | .[-5:]'

# 상세 정보 확인
curl https://pypi.org/pypi/moai-adk/json | jq '.info'

# 다운로드 통계 (최근 7일)
curl "https://pypistats.org/api/packages/moai-adk/recent" | jq '.data | to_entries | .[-7:]'
```

### 1.2 배포 로그 확인

```bash
# GitHub Actions 배포 로그 보기
gh run list -R modu-ai/moai-adk --status failure --limit 5

# 특정 배포 실패 원인 조회
gh run view <RUN_ID> --log
```

### 1.3 사용자 피드백 수집

- GitHub Issues: 문제 보고된 수 및 심각도
- PyPI 댓글: 사용자 반응 및 영향 범위
- 커뮤니티 채널: 실시간 피드백

---

## 🔙 2단계: 롤백 실행

### 2.1 PyPI 렘 삭제 (문제 버전 제거)

```bash
# 경고: 이 작업은 되돌릴 수 없습니다!
# PyPI 웹사이트 또는 twine 사용

# PyPI 웹사이트 방법 (권장 - UI 확인 가능)
# 1. https://pypi.org/project/moai-adk/ 접속
# 2. Release History → 문제 버전 선택
# 3. "Delete this release" 클릭
# 4. 확인 대화상자에서 버전명 입력 후 삭제

# twine을 통한 삭제 (CLI)
pip install twine
twine delete moai-adk==<VERSION> --repository pypi
# 프롬프트: PyPI 토큰 입력
```

### 2.2 Git 히스토리 복구

```bash
# 이전 안정 버전으로 되돌리기
git log --oneline main | head -20

# 문제 버전 직전 커밋으로 태그 생성
git tag -a v0.22.4 <PREVIOUS_STABLE_COMMIT> -m "Rollback to previous stable version"
git push origin v0.22.4

# 현재 main 브랜치 상태 (되돌리지 않음)
git log --oneline main -5
```

### 2.3 동일한 버전으로 재배포 (수정 후)

```bash
# 1. 문제 수정 커밋
# BUGFIX: [설명]
git commit -m "fix: critical issue in v0.22.5"

# 2. 수정된 버전 재배포
/moai:release patch --force

# 또는 완전 수동 과정:
# - pyproject.toml 버전 동일하게 유지
# - 문제 수정 커밋
# - TestPyPI에서 재검증
# - PyPI로 재배포
```

---

## ⚠️ 3단계: 사후 처리

### 3.1 사용자 공지

```markdown
# 긴급 릴리즈 안내

**v0.22.5 회수됨 (2025-11-12)**

긴급 버그가 발견되어 다음과 같이 조치합니다:
- v0.22.5 PyPI에서 제거됨
- v0.22.4로 롤백 권장
- v0.22.5-fixed 예정 (ETA: 2025-11-13)

### 영향 받은 사용자
- v0.22.5 설치한 경우: 즉시 v0.22.4로 다운그레이드
  ```bash
  pip install moai-adk==0.22.4
  ```

### 관련 이슈
- [#123](github-issue-link) 참조

감사합니다.
```

### 3.2 GitHub 이슈 생성

```bash
# 롤백 이유 및 조치 내용 문서화
gh issue create \
  --title "ROLLBACK: v0.22.5 critical bug" \
  --body "$(cat <<'EOF'
## 롤백 사유
- 치명적인 보안 취약점 발견

## 영향 범위
- 사용자: ~150 (PyPI 통계)

## 조치 내용
- v0.22.5 PyPI에서 제거
- v0.22.4로 롤백

## 예정된 재배포
- v0.22.5-fixed (수정 후)
EOF
)" \
  --label "release" \
  --label "critical"
```

### 3.3 릴리스 노트 작성

```markdown
# v0.22.5 [철회됨]

> **주의**: 이 버전은 치명적인 버그로 인해 철회되었습니다. v0.22.4를 사용하세요.

### 문제점
- [#123] 치명적인 보안 취약점
- [#124] 설치 실패 (특정 환경)

### 권장 조치
```bash
pip install moai-adk==0.22.4
```

### 상태
- 📍 v0.22.5-fixed 개발 중 (ETA: 2025-11-13)
```

---

## 🔍 4단계: 근본 원인 분석

### 4.1 문제 원인 파악

```bash
# 빌드 로그 검토
gh run view <FAILED_RUN_ID> --log > build-log.txt

# 테스트 커버리지 확인
pytest --cov=src/moai_adk tests/ --cov-report=html

# 타입 체크 실패 확인
mypy src/src/moai_adk --show-error-codes

# 보안 감시
bandit -r src/moai_adk --format json > security-scan.json
```

### 4.2 발생 원인 분류

| 원인 | 예시 | 예방책 |
|------|------|--------|
| **테스트 부재** | 커버리지 85% 미만 | CI/CD에서 85% 강제 |
| **의존성 충돌** | 호환되지 않는 버전 | 제약 조건 검토 |
| **타입 오류** | mypy 실패 | --strict 모드 사용 |
| **빌드 오류** | 잘못된 패키징 | 사전 빌드 검증 |
| **문서 오류** | 설치 지침 틀림 | 자동 문서 검증 |

### 4.3 예방 계획 수립

```markdown
# 예방 조치

## 즉시 적용 (현재 릴리스)
- [ ] 테스트 커버리지 85% → 90% 상향
- [ ] 통합 테스트 추가 (특정 환경 시뮬레이션)
- [ ] 의존성 호환성 매트릭스 생성

## 단기 (2주 내)
- [ ] CI/CD 검증 단계 강화
- [ ] 수동 QA 체크리스트 추가
- [ ] TestPyPI 검증 자동화

## 중기 (1개월)
- [ ] 자동 회귀 테스트 구축
- [ ] 카나리 배포 전략 도입
- [ ] 사용자 피드백 자동 수집
```

---

## 📋 5단계: 재배포 체크리스트

### 5.1 문제 수정 확인

- [ ] 버그 수정 코드 리뷰 완료
- [ ] 테스트 추가 및 통과 (해당 버그 시나리오)
- [ ] 회귀 테스트 전체 통과 (pytest --cov)
- [ ] 타입 검사 통과 (mypy --strict)
- [ ] 보안 스캔 통과 (bandit)

### 5.2 사전 배포 검증

- [ ] TestPyPI 배포 성공
- [ ] TestPyPI에서 설치 검증
- [ ] 기본 기능 테스트 (설치 후)
- [ ] 문서 링크 확인
- [ ] README 업데이트

### 5.3 배포 승인

- [ ] 팀 리뷰 완료
- [ ] 릴리스 노트 작성
- [ ] GitHub 이슈 종료 준비
- [ ] 사용자 공지 준비

---

## 🛠️ 롤백 스크립트

### rollback.sh

```bash
#!/bin/bash
# 자동 롤백 스크립트
# 사용: ./rollback.sh v0.22.5

set -e

VERSION=$1
ROLLBACK_TO=${2:-v0.22.4}

if [ -z "$VERSION" ]; then
  echo "Usage: ./rollback.sh <VERSION_TO_ROLLBACK> [ROLLBACK_TO_VERSION]"
  exit 1
fi

echo "🚨 롤백 시작: $VERSION → $ROLLBACK_TO"

# 1. 확인
read -p "정말 롤백할까요? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
  echo "롤백 취소됨"
  exit 0
fi

# 2. PyPI에서 제거
echo "PyPI에서 $VERSION 제거 중..."
twine delete moai-adk==$VERSION --repository pypi

# 3. Git 태그 생성
echo "Git 태그 생성: $ROLLBACK_TO"
git tag -a "$ROLLBACK_TO-rollback-$(date +%Y%m%d)" HEAD -m "Rollback to $ROLLBACK_TO"
git push origin --tags

# 4. GitHub 이슈 생성
echo "GitHub 이슈 생성 중..."
gh issue create \
  --title "ROLLBACK: $VERSION" \
  --body "Automatic rollback to $ROLLBACK_TO" \
  --label "release" \
  --label "critical"

echo "✅ 롤백 완료"
echo "다음 단계:"
echo "1. 문제 수정"
echo "2. /moai:release patch 실행"
```

---

## 🎯 빠른 참조

| 상황 | 명령어 | 결과 |
|------|--------|------|
| **현재 배포 확인** | `curl https://pypi.org/pypi/moai-adk/json \| jq '.info.version'` | 최신 버전 표시 |
| **릴리스 이력 확인** | `gh release list -R modu-ai/moai-adk -L 10` | 최근 10개 릴리스 |
| **배포 실패 원인** | `gh run view <RUN_ID> --log` | 빌드 로그 표시 |
| **버전 삭제** | `twine delete moai-adk==<VERSION>` | PyPI에서 제거 |
| **되돌리기** | `git revert <COMMIT>` | 이전 커밋으로 되돌림 |

---

## 📞 긴급 연락처

- **팀 리더**: GoosLab (@goos)
- **보안 담당**: [보안팀 이메일]
- **릴리스 엔지니어**: Alfred Release Manager

---

## 📚 참고 자료

- [PyPI Package Management](https://pypi.org/help/)
- [Twine Documentation](https://twine.readthedocs.io/)
- [Semantic Versioning](https://semver.org/)
- [GitHub Release Management](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository)

---

**마지막 업데이트**: 2025-11-12
**작성자**: Alfred Release Manager
