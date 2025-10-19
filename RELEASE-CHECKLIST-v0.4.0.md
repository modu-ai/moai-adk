# MoAI-ADK v0.4.0 배포 체크리스트

> **배포 날짜**: 2025-10-20
> **작성자**: Alfred SuperAgent
> **버전**: v0.4.0 (Skills Revolution)

---

## 📋 배포 전 체크리스트

### 1️⃣ 코드 및 문서 검증

- [x] ✅ 모든 SPEC 완료 (31개 SPEC, status: completed)
- [x] ✅ 테스트 커버리지 87.66% (목표 85% 달성)
- [x] ✅ CHANGELOG.md 업데이트 완료
- [x] ✅ 신규 Skills 2개 추가
  - [x] moai-alfred-code-reviewer (자동 코드 리뷰)
  - [x] moai-alfred-error-explainer (자동 에러 분석)
- [x] ✅ 템플릿 동기화 완료
  - [x] .claude/skills/ → src/moai_adk/templates/.claude/skills/
- [x] ✅ Git 상태 깨끗 (모든 변경사항 커밋 완료)

---

## 📦 v0.4.0 핵심 내용

### ✨ 신규 추가 (Alfred 전용 Skills)

1. **moai-alfred-code-reviewer**
   - 역할: PR 생성 시 Alfred가 자동으로 코드 리뷰 수행
   - 기능: TRUST 5원칙 + SOLID 원칙 + Code Smells 통합 검증
   - 호출: /alfred:3-sync 완료 후 자동

2. **moai-alfred-error-explainer**
   - 역할: 런타임 에러 발생 시 Alfred가 자동으로 원인 분석
   - 기능: Stack trace 파싱 + SPEC 기반 원인 분석 + 3단계 해결 방법
   - 호출: 에러 발생 시 자동

### 📊 Skills 현황

- **v0.4.0**: 46개 Skills
  - Foundation: 6개
  - Essentials: 4개
  - **Alfred: 2개** ⭐ NEW
  - Domain: 10개
  - Language: 23개
  - Claude Code: 1개

---

## 🚀 배포 절차

### 1. 최종 검증
```bash
# Git 상태 확인
git status

# 버전 확인
grep "^version" pyproject.toml

# Skills 개수 확인
ls -d .claude/skills/*/ | wc -l  # 46개 확인
```

### 2. Git 커밋 및 태그
```bash
git add .
git commit -m "🚀 RELEASE: v0.4.0 - Skills Revolution

- ✅ Skills 46개 제공 (Alfred 전용 2개 추가)
- ✅ moai-alfred-code-reviewer (자동 코드 리뷰)
- ✅ moai-alfred-error-explainer (자동 에러 분석)
- ✅ CHANGELOG.md 업데이트

🤖 Generated with [Claude Code](https://claude.com/claude-code)"

git tag -a v0.4.0 -m "v0.4.0: Skills Revolution"
git push origin develop
git push origin v0.4.0
```

### 3. PyPI 배포
```bash
# 빌드
python -m build

# 배포
twine upload dist/*
```

### 4. GitHub Release 생성
```bash
gh release create v0.4.0 \
  --title "v0.4.0: Skills Revolution 🎯" \
  --notes-file RELEASE-NOTES-v0.4.0.md
```

---

## 📝 다음 단계

- [ ] PyPI 다운로드 모니터링
- [ ] 사용자 피드백 수집
- [ ] Issue #41 해결 여부 확인
- [ ] v0.5.0 계획 검토 (.moai/reports/v0.5.0-future-plan.md)

---

**배포 담당자**: @Goos  
**최종 확인**: Alfred SuperAgent  
**배포 상태**: ✅ 준비 완료
