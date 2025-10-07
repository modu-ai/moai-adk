# MoAI-ADK v0.2.10 배포 체크리스트

📦 **패키지**: moai-adk
🏷️ **버전**: 0.2.10
📅 **배포일**: 2025-10-07

---

## 📋 Phase 1: 사전 검증 (Pre-Deployment Verification)

### 1.1 코드 품질 검증

```bash
cd /Users/goos/MoAI/MoAI-ADK/moai-adk-ts
```

- [ ] **TypeScript 타입 체크**
  ```bash
  bun run type-check
  ```
  **기대 결과**: ✅ No type errors

- [ ] **Biome Lint 검사**
  ```bash
  bun run check:biome
  ```
  **기대 결과**: ✅ All checks passed

- [ ] **테스트 실행**
  ```bash
  bun run test
  ```
  **기대 결과**: ✅ All tests passed

- [ ] **빌드 검증**
  ```bash
  bun run build
  ```
  **기대 결과**:
  - ✅ `dist/index.js` 생성
  - ✅ `dist/index.cjs` 생성
  - ✅ `dist/index.d.ts` 생성
  - ✅ `templates/.claude/hooks/alfred/session-notice.cjs` 생성

### 1.2 버전 정보 검증

- [ ] **package.json 버전 확인**
  ```bash
  cat package.json | grep '"version"'
  ```
  **기대 결과**: `"version": "0.2.10"`

- [ ] **CHANGELOG.md 최신화 확인**
  ```bash
  head -n 50 CHANGELOG.md
  ```
  **기대 결과**: v0.2.10 항목 존재

- [ ] **Release Notes 존재 확인**
  ```bash
  ls -la RELEASE_NOTES_v0.2.10.md
  ```
  **기대 결과**: 파일 존재

### 1.3 Git 상태 확인

- [ ] **변경사항 확인**
  ```bash
  git status
  ```
  **기대 결과**: Working tree clean 또는 모든 변경사항 커밋됨

- [ ] **현재 브랜치 확인**
  ```bash
  git branch --show-current
  ```
  **현재**: `feature/SPEC-INIT-003`
  **권장**: main/master로 머지 후 배포

- [ ] **최신 커밋 확인**
  ```bash
  git log -1 --oneline
  ```

---

## 📦 Phase 2: NPM 배포 (NPM Publish)

### 2.1 배포 스크립트 실행 권한 부여

```bash
chmod +x scripts/publish.sh
```

### 2.2 배포 스크립트 실행

```bash
cd /Users/goos/MoAI/MoAI-ADK/moai-adk-ts
./scripts/publish.sh
```

**스크립트 실행 단계**:
1. ✅ 디렉토리 확인 (package.json 존재)
2. ✅ 버전 검증 (0.2.10)
3. ⚠️ Git 상태 확인 (변경사항 있으면 경고)
4. ✅ 의존성 설치 (`bun install`)
5. ✅ TypeScript 타입 체크
6. ✅ Lint 검사
7. ✅ 테스트 실행
8. ✅ 빌드 실행
9. ✅ 빌드 결과 검증
10. 🤔 배포 확인 프롬프트 → **y 입력**
11. 📤 NPM publish 실행

### 2.3 NPM 배포 검증

- [ ] **NPM 배포 성공 메시지 확인**
  ```
  ✅ 배포 성공!
     패키지: https://www.npmjs.com/package/moai-adk
     버전: v0.2.10
  ```

- [ ] **NPM 웹사이트 확인**
  - 브라우저에서 https://www.npmjs.com/package/moai-adk 접속
  - **Latest version**: 0.2.10 확인

- [ ] **NPM 레지스트리 확인**
  ```bash
  npm view moai-adk version
  ```
  **기대 결과**: `0.2.10`

- [ ] **설치 테스트 (선택적)**
  ```bash
  mkdir /tmp/test-moai-adk-install
  cd /tmp/test-moai-adk-install
  npm init -y
  npm install moai-adk@0.2.10
  npm list moai-adk
  ```
  **기대 결과**: `moai-adk@0.2.10`

---

## 🏷️ Phase 3: Git 태깅 (Git Tagging)

### 3.1 Git 태그 생성

```bash
cd /Users/goos/MoAI/MoAI-ADK
git tag -a v0.2.10 -m "Release v0.2.10: Configuration Schema Enhancement & Auto-Version Management"
```

### 3.2 태그 확인

- [ ] **태그 생성 확인**
  ```bash
  git tag -l "v0.2.10"
  ```
  **기대 결과**: `v0.2.10`

- [ ] **태그 상세 정보 확인**
  ```bash
  git show v0.2.10
  ```
  **기대 결과**: 태그 메시지 및 커밋 정보 표시

### 3.3 태그 푸시

```bash
git push origin v0.2.10
```

- [ ] **푸시 성공 확인**
  ```
  To https://github.com/modu-ai/moai-adk.git
   * [new tag]         v0.2.10 -> v0.2.10
  ```

---

