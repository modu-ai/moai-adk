# Claude Code Statusline 설정 가이드

## 개요

MoAI-ADK statusline은 Claude Code 개발 환경에서 프로젝트 상태, Git 정보, Alfred 워크플로우 진행 상황을 **실시간으로** 한눈에 파악할 수 있게 해줍니다.

---

## 설치 완료 확인

✅ **설치가 완료되었습니다!** 다음 단계만 진행하면 됩니다:

### 1단계: Claude Code 재시작

현재 세션을 종료하고 Claude Code를 **다시 시작**합니다.

```bash
# 또는 Claude Code 앱 종료 후 재실행
```

### 2단계: 상태줄 확인

Claude Code를 재시작하면 터미널 하단에 다음과 같은 상태줄이 표시됩니다:

```
H 4.5 | 5m | MoAI-ADK | 0.20.1 | feature/SPEC-001 | +2 M1 | [PLAN]
```

---

## 표시되는 정보

| 정보 | 예시 | 설명 |
|------|------|------|
| **모델** | `H 4.5` | 현재 사용 중인 AI 모델 (H=Haiku, S=Sonnet) |
| **시간** | `5m`, `1h 30m` | 세션 경과 시간 |
| **디렉토리** | `MoAI-ADK` | 프로젝트명 |
| **버전** | `0.20.1` | MoAI-ADK 현재 설치 버전 |
| **브랜치** | `feature/SPEC-001` | Git branch (색상으로 구분) |
| **상태** | `+2 M1 ?0` | Git 변경 사항 (staged/modified/untracked) |
| **작업** | `[PLAN]`, `[RUN]`, `[SYNC]` | 활성 Alfred 명령 |
| **업데이트** | `⬆️ 0.21.0` | 새 버전 가용성 (선택적) |

---

## 디스플레이 모드

### Compact 모드 (기본, 80자)

```
H 4.5 | 5m | MoAI-ADK | 0.20.1 | feature/SPEC-001 | +2 M1 | [PLAN]
```

표준 터미널에서 가장 많이 사용되는 모드입니다.

### Extended 모드 (120자)

```
Haiku 4.5 | 1h 30m | /Users/goos/MoAI/MoAI-ADK | v0.20.1 | feature/SPEC-AUTH-001 | +5 M3 ?2 | [RUN-GREEN]
```

더 자세한 정보를 원할 때 사용합니다.

### Minimal 모드 (40자)

```
H | 5m | 0.20.1 | feature/001 | +2M
```

극도로 제한된 환경에서 사용합니다.

**모드 변경 방법:**

`.claude/settings.json`에서 다음 부분을 수정합니다:

```json
"statusLine": {
  "type": "command",
  "command": "python3 -m moai_adk.statusline.main --mode compact",
  "padding": 1
}
```

가능한 모드: `compact`, `extended`, `minimal`

---

## 색상 설명

### Git Branch 색상

| 색상 | Branch | 의미 |
|------|--------|------|
| 🟨 노란색 | `feature/*` | 작업 진행 중 |
| 🔵 파란색 | `develop` | 통합 브랜치 |
| 🟢 초록색 | `main` | 릴리스 브랜치 |

### Git 변경사항 색상

| 기호 | 의미 | 색상 |
|------|------|------|
| `+3` | Staged changes | 🟢 초록색 |
| `M2` | Modified files | 🟠 주황색 |
| `?1` | Untracked files | 🔴 빨간색 |

### 업데이트 표시

| 기호 | 의미 |
|------|------|
| `⬆️` 또는 `↑` | 새 버전 가용 |
| `0.20.1 ⬆️ 0.21.0` | 현재 버전 → 최신 버전 |

---

## 성능 특성

MoAI-ADK statusline은 최고의 성능으로 최적화되었습니다:

### 캐싱 전략

| 정보 | 캐시 시간 | 목적 |
|------|---------|------|
| Git 정보 | 5초 | 명령 최소화 |
| 세션 시간 | 10초 | 계산 최적화 |
| Alfred 작업 | 1초 | 빠른 반응 |
| 버전 정보 | 60초 | 파일 읽기 최소화 |
| 업데이트 확인 | 300초 | API 호출 최소화 |

### 리소스 사용량

- **CPU**: < 2% (정상 작동 중)
- **메모리**: < 5MB
- **업데이트 주기**: 300ms (Claude Code API 제약)
- **디스크 I/O**: 최소화 (캐싱)

---

## 트러블슈팅

### 문제: 상태줄이 표시되지 않음

**해결 방법:**

1. Claude Code 재시작 확인
   ```bash
   # Claude Code 완전히 종료하고 다시 실행
   ```

2. 설정 확인
   ```bash
   cat .claude/settings.json | grep statusLine
   ```

