# t838 런북 — 라이브 receipt 400(CauseChain) 1회 계측 절차

카드: t838 · 기반: develop `d416f8162` · 작성: lane-1 (2026-09-14)
실행 주체: **운영자**(GOOS). 이 런북은 레인이 실행하지 않는다.

## 0. 대상과 목적

t707 판정문( `.moai/reports/t707/verdict.md` )의 갭 1: 라이브 CauseChain 400은 "Edit 툴 직후 첫 요청" 공통 패턴을 가지지만 요청 본문 기록이 없어 최종 국소화가 불가했다. 다음 라이브 발생 시 아래 패치를 적용한 바이너리로 1회 재현하면, 게이트웨이 stderr에 **최초로 어긋난 경계의 구조 사실**이 1행으로 찍힌다.

- 패치: `.moai/reports/t838/observability.patch` (v2 — t707의 v1은 26행에서 corrupt + `fmt`/`os` import 누락. v2는 `d416f8162`에서 `git apply --check` 통과·빌드·vet·translate 패키지 테스트 통과를 이 세션에서 확인)
- 출력 형식 (실측 예, 중간 경계 절단 케이스):

```
moai-receipt-divergence: observations=2 candidates=3 first_unmatched_boundary=1 root_unmatched=false observations_items_first=1
```

| 필드 | 의미 |
|---|---|
| `observations` | 재생된 assistant 경계 수 |
| `candidates` | 매니페스트가 가진 수신 후보 수 |
| `first_unmatched_boundary` | 처음으로 어긋난 경계 인덱스(-1이면 전부 일치) |
| `root_unmatched` | 어긋난 것이 0번 경계(뿌리)인지 — 뿌리부터 어긋나면 재생 이력 전체가 다른 바이트 |
| `observations_items_first` | 첫 미일치 경계의 아이템 수 |

메시지 내용·풀 digest·세션 식별자는 출력에 없다(구조 사실만).

## 1. 사전 준비

1. git 추적 수정이 0인 상태에서 시작한다(`git status --porcelain` 확인 — ` M CLAUDE.local.md` 한 건은 §0.4 기준 영구 표식).
2. 핫픽스 작업은 **커밋 없는 워킹 트리**에서 한다. 패치를 커밋에 남기지 않는다(패치 헤더 규약).

## 2. 패치 적용 + 바이너리 교체

```bash
cd /Users/goos/MoAI/moai-adk-go
git apply .moai/reports/t838/observability.patch
go build ./internal/gateway/...          # 적용 확인(통과해야 진행)
make build && make install               # LDFLAGS 포함 설치 — 맨손 go install 금지
~/go/bin/moai version; echo $?           # exit 0 확인(137이면 rm -f ~/go/bin/moai 후 make install 재시도)
```

## 3. 재현 (gpt 세션, Edit 직후 첫 요청)

t707 판정의 4 고유 사례 공통 패턴: **Edit 툴 실행 직후 ~80-90ms 내 첫 요청**에서 400.

1. `MOAI_RECEIPT_DEBUG=1` 를 붙여 gpt 모델 세션을 연다(게이트웨이를 호스팅하는 moai 프로세스의 stderr로 디버그 행이 나온다 — 세션 터미널에서 그대로 보인다).
2. 세션에서 파일을 Edit 한 뒤 즉시 다음 요청을 유도한다(도구 종료 직후 자동 요청이면 더 좋다).
3. 400(`HistoryReplayError` 안내문)이 뜨는 순간 stderr의 `moai-receipt-divergence:` 1행을 그대로 복사해 둔다. 여러 번 재현되면 각각의 행을 모두 모은다.

재현이 잘 안 붙으면: resume 직후·포크(agent_summary) 직후 요청도 t707 분류표에 있으므로 함께 시도한다.

## 4. 포착 후 원복

```bash
cd /Users/goos/MoAI/moai-adk-go
git restore internal/gateway/translate/receipt_history.go
make build && make install               # 비계측 바이너리로 복원
~/go/bin/moai version; echo $?           # exit 0 확인
git status --porcelain                   # 추적 수정 0 복귀 확인
```

## 5. 증거 기록

- 포착한 `moai-receipt-divergence:` 행(들)과 세션·타임스탬프를 `.moai/reports/t707/` 아래 새 케이스 파일로 남긴다(예: `live-divergence-<날짜>.md`).
- 판독 가이드: `root_unmatched=true`면 재생 이력의 뿌리부터 다른 것(클라이언트가 이력 배열을 다시 만든 것) — t707 갭 3의 후보(file-history-delta/토큰 리마인더 삽입) 확인 대상. `first_unmatched_boundary>0`이면 특정 경계부터만 다른 것 — 해당 위치의 봉투/마커 재인코딩 suspect.

## 6. 금지 사항

- 패치가 적용된 상태의 커밋·push 금지(원복이 절차의 일부다).
- 디버그 행 자체를 외부(GitHub 이슈 등)에 올리지 않는다 — 구조 사실만이지만 세션 크기·경계 수가 노출된다.
- `boundaryItems` 등 패치 전용 헬퍼를 다른 용도로 재사용하지 않는다.

## 근거 (이 세션에서 측정)

- `git apply --check` 통과 + 적용 상태에서 `go build ./internal/gateway/...`, `go vet ./internal/gateway/translate/`, `go test -count=1 ./internal/gateway/translate/` → ok 1.849s
- 발화 관측: 적용 상태에서 `MOAI_RECEIPT_DEBUG=1 go test -run TestReceiptHistoryRejectsMidChainTruncationWithGuidance` → `moai-receipt-divergence: observations=2 candidates=3 first_unmatched_boundary=1 root_unmatched=false observations_items_first=1` 출력 확인 후 원복
- 미검증(Gaps): 라이브 400 재현 자체는 운영자 실행 몫 — 이 런북은 준비물과 절차만 검증했다.
