# t1968 Handoff — Claude lane-7 → GLM lane (2026-10-28)

## 카드

t1968 — SPEC-FETCH-WEB-001 (jevfetch port: `moai fetch web` + MCP `fetch_web`, internal/jev client hardening, jevmeasure bench, template packaging). Branch `WT-jevfetch-a2sys-ai`, worktree `.moai/worktrees/t1968`. PR #1940 (remote head **41fe0c5396** = b4622f2434a0bf73..41fe0c53963ea23 pushed + verified; merge = Cloud, leader merges).

## HEAD (작성 시점)

**0fc93ca68696e63919c3753fcc7a7b96ec2eb282** — F2 RED 커밋 직후. 로컬 전용 미푸시 커밋 다수(푸시는 리더 승인 후 1회 원칙).

## 미커밋 파일 (보존 — 건드리지 말 것)

- `internal/fetchweb/scoring.go` — **F2 수리 편집 중**(실행 중인 manager-develop 에이전트 소유; 강제 종료 금지. 에이전트 완료 시 GREEN 커밋 착지 예정: `neighborGroupKey` 경로 길이 포함, `[]`/`[""]` 구분)

## 현재 단계

착지(sync) 재감사 사이클 — **F1+F2 수리 라운드 진행 중**(운영자 승인):

- F1 완료: RED `494da7dfb2` → GREEN `bc714668e3` (NAT64 `64:ff9b::/96` 임베드 IPv4에 기존 refuseNonPublicIP 정책 적용 — 공개 임베드 허용·루프백/사설/CGNAT 임베드 거부)
- F2 진행: RED `0fc93ca686` 착지(`<h2></h2>` 재현 — `[]`/`[""]` 충돌), GREEN 편집 중
- 남은 게이트: fetchweb+jevmeasure 수트·빌드·WIN빌드·gofmt·lint (에이전트가 slot acquire 성공 확인·부하≤48·timeout 규율로 수행)

## 승인 라운드와 범위 (원문 보존)

1. **7건 수리 라운드** — 운영자: "7건 수리 라운드 승인" (범위: bench.go:134·blocks.go:659·scoring.go:574·scoring.go:408·보안 scoring.go:214·ssrf.go:52·output.go:25 P3 — 전부 착지 완료, RED/GREEN 15커밋)
2. **F1·F2 수리 라운드** — 승인 문구 원문: **"리더 세션 AskUserQuestion 응답(주체 미확인) 'F1·F2 수리 라운드 승인'"** (범위: F1·F2 두 건만 — F3·추가 항목 없음)
3. **범위 규칙 (리더 교정, 준수 필수)**: 승인 목록 밖 수리는 **착수 전 리더 사전 통보** (8번째 항목이 범위 초과로 운영자 보고됨)

## 완료 커밋/판정/receipt

- iter14 **PASS(무부채) 0.95** rcpt-8babf16e21e731df3fe546dd @8ce65a38cd (수렴 ok)
- iter15 **FAIL 0.87** rcpt-ed14d689179efd4417f4d4ae @4c4edeccaa — 수리 8+1건 전부 실재·무회귀 확인, 델타 도입 차단 회귀 2건(F1 NAT64 공개 임베드 과차단·F2 neighborKey 충돌) + F3[P3 선택]
- 리더 delta audit PASS: rcpt-101eaacb0ed64b07e49854f4(b4622f2434)·rcpt-c30c01fb2fcf3b5e88b348c0(d5601a9262)
- 판정 사슬: sync-audit.md iter1~15 append-only · card-review.md 8라운드 · SPEC progress.md §E/§5/decision records

## 다음 한 단계 (명령 형태)

1. F1+F2 에이전트 완료 대기 → `git log --oneline 6f15ba4200..HEAD` + `go test -count=1 ./internal/fetchweb/ ./internal/jevmeasure/` (slot·timeout·부하≤48) 레인 검증
2. card-review: `mcp__moai__codex_review {scope:"card", project_root:<본 워크트리>}` → `.moai/reports/t1968/card-review.md` append
3. iter16 독립 감사: sync-auditor 스폰 (delta **41fe0c5396..HEAD**, iter15 형식 그대로) → `.moai/reports/t1968/sync-audit.md` append
4. iter16 판정 수신 → 리더 보고(socket) → 리더 audit_multi → push 승인 대기 → PR #1940 `@codex review`·`@codex security review` 재요청 → Cloud 관측 병합 → complete(merged-pr)

## 멈춤 조건

- iter16 **FAIL이면 즉시 정지** + 리더 보고 (STOP-on-FAIL — iter15에서 발동된 바 있음)
- 승인 목록 밖 수리 금지(사전 통보 의무) · 전체 internal/cli 수트 로컬 금지(CI 소관, Gap 기록 유지)
- push는 리더 승인 후 1회(승인 전 기록 커밋 먼저 마무리하고 그 HEAD로 요청 — 리더 지시) · 로컬 main 병합 금지 · complete는 관측 병합 후 merged-pr 경로만
- 스펨/강제 종료 금지: 실행 중 위임은 종료 시까지 대기 후 기록

## 임대/크론 상태

- 카드 임대: **not-holder**(pr-open 상태 안정 — 관측 병합까지 임대 부여 없음; `moai factory renew t1968 --run tnbn7e` 매 통과 시 refused-not-holder 예상)
- 왓치독 크론: **유지**(Claude 세션 소유 — GLM 세션이 재설정·인수; 크론 삭제 금지)
- go-test 슬롯: **미보유**(필요 시 `moai slot acquire --resource go-test --max-duration 15m` 후 부하 ≤48 확인, acquire 종료 코드 확인 선행 — 선행 라운드에서 redirect로 출력 유실된 사례 있음)
- ENOSPC 보류 유지: /System/Volumes/Data 97% — 삭제·정리 금지, 신규 대량 쓰기 금지
- 메시징: `factory messaging degraded`(수신함 타임아웃) 간헐 관측 — 큐·디스크 증거 경로는 정상

## 주의 (GLM 세션용)

- 판정 파일은 append-only(sync-audit.md) — 기존 iter 블록 수정 금지
- `.moai/reports/*` gitignored — 증거 커밋은 `git add -f` + 명시 pathspec
- 워크트리 가드: 복합 git 명령·`$(...)` 치환 거부 — plain git 단일 명령만
- GLM 사용량 한도 429 사례 다수 — 리셋 시각 확인 후 스폰 (2026-10-27 15:47:38 KST 사례 참조)
- 스폰 모델/effort: 판정 역할 opus/high 원칙 — GLM 세션 판정은 감사자 모델 표기 유지(auditor-model 필드)

## 갱신 (2026-10-29 01:2x — F1+F2 라운드 완료 후)

- **F1+F2 수리 라운드 완료**: F1 RED `494da7dfb2`→GREEN `bc714668e3`, F2 RED `0fc93ca686`→GREEN `32dedccb58`. 게이트 전부 green(3패키지 ok·빌드·WIN빌드·gofmt·lint 0 issues — iter15 대비 신규 0). 범위 4파일 +171−34 정확. 기록 커밋 `530ae989f6`(progress.md).
- **"다음 한 단계" 갱신**: 1번(에이전트 완료 대기) 소진 — GLM 세션은 2번(card-review)부터: `mcp__moai__codex_review {scope:"card", project_root:<본 워크트리>}` → iter16 sync-auditor 스폰(delta **41fe0c5396..HEAD**, 현재 tip `530ae989f6`) → 리더 보고 → audit_multi.
- 멈춤 조건·범위 규칙·임대/크론 상태 불변.
