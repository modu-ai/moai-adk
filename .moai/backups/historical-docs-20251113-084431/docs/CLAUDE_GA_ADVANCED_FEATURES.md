# 🚀 Claude GitHub Actions 고급 기능 분석 & 구현 가이드

> 이슈 라벨 기반 자동 댓글, @claude 트리거 권한 제어, Claude GA 활용 방안

**작성일**: 2025-11-07
**상태**: 분석 완료
**대상**: MoAI-ADK 개발팀

---

## 📋 개요

사용자의 3가지 요청을 분석합니다:

1. **자동 댓글** - Issue 라벨에 따라 자동 댓글 달기
2. **@claude 권한 제어** - 특정 사용자만 @claude 트리거 사용 가능하게
3. **Claude GA 활용 방안** - 최적의 사용 사례 및 전략

---

## 1️⃣ 이슈 라벨 기반 자동 댓글

### 📊 분석

**가능한가?** ✅ **예, 완전히 가능합니다.**

```yaml
# GitHub Actions에서 Issue 라벨 감지 및 자동 댓글
on:
  issues:
    types: [opened, labeled]

jobs:
  auto-comment:
    if: github.event.issue.labels[*].name contains 'bug'
    runs-on: ubuntu-latest
    steps:
      - name: Post auto-comment
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: 'Bug 리포트 감지됨. 다음을 확인하세요:\n- 환경\n- 재현 단계\n- 예상 vs 실제'
            })
```

### 🎯 활용 사례

#### **사례 1: Bug 라벨 → 버그 리포트 템플릿 댓글**

```yaml
name: Auto-Comment on Bug Label

on:
  issues:
    types: [labeled]

jobs:
  auto-comment-bug:
    if: github.event.label.name == 'bug'
    runs-on: ubuntu-latest

    steps:
      - name: 🐛 Post bug checklist
        uses: actions/github-script@v7
        with:
          script: |
            const bugChecklist = `
🐛 **Bug Report Checklist**

필수 정보를 제공해주세요:

- [ ] MoAI-ADK 버전 (예: 0.20.1)
- [ ] 운영체제 및 버전
- [ ] Python 버전
- [ ] 재현 단계 (step-by-step)
- [ ] 예상 결과
- [ ] 실제 결과
- [ ] 에러 메시지 (있으면)

**팁**: \`\`\`
명령어 출력
\`\`\` 으로 감싸면 보기 좋습니다.

---
*🤖 Automated by Claude GitHub Actions*
            `;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: bugChecklist
            })
```

#### **사례 2: Feature Request 라벨 → SPEC 템플릿 댓글**

```yaml
name: Auto-Comment on Feature Request

on:
  issues:
    types: [labeled]

jobs:
  auto-comment-feature:
    if: github.event.label.name == 'feature-request'
    runs-on: ubuntu-latest

    steps:
      - name: 📝 Post SPEC template
        uses: actions/github-script@v7
        with:
          script: |
            const specTemplate = `
📝 **Feature Request Template**

이 요청을 처리하기 위해 다음을 작성해주세요:

## 문제점
사용자가 해결하려는 문제는 무엇인가요?

## 제안 해결책
원하는 기능은 무엇인가요?

## 대체 방안
다른 방법이 있나요?

## SPEC 정보
- [ ] SPEC ID (예: SPEC-AUTH-001)
- [ ] 우선순위 (HIGH/MEDIUM/LOW)
- [ ] 예상 기간

---
*참고: 이 Issue는 SPEC-XXX로 변환되어 개발될 예정입니다.*
            `;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: specTemplate
            })
```

#### **사례 3: 특정 라벨 조합 → Claude GA 자동 활성화**

```yaml
name: Auto-Trigger Claude on Labels

on:
  issues:
    types: [labeled]

jobs:
  trigger-claude:
    # bug + urgent 라벨이 모두 있으면 Claude 자동 활성화
    if: |
      contains(github.event.issue.labels.*.name, 'bug') &&
      contains(github.event.issue.labels.*.name, 'urgent')
    runs-on: ubuntu-latest

    steps:
      - name: 🤖 Trigger Claude Auto-Fix
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '@claude emergency hotfix: analyze and create PR for this critical bug'
            })
```

### 💡 MoAI-ADK 맞춤 구현

**MoAI-ADK용 자동 댓글 워크플로우:**

```yaml
name: MoAI Auto-Comment on Issues

on:
  issues:
    types: [opened, labeled]

jobs:
  auto-comment:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      # BUG 라벨: 버그 리포트 가이드
      - name: Comment on bug
        if: contains(github.event.issue.labels.*.name, 'bug')
        uses: actions/github-script@v7
        with:
          script: |
            const bugGuide = `🐛 **Bug Report**\n\n
필수 정보:\n
- [ ] moai-adk 버전\n
- [ ] 재현 단계\n
            github.rest.issues.createComment({...})

      # FEATURE-REQUEST: SPEC 템플릿
      - name: Comment on feature request
        if: contains(github.event.issue.labels.*.name, 'feature-request')
        uses: actions/github-script@v7
        with:
          script: |
            const specTemplate = `📝 **Feature Request**\n\n
