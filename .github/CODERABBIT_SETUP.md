# 🐰 CodeRabbit AI Setup Guide

CodeRabbit을 이용한 자동 AI PR 리뷰 및 자동 승인 시스템 구성 가이드입니다.

## 📋 개요

- **도구**: [CodeRabbit AI](https://github.com/marketplace/coderabbitai)
- **설치 방식**: GitHub Marketplace 앱 (GitHub App)
- **기능**: 자동 코드 리뷰, Agentic Chat, 자동 수정, PR 요약
- **가격**: 오픈소스 무료 Pro (본 프로젝트 적용)
- **상태**: ✅ 설정 완료

## 🚀 설치 방법

### 1단계: GitHub Marketplace에서 설치

```
1. https://github.com/marketplace/coderabbitai 접속
2. "Install it for free" 또는 "Setup & Pricing" 클릭
3. 조직 선택: modu-ai
4. 저장소 선택:
   ✅ "Only select repositories" 선택
   ✅ moai-adk 선택
5. "Install" 클릭
6. GitHub 인증 허용
```

### 2단계: CodeRabbit 계정 연결 (처음 설치 시)

CodeRabbit 대시보드에서:

```
1. https://app.coderabbit.ai 방문
2. GitHub로 로그인
3. 저장소 인증 허용
4. moai-adk 저장소 선택
5. 설정 완료
```

### 3단계: 저장소 설정 확인

**CodeRabbit Dashboard → Settings → moai-adk:**

```yaml
# 기본 설정
✅ Enable auto-review: 활성화
✅ Enable PR description update: 활성화 (선택)
✅ Enable auto-suggestions: 활성화

# Pro 기능
✅ Enable auto-approval: 활성화 (Pro)
✅ Enable Agentic Chat: 활성화 (Pro)
✅ Enable auto-fix suggestions: 활성화 (Pro)

# 리뷰 정책
📋 Review mode: Detailed
📋 Languages: Auto-detect
📋 Ignore patterns: node_modules/, dist/, build/
```

## ✨ 주요 기능

### 1. 자동 PR 리뷰

PR이 생성되면 자동으로:

```
📝 PR 요약 생성
├─ 변경사항 분석
├─ 영향 범위 파악
└─ 변경 의도 설명

🔍 코드 리뷰 수행
├─ 코드 품질 분석
├─ 보안 이슈 검출
├─ 성능 최적화 제안
└─ 테스트 커버리지 확인

💬 리뷰 코멘트 게시
├─ Line-by-line 피드백
├─ 자동 수정 제안 (클릭 1번)
└─ 컨텍스트 기반 설명
```

### 2. Agentic Chat (Pro 기능) 💬

PR 내에서 CodeRabbit과 대화:

```
사용자: "@coderabbit, explain this logic"
CodeRabbit: 상세 설명 제공

사용자: "@coderabbit, suggest better approach"
CodeRabbit: 최적화 방안 제시

사용자: "@coderabbit, fix this issue"
CodeRabbit: 자동 수정 제안 (클릭으로 적용)
```

### 3. 자동 승인 (Pro 기능) 🚀

설정된 품질 기준 충족 시:

```
✅ 자동 승인 조건:
  - 코드 품질 기준 충족
  - 보안 이슈 없음
  - 테스트 케이스 포함
  - 문서 업데이트 있음

🎯 자동 승인 결과:
  - PR이 자동으로 승인됨
  - 병합 조건 만족
  - CI/CD 파이프라인 실행 가능
```

## 🎯 사용 흐름

### 개발자 관점

```
1. Feature 브랜치 생성
   git checkout -b feature/my-feature

2. 코드 작성 및 커밋
   git add . && git commit -m "feat: implement feature"

3. PR 생성
   git push origin feature/my-feature
   gh pr create --title "feat: implement feature"

4. CodeRabbit 자동 리뷰 대기 (1-2분)
   📧 PR 수신
   🤖 CodeRabbit 리뷰 시작
   💬 리뷰 코멘트 게시

5. 피드백 확인 및 대응
   - 자동 수정 제안 클릭
   - Agentic Chat으로 설명 요청
   - 필요시 코드 수정

6. 품질 기준 충족 시 자동 승인
   ✅ CodeRabbit 자동 승인
   ✅ merge 가능 상태

7. PR 병합
   gh pr merge <pr-number>
```

## 📊 리뷰 기준

CodeRabbit이 검토하는 항목:

| 범주 | 평가 항목 | Python 특화 |
|------|---------|-----------|
| **코드 품질** | 가독성, 유지보수성, 디자인 패턴 | PEP-8, Type hints |
| **보안** | OWASP Top 10, 취약점, 암호화 | SQL Injection, XSS |
| **성능** | 복잡도, 알고리즘 최적화 | O(n) 분석 |
| **테스트** | 커버리지, 엣지 케이스, 픽스처 | pytest, coverage |
| **문서** | Docstring, 주석, README | Sphinx 호환 |
| **릴리즈** | 버전 관리, 변경로그 | Semantic Versioning |

## 🔧 고급 설정

### `.coderabbit.yaml` (저장소 루트)

선택사항: 더 세밀한 제어를 위해 설정 파일 생성

```yaml
# .coderabbit.yaml
coderabbit:
  language: python
  framework: general

  # 리뷰 정책
  rules:
    - name: "Python Code Quality"
      files: "*.py"
      instructions: |
        - Check for PEP-8 compliance
        - Verify type hints are present
        - Ensure docstrings for functions
        - Validate error handling

    # 무시할 파일
    - name: "Ignore patterns"
      ignore:
        - "*.md"
        - "docs/"
        - "build/"
        - "dist/"

  # 자동 승인 조건
  auto_approve:
    enabled: true
    min_score: 0.80  # 80% 품질 점수
    conditions:
      - tests_pass: true
      - coverage_above: 0.85
      - no_security_issues: true

  # 리뷰 대상 언어
  languages:
    - python
```

저장 위치: `.coderabbit.yaml` (저장소 루트)

**설정 후:**
```bash
git add .coderabbit.yaml
git commit -m "config: add CodeRabbit configuration"
git push origin feature-branch
```

## 🎓 사용 팁

### 1. Agentic Chat 활용

PR 코멘트에서:

```
# 코드 설명 요청
@coderabbit explain this function

# 최적화 제안
@coderabbit suggest a more efficient approach

# 버그 찾기
@coderabbit find potential bugs

# 테스트 생성
@coderabbit write unit tests for this

# 자동 수정
@coderabbit fix this issue
```

### 2. 자동 수정 적용

CodeRabbit 코멘트의 "Fix" 버튼 클릭:

```
CodeRabbit Comment:
"This function should include type hints"
[Fix] ← 클릭하면 자동으로 수정됨
```

### 3. 리뷰 피드백 활용

```
✅ 자동 수정 제안 → 1-click 적용
💬 설명 요청 → Agentic Chat으로 상세 설명 받기
🔄 재리뷰 → 수정 후 재리뷰 자동 실행
```

## 🐛 트러블슈팅

### Q: CodeRabbit이 PR을 리뷰하지 않음

**A: 다음을 확인하세요:**

1. GitHub App 설치 확인
   ```bash
   # Settings → Installed GitHub Apps
   # CodeRabbit이 있는지 확인
   ```

2. 저장소 권한 확인
   ```bash
   # CodeRabbit Settings → Repositories
   # moai-adk이 선택되어 있는지 확인
   ```

3. 워크플로우 상태 확인
   ```bash
   # Actions 탭에서 실패한 워크플로우 확인
   ```

### Q: 자동 승인이 안 됨

**A: Pro 기능 확인:**

1. CodeRabbit 계정이 Pro인지 확인
2. Dashboard → Settings → Auto-approval 활성화 확인
3. 최소 품질 점수 확인 (기본값: 80%)

### Q: 리뷰 코멘트가 너무 많음

**A: 설정 조정:**

```yaml
# .coderabbit.yaml
coderabbit:
  review_mode: "summary"  # "detailed" 대신 사용
  # 또는 Dashboard에서 "Review mode"를 "Summary"로 변경
```

### Q: 특정 파일은 리뷰 안 하고 싶음

**A: 무시 패턴 설정:**

```yaml
# .coderabbit.yaml
coderabbit:
  ignore:
    - "*.md"
    - "docs/"
    - "migrations/"
    - "build/"
```

## 📞 지원 및 문서

- **CodeRabbit 공식 문서**: https://docs.coderabbit.ai
- **GitHub Issues**: https://github.com/coderabbit/issues
- **커뮤니티 Slack**: [CodeRabbit Community](https://slack.coderabbit.ai)
- **MoAI-ADK CLAUDE.md**: ../../CLAUDE.md

## 💡 MoAI-ADK 통합

### TRUST 5 원칙과의 연결

```
🤖 CodeRabbit (자동 리뷰)
   ↓
✅ Test First       → pytest 커버리지 검증
✅ Readable         → 코드 가독성 확인
✅ Unified          → 스타일 일관성 확인
✅ Secured          → 보안 이슈 검출
```

### Alfred와의 연동

```
User creates PR (feature branch)
   ↓
GitHub Actions (moai-gitflow.yml)
   ├─ Run tests
   └─ Validate TRUST 5
   ↓
CodeRabbit (자동 실행)
   ├─ AI 코드 리뷰
   ├─ 보안 검사
   └─ 자동 승인 (Pro)
   ↓
PR Ready for Merge
```

## 🎉 완료!

CodeRabbit이 모든 PR을 자동으로 리뷰하고, Pro 기능으로 자동 승인까지 처리합니다!

```
이전 (수동 리뷰):
PR 생성 → 대기 → 수동 리뷰 → 승인 → 병합

이후 (CodeRabbit):
PR 생성 → 1-2분 → 자동 리뷰 + 자동 승인 → 병합 가능 ✨
```

---

✨ **다음 단계**: 테스트 PR을 생성하여 CodeRabbit이 제대로 작동하는지 확인하세요!
