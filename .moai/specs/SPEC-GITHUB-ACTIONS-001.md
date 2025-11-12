---
id: SPEC-GITHUB-ACTIONS-001
title: Claude Code GitHub Actions 통합
status: completed
created: 2025-11-07T17:00:00+09:00
updated: 2025-11-07T17:00:00+09:00
priority: high
version: 1.0.0
---


# SPEC-GITHUB-ACTIONS-001: Claude Code GitHub Actions 통합

> GitHub Actions에서 Claude가 자동으로 PR을 처리하고 CodeRabbit과 함께 작업하도록 통합

## 📋 목표

- GitHub Actions에서 Claude Code를 활용하여 Issue → PR 자동 생성
- CodeRabbit과 Claude GA의 역할 분담 및 통합
- MoAI-ADK TRUST 5 원칙 준수하며 자동화 확대
- 점진적 전개: 기본 인프라 구축 후 기능 확장

---

## 📋 EARS 요구사항

### Ubiquitous (항상 적용되는 요구사항)

**REQ-001**: GitHub Actions 워크플로우는 MoAI-ADK 표준을 따라야 한다.
```
Given: 어떤 워크플로우든
When: .github/workflows/ 디렉토리에 생성됨
Then: 다음을 준수해야 함
  • CLAUDE.md 스타일 준용
  • TRUST 5 원칙 검증
  • CodeRabbit과 충돌 없음
```

**REQ-002**: GitHub Secrets는 안전하게 보관되어야 한다.
```
Given: API Key 또는 민감한 정보
When: GitHub Actions에서 사용됨
Then: 다음을 보장해야 함
  • Secrets로 암호화되어 저장
  • 워크플로우 로그에 마스킹됨
  • 코드에 하드코딩되지 않음
  • 정기적 로테이션 (90일)
```

**REQ-003**: Claude GA와 CodeRabbit은 함께 작동해야 한다.
```
Given: Issue 또는 PR 생성
When: GitHub Actions 실행
Then: 다음 순서로 진행
  1. Claude GA: PR 생성 또는 분석
  2. CodeRabbit: 자동 리뷰
  3. MoAI-ADK: 검증 + 문서화
  4. 자동 병합 또는 수동 병합
```

### Event-driven (이벤트 기반)

**REQ-004**: Issue 코멘트에서 @claude 감지 시 PR을 생성해야 한다.
```
Given: Issue 코멘트에 "@claude <command>"
When: GitHub Actions Issue Handler 실행
Then: 다음을 수행
  • 코멘트 파싱 및 명령어 추출
  • Claude API 호출 (향후)
  • feature/SPEC-XXXX 브랜치 생성
  • 코드 생성 및 커밋
  • Draft PR 자동 생성
  • 상태 코멘트 게시
```

**REQ-005**: PR 생성 시 자동으로 분석해야 한다.
```
Given: Pull Request 생성됨
When: GitHub Actions PR Validator 실행
Then: 다음을 분석하고 코멘트 게시
  ✓ SPEC 문서 확인
  ✓ 테스트 포함 여부
  ✓ TRUST 5 원칙 준수 확인
  ✓ 파일 변경사항 분석
```

**REQ-006**: Draft PR → Ready for Review 변경 시 자동 SYNC를 트리거해야 한다.
```
Given: Draft PR이 Ready for Review로 변경됨
When: GitHub Actions Auto-Sync Trigger 실행
Then: 다음을 자동으로 실행
  • 문서 동기화 (향후)
  • CHANGELOG 생성 (향후)
  • 병합 준비 확인
  • 상태 코멘트 게시
```

### State-driven (상태 기반)

**REQ-007**: CI/CD 상태에 따라 병합 준비 상태를 결정해야 한다.
```
Given: PR이 다음 상태임
  • 모든 체크 완료
  • CodeRabbit 승인
  • TRUST 5 검증 완료
When: GitHub Actions Merge Readiness Check 실행
Then: 병합 준비 완료 상태로 표시하고
  • PR을 merge-ready 레이블 추가
  • 개발자에게 수동 병합 가능함을 알림
  • (향후) 자동 병합 실행
```

