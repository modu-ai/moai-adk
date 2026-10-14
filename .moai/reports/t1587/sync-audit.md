# sync-audit — card t1587 (lane-7, 2026-10-08)

verdict: PASS
audited_sha: 423508cd07f07ccdec36fcc93443cd3177db91bc

(감사 실행 시점의 코드 팁은 `e588f9882e2dbd239fe94bba36112adaaac8d940` — 기록 증거 커밋 423508cd0과의 델타는 본 파일·card-review.md 증거 2파일뿐이다. 코드 diff는 e588f9882 기준으로 감사됐다.)

- owner: sync-audit-4dim workflow (binding owner — PASS·dim 0 없음·INCOMPLETE 아님), run wf_566dbfa8-a06 재개 실행(1차는 구독 429로 INCOMPLETE — 캐시 재개로 완주)
- **verdict: PASS** — harmonic mean **0.9016** (threshold 0.85, tier M)
- scores: Functionality 0.90 · Security 0.88 · Craft 0.88 · Consistency 0.95
- 발견: minor 4건 이상 — **전부 카드 diff 밖 파일**(internal/constitution/validator.go·zone-registry.md·registry_sync_test.go·internal_content_leak_test.go 등). 카드 커밋(`db6d875f6`·`e588f9882`: internal/cli/init.go·init_user_asset_retry_test.go·.github/test-input-filters.yml·test_input_filter_parity_test.go)에 대한 발견 **0건**.

## 범위 주의(스코프 케이브앳) — 판독자 필독

- 워크플로우 Context 단계가 AC 표면을 **SPEC-ZONE-REGISTRY-RESYNC-001**(트리 안 존재 SPEC)로 고정했고 감사 diff 창(`1ae6e5c36..HEAD`)이 카드보다 넓었다 — 본 카드는 SPEC 없는 Class B라 AC 매핑은 부분 소음.
- 그럼에도 4차원 판독 자체는 카드 커밋을 포함한 트리 실측(grep·go test·diff 판독)이며, 카드 diff 파일에 대한 지적은 0건 — 카드 diff 품질의 1차 근거는 codex card-review(diff 발견 0건) + RED→GREEN 패밀리 재측정(§Run-phase Evidence)이 대신한다.
- Functionality minor 1건은 `go test ./internal/template/` 기본 10m 타임아웃 flaky(751.8s, -timeout 25m에서 ok) — 카드 변경과 무관한 기존 패키지 특성, CI 예산 25m로 흡수됨.

decision record: decided_by=sync-audit-4dim(바인딩 소유자) evidence_refs=워크플로우 완료 통지 전문(PASS 0.9016·scores·findings)·journal.jsonl 5결과 ladder_path=T13 guardVerdictPass 충족(범위 케이브앳 병기)
