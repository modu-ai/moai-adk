---
id: SPEC-SEARCH-001
type: research
---

# SPEC-SEARCH-001: 리서치 문서

## 현재 상태 분석

### JSONL 세션 파일 현황
- 위치: `~/.claude/projects/{project-hash}/*.jsonl`
- 파일 수: 758개 이상 (2026-03 기준)
- 평균 크기: ~2-22MB
- 최대 크기: 22MB (1677개 assistant 레코드)
- 형식: 줄 단위 JSON (JSONL)

### 현재 검색 방법의 한계
- `grep -r` 기반: JSONL 구조 무시, JSON 이스케이프 문자 매칭 불가
- 속도: 758개 파일 전체 grep에 수십 초 소요
- 정확도: 구조화되지 않은 텍스트 매칭으로 노이즈 다수
- 필터링: 날짜, 브랜치, 역할별 필터링 불가

## JSONL 레코드 구조 분석

### 레코드 유형
```json
{"type": "user", "message": {"role": "user", "content": "검색할 텍스트"}, "timestamp": "2026-03-06T..."}
{"type": "assistant", "message": {"role": "assistant", "content": [{"type": "text", "text": "응답 텍스트"}]}, "timestamp": "2026-03-06T..."}
{"type": "tool_use", "message": {"role": "assistant", "content": [{"type": "tool_use", ...}]}}
{"type": "tool_result", "message": {"role": "user", "content": [{"type": "tool_result", ...}]}}
```

### content 필드 구조
- `user` 메시지: `content`가 문자열 (직접 텍스트)
- `assistant` 메시지: `content`가 배열 (`[{"type": "text", "text": "..."}]`)
- `tool_use`/`tool_result`: 도구 호출/결과 데이터 (검색 대상 아님)

### 노이즈 패턴 (필터링 대상)
- `<local-command-caveat>...</local-command-caveat>`: Claude Code 내부 주의사항
- `<command-name>...</command-name>`: 명령 이름 메타데이터
- `<system-reminder>...</system-reminder>`: 시스템 프롬프트 리마인더
- `<function_calls>...</function_calls>`: 도구 호출 XML
- `<skill-format>true</skill-format>`: 스킬 로딩 마커

## SQLite FTS5 trigram 토크나이저 선택 근거

### 토크나이저 비교

| 토크나이저 | 한국어 지원 | 인덱스 크기 | 검색 정확도 |
|-----------|-----------|-----------|-----------|
| unicode61 | X (공백 기반) | 작음 | 영어 우수 |
| porter | X (영어 스테밍) | 작음 | 영어 우수 |
| trigram | O (문자 3-gram) | 큼 | 부분 문자열 매칭 |
| ICU | O (형태소) | 보통 | 우수하나 설정 복잡 |

### trigram 선택 이유
1. **CJK 호환성**: 형태소 분석 없이 한국어/일본어/중국어 검색 가능
2. **설정 간소화**: 외부 사전/형태소 분석기 불필요
3. **부분 매칭**: "인증"으로 "인증 구현", "사용자 인증", "JWT 인증" 모두 매칭
4. **인덱스 크기 절충**: 다소 큰 인덱스이나, 세션 데이터 규모에서 수용 가능

### trigram 한계
- "인"만으로는 매칭 불가 (최소 3바이트 필요, 한글 1글자 = 3바이트)
- BM25 랭킹 정확도가 형태소 기반보다 낮음
- 인덱스 크기가 원본 데이터의 3-5배

## 비동기 훅 패턴 선택 근거

### 대안 비교

| 패턴 | 훅 블로킹 | 복잡도 | 인덱싱 보장 |
|------|---------|--------|-----------|
| 동기 실행 | O (타임아웃 위험) | 낮음 | 높음 |
| 고루틴 | X | 낮음 | 낮음 (프로세스 종료 시 손실) |
| 서브프로세스 | X | 보통 | 높음 |
| 메시지 큐 | X | 높음 | 매우 높음 |

### 서브프로세스 선택 이유
1. **훅 타임아웃 안전**: `cmd.Start()` 즉시 반환, 훅 5초 제한 준수
2. **프로세스 독립성**: 훅 프로세스 종료 후에도 인덱싱 계속
3. **단순성**: 고루틴 동기화나 큐 인프라 불필요
4. **재실행 가능**: 실패 시 수동 `moai search --index-session` 가능

### 고루틴 미선택 이유
- `session_end.go`의 `Handle()`은 동기 함수
- 훅 프로세스 종료 시 고루틴도 종료
- 인덱싱 완료 보장 불가

## modernc.org/sqlite 선택 근거

### SQLite Go 드라이버 비교

| 드라이버 | CGO 필요 | FTS5 | 크로스 컴파일 | 유지보수 |
|---------|---------|------|-------------|---------|
| mattn/go-sqlite3 | O | O (빌드 태그) | 어려움 | 활발 |
| modernc.org/sqlite | X | O (기본) | 쉬움 | 활발 |
| crawshaw.io/sqlite | O | O | 어려움 | 중단 |

### modernc.org/sqlite 선택 이유
1. **CGO 불필요**: `moai` 바이너리의 크로스 컴파일 유지
2. **FTS5 기본 활성화**: 별도 빌드 태그 불필요
3. **호환성**: `database/sql` 표준 인터페이스 준수
4. **성능**: CGO 버전 대비 약 10-20% 느리나 CLI 사용에 충분

### 빌드 영향
- `go.mod`에 약 15개 간접 의존성 추가
- 초기 빌드 시간 약 30-60초 증가 (캐시 후 무시 가능)
- 바이너리 크기 약 10-15MB 증가

## 코드베이스 참조 패턴

### Cobra CLI 패턴 (기존 명령 참조)

```go
// internal/cli/init.go 패턴
func newInitCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "init [project-name]",
        Short: "Initialize a new MoAI project",
        RunE:  runInit,
    }
    cmd.Flags().StringP("template", "t", "", "template name")
    return cmd
}
```

### SessionEnd 훅 패턴 (기존 핸들러 참조)

```go
// internal/hook/session_end.go 패턴
func (h *SessionEndHandler) Handle(input []byte) error {
    // 기존 cleanup 로직...
    // 여기에 triggerSessionIndex() 추가
    return nil
}
```

### lipgloss 출력 패턴 (기존 UI 참조)

CLI 출력은 프로젝트의 기존 lipgloss `renderCard` 패턴을 따름.

### 테스트 패턴 (기존 테스트 참조)
- table-driven tests 사용
- `t.TempDir()`로 격리
- `t.Parallel()` 적용
- testify 미사용 (stdlib만)

## 추적성

- SPEC: SPEC-SEARCH-001
- Spec: `.moai/specs/SPEC-SEARCH-001/spec.md`
- Plan: `.moai/specs/SPEC-SEARCH-001/plan.md`
- Acceptance: `.moai/specs/SPEC-SEARCH-001/acceptance.md`
