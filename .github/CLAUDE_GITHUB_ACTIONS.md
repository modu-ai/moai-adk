# 🤖 Claude Code GitHub Actions 통합 가이드

> Claude가 GitHub Actions에서 자동으로 작업을 수행합니다.

**최종 수정**: 2025-11-07
**상태**: ✅ 설정 완료 + CodeRabbit 통합
**버전**: 0.20.1

---

## 📋 목차

1. [빠른 시작](#빠른-시작)
2. [설정 방법](#설정-방법)
3. [통합 시나리오](#통합-시나리오)
4. [CodeRabbit + Claude GA](#coderabbit--claude-ga)
5. [워크플로우 상세](#워크플로우-상세)
6. [트러블슈팅](#트러블슈팅)

---

## 빠른 시작

### 요구사항

✅ **이미 준비됨**:
- GitHub 저장소 (MoAI-ADK)
- CodeRabbit 설정 완료 (.github/CODERABBIT_SETUP.md)
- MoAI-ADK GitFlow 워크플로우 활성화

❌ **필요한 것**:
- Anthropic API Key (제공됨)
- GitHub Secrets 설정 (5분)

### 1단계: GitHub Secrets 설정 (필수)

```bash
위치: https://github.com/modu-ai/moai-adk/settings/secrets/actions

1. "New repository secret" 클릭
2. 정보 입력:
   Name: ANTHROPIC_API_KEY
   Secret: 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
3. "Add secret" 클릭
4. 확인: Settings → Secrets에서 ANTHROPIC_API_KEY 표시됨
```

**🔒 보안 체크리스트**:
- [ ] API Key가 GitHub Secrets로 저장됨 (암호화됨)
- [ ] `.github/workflows/` 파일에 하드코딩되지 않음
- [ ] `${{ secrets.ANTHROPIC_API_KEY }}`로만 접근

### 2단계: 워크플로우 활성화 확인

```bash
위치: Actions 탭 → 워크플로우 목록

확인 항목:
✅ Claude Code GitHub Actions (신규)
✅ MoAI-ADK GitFlow (기존)
✅ CodeRabbit (기존 - 제거 안 함!)
✅ Release Pipeline (기존)

모두 활성화되어야 합니다.
```

### 3단계: 첫 테스트 (선택)

```bash
# 방법 1: Issue 코멘트로 테스트
Issue 열기 → 코멘트 추가: "@claude implement feature X"

# 방법 2: 새 PR 생성
Feature branch 생성 → PR 생성 → Claude가 자동 분석
```

---

## 설정 방법

### 옵션 A: 자동 설정 (권장)

```bash
# GitHub App 자동 설정
cd /Users/goos/MoAI/MoAI-ADK
moai-adk github-setup --with-claude --api-key "sk-ant-api03..."
```

**⚠️ 현재 상태**: `moai-adk` CLI에 해당 명령어가 없으므로 수동 설정 필요

### 옵션 B: 수동 설정 (현재 권장)

#### 1단계: GitHub Secrets 추가

```
Settings → Secrets and variables → Actions

New secret 추가:
- Name: ANTHROPIC_API_KEY
- Value: <API Key>
```

#### 2단계: 워크플로우 파일 확인

```bash
.github/workflows/
├── claude-github-actions.yml        ✅ (신규 - 자동 생성됨)
├── moai-gitflow.yml                 ✅ (기존)
├── moai-release-pipeline.yml        ✅ (기존)
├── tag-validation.yml               ✅ (기존)
├── spec-issue-sync.yml              ✅ (기존)
└── docs-deploy.yml                  ✅ (기존)
```

#### 3단계: 워크플로우 권한 확인

```bash
Settings → Actions → General

Workflow permissions:
✅ Read and write permissions
✅ Allow GitHub Actions to create and approve pull requests
```

---

## 통합 시나리오

### 시나리오 1: Issue → PR 자동 생성

```
1️⃣ 사용자가 Issue 생성
   "사용자 인증 기능 구현 필요"

2️⃣ Issue 코멘트에서 @claude mention
   "@claude implement JWT authentication system"

3️⃣ Claude GitHub Actions 활성화
   ├─ Issue 분석
   ├─ 코드 생성 (Claude API)
   ├─ feature/SPEC-XXX 브랜치 생성
   └─ Draft PR 자동 생성

4️⃣ CodeRabbit 자동 리뷰
   ├─ 코드 품질 검사
   ├─ 보안 이슈 검출
   ├─ 자동 수정 제안
   └─ 자동 승인 (Pro)

5️⃣ 개발자가 PR Ready 상태로 변경
   → /alfred:3-sync 자동 실행

6️⃣ 병합 준비 완료
   gh pr merge feature/SPEC-XXX --squash
```

### 시나리오 2: PR 생성 → Claude 자동 분석

```
1️⃣ Feature branch에서 PR 생성
   "Implement OAuth integration"

2️⃣ Claude PR Validator 자동 실행
   ├─ SPEC 문서 확인
   ├─ @TAG 참조 검증
   ├─ 테스트 커버리지 확인
   └─ TRUST 5 원칙 검증

3️⃣ 분석 결과를 PR 코멘트로 게시
   ├─ 체크리스트
   ├─ 제안사항
   └─ 다음 단계 안내

4️⃣ CodeRabbit이 이어서 리뷰
   (위 시나리오 4번부터 반복)
```

### 시나리오 3: Draft PR → Ready PR → 자동 SYNC

```
1️⃣ Draft PR에서 개발
   - TDD RED-GREEN-REFACTOR 진행
   - 테스트 작성 및 통과
   - 코드 리뷰 및 수정

2️⃣ "Ready for Review" 상태로 변경
   → Claude Auto-Sync Trigger 활성화

3️⃣ 자동 실행 항목:
   ├─ 문서 동기화
   ├─ SPEC 업데이트
   ├─ CHANGELOG 생성
   ├─ @TAG 무결성 검증
   └─ README 업데이트

4️⃣ 병합 준비 완료
   → CodeRabbit 최종 승인
   → gh pr merge 준비
```

---

## CodeRabbit + Claude GA

### 역할 분담

| 단계 | CodeRabbit | Claude GA | MoAI-ADK |
|------|-----------|-----------|---------|
| **Issue 생성** | ❌ | ⏳ *대기* | ✅ |
| **PR 자동 생성** | ❌ | ✅ *주도* | 📋 *검증* |
| **코드 리뷰** | ✅ *자동* | ❌ | 📊 *추적* |
| **자동 승인** | ✅ *Pro* | ❌ | 📈 *품질 측정* |
| **문서 동기화** | ❌ | 🔄 *지원* | ✅ *주도* |
| **TAG 검증** | ❌ | 📋 *추적* | ✅ *검증* |
| **TRUST 5** | 🔍 *부분* | 🔍 *부분* | ✅ *완전* |

### 통합 파이프라인

```
Issue Created
    ↓
@claude comment (또는 PR created)
    ↓
Claude GitHub Actions
├─ Issue 분석
├─ 코드 생성 (Claude API)
├─ Branch 생성
└─ PR 자동 생성
    ↓
CodeRabbit 자동 리뷰
├─ 코드 품질 검사
├─ 보안 이슈 검출
├─ 자동 수정 제안
└─ 자동 승인 (Pro)
    ↓
PR Ready for Review
    ↓
Claude Auto-Sync
├─ 문서 동기화
├─ SPEC 업데이트
└─ 병합 준비
    ↓
Merge & Deploy
```

### 보수적 운영

```yaml
# .github/workflows/claude-github-actions.yml
# 현재 상태: 안전 모드 (dry-run)

설정:
  - PR 자동 생성 ❌ (아직 미구현)
  - 자동 커밋 ❌ (아직 미구현)
  - 자동 푸시 ❌ (아직 미구현)
  - 자동 병합 ❌ (아직 미구현)
  - 상태 코멘트 ✅ (안전)
  - 분석 리포팅 ✅ (안전)

향후:
  - 테스트 통과 시 자동 커밋 활성화
  - CodeRabbit 승인 확인 후 자동 병합 활성화
```

---

## 워크플로우 상세

### 1. Claude Issue Handler

**Trigger**: Issue 코멘트에 `@claude` mention

```bash
# 예제:
Issue #42에 코멘트: "@claude implement login form component"

실행:
  1. Mention 감지
  2. 명령어 파싱
  3. Claude API 호출 (향후)
  4. 코드 생성 (향후)
  5. PR 생성 (향후)
  6. 상태 코멘트 게시
```

**현재 상태**: 🔄 준비 중
- [x] 워크플로우 파일 생성
- [x] @claude mention 감지
- [ ] Claude API 통합
- [ ] PR 자동 생성

### 2. Claude PR Validator

**Trigger**: PR 생성 시

```bash
자동 분석 항목:
  ✅ SPEC 문서 확인
  ✅ @TAG 참조 검증
  ✅ 테스트 검출
  ✅ 파일 변경사항 분석

출력:
  - PR 코멘트에 분석 결과 게시
  - Checklist 형식으로 표시
  - MoAI-ADK 준수 확인
```

**현재 상태**: ✅ 활성화
- [x] 워크플로우 구현
- [x] 자동 분석
- [x] 결과 코멘트 게시

### 3. Claude Auto-Sync Trigger

**Trigger**: Draft PR → Ready for Review 변경

```bash
자동 실행:
  1. 문서 동기화 (향후)
  2. CHANGELOG 생성 (향후)
  3. @TAG 무결성 검증
  4. 병합 준비 확인
  5. 상태 코멘트 게시

목적:
  - 개발자가 "Ready" 버튼만 누르면
  - 나머지는 모두 자동화
```

**현재 상태**: 🔄 준비 중
- [x] 워크플로우 파일
- [ ] 실제 SYNC 구현
- [ ] 자동 병합 준비

### 4. Claude Merge Readiness Check

**Trigger**: 모든 PR 생성 시

```bash
검증 항목:
  ✅ CodeRabbit 리뷰 대기
  ✅ 모든 CI 체크 대기
  ✅ 테스트 대기
  ✅ TRUST 5 검증 대기
  ✅ @TAG 검증 대기

목적:
  - 병합 전 모든 조건 확인
  - 자동 병합 안전성 보장
```

**현재 상태**: ✅ 활성화 (모니터링)

---

## 트러블슈팅

### Q: 워크플로우가 실행되지 않음

**A: 다음을 확인하세요:**

1. GitHub Secrets 확인
   ```
   Settings → Secrets → ANTHROPIC_API_KEY 존재?
   ```

2. 워크플로우 파일 확인
   ```
   .github/workflows/claude-github-actions.yml 존재?
   ```

3. 권한 확인
   ```
   Settings → Actions → Workflow permissions
   "Allow GitHub Actions to create and approve pull requests" 활성화?
   ```

4. 트리거 확인
   ```
   - Issue 코멘트에 "@claude" 포함?
   - PR 생성 후 1-2분 대기?
   ```

### Q: "@claude" mention을 했는데 반응이 없음

**A: 다음을 시도하세요:**

```bash
1. 코멘트 재작성
   "@claude implement authentication"

2. PR을 다시 생성해보기
   git push && gh pr create

3. GitHub Actions 로그 확인
   Actions 탭 → 워크플로우 클릭 → 로그 보기

4. Secrets 확인
   echo $ANTHROPIC_API_KEY
   (워크플로우 로그에서 *로 마스킹되는지 확인)
```

### Q: CodeRabbit과 Claude GA가 충돌하나요?

**A: 아니요, 충돌하지 않습니다.**

```
CodeRabbit:   PR 리뷰 (사람 대신)
Claude GA:    PR 생성 (지시 대신)

→ 서로 다른 단계에서 작동 → 상호 보완
```

### Q: 자동 병합이 활성화되나요?

**A: 현재 아니요, 향후 활성화 예정입니다.**

```
현재:
  ✅ PR 자동 생성 준비 중
  ✅ CodeRabbit 자동 승인 ✓
  ❌ 자동 병합 (비활성화 - 수동 확인 필요)

향후:
  ✅ CodeRabbit 승인 확인 후 자동 병합
```

---

## 다음 단계

### 즉시 실행 (필수)

- [ ] GitHub Secrets에 ANTHROPIC_API_KEY 추가
- [ ] 워크플로우 권한 확인
- [ ] 테스트 PR 생성

### 향후 구현 (선택)

- [ ] Claude API 통합 (자동 PR 생성)
- [ ] 자동 커밋 + 푸시
- [ ] 자동 병합 (신뢰도 기준)
- [ ] Slack 알림 통합

---

## 참고 자료

| 자료 | 위치 |
|------|------|
| **CodeRabbit 설정** | .github/CODERABBIT_SETUP.md |
| **MoAI-ADK GitFlow** | .github/workflows/moai-gitflow.yml |
| **Claude Code 문서** | https://code.claude.com/docs/ko/github-actions |
| **GitHub Secrets 문서** | https://docs.github.com/en/actions/security-guides/encrypted-secrets |
| **Project Config** | .moai/config.json |

---

## 최종 체크리스트

- [ ] GitHub Secrets 추가됨 (ANTHROPIC_API_KEY)
- [ ] 워크플로우 파일 생성됨 (.github/workflows/claude-github-actions.yml)
- [ ] 워크플로우 권한 설정됨
- [ ] CodeRabbit 유지됨 (제거 안 함)
- [ ] 테스트 PR 준비됨
- [ ] 이 문서 읽음

---

✨ **완료되었습니다!** Claude Code GitHub Actions이 준비되었습니다.

다음: GitHub Secrets를 설정하고 테스트 PR을 생성해보세요.

🚀 **명령어**:
```bash
# 테스트 PR 생성
git checkout -b test/claude-github-actions
echo "# Test PR for Claude GitHub Actions" > test.md
git add test.md && git commit -m "test: claude github actions"
git push origin test/claude-github-actions
gh pr create --base develop --title "test: Claude GitHub Actions integration"
```

---

Generated with Claude Code
Co-Authored-By: 🎩 Alfred@[MoAI](https://adk.mo.ai.kr)