/alfred:1-plan "feature name" 로 SPEC을 생성하세요`;
            github.rest.issues.createComment({...})

      # DOCUMENTATION: 문서 기여 가이드
      - name: Comment on documentation
        if: contains(github.event.issue.labels.*.name, 'documentation')
        uses: actions/github-script@v7
        with:
          script: |
            const docGuide = `📚 **Documentation**\n\n
문서 위치: .moai/docs/\n
SPEC: .moai/specs/SPEC-*/`;
            github.rest.issues.createComment({...})
```

---

## 2️⃣ @claude 트리거 권한 제어

### 📊 분석

**@claude 트리거를 특정 사용자만 사용하게 할 수 있는가?** ✅ **예, 권한 제어 가능합니다.**

### 🔐 권한 제어 방식

#### **방법 1: GitHub Organization Role 기반 (권장)**

```yaml
name: Claude GitHub Actions with Permission Check

on:
  issue_comment:
    types: [created]

jobs:
  claude-handler:
    if: contains(github.event.comment.body, '@claude')
    runs-on: ubuntu-latest

    steps:
      - name: 🔐 Check Permission (Admin/Maintainer)
        id: check-permission
        uses: actions/github-script@v7
        with:
          script: |
            const { data: permission } = await github.rest.repos.getCollaboratorPermissionLevel({
              owner: context.repo.owner,
              repo: context.repo.repo,
              username: context.actor
            });

            // admin, maintain, write = 허용
            // triage, pull, none = 거부
            const allowed = ['admin', 'maintain', 'write'].includes(permission.permission);

            if (!allowed) {
              console.log(`❌ Permission denied for ${context.actor}`);
              core.setFailed('Not authorized to use @claude');
              return;
            }

            console.log(`✅ Permission granted for ${context.actor}`);

      - name: 🤖 Process @claude command
        if: steps.check-permission.outcome == 'success'
        run: |
          echo "Processing @claude command..."
          # Claude GA 로직 실행
```

#### **방법 2: 특정 사용자 화이트리스트 (엄격함)**

```yaml
name: Claude GitHub Actions - Whitelist Mode

on:
  issue_comment:
    types: [created]

env:
  AUTHORIZED_USERS: |
    goos
    alfred-bot

jobs:
  claude-handler:
    if: contains(github.event.comment.body, '@claude')
    runs-on: ubuntu-latest

    steps:
      - name: 🔐 Check Whitelist
        id: check-auth
        run: |
          AUTHORIZED="${{ env.AUTHORIZED_USERS }}"
          ACTOR="${{ github.actor }}"

          if echo "$AUTHORIZED" | grep -q "^$ACTOR$"; then
            echo "authorized=true" >> $GITHUB_OUTPUT
            echo "✅ User authorized: $ACTOR"
          else
            echo "authorized=false" >> $GITHUB_OUTPUT
            echo "❌ User NOT authorized: $ACTOR"
          fi

      - name: 📝 Post denial comment (if unauthorized)
        if: steps.check-auth.outputs.authorized == 'false'
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `❌ @claude command는 권한이 있는 사용자만 사용할 수 있습니다.\n\n현재 사용자: @${{ github.actor }}\n\n자세한 내용은 MAINTAINERS에게 문의하세요.`
            })

      - name: 🤖 Process @claude command (if authorized)
        if: steps.check-auth.outputs.authorized == 'true'
        run: |
          echo "✅ Processing @claude command for authorized user..."
```

#### **방법 3: GitHub Teams 기반 (조직 규모)**

```yaml
name: Claude GitHub Actions - Teams Mode

on:
  issue_comment:
    types: [created]

jobs:
  claude-handler:
    if: contains(github.event.comment.body, '@claude')
    runs-on: ubuntu-latest

    steps:
      - name: 🔐 Check Team Membership
        id: check-team
        uses: actions/github-script@v7
        with:
          script: |
            try {
              const team = await github.rest.teams.getMembershipForUserInOrg({
                org: 'modu-ai',
                team_slug: 'developers',  // Team 이름
                username: context.actor
              });

              // active 멤버만 허용
              if (team.data.state === 'active') {
                console.log(`✅ Team member: ${context.actor}`);
                core.setOutput('authorized', 'true');
              } else {
                console.log(`❌ Not active team member: ${context.actor}`);
                core.setOutput('authorized', 'false');
              }
            } catch (error) {
              console.log(`❌ Not in team: ${context.actor}`);
              core.setOutput('authorized', 'false');
            }

      - name: 🤖 Process @claude
        if: steps.check-team.outputs.authorized == 'true'
        run: echo "✅ Processing @claude for team member..."
```

