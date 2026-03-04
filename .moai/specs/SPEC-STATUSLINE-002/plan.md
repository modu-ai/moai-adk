# SPEC-STATUSLINE-002: 구현 계획

## 상태: Draft

---

## 1. 구현 방법론

**개발 방법**: TDD (Red-Green-Refactor)
- 캐시 파일 I/O, 집계 로직, 렌더링 세 영역 각각에 대해 테스트 우선 작성
- `t.TempDir()` 기반 격리된 파일 시스템 테스트

---

## 2. 의존성 분석

```
SPEC-STATUSLINE-001 (완료)
  └── statusline.yaml 세그먼트 설정 시스템
        └── SPEC-STATUSLINE-002 (신규)
              ├── token_history.go (캐시 I/O)
              ├── types.go 확장 (TokenUsageWindow)
              ├── builder.go 확장 (스토어 통합)
              └── renderer.go 확장 (새 세그먼트)
```

---

## 3. 구현 단계

### Phase 1: 데이터 계층 (token_history.go)

**목표**: 세션 토큰 데이터의 로컬 캐시 읽기/쓰기 및 시간 윈도우 집계

**작업 목록**:
1. `TokenHistoryCache`, `SessionRecord`, `TokenUsageData` 타입 정의
2. `fileTokenHistoryStore` 구현:
   - `loadCache()`: JSON 파일 읽기 + 파싱 (파일 없으면 빈 캐시 반환)
   - `saveCache()`: 원자적 파일 쓰기 (tmpfile + rename)
   - `UpsertSession()`: 세션 레코드 upsert
   - `Aggregate()`: 시간 윈도우별 토큰 합산
   - `Purge()`: 90일 초과 레코드 정리 (24h 쿨다운)
3. `NewFileTokenHistoryStore(cacheDir string) TokenHistoryStore` 생성자

**테스트 케이스**:
- 캐시 파일 없을 때 Aggregate → 결과 0, 오류 없음
- 캐시 파일 손상 시 Aggregate → 빈 결과, 오류 없음
- 세션 upsert → 파일 생성 확인
- 동일 session_id 재upsert → 토큰 값 갱신
- 5h 집계: 4시간 전 세션 포함, 6시간 전 세션 제외
- 7d 집계: 6일 전 세션 포함, 8일 전 세션 제외
- Purge: 90일 초과 세션 제거, 이내 세션 보존
- Purge 쿨다운: 24h 이내 재실행 시 스킵

### Phase 2: types.go 확장

**작업 목록**:
1. `TokenUsageWindow` 타입 추가
2. `StatusData`에 `TokenUsage5h`, `TokenUsage7d` 필드 추가
3. `SegmentTokenUsage5h = "token_usage_5h"`, `SegmentTokenUsage7d = "token_usage_7d"` 상수 추가

### Phase 3: builder.go 통합

**작업 목록**:
1. `Options`에 `TokenHistoryStore` 필드 추가 (nil이면 자동 생성)
2. `Build()` 내 데이터 수집 단계에서 UpsertSession + Aggregate 호출
3. 결과를 `StatusData.TokenUsage5h`, `TokenUsage7d`에 할당

**설계 고려사항**:
- TokenHistoryStore는 인터페이스이므로 테스트에서 mock 주입 가능
- cacheDir 경로 결정: stdin.Workspace.ProjectDir + "/.moai/cache" 우선

### Phase 4: renderer.go 확장

**작업 목록**:
1. `renderCompact()`, `renderDefault()` 등에 `token_usage_5h`, `token_usage_7d` 세그먼트 처리 추가
2. `formatTokenUsage()` 헬퍼 함수: 토큰 수 → "5h:12K" 형식 변환
3. 임계값 비교 후 ANSI 색상 적용 (`levelWarn` → 노란색, `levelError` → 빨간색)
4. `NO_COLOR` / `MOAI_NO_COLOR` 환경 변수 존중

**렌더링 우선순위 (기존 세그먼트 뒤에 추가)**:
```
model | context | output_style | directory | git_status | claude_version | moai_version | git_branch | [token_usage]
```

### Phase 5: 설정 및 템플릿

**작업 목록**:
1. `internal/template/templates/.moai/config/sections/statusline.yaml` 수정:
   - `segments`에 `token_usage_5h: true`, `token_usage_7d: true` 추가
   - `token_usage` 섹션 추가 (기본값 포함 주석)
2. `.moai/config/sections/statusline.yaml` (로컬 프로젝트) 동일하게 수정
3. `make build` 실행하여 `embedded.go` 재생성

### Phase 6: CLI 통합 (internal/cli/statusline.go)

**작업 목록**:
1. `statusline.yaml` 로드 시 `token_usage` 설정 파싱
2. `Options.TokenHistoryStore` 초기화 (cacheDir 전달)
3. `Options.SegmentConfig`에 새 세그먼트 키 반영

---

## 4. 파일 변경 요약

| 파일 | 변경 유형 | 설명 |
|------|-----------|------|
| `internal/statusline/token_history.go` | 신규 | 캐시 I/O + 집계 로직 |
| `internal/statusline/token_history_test.go` | 신규 | 단위 테스트 |
| `internal/statusline/types.go` | 수정 | TokenUsageWindow, StatusData 확장 |
| `internal/statusline/builder.go` | 수정 | TokenHistoryStore 통합 |
| `internal/statusline/renderer.go` | 수정 | 새 세그먼트 렌더 |
| `internal/cli/statusline.go` | 수정 | token_usage 설정 로드 |
| `internal/template/templates/.moai/config/sections/statusline.yaml` | 수정 | token_usage 섹션 추가 |
| `.moai/config/sections/statusline.yaml` | 수정 | 로컬 프로젝트 설정 동기화 |

---

## 5. 위험 요소 및 완화 방안

| 위험 | 가능성 | 영향 | 완화 방안 |
|------|--------|------|-----------|
| 캐시 파일 I/O로 인한 statusline 지연 | 중 | 중 | 50ms 타임아웃 적용, 읽기 실패 시 우아한 폴백 |
| session_id 미제공 시 집계 누락 | 저 | 저 | session_id 없는 세션은 캐시에 기록하지 않고 현재 세션만 UI에 미표시 |
| 캐시 파일 무한 증가 | 저 | 중 | 90일 Purge + 24h 쿨다운 |
| 병렬 moai 프로세스의 캐시 파일 충돌 | 저 | 중 | sync.RWMutex + 원자적 파일 교체 |
| 기존 statusline 렌더 성능 저하 | 저 | 고 | TokenHistoryStore를 goroutine으로 비동기 upsert (렌더 블록 최소화) |

---

## 6. 비고

- `moai statusline` 명령은 Claude Code가 매 응답마다 호출하므로 지연에 매우 민감
- `.moai/cache/` 디렉토리는 `.gitignore`에 추가 필요 (`moai init` 시 자동 처리)
- 향후 확장: 비용(USD) 기반 집계 세그먼트 (`cost_5h`, `cost_7d`) 추가 가능
