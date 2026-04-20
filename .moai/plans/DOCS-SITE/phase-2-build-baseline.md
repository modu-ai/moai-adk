# Phase 2 — Hugo 빌드 시간 베이스라인 기록

> 측정일: 2026-04-20
> 목적: AC-G1.5-05, AC-G2-11 (Gap 3), D-017 (Vercel timeout 여유 계산)
> 환경: Hugo v0.160.1+extended darwin/arm64 (Homebrew), Apple Silicon

## 측정 결과

### Phase 2 완료 시점 (빈 콘텐츠 상태)

```
hugo --minify --gc
Total in 87 ms
real: 0.11s user / 0.07s system / 158% cpu / 0.110s total
```

- 페이지 수: KO 9 / EN 7 / JA 7 / ZH 7 (Hugo 빌드 페이지 + Hextra 생성 페이지)
- 콘텐츠 소스: 없음 (content/ 디렉토리 미존재)

### Vercel 빌드 timeout 여유 계산

| 항목 | 값 |
|------|-----|
| Vercel 기본 빌드 timeout | 10분 (600초) |
| Phase 2 빌드 시간 (빈 콘텐츠) | 0.11초 |
| 예상 Phase 3 후 빌드 시간 (219페이지 추정) | 5~15초 (Hugo 는 페이지당 약 50~100ms 추가) |
| 안전 여유 | >590초 (약 98% 여유) |

Hugo 정적 사이트 생성기의 빌드 속도는 Node.js 기반 Nextra 대비 10~100배 빠르므로 Vercel timeout 문제 발생 위험이 없다고 판단됨.

## WARNINGs 기록

Hugo v0.160.1에서 Hextra v0.12.2 빌드 시 다음 deprecated API 경고 발생:

```
WARN  deprecated: .Site.Data was deprecated in Hugo v0.156.0
WARN  deprecated: .Site.Languages was deprecated in Hugo v0.156.0
```

- 원인: Hextra 내부 템플릿이 Hugo v0.156 이전 API를 사용
- 영향: 빌드 성공에 영향 없음, 향후 Hextra 업데이트 시 자동 해결 예정
- Phase 2 G1.5 통과 기준: 빌드 성공 (exit 0) — 충족

## Phase 3 이후 재측정 예정

219페이지 변환 완료 후 `hugo --minify --gc` 재실행하여 이 문서에 결과 추가 예정.