### 🎯 권장 구성 (MoAI-ADK)

```yaml
# .github/workflows/claude-github-actions-with-auth.yml

name: Claude GitHub Actions with Auth

on:
  issue_comment:
    types: [created, edited]

env:
  # Owner만 사용 가능
  AUTHORIZED_USERS: goos

jobs:
  claude-handler:
    if: contains(github.event.comment.body, '@claude')
    runs-on: ubuntu-latest

    permissions:
      contents: write
      issues: write
      pull-requests: write

    steps:
      - name: 🔐 Verify Authorization
        id: auth
        run: |
          AUTHORIZED_USERS="${{ env.AUTHORIZED_USERS }}"
          CURRENT_USER="${{ github.actor }}"

          echo "User: $CURRENT_USER"
          echo "Authorized: $AUTHORIZED_USERS"

          if [[ "$AUTHORIZED_USERS" == *"$CURRENT_USER"* ]]; then
            echo "authorized=true" >> $GITHUB_OUTPUT
            echo "✅ Authorization PASSED"
          else
            echo "authorized=false" >> $GITHUB_OUTPUT
            echo "❌ Authorization FAILED"
          fi

      - name: 📛 Reject Unauthorized
        if: steps.auth.outputs.authorized == 'false'
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `🚫 **Permission Denied**\n\n@claude는 권한이 있는 사용자만 사용할 수 있습니다.\n\nUser: @${{ github.actor }}\n\nMaintainer에게 문의하세요.`
            })
            core.setFailed('Unauthorized @claude usage')

      - name: 🤖 Process @claude
        if: steps.auth.outputs.authorized == 'true'
        uses: actions/github-script@v7
        with:
          script: |
            const comment = '${{ github.event.comment.body }}';
            const command = comment.replace('@claude ', '').trim();

            console.log('🎯 Processing command:', command);
            console.log('👤 Authorized user:', '${{ github.actor }}');

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `✅ **Claude Processing**\n\nCommand: \`${command}\`\nUser: @${{ github.actor }}\n\nProcessing...`
            })
```

### 📊 권한 제어 비교표

| 방법 | 난이도 | 보안성 | 관리성 | 권장도 |
|------|--------|--------|--------|---------|
| **Organization Role** | 낮음 | 높음 | 높음 | ⭐⭐⭐⭐⭐ (최고) |
| **화이트리스트** | 낮음 | 높음 | 중간 | ⭐⭐⭐⭐ |
| **Teams 기반** | 중간 | 매우높음 | 높음 | ⭐⭐⭐⭐ |
| **하이브리드** | 높음 | 매우높음 | 높음 | ⭐⭐⭐⭐⭐ |

**MoAI-ADK 권장**: Organization Role 기반 (간단하면서 안전)

---

## 3️⃣ Claude GA 활용 방안 - 최적 전략

### 📊 Claude GA의 강점

```
Claude Code GitHub Actions의 특화 영역:

✅ Issue 분석 및 이해
   • 자연어 처리로 복잡한 요구사항 파악
   • CLAUDE.md, TRUST 5 규칙 자동 적용

✅ 코드 생성
   • MoAI-ADK 패턴 학습 및 적용
   • SPEC → Code → Test → Doc 연쇄
   • 테스트 코드도 함께 생성

✅ 문서 작성
   • README, API 문서 자동 생성
   • CHANGELOG 자동 작성
   • SPEC 문서 완성도 개선

✅ 코드 분석
   • 성능 병목 지점 감지
   • 보안 취약점 분석
   • 코드 스멜 감지

✅ 자동화
   • 반복 작업 자동화
   • 보일러플레이트 코드 생성
   • CI/CD 파이프라인 최적화
```

### 🎯 실제 활용 사례

#### **사례 1: Issue → SPEC 자동 변환**

```
Flow:
Issue 생성 (자연어)
    ↓
@claude convert to SPEC
    ↓
Claude GA:
  1. Issue 분석
  2. SPEC-XXXX YAML 생성
  3. EARS 요구사항 작성
  4. feature/SPEC-XXXX 브랜치 생성
  5. .moai/specs/SPEC-XXXX.md 커밋
  6. Draft PR 생성
    ↓
개발자: SPEC 검토 → SPEC 승인 → /alfred:2-run 실행
```

#### **사례 2: Bug Fix 자동화**

```
Flow:
Bug Report Issue 생성
    ↓
@claude hotfix: [description]
    ↓
Claude GA:
  1. 버그 분석
  2. 테스트 케이스 작성 (RED)
  3. 최소 수정 코드 작성 (GREEN)
  4. 코드 정리 (REFACTOR)
  5. feature/BUG-XXXX 브랜치 생성
  6. Draft PR 생성 (모든 테스트 통과)
    ↓
