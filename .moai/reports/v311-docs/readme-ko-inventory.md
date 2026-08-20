# README.ko.md inventory (778 lines, full read)

> Captured 2026-08-20. Parent-persisted (surveying agent had no write tools).
> Line numbers are identical across README.md / .ko / .ja / .zh (perfect parity).

## Section outline

| Lines | Section | Type | Version-sensitive content |
|---|---|---|---|
| 1-38 | Header / badges / hero | reference | Go 1.26+ (23), `Release-v3.1.1` badge (24), Apache-2.0 (25), 4 locale links (12-15), adk.mo.ai.kr + book + Discord (29-31) |
| 40-53 | `## v3.1 새 기능 — 칸반 모드` | feature | "v3.1" (40, 42), Aug 15 release date (42), 1→4 terminals, review absorbed by sync gate (46); Opus 5 high / GLM 5.2 xhigh (52) |
| 54-64 | `### 시작하기` | how-to | `moai cc -k`, `--name plan/run/sync` (57-60), `moai glm` (63) |
| 65-68 | `### 어느 백엔드로 나눌까` | how-to | backend split recommendation; `judge` session; 429 rate-limit spreading (67) |
| 69-92 | `### 팩토리 모드` | feature+how-to | `-f`, lane-1..N (71), `moai cc -f 4` (75), max 10 Agent()/lane, `-k`+`-f` error, `moai cg` rejects factory (80), 5 columns (84), `/moai todo` (87-88), progress.md + `/clear` (91) |
| 93-118 | `### 네 세션이 쓰는 말` | glossary | 9-row term table (107-115), `WT-<slug>` branch naming (114) |
| 119-128 | `### 보드를 눈으로 보기` | feature | `moai web`, Overview/Specs/Monitor/Settings (121) |
| 131-142 | `## 왜 moai-adk인가요?` | rationale | six identity pillars (135), three axes (141) |
| 143-155 | `### 여덟 가지 차별점` | feature table | 8 differentiators; turn cap 30 + 4 hard boundaries (148); 60-70% CG (151); 16 languages (152); 4 locales (154) |
| 156-178 | `### 무엇이 다른가` | comparison | 5-section evidence format (160), 3-phase + Tier S/M/L (161), mermaid (168-178) |
| 180-189 | `### 세 핵심이 서로를 지탱한다` | rationale | — |
| 190-219 | `### 비용은 단가가 아니라 배정이…` | benchmark | 98%/320% (192); Uber 5,000 eng (194); DeepSWE 113 tasks (198,218); model x effort cost table (200-208); 58 vs 54, 1/16, 268 vs 36 steps (210); DeepSWE v1.1 2026-07-25 (218) |
| 222-247 | `## 빠르게 시작` / `### 설치` | how-to | install.sh / install.ps1 (229,235), PowerShell 7.x+ (232), Go 1.26+ (238), `make build` (242), `moai update` (245), GLM-4.7-Flash/4.5-Flash (247) |
| 249-257 | `### 프로젝트 초기화` | how-to | `moai init my-project` (252) |
| 258-271 | `### 첫 워크플로우` | how-to | `claude` / `moai cc` (261), `/moai plan\|run\|sync` (265-267) |
| 272-283 | `### 요구사항` | reference | OS table, PowerShell 7.x+ (278), gh/tmux/golangci-lint (282) |
| 286-293 | `### 단일 진입점 /moai` | reference | **16 subcommands** (290); **4 retired**: design, brain, coverage, security (292) |
| 294-309 | `### MCP 서버` | reference | 1 active MCP entry, **21 MoAI tools in 6 groups**, 4 inactive (296); tool table (298-305); `~/.moai/.env.glm`, `~/.codex/auth.json`, fail-open inconclusive (307) |
| 311-314 | `### goal 엔진` | feature | `--max-turns 0`, `--max-duration` (313) |
| 315-318 | `### 병렬 worktree` | feature | `moai cc -w <name>`, `--spawn` (317) |
| 319-332 | `### 칸반 모드` | feature | `--kanban`/`-k` (321), `.moai/state/chain/events.jsonl` (325), WorktreeNode **13 fields** (326), `moai chain`, `moai todo` verbs (330) |
| 334-340 | `### CG 모드` | feature | 60-70% (336) |
| 342-345 | `### 16가지 언어` | feature | 16 languages (344) |
| 346-349 | `### 자동 품질 게이트` | feature | TRUST 5, `/moai gate`, 4-dimension (348) |
| 350-353 | `### @MX 태그` | feature | — |
| 354-357 | `### Navigator` | feature | `@NAV:DEC`/`@NAV:SYM`/`@MX:SPEC`, `nav-graph.json` (356) |
| 358-361 | `### 세션 핸드오프` | feature | **6-block** resume (360) |
| 362-365 | `### loop / fix` | feature | (364) |
| 366-369 | `### review --deep` | feature | (368) |
| 370-373 | `### 4-로케일 문서` | feature | parity check in build gate (372) |
| 374-381 | `### moai web 콘솔` | feature | alt text **10 tabs** (377) vs body **11 tabs** + 5 screens (380) — CONFLICT |
| 382-385 | `### ref / domain 스킬` | reference | 10 ref + **5 domain** (384) |
| 386-389 | `### 크로스 플랫폼` | feature | Go single binary (388) |
| 392-428 | `### SPEC 3-페이즈` | feature+ref | Tier S/M/L, GEARS (396), mermaid (398-407), coverage **10%** TDD/DDD switch (417,427) |
| 429-447 | `### 12-에이전트 카탈로그` | reference | **12 agents** (432-444); depth-2 seal manager-lead (438); sync-auditor **40/25/20/15** (440); `moai model profile` (446) |
| 448-453 | `### trust-but-verify` | feature | **7 read-only verifications** (450); 5-section format (452) |
| 454-459 | `### 검증 비용을 줄이고…` | feature | tail max **50 lines**, cache read **0.1x**, **1M 50% / 200K 90%** (456); hard limit **90%** (458) |
| 460-485 | `### 스테이터스라인 읽기` | reference | sample `cc v2.1.212`, `v3.1.1`, 87/88/45/82%, PR #1042 (463-466); 13-row element table (470-482) |
| 488-499 | `### 새 기능 만들기 (TDD)` | how-to | SPEC-PROFILE-001 (493-495) |
| 500-509 | `### 장시간 돌리기 (goal)` | how-to | "20 turns" example (505); default 30, 1M 50%/200K 90% (508) |
| 510-526 | `### 병렬로 돌리기 (worktree)` | how-to | `moai cc -w feature-auth`, `--spawn` (512-514) |
| 527-539 | `### 비용 줄이기 (CG)` | how-to | `moai glm sk-...`, `moai cg` (530-531), 60-70% (538) |
| 540-547 | `### 버그 자동으로 잡기 (loop)` | how-to | (543,546) |
| 550-566 | `### .moai/config/sections/` | reference | **6 YAML sections**: language/quality/harness/workflow/lsp/user (558-563); harness minimal/standard/thorough (560) |
| 567-582 | `### 모델 프로파일` | reference | **11 agents x 3 profiles = 33 cells** (569) — conflicts with 12-agent catalog; No-Haiku 3-tier (581) |
| 583-591 | `### settings.json 분리` | reference | settings.json / settings.local.json, `git rm --cached` (587-590) |
| 594-606 | `### 16가지 언어` | reference | 4x4 grid (600-603); "flutter" canonical (605) |
| 607-617 | `### 4-로케일 문서` | reference | adk.mo.ai.kr/{ko,en,ja,zh} (611-614) |
| 618-625 | `### 운영체제` | reference | PowerShell 7.x+, no native cmd.exe (624) |
| 626-648 | `### Claude + GLM` | reference | mode table ~70%/~60% (632-634); GLM Coding Plan from **$10/mo**, glm-5.3/4.7/4.5-air + free Flash (636); `ANTHROPIC_DEFAULT_*_MODEL` (638); Opus/Sonnet/Haiku/**Fable** -> glm-5.3, 1M (640-645) |
| 651-671 | `### 공식 문서` | reference | **12 doc sections** (655), table (658-670); "전체 36개" CLI (663); Harness v4 Builder (669) |
| 672-675 | `### 도서` | reference | book.mo.ai.kr (674) |
| 676-696 | `### CLI 명령표 (자주 쓰는 14개)` | reference | **14 rows** (680-693) incl. `moai graph`, `edges.jsonl` (684), worktree/session/spec/goal/harness/handoff/preference subverbs, `moai web` 5 screens + 11 tabs (693); **36 total** (695) |
| 697-702 | `### ref / domain 스킬` | reference | 10 ref (699); **8 domain** (701) — conflicts with L384's 5 |
| 703-706 | `### CHANGELOG` | reference | (705) |
| 707-710 | `### 코드 품질 요구사항` | reference | **85% coverage**, 0 lint, 0 type errors (709) |
| 713-718 | FAQ @MX | FAQ | — |
| 719-726 | FAQ 스테이터스라인 버전 | FAQ | `v3.1.0 -> v3.1.1` (722), `moai update` (725) |
| 727-730 | FAQ GLM 없이 | FAQ | (729) |
| 731-734 | FAQ 기존 프로젝트 | FAQ | coverage **10%** (733) |
| 737-750 | `### 기여하기` | how-to | `git checkout -b` (744), make test/lint/fmt (746), 85% (749) |
| 751-754 | `### 피드백` | how-to | `/moai feedback` (753) |
| 755-759 | `### 커뮤니티` | reference | Discord, GitHub Issues |
| 760-763 | `### 라이선스` | reference | Apache 2.0 (762) |
| 766-778 | `## 스타 히스토리` | reference | star-history embed |

## A) Version strings (release touch-points in bold)

23 `Go-1.26+` · **24 `Release-v3.1.1` badge** · 25 Apache-2.0 · **40, 42 `v3.1` heading** ·
52 Opus 5 / GLM 5.2 · 200-208 opus-5, glm-5.2, sonnet-5 · 218 DeepSWE v1.1 (2026-07-25) ·
232/278/624 PowerShell 7.x+ · 238 Go 1.26+ · 247 GLM-4.7-Flash / GLM-4.5-Flash ·
**330 cross-ref to "v3.1 새 기능"** · **463 `cc v2.1.212`, `v3.1.1` statusline sample** ·
636 glm-5.3 / glm-4.7 / glm-4.5-air · 640-645 glm-5.3 x4 · 669 Harness v4 Builder ·
**722 `v3.1.0 -> v3.1.1` — stale pair, must move as a unit** · 762 Apache 2.0

Release-touch points: **24, 40, 42, 330, 463, 722**.

## B) CLI surface referenced

**Launcher/session**: `moai cc -k` (57,330) · `--name plan|run|sync` (58-60,67) · `moai glm -k` (67,330) ·
`moai cc` (63,261,590,632,685,729) · `moai glm` (63,590,633,685,729) · `moai glm sk-<key>` (530) ·
`moai cg` (80,531,590,634,685,729) · `--kanban`/`-k` (80,321,330) · `-f` factory (71,80) ·
`moai cc -f` (74) · `-f 4` (75) · `-f lane-1` (76) · `moai glm -f lane-3` (77) ·
`moai cc -w <name>` (317) · `-w feature-auth` (512) · `--spawn` (317,514)

**Core CLI**: `moai init` (252,296,413,498,680,733) · `moai doctor` (681) · `moai status` (682) ·
`moai update` (245,683,725) · `moai graph build|query` (684) · `moai web` (121,374,380,693) ·
`moai worktree sync|done|remove|clean|recover|snapshot|verify|restore` (686) ·
`moai session list|register|current` (687) · `moai spec audit|archive|lint|list|new` (688) ·
`moai goal arm|status|clear` (689) · `moai harness status|apply|rollback|disable` (690) ·
`moai handoff save|list` (691) · `moai preference list|decay-scan|toggle` (692) ·
`moai chain status|lineage|back|list|prune` (330) · `moai todo add|list|next|done|unpick` (330) ·
`moai model profile` (446,569) · `moai mcp add|remove|list`, `moai mcp-server` (296) · "36 commands" (663,695)

**Slash**: `/moai` NL (270) · plan (265,290,493,502,661) · run (266,290,494,503,519,522,535,661) ·
sync (267,290,495,661) · todo (87,88,107,290,330,662) · goal (148,290,505,662) · loop (290,364,543,662) ·
fix (290,364,546,662) · gate (290,348,662) · review (290,662) · `review --deep` (368) · clean (290,662) ·
codemaps (290,662) · e2e (290,662) · mx (290) · feedback (290,662,753) · project (290) · harness (290) ·
retired: design/brain/coverage/security (292) · `/clear` (44,91,150,164,321,360,456,474,508)

**goal flags**: `--max-turns 0`, `--max-duration` (313)

**Config paths**: `.mcp.json` 296 · `~/.moai/.env.glm`, `~/.codex/auth.json` 307 ·
`.moai/state/chain/events.jsonl` 325 · `nav-graph.json` 356 · `progress.md` 91,150,458,508 ·
`.moai/config/sections/{language,quality,harness,workflow,lsp,user}.yaml` 552-563 ·
`.claude/settings.json` 587 · `.claude/settings.local.json` 588,590 · `edges.jsonl` 684 ·
`ANTHROPIC_DEFAULT_*_MODEL` 638

## C) Numeric assertions

4 terminals (46) · max 10 Agent()/lane (80) · 5 board columns (84,108) · 6 identity pillars (135) ·
3 axes (141,180) · 8 differentiators (143) · turn cap 30 + 4 hard boundaries (148,508) ·
60-70% CG (151,336,538) · 16 languages (152,165,342,344,596,600-603) · 4 locales (154,370-372,607-616) ·
5-section evidence (160,452) · 3-phase + S/M/L (161,394) · -98%/+320% (192) · Uber 5,000 eng (194) ·
DeepSWE 113 tasks (198,218) · cost table 58%/$1.66 .. 74%/$11.84, 44%/$3.92, 54%/$26.40 (200-208) ·
"+4 points, 1.8x cost" (204) · 58 vs 54, 1/16 cost, 268 vs 36 steps (210) · Go 1.26+ (23,238) ·
PowerShell 7.x+ (232,278,624) · 16 subcommands (290) · 4 retired (292) ·
1 active MCP entry / 21 tools / 6 groups / 4 inactive (296) · WorktreeNode 13 fields (326) ·
4 dimensions 40/25/20/15 (348,440) · 3 token families (356) · 6-block resume (360) ·
10 tabs alt vs 11 tabs body (377,380,693) · 10 ref + 5 domain (384) · coverage 10% switch (417,427,733) ·
12-agent catalog (429,432-444) · depth-2 seal (438) · E1-E4 (442) · 7 read-only verifications (450) ·
tail 50 lines / cache 0.1x / 1M 50% / 200K 90% (456) · hard limit 90% (458) ·
statusline sample 87/88/45/82%, PR #1042, TODO 1/3, 2/1 (463-466) · goal 20-turn example (505) ·
6 config sections (558-563) · 11 agents x 3 profiles = 33 cells (569) · 3 profiles (575-579) ·
No-Haiku 3-tier (581) · ~70% / ~60% (633-634) · GLM Coding Plan $10/mo (636) ·
4 Claude tiers -> glm-5.3, 1M each (640-645) · 12 doc sections (655,658-670) · 36 CLI commands (663,695) ·
14 frequently-used CLI (676,680-693) · 10 ref skills (699) · 8 domain skills (701) ·
85% coverage / 0 lint / 0 type errors (709,749)

## Internal inconsistencies (pre-existing, release-relevant)

1. **Settings tabs 10 vs 11** — image alt L377 says 10; body L380 and L693 say 11 (11 named).
2. **Agent count 12 vs 11** — L429 "12-에이전트 카탈로그" (12 rows) vs L569 "11 agents x 3 profiles = 33 cells".
   Possibly intentional (Explore is built-in, inherits session model) but never reconciled in text.
3. **Domain skills 5 vs 8** — L384 lists 5; L701 lists 8 (adds html-report, humanize, svg-infographic).
4. **Utility-command list drift** — L290 names 13 peripheral subcommands; L662 doc row lists only 10
   (omits mx, project, harness).
5. **L722 version pair** (`v3.1.0 -> v3.1.1`) is an illustrative current->available pair; must move as a
   unit with L24/L463 or it reads as the shipping version.
6. **L463 shows v3.1.1 while L722 "installed" shows v3.1.0** — coherent as separate examples, easy to misread.
