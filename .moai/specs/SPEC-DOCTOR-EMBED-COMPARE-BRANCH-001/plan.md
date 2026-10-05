# plan.md — SPEC-DOCTOR-EMBED-COMPARE-BRANCH-001

## §A Context

- **카드**: t356. **워크트리**: `.claude/worktrees/t356`, 브랜치 `WT-embed-compare-branch`,
  HEAD `c6aa61346` (= origin/develop).
- **Tier**: S — 단일 마일스톤, 테스트 함수 1개 추가, 프로덕션 diff 0.
- **산출물**: `.moai/specs/SPEC-DOCTOR-EMBED-COMPARE-BRANCH-001/{spec,plan,acceptance,progress}.md`
- **결함 한 줄**: `doctor_agentemit_embed.go:146`(comparison-failure) 분기에 전용 비회귀 테스트가 없는데
  CHANGELOG(t346 `c2b51293e`)는 네 분기 전부의 보존을 약속한다.
- **PRESERVE**: `internal/cli/doctor_agentemit_embed_test.go` 이외 **모든 파일**. 특히
  `doctor_agentemit_embed.go`(변형은 mutant 관측 중에만, 반드시 원복), `doctor.go`, `uikit/**`,
  `internal/template/**`, `.github/workflows/**`, `CHANGELOG.md`.

## §B Known Issues

- **B-분기식별**: 형제 세 분기가 모두 `CheckFail`을 내므로 상태 단언만으로는 회귀를 못 잡는다.
  메시지 접두(`comparison failed`)까지 단언해야 가드가 성립한다 (AC-DECB-002).
- **B-픽스처순서**: `compareEmission`은 추출 대응물을 **먼저** 읽는다. 대응물이 없으면 err가 아니라
  `uncompared`로 흘러 기수 부족(:155)이 걸린다 — 픽스처는 추출 대응물을 정상 파일로 반드시 공급해야 한다.
- **B-Glob성질**: 디렉터리를 매치하는 `filepath.Glob`은 여기서 **지렛대**이지 수리 대상이 아니다.
  이 성질이 훗날 수리되면 본 테스트의 픽스처 구성법도 함께 바뀌어야 한다 — 테스트 주석에 이 결합을 명시한다.
- **B-증거소재**: 워크트리의 `.moai/reports/`는 gitignored이며 워크트리 폐기 시 유실된다
  (기존 관측 사례 다수). 판정 증거는 primary 체크아웃에 쓴다.

## §C Pre-flight

```bash
git -C <worktree> branch --show-current                  # WT-embed-compare-branch
git -C <worktree> rev-parse --short HEAD                 # c6aa61346 (착수 시 재확인)
go vet ./internal/cli/                                   # baseline
go test ./internal/cli/ -run 'TestAgentEmitEmbed' -count=1   # 기존 임베드 테스트 baseline GREEN
```

## §D Constraints

- 프로덕션 코드 영구 변경 금지. mutant는 **관측 목적의 일시 변형**이며, RED 관측 직후 원복하고
  `git diff --stat -- internal/cli/doctor_agentemit_embed.go`가 빈 출력임을 확인한 뒤 커밋한다.
  mutant 상태의 트리는 절대 커밋·푸시하지 않는다.
- 신규 헬퍼 도입 금지 — 기존 4종 픽스처 재사용 (REQ-DECB-006).
- 로컬 전체 스위트 금지. 판정면은 `./internal/cli/...` + 원격 CI.
- 커밋: Conventional Commits, 제목에 카드 id `t356` 병기. `--no-verify`/`--amend`/force-push 금지.
- 커밋 직전 `git rev-parse --short HEAD` + `git branch --show-current` 재판독 (AGENTS.md §2).
- 스테이징은 명시 pathspec만. sweep(`git add -A`/`.`) 금지.

## §E Self-Verification