**REQ-008**: 실패 상태를 감지하고 알려야 한다.
```
Given: 워크플로우 단계에서 오류 발생
When: GitHub Actions 실행 중
Then: 다음을 수행
  • PR 코멘트로 오류 상세 내용 게시
  • 개발자에게 액션 항목 제시
  • 오류 로그 링크 제공
  • MoAI 커뮤니티에 에스컬레이션 (심각한 경우)
```

### Optional (선택적 기능)

**REQ-009**: Slack/Discord 알림을 지원해야 한다. (선택)
```
Given: PR 또는 워크플로우 중요 이벤트
When: GitHub Actions 실행 완료
Then: (선택적) 다음을 수행
  • Slack 채널에 알림
  • Discord 서버에 알림
  • 커스텀 웹훅 호출
```

**REQ-010**: 자동 병합을 지원해야 한다. (선택)
```
Given: 모든 조건 충족
When: CodeRabbit이 승인하고 모든 체크 통과
Then: (선택적) 다음을 수행
  • feature branch를 develop으로 자동 병합
  • develop을 main으로 자동 병합 (릴리즈 시)
  • 브랜치 자동 삭제
```

### Unwanted Behaviors (제거해야 할 동작)

**UB-001**: CodeRabbit 제거
```
❌ 하면 안 됨: CodeRabbit을 제거하고 Claude GA만 사용
이유:
  • CodeRabbit은 "자동 리뷰"에 특화
  • Claude GA는 "자동 PR 생성"에 특화
  • 역할이 다르므로 함께 운영해야 함
  • 둘 다 활성화되면 완전한 자동화 달성 가능
```

**UB-002**: API Key 하드코딩
```
❌ 하면 안 됨: 워크플로우에 API Key를 직접 입력
올바른 방법:
  • GitHub Secrets로 저장
  • ${{ secrets.ANTHROPIC_API_KEY }}로만 접근
  • 코드 리뷰 시 민감도 높음
```

**UB-003**: 자동 병합 without 검증
```
❌ 하면 안 됨: 모든 조건을 확인하지 않고 자동 병합
필수 조건:
  • 모든 테스트 통과
  • CodeRabbit 자동 승인
  • TRUST 5 원칙 준수
  • 코드 리뷰 완료
```

**UB-004**: MoAI-ADK 표준 무시
```
❌ 하면 안 됨: GitHub Actions 워크플로우가 CLAUDE.md, TRUST 5를 무시
올바른 방법:
  • SPEC 문서와 CODE 링크
  • 커밋 메시지에 코드, 스펙, 테스트 참조
```

---

## 🎯 구현 계획

### Phase 1: 기본 인프라 (완료)

- [x] GitHub Actions 워크플로우 파일 생성
  - claude-github-actions.yml
  - 4개 주요 Job 정의

- [x] 설정 및 가이드 문서 작성
  - CLAUDE_GITHUB_ACTIONS.md
  - 상세한 트러블슈팅

- [x] CodeRabbit 통합 분석
  - 역할 분담 정의
  - 충돌 회피 전략

### Phase 2: GitHub Secrets 설정 (수동)

- [ ] GitHub Repository Settings 접근
- [ ] ANTHROPIC_API_KEY Secret 추가
- [ ] 워크플로우 권한 확인

### Phase 3: Claude API 통합 (향후)

- [ ] Issue 분석 자동화
- [ ] 코드 생성 로직
- [ ] Branch 생성 및 커밋
- [ ] PR 자동 생성

### Phase 4: 자동 병합 (향후)

- [ ] 자동 병합 조건 정의
- [ ] CodeRabbit 승인 확인
- [ ] 안전 메커니즘 구현
- [ ] 롤백 절차 수립

---

## 💻 기술 스택

| 항목 | 선택 | 이유 |
|------|------|------|
| **CI/CD 플랫폼** | GitHub Actions | GitHub 네이티브, 무료 |
| **AI 모델** | Claude (Anthropic) | 한국어 지원, CLAUDE.md 준수 |
| **코드 리뷰** | CodeRabbit | 자동 리뷰 + 자동 승인 (Pro) |
| **버전 관리** | Git (GitFlow) | MoAI-ADK 표준 |
| **SPEC 관리** | YAML + Markdown | MoAI-ADK 표준 |