CodeRabbit: 자동 리뷰 → 자동 승인
    ↓
자동 병합 (향후)
```

#### **사례 3: Documentation 생성**

```
Flow:
새로운 기능 구현 완료
    ↓
@claude generate docs
    ↓
Claude GA:
  1. 코드 분석
  2. API 문서 생성
  3. 사용 예제 작성
  4. README 업데이트
  5. 다국어 문서 생성 (선택)
  6. docs/ 폴더에 커밋
  7. PR 생성
```

### 🚀 고급 활용 전략

#### **전략 1: SPEC-First Automation**

```yaml
# SPEC → Code → Test → Doc → Deploy 자동화

Trigger: @claude implement SPEC-001

Steps:
1. SPEC 파일 읽기
2. EARS 요구사항 파싱
3. 테스트 케이스 자동 생성
4. 구현 코드 자동 생성
5. 문서 자동 생성
6. CHANGELOG 자동 생성
7. Deploy 준비 (선택적)
```

#### **전략 2: Multi-Language Support**

```yaml
# 다국어 문서 자동 생성

Trigger: @claude translate docs to ja,es,zh

Steps:
1. 영문 문서 분석
2. Claude 다국어 생성
3. docs/ja/, docs/es/, docs/zh/ 생성
4. PR 자동 생성
```

#### **전략 3: Code Review Automation**

```yaml
# Claude가 코드 리뷰

Trigger: @claude review code

Steps:
1. 변경사항 분석
2. TRUST 5 검사
3. 성능 분석
4. 보안 분석
5. 개선 제안
6. 자동 수정 코드 생성 (선택)
```

### 💡 MoAI-ADK 최적 활용 방안

```
전체 Flow:

Issue 생성 (사용자)
    ↓
@claude [command] (사용자 또는 Automation)
    ↓
Claude GA (자동)
├─ 요구사항 분석
├─ SPEC 생성 또는 업데이트
├─ Feature Branch 생성
└─ Draft PR 생성
    ↓
MoAI GitFlow (자동)
├─ TRUST 5 검증
└─ 테스트 실행
    ↓
CodeRabbit (자동)
├─ 코드 리뷰
├─ 자동 수정 제안
└─ 자동 승인 (Pro)
    ↓
개발자 (선택)
├─ PR 검토 (선택)
├─ Ready for Review 전환
└─ 병합 (또는 자동 병합)
    ↓
Deploy (자동 또는 수동)
```

---

## 🔧 실제 구현 추천사항

### 즉시 구현 (1주)

1. **자동 댓글** (위의 사례 코드 활용)
   - Bug 라벨 → 버그 리포트 가이드
   - Feature Request → SPEC 템플릿
   - Documentation → 문서 기여 가이드

2. **권한 제어** (Organization Role 기반)
   - Admin/Maintainer만 @claude 사용 가능
   - 로그 기록 및 감시

### 단기 구현 (2-4주)

3. **Issue → SPEC 자동 변환**
   - Claude가 SPEC 문서 자동 생성
   - EARS 요구사항 작성
   - Feature branch 자동 생성

4. **Code Generation**
   - 테스트 코드 자동 생성
   - 구현 코드 자동 생성
   - 문서 자동 생성

### 장기 구현 (1-3개월)

5. **완전 자동화 파이프라인**
   - Issue → SPEC → Code → Test → Doc → Deploy
   - 자동 병합 (신뢰도 기준)
   - 자동 배포

---

## 📋 체크리스트

### 즉시 (오늘)

- [ ] GitHub Secrets 설정 완료 (ANTHROPIC_API_KEY)
- [ ] 워크플로우 권한 확인

### 단기 (이번 주)

- [ ] 자동 댓글 워크플로우 추가
- [ ] @claude 권한 제어 구현
- [ ] 테스트 PR 생성

### 중기 (1-2주)

- [ ] Claude API 통합
- [ ] 자동 PR 생성 구현

### 장기 (3-4주)

- [ ] 완전 자동화 파이프라인
- [ ] 자동 병합 구현

---

## 🎓 참고 자료

| 자료 | 위치 |
|------|------|
| **GitHub Actions 문서** | https://docs.github.com/en/actions |
| **GitHub Script 예제** | https://github.com/actions/github-script |
| **기존 워크플로우** | .github/workflows/claude-github-actions.yml |
| **설정 가이드** | .github/CLAUDE_GITHUB_ACTIONS.md |
| **SPEC 문서** | .moai/specs/SPEC-GITHUB-ACTIONS-001.md |

---

🤖 Generated with Claude Code

Co-Authored-By: 🎩 Alfred@[MoAI](https://adk.mo.ai.kr)
