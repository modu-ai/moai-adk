# t280 sync-audit 판정 (SPEC-INBOX-DRAIN-GAP-001)

- Auditor: sync-auditor (독립 컨텍스트, lane-15 디스팟)
- Tree: `WT-inbox-drain-gap` @ `e7dc2627d`
- Date: 2026-09-03
- 이 문서는 감사자 판정을 lane이 기록한 것

## Verdict

**PASS — 0.96/1.00** (harmonic: Functionality 100 / Security 100 / Craft 90 / Consistency 95;
must-pass Functionality·Security 독립 통과). 결함 3건 전부 MINOR·optional, blocking 0.

## AC 재관측 (감사자 자체 측정, 10/10 GREEN)

| AC | 감사자 관측 |
|---|---|
| AC-IBX-001 (캡 회전) | ok 0.595s rc=0 |
| AC-IBX-002 (stand-down + byte-snapshot) | ok 0.498s rc=0 |
| AC-IBX-003 (retention ≤2세대) | ok 0.484s rc=0 |
| AC-IBX-004 (CLI status) | 2/2 PASS rc=0 |
| AC-IBX-005 (CLI drain 회전) | ok 0.821s rc=0 |
| AC-IBX-006 (curator 거절) | ok 0.808s rc=0 |
| AC-IBX-007 (스키마 v:1) | PASS ×2 rc=0 |
| AC-IBX-008 (fail-open) | ok 0.482s rc=0 |
| AC-IBX-009 (scope guard, 최종 병합 베이스 재생성) | grep rc=1 — 매치 0 |
| AC-IBX-010 (-race) | ok 1.740s, DATA RACE 0 |

엣지 4건(boundary / pre-era / empty-inbox / root-registration) 전부 PASS.
보조: golangci-lint 3패키지 0 issues · GOOS=windows vet+build rc=0 · 커밋 체인 10건 Conventional ·
mutant 잔재 0 · 트리 클린.

## NFC-4 딥다이브 — 통과

`enforceInboxCap`(inbox_lifecycle.go:152-167): stat(inbox) → 미달 즉시 반환 → 초과 시에만
`LSELDrainMarkerPresent` = `.moai/state/lsel` **단 1회 os.Stat**. MkdirAll·ReadDir·파일 읽기·
offset 접촉 0. 역방향 보호: AC-IBX-001이 marker-absent 설치에서 `.moai/state` 미생성 단언.
AC-IBX-002 byte-unchanged 단언은 실재(재귀 walk + sha256 + 공허 가드; m2-mutant6 RED 직독).
fan_in=3 직접 계수 일치(enforceInboxCap:160, inbox.go:91, inbox.go:110).

## 회전 역학 — 통과

모든 링크 delete-then-rename. pre-era 잔여 아카이브 흡수(시프트 단언 포함). 사보타주 시나리오에서
에러 → slog.Warn 후 반환, append는 기존 live에 착지(회전 부재 + stub 착지 양쪽 단언). swallow 없음.

## 공허성 — 회귀를 잡는다

세 방향 독립 포착: 회전 부재(AC-IBX-001) / stand-down 과잉(AC-IBX-002) / race 테스트 bounded 최종
크기. 무장·해제 양방향 테스트. 경계는 포함 규칙(`>=`→`>` 뮤턴트 RED로 이빨 확인). NFC-5는 처음부터
CI 소관으로 선언됨 — 공백 아님.

## 결함 (전부 MINOR·optional)

- **F1** CHANGELOG.md:12 "byte-identical to pre-SPEC" — stub 줄 기준 과대(신규 줄도 v:1 획득).
  기능 영향 없음(drain jq는 event_key만 검사). → lane이 마이크로 수리 지시(F1 한정 문구 교정)
- **F2** inbox_lifecycle.go:30 LessonsInboxPath godoc "단일 경로 생성" 주장 — collector는 inline
  filepath.Join 자체 생성(바이트 동일, 동작 문제 없음). → 기록 자문으로 잔존
- **F3** §E.4에 AC-IBX-009 sync 재생성 흔적 미기록 — 감사자가 재측정으로 대체 확보.
  → lane이 §E.4 추적 라인 추가 지시

## 리드용 잔여 위험

curator 머신에서 `.moai/state/lsel/` 삭제 시 캡이 무장돼 다음 세션 시작까지 미독 stub이 회전됨 —
노출은 캡 1세대(~4.2k 줄)로 유계, `session_drain.sh:85`가 재소유. 전이는 기록만 되고 가드되지 않음
(acceptance §B 명시). 세션 시작 배선이 깨진 머신에서는 이 창이 반복 열림.

## Gaps

CI windows 매트릭스(NFC-5 최종 판정)·전체 스위트는 origin/develop CI 미관측(레인 계약상 정상).

## Scope diff

131daa290..e7dc2627d 39파일: internal/hook · internal/cli · internal/config/defaults.go ·
SPEC 산출물 4 · CHANGELOG 1줄 · 증거 로그. **internal/cli/update/ 0건, 템플릿 트리 0건.**
