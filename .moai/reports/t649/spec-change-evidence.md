# t649 SPEC 변경 근거

## Claim

SPEC-MOAI-GATEWAY-001 0.10.0은 제공자별 모델 선택·요청 강제·상태 격리와 실제 제품 검증을 현재 계약으로 삼는다.
기존 번호를 재사용하지 않고 REQ 24 활성/2 묘비, AC 24 활성/1 묘비를 유지했다.
변경 소유 파일은 spec.md, plan.md, acceptance.md, design.md 네 개다. progress.md와 구현 파일은 수정하지 않았다.

## Evidence

명령: `go run ./cmd/moai spec lint SPEC-MOAI-GATEWAY-001`

```text
✓ No findings — all SPEC documents are valid
```

명령: `git rev-parse --short HEAD` 및 `git branch --show-current`

```text
81c1d58f9
WT-unified-gateway
```

명령: Python 정규식으로 본문 굵은 REQ/AC 기본 ID의 집합과 `[RETIRED]` 집합의 차집합을 계산했다.

```text
spec.md: unique IDs=26 active=24 retired=2
acceptance.md: unique IDs=25 active=24 retired=1
```

## Baseline-attribution

2026-09-12, `.claude/worktrees/moai-proxy-unified`의 현 작업 트리. HEAD 위 미커밋 선행 작업을 포함한다.
source_session_id: 01a08e7b-6aa0-7361-ab7e-ea8da1f02228. 카드: t649.
설치본이 아닌 현 소스 `go run`의 lint 출력을 사용했다. 현재 변경을 완료·배포로 표시하지 않았다.

## Gaps

이 검사는 문서 구조 검사이며 독립 의미 감사나 실행 결과가 아니다. 실제 MoAI 세 제공자 picker,
GPT-6 Astra 응답·tool 후속·resume, family/credential 귀속 음성, 공유 설정 격리 및 Windows CI는
구현 담당과 오케스트레이터가 실제 증거로 판정해야 한다. 로그인 상태나 직접 Claude UI 프로브로 대체할 수 없다.

## Residual-risk

과거 HISTORY·결정 기록은 보존하여 과거 PICKER 이관 문구가 남는다. 현재 계약의 회수 범위는
0.10.0 HISTORY, REQ-MG-019, plan/design 앞머리와 t649 AC에 명시했다.
reasoning 완료 item 확보 뒤 SSE를 송신하는 경로는 본문 표시가 늦을 수 있으며 실제 지연 수치는 아직 없다.
독립 감사·권한·유실 음성 게이트는 여전히 유효하다. 승인되지 않은 push·PR·병합·워크트리 제거를 수행하지 않았다.