| # | 항목 | 명령 |
|---|---|---|
| E1 | AC 매트릭스 | acceptance.md AC-DECB-001..007 개별 판정 |
| E2 | mutant RED | `go test ./internal/cli/ -run '<새 테스트>' -count=1 -v` (mutant 상태) |
| E3 | 원복 GREEN | 동일 명령 (원복 상태) + `git diff --stat -- internal/cli/doctor_agentemit_embed.go` 빈 출력 |
| E4 | 패키지 비회귀 | `go test ./internal/cli/... -count=1` |
| E5 | lint | `go vet ./internal/cli/` |
| E6 | 증거 소재 | `ls /Users/goos/MoAI/moai-adk-go/.moai/reports/t356/verdict.md` |

## §F Milestones

### M1 — comparison-failure 분기 비회귀 테스트 (단일 마일스톤)

가장 뒤집히기 쉬운 결정을 먼저 둔다.

1. **[결정] 분기 식별 단언의 형태** — 메시지 접두 `comparison failed` 포함 **and** 형제 3접두
   (`could not extract` / `compared ` / `embeds stale`) 비포함. 이것이 "상태만 보는 단언은 불충분"이라는
   운영자 HARD 조건의 구체화다. 대안(에러 타입 단언)은 `check`가 문자열 메시지만 노출하므로 불가.
2. **[결정] 픽스처 구성법** — `newEmbedFixtureRoot(t, "manager-git.toml")`로 커밋 세트를 만든 뒤,
   그 `.toml` **파일을 지우고 같은 경로에 디렉터리를 생성**한다(`os.Remove` → `os.MkdirAll`).
   `writeFakeBinary(t, root)`로 판정 대상 존재를 만족시키고,
   `newExtractedDir(t, map[string]string{"manager-git.toml": …})` + `staticExtractor`로 정상 파일
   대응물을 공급한다. 이 순서가 `:146`을 겨냥하는 유일한 구성이다(§B-픽스처순서).
3. **[기계] 테스트 함수 작성** — `TestAgentEmitEmbed_ComparisonErrorFails` (형제 명명 관례 준수),
   `doctor_agentemit_embed_test.go`의 `TestAgentEmitEmbed_ExtractionErrorFails` 직후에 배치.
   주석에 (a) 겨냥 분기(:146), (b) `compareEmission`의 유일 err 경로, (c) Glob-디렉터리 결합(§B-Glob성질)을 명시.
4. **[기계] mutant RED 확립** — `:146`의 `CheckFail` → `CheckOK` 일시 변형 → E2 관측 → 원복 → E3 관측.
5. **[기계] 증거 기록** — E2/E3의 축자 명령·출력을 primary의 `.moai/reports/t356/verdict.md`에 기록.
6. **[기계] 커밋** — `test(SPEC-DOCTOR-EMBED-COMPARE-BRANCH-001): guard the comparison-failure branch (t356)`.

## §G Anti-Patterns

- **상태만 단언**: `if c.Status != uikit.CheckFail` 만 두고 끝내기 — 형제 세 분기가 같은 상태를 내므로
  이 가드는 :146이 다른 분기로 대체돼도 초록이다(공허한 초록).
- **mutant 없이 완료 선언**: "테스트를 추가했다"는 완료가 아니다. 분기가 **실패하는 것을 관측**해야 한다.
- **chmod 기반 재현**: root 실행·플랫폼 차이에 취약. 디렉터리형 항목이 이식성 있는 유일 경로다.
- **mutant 커밋**: 프로덕션 변형이 섞인 커밋. 커밋 전 `git diff --stat` 필수.
- **워크트리 내부 증거**: 폐기 시 유실.

## §H Cross-References

- `internal/cli/doctor_agentemit_embed.go` — 4개 fail 분기(:139/:146/:155/:162), `compareEmission`(:253),
  `committedEmissionSet`(:230)
- `internal/cli/doctor_agentemit_embed_test.go` — 픽스처 4종(:18/:38/:55/:68), 형제 테스트 3종
- `CHANGELOG.md:13` — t346의 "all four fail branches" 약속 (커밋 `c2b51293e`)
- SPEC-CI-DOCTOR-BIN-001 (t346), SPEC-AGENT-EMIT-LINEAGE-001 (t317)
