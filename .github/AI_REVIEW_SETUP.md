# 🤖 AI Code Review Setup Guide

AI-powered PR review 및 자동 승인 시스템 구성 가이드입니다.

## 📋 개요

- **도구**: [PR-Agent](https://github.com/qodo-ai/pr-agent) (Qodo)
- **LLM**: Claude 3.5 Sonnet (Anthropic) 또는 GPT-4 (OpenAI)
- **기능**: 자동 코드 리뷰, PR 승인, 코멘트 작성
- **상태**: ✅ 워크플로우 설정 완료

## 🚀 빠른 시작

### 1단계: GitHub Secrets 설정

#### Option A: Claude 사용 (권장)

```bash
# GitHub Repository Settings → Secrets → New repository secret
# 또는 gh CLI 사용:
gh secret set ANTHROPIC_API_KEY --body "sk-ant-..."
```

**필요한 정보:**
- [Anthropic API Key](https://console.anthropic.com/account/keys) 발급
- MoAI-ADK 저장소의 Secrets에 `ANTHROPIC_API_KEY` 등록

#### Option B: OpenAI GPT-4 사용

```bash
gh secret set OPENAI_API_KEY --body "sk-..."
```

**필요한 정보:**
- [OpenAI API Key](https://platform.openai.com/account/api-keys) 발급
- MoAI-ADK 저장소의 Secrets에 `OPENAI_API_KEY` 등록

**주의**: `.github/workflows/ai-review.yml`에서 모델 설정 변경 필요

### 2단계: 워크플로우 확인

```bash
# 워크플로우 파일 확인
cat .github/workflows/ai-review.yml
```

**현재 설정:**
- ✅ LLM: Claude 3.5 Sonnet
- ✅ 자동 승인: Draft 아닌 PR (준비된 PR)
- ✅ 승인 임계값: 80% 품질 점수
- ✅ 리뷰 대상: 모든 PR (자동 실행)

### 3단계: 테스트 PR 생성

새로운 PR을 생성하면 자동으로 AI 리뷰가 시작됩니다:

```bash
# 테스트 PR 생성
git checkout -b test/ai-review
echo "test" > test-file.txt
git add .
git commit -m "test: AI review workflow"
git push origin test/ai-review
gh pr create --title "test: AI review" --body "Testing AI review workflow"
```

**예상 동작:**
1. AI 리뷰 워크플로우 자동 실작 (1-2분)
2. PR에 리뷰 코멘트 게시
3. 품질 점수 80% 이상 → 자동 승인 (Draft 아닐 경우)

## 🎯 리뷰 기준

AI가 검토하는 항목:

| 범주 | 가중치 | 평가 항목 |
|------|--------|---------|
| **코드 품질** | 25% | 디자인 패턴, 가독성, 유지보수성 |
| **보안** | 30% | OWASP Top 10, 취약점, 암호화 |
| **테스트** | 25% | 커버리지, 엣지 케이스, 픽스처 |
| **문서화** | 15% | Docstring, 주석, 설명 |
| **성능** | 5% | 최적화, 복잡도 |

### Python 프로젝트 추가 항목

- Type hints 및 mypy 준수
- PEP-8 스타일 및 ruff 린팅
- pytest 커버리지 (85%+ 목표)
- Async/await 패턴
- Exception handling

## 🔧 워크플로우 설정 파일

### 파일 위치
```
.github/workflows/ai-review.yml
```

### 주요 설정 옵션

```yaml
# LLM 선택
model: "claude-3-5-sonnet"  # Claude
# model: "gpt-4"            # OpenAI

# 자동 승인
auto_approve: true           # 자동 승인 활성화
approval_threshold: 0.80    # 80% 이상 품질 점수 시 승인

# 코멘트 설정
comment_on_pr: true         # PR에 리뷰 코멘트
persistent_comment: true    # 영구 코멘트 (업데이트됨)
```

## 📊 리뷰 결과 확인

### PR의 리뷰 코멘트 확인

1. GitHub PR 페이지 → **Conversation** 탭
2. AI 리뷰 코멘트 확인 (구조화된 리뷰)
3. 추천 개선사항 확인

### 자동 승인 확인

- Draft 아닌 PR: 자동 승인됨 (✅ Approved 배지)
- Draft PR: 수동 승인 필요 (테스트 중 변경 가능)

### 트러블슈팅

**Q: AI 리뷰 코멘트가 없음**
- Secrets 설정 확인: `ANTHROPIC_API_KEY` 또는 `OPENAI_API_KEY` 존재?
- 워크플로우 로그 확인: Actions 탭 → ai-review 워크플로우 → 오류 확인

**Q: 자동 승인이 안 됨**
- Draft PR 확인: Draft PR은 자동 승인 안 함
- 임계값 확인: 80% 미만 품질이면 승인 안 함
- 토큰 권한 확인: `pull-requests: write` 권한 필요

**Q: 비용 관련 문제**
- **Claude (Anthropic)**: 사용량 기반 요금 (매월 청구)
- **GPT-4 (OpenAI)**: 사용량 기반 요금 (매월 청구)
- 예상 비용: PR당 $0.01-0.05

## 🔐 보안 고려사항

### 1. Secrets 보호
- ✅ `ANTHROPIC_API_KEY`는 Secrets로 관리 (노출 방지)
- ✅ 로그에 API 키 노출 안 함 (자동)

### 2. 토큰 권한 제한
```yaml
permissions:
  pull-requests: write  # PR 리뷰 & 승인만
  contents: read        # 코드 읽기만
```

### 3. 자동 승인 제한
- Draft PR은 자동 승인 안 함 (임계값 설정)
- 80% 품질 미만은 승인 안 함
- 수동 리뷰는 여전히 필요할 수 있음

## 📈 모범 사례

### 1. 점진적 도입
```
Week 1: Review only (자동 승인 OFF)
        → 워크플로우 안정성 확인
Week 2: Review + Auto-comment (코멘트만)
        → 개발자 피드백 수집
Week 3: Review + Auto-approve (본격 적용)
        → 최적화 시작
```

### 2. 품질 기준 조정
초기 임계값은 보수적으로 설정:
```yaml
approval_threshold: 0.90  # 첫 주: 90%
approval_threshold: 0.85  # 둘째 주: 85%
approval_threshold: 0.80  # 셋째 주: 80% (안정화)
```

### 3. 리뷰 규칙 커스터마이징
```yaml
review_pr_instructions: |
  - Check for TRUST 5 principles compliance
  - Validate @TAG system usage
  - Ensure test coverage ≥ 85%
  - Verify documentation updates
```

## 🎓 학습 자료

- [PR-Agent Documentation](https://github.com/qodo-ai/pr-agent)
- [Anthropic API Docs](https://docs.anthropic.com/)
- [GitHub Actions](https://docs.github.com/en/actions)
- [MoAI-ADK CLAUDE.md](../../CLAUDE.md)

## 📞 지원

문제 발생 시:
1. `.github/workflows/ai-review.yml` 로그 확인
2. GitHub Actions 탭 → ai-review 워크플로우 → 세부 정보
3. PR-Agent [GitHub Issues](https://github.com/qodo-ai/pr-agent/issues) 확인

---

✨ AI 리뷰 시스템이 MoAI-ADK의 품질 자동화를 한 단계 업그레이드했습니다!
