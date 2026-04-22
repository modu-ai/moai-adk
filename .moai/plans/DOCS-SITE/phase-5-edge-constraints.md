# Phase 5 — Vercel Edge Function Platform Constraints Review

SPEC-DOCS-SITE-001, Phase 5 (D-007 사전 검토 의무)
생성일: 2026-04-17

---

## 1. Vercel Edge Runtime 제약 요약

| 항목 | 제한값 | 비고 |
|------|--------|------|
| 최대 실행 시간 | 25ms (Hobby) / 50ms (Pro) | cold start 포함 |
| 메모리 | 128MB | 4MB JavaScript heap 제한 아님, 전체 128MB |
| 번들 크기 | 4MB | gzipped 기준 |
| Node.js API 지원 여부 | 불가 | V8 isolate 환경 |
| 사용 가능 API | Web Platform APIs 전용 | `URL`, `Headers`, `Request`, `Response`, `crypto` 등 |
| Cold start 추가 지연 | 50-100ms (첫 요청) | 이후 warm 상태 유지 |
| 배포 리전 | Vercel Edge Network (전역) | 별도 region 지정 불필요 |
| npm 패키지 호환성 | Edge 호환 패키지만 가능 | Node.js 전용 패키지 사용 불가 |

참조: https://vercel.com/docs/concepts/functions/edge-functions/limitations

---

## 2. `docs-site/api/i18n-detect.ts` 구현 분석

### 2.1 코드 크기

- 예상 번들 크기: < 10KB (4MB 제한 대비 0.25% 미만)
- 외부 의존성: 없음 (표준 Web API만 사용)

### 2.2 실행 시간 분석

| 처리 단계 | 예상 소요 시간 |
|-----------|---------------|
| URL 파싱 (`new URL`) | < 0.01ms |
| 경로 분석 (regex, string split) | < 0.1ms |
| 쿠키 파싱 (regex match) | < 0.1ms |
| Accept-Language 파싱 (string split, sort) | < 0.5ms |
| Response 생성 | < 0.1ms |
| 합계 (warm) | < 1ms |
| 합계 (cold start) | 50-100ms |

warm 상태 기준 총 처리 시간은 **< 1ms**로 50ms 제한 대비 충분한 여유를 가진다.

### 2.3 Node.js 전용 API 의존 여부

`middleware.ts` 원본(164 LOC)과 비교:

| 원본 API | Edge 호환 여부 | 이식 방법 |
|----------|---------------|-----------|
| `NextRequest` | Edge 호환 | 표준 `Request`로 대체 |
| `NextResponse.next()` | Edge 전용 아님 | `new Response(null, { status: 200 })`로 대체 |
| `NextResponse.redirect()` | Edge 전용 아님 | `new Response(null, { status: 307, headers: { Location } })`로 대체 |
| `request.cookies.get()` | Next.js 전용 | 직접 쿠키 헤더 파싱으로 대체 |
| `request.nextUrl` | Next.js 전용 | `new URL(req.url)`로 대체 |
| `path`, `fs`, `crypto` | 미사용 | 해당 없음 |

**결론**: Node.js 전용 API 의존 없음. 전체 로직이 Web Platform APIs만으로 구현 가능.

### 2.4 Edge 런타임 선언

파일 상단에 `export const config = { runtime: "edge" };` 명시. 이 선언이 없으면 Vercel은 기본적으로 Node.js Serverless Function으로 배포하므로 필수.

---

## 3. Option A (Edge Function + vercel.json rewrites) 채택 근거

### 검토한 옵션

**Option A: Edge Function + vercel.json rewrites**

```json
"rewrites": [{ "source": "/", "destination": "/api/i18n-detect" }]
```

- `/` 요청만 Edge Function으로 라우팅
- locale URL (e.g. `/ko/...`)은 Hugo 정적 HTML 직접 서빙
- 장점: 단순, 명확한 경계
- 단점: `/` 경로만 처리 (subdirectory 없는 unlocalized paths는 Hugo가 404 처리)

**Option B: Next.js Middleware (`middleware.ts`)**

- Hugo static 빌드와 충돌: Next.js 런타임 없이 `middleware.ts` 실행 불가
- Vercel은 `middleware.ts`를 Next.js 프레임워크로 인식 → Hugo 빌드와 공존 불가
- 폐기.

### 결정: Option A 채택

Hugo 정적 빌드에서 `/ko/`, `/en/`, `/ja/`, `/zh/` 경로는 직접 서빙된다. 루트 `/` 에 대한 첫 요청만 Edge Function이 처리하여 적절한 locale로 307 redirect. 이후 모든 페이지 탐색은 Hugo HTML 직접 접근.

---

## 4. Accept-Language 파싱 복잡도

- 알고리즘: `string.split(",")` → `string.split(";q=")` → `parseFloat` → `.sort` → 첫 매칭 탐색
- 시간 복잡도: O(n) (n = Accept-Language 헤더의 locale 항목 수, 일반적으로 3-8개)
- 공간 복잡도: O(n)
- 최악 케이스 헤더 길이: ~200 bytes

Accept-Language 파싱은 Cold start 포함 전체 실행 시간에 비해 무시할 수준.

---

## 5. Cold Start 영향 평가

- 첫 요청 추가 지연: ~50-100ms (Vercel 공식 자료 기준)
- 이후 warm 상태: < 1ms 처리
- 사용자 체감 영향: 최초 방문자 1회만 영향, 반복 방문자 무영향
- 완화 방법: Vercel Edge는 글로벌 배포로 cold start 발생 빈도가 낮음

---

## 6. 외부 의존성 검토

`i18n-detect.ts`는 외부 npm 패키지를 사용하지 않는다.

| 사용 API | 출처 |
|---------|------|
| `URL` | Web Platform (WHATWG) |
| `Request`, `Response` | Fetch API |
| `RegExp` | JavaScript 표준 |
| `String.prototype.split/match` | JavaScript 표준 |
| `Array.prototype.sort/filter` | JavaScript 표준 |

Edge 호환성 문제 없음.

---

## 7. 결론

`docs-site/api/i18n-detect.ts`는 Vercel Edge Function 플랫폼 제약을 모두 충족한다:

- 실행 시간: warm < 1ms (50ms 제한 대비 충분한 여유)
- 메모리: < 1MB (128MB 제한 대비)
- 번들 크기: < 10KB (4MB 제한 대비)
- Node.js API 의존: 없음
- Edge 런타임 선언: `export const config = { runtime: "edge" }` 명시

Phase 5 진행에 플랫폼 제약상 blocking issue 없음.
