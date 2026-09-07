# t239 sync-audit 판정 (SPEC-LLMCFG-PRESERVE-001)

- Auditor: sync-auditor (독립 컨텍스트, lane-15 디스팟)
- Tree: `WT-llm-yaml-preserve` @ `734602707`
- Date: 2026-09-02
- 이 문서는 감사자의 메시지 판정을 lane이 기록한 것 — 원문 전문은 감사자 반환값 기준

## Verdict

**PASS — 0.97/1.00** (harmonic: Functionality 100 / Security 100 / Craft 95 / Consistency 95;
must-pass Functionality·Security 각각 독립 통과). 결함 5건 전부 MINOR·optional.

## AC 재관측 (감사자 자체 측정, 6/6 일치)

| AC | 감사자 명령 | 관측 | 일치 |
|---|---|---|---|
| AC-LCP-001 | `go test -run TestUpdateLLMYAMLPreserveTemplateSync -count=1 ./internal/cli/` | ok 1.079s exit 0 | y |
| AC-LCP-002 | `go test -run TestUpdateLLMYAMLNewKeyDelivery …` | ok 0.994s exit 0 | y |
| AC-LCP-003 | `go test -run TestUpdateLLMYAMLFirstDeployCalm …` | ok 1.064s exit 0 | y |
| AC-LCP-004 | `go test -run TestUpdateLLMYAMLCommentsSurvive …` | ok 1.031s exit 0 | y |
| AC-LCP-005 | `go test -run TestCleanReinstallLLMYAMLPreserved …` | ok 0.811s exit 0 | y |
| AC-LCP-006 | grep 스캔 2종 | update scope 0 / hook tier 2파일만 | y |

셀렉터 0매치 경계: `-v` 재실행으로 5개 테스트 `=== RUN`+`--- PASS` 실측.
§C 게이트: backup `ok 0.695s` / merge `ok 0.912s`, `go vet` exit 0,
`golangci-lint run ./internal/cli/...` → `0 issues`, `GOOS=windows go vet` exit 0.

## Teeth 감사 — 진짜 판정

- mutant 로그 4개의 FAIL 라인이 테스트 소스와 줄번호·메시지까지 일치
  (`update_llm_preserve_test.go:228/:230`, `:272`, `update_clean_install_config_preserve_test.go:238/:240`)
- 변위 지점 실재 확인: `update_template_sync.go` "Restore Settings"의
  `if configBackupPath != ""` 게이트, `update_clean_install.go:481` 부근 `RestoreMoaiConfig` 호출
- 머지레벨 프로브는 실존 함수로 민감도 계산(`DeepMerge3Way` merge.go:85, `MergeYAML3Way` merge.go:26)
- 스크래치 미커밋 확인: porcelain 클린, zz_scratch 트래킹 0

## Vacuity — 회귀를 잡는다 (판정문 요지)

픽스처가 진짜 임베디드 템플릿 바이트(`template.EmbeddedTemplates`)에서 파생되고,
must-replace 시맨틱이 템플릿 드리프트를 픽스처 빌더 실패로 승격시키며,
갱신은 프로덕션 진입점 `runTemplateSyncWithReporter`로 전체 주기를 돈다.
어설션은 파싱 키값+코멘트 존재 단위고 byte-equality는 AC-LCP-003 단 한 곳.

## 결함 (전부 MINOR·optional)

- **D1** acceptance.md AC-LCP-001 mutant 셀 "all three assertions" — 실제 로그는 2어설션 후 t.Fatalf. 문구 수리만.
- **D2** CHANGELOG "each observed RED" — AC-001/003/005는 e2e 명령 RED, AC-002/004는 머지레벨 프로브. 조건절 압축.
- **D3** plan.md:27-28 경로 없는 인용 — cosmetic (인지된 advisory).
- **D4** FirstDeployCalm의 `!strings.Contains(out, "llm.yaml")` — 미래 정당 배포 로그 오타 가능. optional 경비 강화.
- **D5** AC-LCP-005가 overwritingDeployer 더블로 clobber 재현(가족 패턴과 일치하는 선택) — 실 force-deploy 정지 시 이 테스트는 모름.

## 리드 플래그 잔여 위험

머지 엔진 측정 동작상 **템플릿이 가진 키에 단 사용자 코멘트는 갱신 시 템플릿 코멘트로 교체**된다
(progress.md 설계 관찰에 기록됨 — 마커를 사용자추가 키에 얹는 것이 회피책).
이 SPEC이 핀한 계약은 "사용자 값 편집 보존"이며 "템플릿 키 위 사용자 코멘트 손실"은 미핀 잔여 —
후속 카드 후보.

## Scope diff

b7462203a..HEAD = 테스트 2파일 + SPEC 아티팩트 4 + CHANGELOG +1불릿 + 증거 로그 17.
§A.5 PRESERVE 경로 수정 0, template .yaml 편집 0, t280·t230 크리프 0.
브랜치 선두 t255 판정 커밋 `874b2d20a`는 lane-15 선행 카드 산출물(의도된 증거 커밋).