---

## 🔐 보안 고려사항

### API Key 관리

```
위험도: 높음
위치: ANTHROPIC_API_KEY
보호:
  ✅ GitHub Secrets 암호화
  ✅ 워크플로우 로그 자동 마스킹
  ✅ 액세스 제한 (Org members only)
  ✅ 정기 로테이션 (90일 권장)

감시:
  • 키 사용 패턴 모니터링
  • 비정상 API 호출 감지
  • 월간 보안 감사
```

### 권한 관리

```
GitHub Actions 권한:
  ✅ contents: write (코드 작성 권한)
  ✅ issues: write (Issue 코멘트)
  ✅ pull-requests: write (PR 생성/수정)
  ❌ admin: N/A (필요 없음)

제한:
  • 특정 브랜치에만 적용 (develop, feature/*)
  • PR merge는 수동 또는 조건부
```

### 코드 검증

```
자동 검증:
  ✅ TRUST 5 원칙 확인
  ✅ 테스트 커버리지 검증
  ✅ CodeRabbit 자동 리뷰
```

---

## 📊 성공 기준

### 기본 기준

- [x] GitHub Actions 워크플로우 생성됨
- [x] CodeRabbit과 충돌 없음
- [x] 설정 및 가이드 문서 완성됨
- [x] SPEC 문서 작성됨
- [ ] GitHub Secrets 설정됨 (수동 작업)
- [ ] 테스트 PR 생성됨 (수동 작업)

### 고급 기준

- [ ] @claude mention 자동 감지 ✓
- [ ] Issue → PR 자동 생성 (향후)
- [ ] 자동 코드 생성 (향후)
- [ ] 자동 병합 (향후)

---

## 📚 참고 자료

| 자료 | 위치 |
|------|------|
| **설정 가이드** | .github/CLAUDE_GITHUB_ACTIONS.md |
| **워크플로우** | .github/workflows/claude-github-actions.yml |
| **CodeRabbit 가이드** | .github/CODERABBIT_SETUP.md |
| **MoAI-ADK CLAUDE.md** | ./CLAUDE.md |
| **GitHub Actions 문서** | https://docs.github.com/en/actions |
| **Claude Code 문서** | https://code.claude.com/docs/ko/github-actions |

---

## 🔄 HISTORY

### v1.0.0 (2025-11-07) - INITIAL

**작업 내용:**
- GitHub Actions 기본 인프라 구축
- CodeRabbit 통합 분석
- 4개 주요 워크플로우 정의
  1. Claude Issue Handler
  2. Claude PR Validator
  3. Claude Auto-Sync Trigger
  4. Claude Merge Readiness Check
- 상세 설정 가이드 작성

**상태:** 완료 ✅

---

## 🎯 다음 단계 (우선순위)

### 즉시 (1주일 내)

1. **GitHub Secrets 설정** (5분)
   ```bash
   Settings → Secrets → New secret
   Name: ANTHROPIC_API_KEY
   Value: sk-ant-api03-...
   ```

2. **테스트 PR 생성**
   ```bash
   git checkout -b test/claude-github-actions
   echo "# Test" > test.md
   git commit -m "test: claude github actions"
   gh pr create --base develop
   ```

3. **워크플로우 실행 확인**
   - Actions 탭에서 실행 로그 확인
   - 워크플로우 성공 여부 검증

### 단기 (1-2주)

4. **Claude API 통합** (향후 Phase 2)
   - Issue 분석 자동화
   - 코드 생성 로직 추가

5. **자동 커밋** 기능 (향후 Phase 3)
   - 자동 Branch 생성
   - 자동 코드 작성

### 장기 (3-4주)

6. **자동 병합** 기능 (향후 Phase 4)
   - 조건부 자동 병합
   - 롤백 메커니즘

---

📌 **SPEC 완료**: Claude Code GitHub Actions 통합 기본 인프라가 완성되었습니다.

🚀 **다음**: GitHub Secrets 설정 후 테스트 PR을 생성하세요.

---

🤖 Generated with Claude Code

Co-Authored-By: 🎩 Alfred@[MoAI](https://adk.mo.ai.kr)
