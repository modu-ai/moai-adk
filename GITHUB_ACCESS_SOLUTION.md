# 🔧 GitHub 저장소 접근 문제 해결 가이드

**문제**: `modu-ai/moai-adk` 저장소 접근 실패
**상태**: 원인 분석 완료, 해결 방법 제시

---

## 🔍 **문제 원인 분석 결과**

### **1. 확인된 사실들**
- ✅ **modu-ai**: GitHub 조직(Organization)으로 존재함
- ✅ **현재 계정**: GoosLab으로 인증됨
- ❌ **저장소 접근**: `moai-adk` 저장소에 접근 불가 (404 에러)
- ❌ **토큰 권한**: "Bad credentials" 에러 발생

### **2. 가능한 원인들**

#### **A. 저장소가 실제로 존재하지 않음**
- modu-ai 조직의 공개 저장소 목록에 `moai-adk`가 없음
- 다른 이름으로 저장소가 존재할 가능성

#### **B. 프라이빗 저장소 + 권한 없음** (가장 가능성 높음)
- 저장소는 존재하지만 프라이빗 설정
- GoosLab 계정이 modu-ai 조직 멤버가 아님
- 또는 저장소별 접근 권한이 부여되지 않음

#### **C. GitHub 토큰 문제**
- 현재 토큰이 만료되었거나 잘못됨
- 필요한 스코프(`repo`, `workflow`)가 부족함

---

## 🛠️ **해결 방법 (우선순위별)**

### **방법 1: 조직 초대 요청 (가장 확실한 방법)**

#### **modu-ai 조직 관리자가 해야 할 작업:**
```
1. https://github.com/orgs/modu-ai/people 접속
2. "Invite member" 버튼 클릭
3. 사용자명: GoosLab 입력
4. 권한: Member 또는 Owner 선택
5. 초대 전송

또는 저장소별 직접 초대:
1. https://github.com/modu-ai/moai-adk/settings/access 접속
2. "Invite a collaborator" 클릭
3. 사용자명: GoosLab 입력
4. 권한: Write 또는 Admin 선택
```

#### **GoosLab이 해야 할 작업:**
```
1. 이메일에서 초대 수락
2. 또는 https://github.com/notifications에서 수락
3. 권한 확인: gh repo view modu-ai/moai-adk
```

### **방법 2: GitHub 토큰 재생성**

#### **단계별 실행:**
```bash
# 1. 기존 인증 제거
gh auth logout

# 2. 새로운 토큰으로 로그인
gh auth login
# → GitHub.com 선택
# → HTTPS 선택
# → 브라우저에서 로그인
# → repo, workflow 스코프 선택

# 3. 확인
gh auth status
gh repo view modu-ai/moai-adk
```

### **방법 3: 저장소 이름 확인**

실제 저장소 이름이 다를 수 있습니다:

```bash
# modu-ai 조직의 모든 저장소 확인
curl -s "https://api.github.com/orgs/modu-ai/repos?per_page=100" | \
  jq '.[] | .name' | grep -i moai

# 가능한 이름들:
# - MoAI-ADK
# - moai_adk
# - moai-development-kit
# - moai-agentic-development-kit
```

### **방법 4: 수동 저장소 생성 (최후 수단)**

저장소가 실제로 존재하지 않는다면:

```bash
# 1. modu-ai 조직에서 새 저장소 생성
# 2. 이름: moai-adk
# 3. Private 설정
# 4. 현재 로컬 코드 푸시

git remote set-url origin https://github.com/modu-ai/moai-adk.git
git push -u origin main
```

---

## 🚀 **즉시 실행 가능한 검증 스크립트**

### **저장소 접근 테스트:**
```bash
#!/bin/bash
echo "🔍 GitHub 저장소 접근 테스트 시작..."

# 1. 현재 인증 상태 확인
echo "1. 인증 상태:"
gh auth status

# 2. 조직 멤버십 확인
echo "2. 조직 멤버십:"
gh api user/orgs | jq '.[] | select(.login == "modu-ai")'

# 3. 저장소 접근 시도
echo "3. 저장소 접근:"
gh repo view modu-ai/moai-adk 2>&1

# 4. 조직의 저장소 목록 (moai 관련)
echo "4. modu-ai 조직의 moai 관련 저장소:"
gh api orgs/modu-ai/repos | jq '.[] | select(.name | contains("moai")) | .name'

echo "✅ 테스트 완료"
```

### **문제 해결 후 PR 생성:**
```bash
#!/bin/bash
echo "🚀 PR 생성 시작..."

# 저장소 접근 확인
if gh repo view modu-ai/moai-adk &>/dev/null; then
    echo "✅ 저장소 접근 성공"

    # PR 생성
    gh pr create \
        --title "🚀 SPEC-003: Package Optimization System 구현 완료" \
        --body-file "PR_TEMPLATE_SPEC-003.md" \
        --label "enhancement,spec-003" \
        --assignee "@me"

    echo "✅ PR 생성 완료"
else
    echo "❌ 저장소 접근 실패 - 위의 해결 방법 실행 필요"
fi
```

---

## 📋 **권장 실행 순서**

### **즉시 실행 (GoosLab):**
1. GitHub 토큰 재생성 (방법 2)
2. 접근 테스트 실행
3. 결과에 따라 추가 조치

### **조직 관리자 요청:**
1. modu-ai 조직 초대 요청 (방법 1)
2. 또는 저장소 이름 확인 요청 (방법 3)

### **해결 완료 후:**
1. PR 생성 스크립트 실행
2. SPEC-003 80% 최적화 성과 공유
3. 팀 리뷰 및 병합 진행

---

## 🎯 **결론**

**가장 가능성 높은 원인**: modu-ai 조직의 프라이빗 저장소에 대한 접근 권한 부족

**권장 해결 순서**:
1. GitHub 토큰 재생성 시도
2. 조직 관리자에게 초대 요청
3. 해결 즉시 PR 생성하여 SPEC-003 성과 공유

**현재 준비 상태**: PR 생성을 위한 모든 자료 완비, 권한 해결 즉시 실행 가능

---

**📞 문의 사항**: modu-ai 조직 관리자에게 GoosLab 계정 초대 요청
**🚀 목표**: SPEC-003의 혁신적인 80% 패키지 최적화 성과를 팀과 공유