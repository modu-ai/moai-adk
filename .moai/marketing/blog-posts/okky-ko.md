---
title: "Claude Code 혼자는 부족하다. MoAI-ADK가 필요한 이유"
description: "SPEC-First 개발로 품질 보증하기"
---

# Claude Code 혼자는 부족하다. MoAI-ADK가 필요한 이유

지난달 Claude Code로 세 개 프로젝트를 완성했습니다. 초반 30%는 빨랐어요. 초기 구조 잡고, 핵심 기능 구현되는 시점까지. 그런데 50%를 넘어가면서 이상함을 느꼈어요.

코드는 돌아가는데... 뭔가 이상했거든요.

테스트는 제대로 없고. 왜 이렇게 구현했는지 아무도 모르고. PR 리뷰할 때마다 "왜 이건 이렇게 했어?" 하는 질문이 반복됐어요. 한두 군데 버그 나면 전체 구조가 흔들렸어요.

그래서 찾게 된 게 **MoAI-ADK**입니다.

## Claude Code만으로는 모자란 이유

Claude Code는 정말 강력해요. 근데 문제가 있어요:

- **개발 워크플로우가 없어요**: 자유도는 높은데, 이게 오히려 더 위험해요. 테스트? 선택사항. 요구사항? 막연해요.
- **컨텍스트 관리 지옥**: 매번 파일 다시 로드하고. 토큰 낭비 심해요.
- **품질 보증이 없어요**: 커버리지 체크? Linting? 보안? 다 수동이에요.
- **언어 자동 감지 안 돼요**: Python으로 하려면 pytest, ruff 설정을 직접 해야 해요.
- **진행 상태 추적이 없어요**: 세션 끊기면 어디까지 했는지 몰라요.

MoAI-ADK는 이 모든 걸 해결해요.

## 핵심은 "하네스 엔지니어링"

MoAI-ADK의 아이디어가 단순하지만 강력해요:

**인간이 뭘 만들지 정하면, AI가 어떻게 만들지 정한다.**

24개 전문 AI 에이전트와 52개 스킬이 협력해요.

- **Plan → Run → Sync** 워크플로우는 필수, 선택이 아니에요.
- **TDD/DDD 자동 선택**: 새 프로젝트면 TDD(테스트 먼저), 기존 프로젝트면 DDD(분석 먼저).
- **16개 언어 지원**: Go, Python, TypeScript, Rust... 자동 감지돼요.
- **자동 품질 보증**: 테스트, 린팅, 커버리지 체크가 자동으로 실행돼요.

## 실전 예시: `/moai plan` → `/moai run`

Go API에 JWT 인증을 추가한다고 해봅시다.

### 1단계: SPEC 문서 작성

```bash
/moai plan "사용자 로그인에 JWT 인증 추가"
```

MoAI가 물어봅니다:
- 새 기능인가요? 기존 기능 수정인가요?
- 인증 방식은 뭔가요?
- 어느 엔드포인트가 보호되어야 하나요?

답하면 자동으로 SPEC 문서가 생성돼요.

### 2단계: 구현하기

```bash
/moai run SPEC-AUTH-001
```

백엔드 전문 에이전트가 자동으로:
- 기존 코드 로드
- handler.go, models.go 분석
- JWT 로직 구현
- 테스트 작성
- 린팅 실행
- 커버리지 체크

코드가 생성되고:

```go
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	json.NewDecoder(r.Body).Decode(&req)
	
	user, err := h.db.ValidateCredentials(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	
	tokenString, _ := token.SignedString([]byte(h.jwtSecret))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
```

테스트도 자동 생성돼요:

```go
func TestLogin_ValidCredentials(t *testing.T) {
	handler := &Handler{
		db:        mockDB{valid: true},
		jwtSecret: []byte("test-secret"),
	}
	
	req := httptest.NewRequest("POST", "/login",
		strings.NewReader(`{"email":"user@test.com","password":"secret"}`))
	w := httptest.NewRecorder()
	
	handler.Login(w, req)
	
	require.Equal(t, http.StatusOK, w.Code)
}
```

실행하면:

```bash
go test ./... -v -cover
--- PASS: TestLogin_ValidCredentials (0.03s)
coverage: 94.2%
```

### 3단계: 문서화

```bash
/moai sync SPEC-AUTH-001
```

자동으로:
- API 문서 업데이트
- 보안 가이드라인 작성
- 예제 curl 명령 생성
- 아키텍처 다이어그램 생성
- PR 생성

## MoAI-ADK의 특징

- **937+ GitHub Stars** (2026년 4월 기준)
- **24개 전문 에이전트**: 백엔드, 프론트엔드, 보안, 디버깅, 테스트, 문서화
- **52개 재사용 가능한 스킬**: 인증, 테스팅, 데이터베이스 마이그레이션, 디자인 시스템
- **16개 프로그래밍 언어**: Go, Python, TypeScript, Rust 등 자동 감지
- **4개 언어 문서**: 한국어, 영어, 일본어, 중국어 (https://adk.mo.ai.kr)
- **Go 단일 바이너리**: 의존성 0개. 설치 최소화.

## 워크플로우의 변화

이전: 자유도 높음 → 품질 낮음 → 수동 수정 → 데드라인 증가

지금: SPEC 작성 → 자동 구현 → 자동 테스트 → 자동 품질 보증 → PR 생성

제 경험상 SPEC 작성에 30분 투자하면, 구현은 1시간 안에 끝나고, 버그는 거의 없어요.

## 10분 안에 시작하기

```bash
# 설치
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash

# 프로젝트 초기화
moai init my-project
cd my-project

# Claude Code 실행
claude

# Claude Code에서
/moai plan "구현하고 싶은 기능"
/moai run SPEC-001
/moai sync SPEC-001
```

언어는 자동 감지돼요. 수동 설정 거의 없어요.

## 마지막 조언

MoAI-ADK는 완벽하진 않아요:

- 자유도가 낮아요. Plan → Run → Sync는 필수예요. 그래서 품질이 보증돼요.
- SPEC을 잘 써야 합니다. 요구사항이 명확할수록 코드 품질도 높아요.
- Claude Code v2.1.110 이상 필요해요.

하지만 테스트를 빠뜨리고 싶지 않은 분들, 품질을 포기하고 싶지 않은 분들에겐 정말 추천해요.

**[GitHub에서 별 주세요](https://github.com/modu-ai/moai-adk) | [한국어 문서 읽기](https://adk.mo.ai.kr/ko)**

---

**단어: 950 | 읽기 수준: 대학 수준**