## 🐙 Phase 4: GitHub Release 생성

### 4.1 GitHub CLI 인증 확인

```bash
gh auth status
```

**기대 결과**: ✅ Logged in to github.com

### 4.2 GitHub Release 생성

#### 방법 1: CLI로 Release Notes 첨부

```bash
cd /Users/goos/MoAI/MoAI-ADK/moai-adk-ts
gh release create v0.2.10 \
  --title "MoAI-ADK v0.2.10 - Configuration Schema Enhancement" \
  --notes-file RELEASE_NOTES_v0.2.10.md \
  --repo modu-ai/moai-adk
```

#### 방법 2: 대화형으로 생성

```bash
gh release create v0.2.10 \
  --title "MoAI-ADK v0.2.10 - Configuration Schema Enhancement" \
  --generate-notes \
  --repo modu-ai/moai-adk
```

### 4.3 GitHub Release 검증

- [ ] **Release 페이지 확인**
  ```bash
  gh release view v0.2.10 --repo modu-ai/moai-adk
  ```

- [ ] **웹 브라우저 확인**
  - https://github.com/modu-ai/moai-adk/releases/tag/v0.2.10 접속
  - Release notes 내용 확인
  - Assets 확인 (Source code zip/tar.gz 자동 생성)

---

## 📚 Phase 5: 문서 업데이트

### 5.1 README.md 업데이트 (필요 시)

- [ ] **설치 명령어 최신 버전 반영**
  ```bash
  # 이전
  npm install moai-adk@0.2.6

  # 현재
  npm install moai-adk@0.2.10
  # 또는
  npm install moai-adk@latest
  ```

### 5.2 문서 사이트 배포 (필요 시)

```bash
cd /Users/goos/MoAI/MoAI-ADK/moai-adk-ts
bun run docs:build
bun run docs:preview
```

- [ ] **문서 빌드 성공 확인**
- [ ] **문서 사이트 접속**: https://moai-adk.vercel.app
- [ ] **버전 정보 최신화 확인**

---

## ✅ Phase 6: 사후 검증 (Post-Deployment Verification)

### 6.1 전체 배포 파이프라인 검증

- [ ] **NPM 설치 테스트**
  ```bash
  npm install -g moai-adk@0.2.10
  moai --version
  ```
  **기대 결과**: `0.2.10` 또는 버전 정보 표시

- [ ] **신규 프로젝트 초기화 테스트**
  ```bash
  mkdir /tmp/test-moai-init
  cd /tmp/test-moai-init
  moai init
  # 프롬프트 입력 후
  cat .moai/config.json | grep -A2 '"moai"'
  ```
  **기대 결과**:
  ```json
  "moai": {
    "version": "0.2.10"
  }
  ```

- [ ] **Session-Start Hook 동작 확인**
  - Claude Code에서 프로젝트 열기
  - 세션 시작 시 버전 정보 확인
  ```
  📦 버전: v0.2.10 (최신)
  ```

### 6.2 커뮤니티 공지 (선택적)

- [ ] **GitHub Discussions 게시**
  - Release 공지
  - 주요 변경사항 설명
  - Migration guide 링크

- [ ] **NPM README 업데이트 확인**
  - https://www.npmjs.com/package/moai-adk
  - README가 최신 버전 반영 확인

---

## 🚨 롤백 계획 (Rollback Plan)

만약 배포 후 치명적인 문제 발견 시:

### NPM 롤백

```bash
# 이전 버전을 latest로 복구
npm dist-tag add moai-adk@0.2.6 latest

# 0.2.10을 deprecated 마킹
npm deprecate moai-adk@0.2.10 "Critical bug - use 0.2.6 instead"
```

### GitHub Release 롤백

```bash
# Release 삭제 (태그는 유지)
gh release delete v0.2.10 --repo modu-ai/moai-adk --yes

# 태그 삭제 (필요 시)
git tag -d v0.2.10
git push origin :refs/tags/v0.2.10
```

---

## 📊 배포 요약 (Deployment Summary)

배포 완료 후 작성:

```
✅ NPM 배포: https://www.npmjs.com/package/moai-adk/v/0.2.10
✅ Git 태그: v0.2.10
✅ GitHub Release: https://github.com/modu-ai/moai-adk/releases/tag/v0.2.10
✅ 문서 사이트: https://moai-adk.vercel.app

배포 시간: [YYYY-MM-DD HH:MM:SS]
배포자: @Goos
총 소요 시간: [XX분]
```

---

## 🎉 다음 단계 (Next Steps)

배포 후 권장 작업:

1. [ ] **모니터링 설정**
   - NPM download 통계 확인
   - GitHub star/fork 변화 추적

2. [ ] **사용자 피드백 수집**
   - GitHub Issues 모니터링
   - NPM weekly downloads 확인

3. [ ] **다음 버전 계획**
   - v0.2.11 또는 v0.3.0 로드맵 작성
   - 새 기능/개선 사항 정리

---

🗿 Generated with MoAI-ADK v0.2.10
