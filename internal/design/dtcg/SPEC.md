# DTCG 2025.10 Validator — 스냅샷 참조

## 버전 고정

| 항목 | 값 |
|------|----|
| 사양 이름 | W3C Design Tokens Community Group (DTCG) Format |
| 스냅샷 날짜 | 2025년 10월 (Editor's Draft 2025-10) |
| 소스 URL | https://tr.designtokens.org/format/ |
| 구현 버전 | SPEC-V3R3-DESIGN-PIPELINE-001 Phase 3 (Wave C.3) |
| 구현 날짜 | 2026-04-27 |

> [HARD] 이 파일은 버전 업그레이드 추적을 위해 존재합니다. 검증기를 새 버전으로
> 업그레이드할 때는 반드시 이 파일의 스냅샷 날짜와 소스 URL을 갱신해야 합니다.

## 지원 카테고리 (DTCG 2025.10 §8 기준)

| 카테고리 | 설명 | 구현 파일 |
|---------|------|---------|
| `color` | sRGB 색상 (hex, rgb(), hsl(), named) | `categories/color.go` |
| `dimension` | 크기 단위 (px, rem, em, %) | `categories/dimension.go` |
| `fontFamily` | 폰트 패밀리 목록 | `categories/font_family.go` |
| `fontWeight` | 폰트 두께 (수치 또는 named) | `categories/font_weight.go` |
| `font` | 복합 폰트 토큰 | `categories/font.go` |
| `typography` | 복합 타이포그래피 토큰 | `categories/typography.go` |
| `duration` | 시간 단위 (ms, s) | `categories/duration.go` |
| `cubicBezier` | 베지어 곡선 [x1,y1,x2,y2] | `categories/cubic_bezier.go` |
| `number` | 숫자 원시값 | `categories/number.go` |
| `strokeStyle` | 선 스타일 (enum 또는 복합) | `categories/stroke_style.go` |
| `border` | 복합 테두리 토큰 | `categories/border.go` |
| `transition` | 복합 전환 토큰 | `categories/transition.go` |
| `shadow` | 복합 그림자 토큰 (단일 또는 다층) | `categories/shadow.go` |
| `gradient` | 그라디언트 배열 | `categories/gradient.go` |

## 에일리어스 문법 (DTCG 2025.10 §7)

```json
{ "$value": "{group.subgroup.token-name}" }
```

- 중괄호 표기법으로 다른 토큰을 참조합니다.
- 순환 참조(cyclic alias)는 오류로 처리됩니다.
- 에일리어스는 동일하거나 호환 가능한 `$type`을 가진 토큰만 참조해야 합니다.

## 검증 규칙 (DTCG 2025.10 §9 기반)

1. `$value` MUST 존재하고 타입 적합해야 함
2. `$type`은 등록된 카테고리 중 하나여야 함
3. 에일리어스는 호환 가능한 `$type` 토큰으로 해결돼야 함
4. 복합 타입의 중첩 필드는 값 또는 호환 타입 에일리어스여야 함
5. 순환 에일리어스는 반드시 거부돼야 함

## 향후 버전 업그레이드 절차

1. https://tr.designtokens.org/format/ 에서 새 스냅샷 날짜 확인
2. 이 파일의 스냅샷 날짜 및 소스 URL 갱신
3. 새로 추가된 카테고리에 대응하는 `categories/*.go` 파일 추가
4. `validator.go`의 `categoryValidators` 맵 업데이트
5. `go test -race ./internal/design/dtcg/...` 전체 통과 확인