3. Python 경로 확인
   ```bash
   which python3
   python3 --version  # Python 3.10 이상 필요
   ```

4. 모듈 설치 확인
   ```bash
   python3 -c "from moai_adk.statusline.main import main; print('OK')"
   ```

### 문제: 오류 메시지 표시 ([ERROR] ...)

**가능한 원인과 해결:**

| 오류 | 원인 | 해결 방법 |
|------|------|---------|
| `Git repository not found` | Git 저장소 미초기화 | `git init` 실행 |
| `Permission denied` | 파일 접근 권한 오류 | 권한 확인: `ls -la .moai/` |
| `Module not found` | Python 패키지 미설치 | `pip install moai-adk` 재실행 |
| `Network error` | 업데이트 확인 실패 | 네트워크 연결 확인 (오류 무시됨) |

### 문제: 상태줄이 깨져 보임

**가능한 원인:**

- 터미널이 ANSI 256-color를 지원하지 않음
- 터미널 인코딩이 UTF-8이 아님

**해결 방법:**

```bash
# 터미널 색상 지원 확인
echo $TERM
# xterm-256color 또는 screen-256color 이상 필요

# 인코딩 확인
echo $LANG
# UTF-8 필요
```

---

## 고급 사용법

### 환경변수로 모드 변경

```bash
export MOAI_STATUSLINE_MODE=extended
# 다음 Claude Code 세션에서 Extended 모드 사용
```

### 프로젝트별 설정

`.moai/config.json`에 statusline 섹션을 추가하여 프로젝트별로 설정할 수 있습니다:

```json
{
  "statusline": {
    "enabled": true,
    "mode": "compact",
    "colors": {
      "enabled": true,
      "theme": "auto"
    }
  }
}
```

### 업데이트 확인 비활성화

업데이트 확인을 비활성화하려면 PyPI API 호출을 제거할 수 있습니다:

```python
# src/moai_adk/statusline/main.py에서 다음 줄 주석 처리
# update_available, latest_version = safe_check_update(version)
```

---

## 자주 묻는 질문 (FAQ)

### Q: 상태줄이 너무 느립니다

**A:** 캐싱 시간을 확인하세요. 기본 캐싱이 충분히 최적화되어 있습니다.
- Git 캐싱: 5초
- 업데이트 확인: 300초 (가장 오래 - 수정 가능)

### Q: 색상이 내 테마와 맞지 않습니다

**A:** `.claude/statusline-config.yaml`에서 색상을 커스터마이즈할 수 있습니다:

```yaml
colors:
  palette:
    model: "38;5;33"  # 원하는 ANSI 색상 코드로 변경
```

### Q: 왜 업데이트 아이콘(⬆️)이 표시되지 않습니다?

**A:** 몇 가지 이유가 있을 수 있습니다:
- 최신 버전을 사용 중입니다 (업데이트 필요 없음)
- 인터넷 연결 문제 (오류는 무시되고 표시 안 됨)
- 터미널이 이모지를 지원하지 않습니다 (기호 `↑`로 표시됨)

### Q: Git 커맨드 timeout을 조정할 수 있습니까?

**A:** `src/moai_adk/statusline/git_collector.py`에서 timeout을 수정할 수 있습니다:

```python
result = subprocess.run(
    ["git", "status", "-b", "--porcelain"],
    timeout=10,  # 기본: 5초, 필요시 변경
    capture_output=True,
    text=True
)
```

### Q: Claude Code 외의 다른 에디터에서 사용할 수 있습니까?

**A:** 현재는 Claude Code 전용입니다. 하지만 `main.py`를 다른 에디터의 statusline API에 맞게 수정할 수 있습니다.

---

## 피드백 및 버그 보고

상태줄에 문제가 있거나 개선 제안이 있으시면:

1. GitHub Issues에 보고: https://github.com/MoAI-ADK/issues
2. 버그 분류:
   - 표시 오류: 상태줄에 잘못된 정보 표시
   - 성능 문제: CPU/메모리 사용량 초과
   - 호환성: 특정 환경에서 작동 안 함

---

## 더 알아보기

- **SPEC 문서**: `.moai/specs/SPEC-CLAUDE-STATUSLINE-001/spec.md`
- **구현 코드**: `src/moai_adk/statusline/`
- **테스트**: `tests/statusline/`
- **설정**: `.claude/statusline-config.yaml`

---

**최종 확인 체크리스트:**

- [ ] Claude Code 재시작 완료
- [ ] 상태줄이 표시됨
- [ ] 7가지 정보가 모두 보임 (모델, 시간, 디렉토리, 버전, 브랜치, 상태, 작업)
- [ ] 색상이 정상 표시됨
- [ ] Git 변경사항이 정확함

모든 항목을 확인했다면 설치가 완료되었습니다! 🎉
