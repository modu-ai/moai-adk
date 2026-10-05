# SPEC-INIT-HARNESS-001 — 설계

run 단계가 따를 설계 방향. 함수·파일·테스트 이름은 제안이며, run 단계가 기존 관례에 맞춰 확정한다(이름이 바뀌면 acceptance.md 선택자도 함께 고친다). 인용 행번호는 워크트리 HEAD `a404132e7` 기준.

## §1 목표와 제약

- 하니스 3값이 배포 수준에서 서로 다른 결과를 만든다: claude=현행(파일집·로직 기준 — 내용 갱신은 REQ-IH-002/008), both=현행(같은 기준), codex=AGENTS.md+codex 표면만.
- 위저드는 4문항을 유지한다 — 문항 추가·삭제 없이 `agent_wiring`의 설명만 배포 결과를 서술하게 고친다.
- AC-CW-004와 기존 codex 배선 계약(SPEC-CODEX-WIRING-001)을 한 줄도 깨지 않는다.
- 기존 프로젝트(키 없음)는 claude로 해석되어 행동이 변하지 않는다 — 마이그레이션 없음.

## §2 결정

### D1 — 새 질문 없음. `agent_wiring`(3번째 문항)이 하니스 선택을 운반한다

조사의 "질문 신설"은 t393(SPEC-INIT-HARNESS-PROMPT-001)이 이미 해냈고 t583이 4문항 프레임에 남겼다. 남은 일은 (1) 옵션 Label·Desc를 배포 결과를 서술하게 고치고 (2) ko/ja/zh 번역 3블록(`translations.go:107/:193/:279`)을 동기화하는 것.

- **옵션 값은 동결**(claude/codex/both) — `WizardResult.AgentWiring`, `--llm` 검증, 기존 테스트 전체가 이 문자열에 의존.
- **Label·Desc 변경**(제안):
  - claude → "Claude only (Recommended)" — "Deploy the .claude/ surface plus AGENTS.md (today's default behavior)"
  - both → "Claude + Codex" — "Same claude deployment plus .codex/ wiring; .mcp.json provisioning forced on"
  - codex → "Codex only" — "AGENTS.md and Codex surfaces only — no .claude/ tree, no CLAUDE.md, no .mcp.json"
- Title도 "Select the agent harness to wire" → 배포를 포함한 서술로("Select the agent harness to deploy and wire", 제안).
- 기각: 새 질문 `harness` 추가 — 위저드 5문항이 되어 t583 프레임과 t753 목표 형상을 동시에 깬다. 기각.
- 기각: `agent_wiring` ID를 `harness`로 개명 — 이득 0, 호출부·번역·테스트 전파. 기각.

### D2 — `--llm codex`의 의미를 codex 단독 배포로 재정의한다

판정 축(조사 4축의 넷째): 값 의미 변경 vs 신규값.

- **선택: 의미 변경.** 근거 4개 — (1) 조사 운영자 결정이 위저드 3옵션의 뜻을 "claude 단독/둘 다/codex 단독"으로 이미 정의했고 플래그는 그 질문의 비대화형 경로다(플래그 > 위저드 우선이 같은 값집합 위에서 작동). (2) 값의 나이가 2주(SPEC-CODEX-WIRING-001, 2026-09-01)라 파괴면이 작다. (3) "Codex"를 고른 사용자가 기대하는 것은 codex 프로젝트지 "claude 전체 + codex 훅 한 겹"이 아니다 — 오늘의 실제 효과는 질문이 약속하는 것과 갈라져 있다(그 갈라짐이 이 SPEC의 존재 이유). (4) 넷째 값을 추가하면(예: `codex-only`) 위저드도 4옵션이 되어 조사가 확정한 3옵션을 깬다.
- **파괴면의 크기(측정)**: `--llm codex`로 init한 기존 프로젝트는 `.claude/`가 이미 있다. 이 SPEC 후에도 키가 없어 claude로 해석되므로 update 행동이 변하지 않는다(research.md §4). 변하는 것은 "앞으로 `--llm codex`로 새로 init하는 프로젝트"의 파일집뿐. CHANGELOG Breaking 라벨로 sync 단계에서 공지.
- **유지되는 핀**: AC-CW-004(플래그 부재 ≡ claude ≡ `.codex/` 배선 0) — `init_agent_flag_test.go:94-118` 무변경 통과. `--llm codex` ⇒ `.codex/` 배선 3종 존재 + `.mcp.json` 프로비저닝 생략 — 기존 단정은 유지되고 **부정 단정(`.claude/` 부재)이 추가**된다.
- `--llm` help 문자열(`init.go:132`)을 "claude, codex, or both (codex deploys AGENTS.md + Codex surfaces only — no .claude/ tree)"로 다시 쓴다.

### D3 — 배포 필터는 deployer 본문 분기가 아니라 FS 래퍼다 (slimFS 선례)

```
templates(임베디드 FS)
  └─ slimFS (카탈로그 등급 — SPEC-V3R4-CATALOG-002, slim_fs.go)   ← 기존
  └─ harnessFS (하니스 축 — 이 SPEC 신설)                          ← 같은 자리에 중첩
        deployer는 fs.FS 인터페이스만 본다 — 무 changed
```

- `harnessFS(codex)`의 두 가지 일:
  1. **은닉**: claude 전용 접두어(`.claude/`, `CLAUDE.md`, `.mcp.json`, `.claudeignore`, `.moai/status_line.sh`)를 WalkDir에서 보이지 않게 한다 — `slim_fs.go` `isHidden`과 같은 모양.
  2. **재매핑**: `.claude/skills/<name>/<rest>` 읽기를 `.agents/skills/<name>/<rest>` 대상 경로로 대응시킨다(내용은 원본). 배포기는 자기가 `.agents/skills`를 쓰고 있다고 알 필요가 없다.
- 선택 위치: `internal/core/project/initializer.go`의 deployer 생성부(`:375-393` 직전) — opts에 하니스를 실어 initializer까지 전달. claude·both는 오늘과 같은 FS(래퍼 없음 또는 항등 래퍼).
- 미러 끄기: codex 단독에서는 `WithSkillMirror(false)`(`skill_mirror.go:138-140`) — 링크가 가리킬 `.claude/skills`가 없으므로.
- 기각: deployer `Deploy` 루프 안에 하니스 분기 — 필터 규칙이 walk 콜백에 섞여 slimFS와 겹치고, 재매핑은 경로 치환이라 walk 분기보다 FS 층이 정확하다. 기각.
- 기각: codex 단독용 별도 템플릿 나무 — 3벌 포크는 조사가 명시 기각한 설계("3벌 포크 금지 — 베이스 + 합성", §2차 결정). 기각.

### D4 — 표면 소속 정본 표 (codex 단독 필터 명세)

| templates/ 경로 | claude | both | codex | 근거 |
|---|---|---|---|---|
| `AGENTS.md` | 배포 | 배포 | **배포** | 보편 (조사 :70) |
| `.moai/config/sections/*`, `.moai/` 스캐폴드 | 배포 | 배포 | **배포** | 보편 |
| `.gitignore`, `.git_hooks/`, `.github/`, `.worktreeinclude` | 배포 | 배포 | **배포** | git 인프라·보편 |
| `.codex/agents/moai/*.toml` (11) | 배포(현행) | 배포 | **배포** | codex 기착 (조사 :71) |
| `.agents/skills/moai-<command>/` (16 게시) | 배포 | 배포 | **배포** | 게시 실파일 — R-011 보호 유지 |
| `.claude/skills/**` (카탈로그 34) | 배포 | 배포 | **재매핑 배포 → `.agents/skills/<name>/`** | REQ-IH-006 |
| `.claude/**` (skills 외 전부) | 배포 | 배포 | **제외** | claude 전용 (조사 :72) |
| `CLAUDE.md` | 배포 | 배포 | **제외** | claude 전용 |
| `.mcp.json` | 배포 | 배포 | **제외** | codex는 config.toml — 현행도 프로비저닝 생략 |
| `.claudeignore` | 배포 | 배포 | **제외** | Claude Code 표면 |
| `.moai/status_line.sh(.tmpl)` | 배포 | 배포 | **제외** | Claude statusline — codex는 네이티브 statusline |
| `.codex/hooks.json`·`config.toml`·사이드카 | — (claude 배선 없음) | 배포 | **배포** | codexwiring.Wire 런타임 |

- `.claude/skills/moai/`(부모 스킬)도 재매핑 대상이다 — 이름 충돌 없음(research.md §4).
- 판정 불가 항목은 없다 — 14개 루트 엔트리 전수 판정(`ls templates/` 2026-09-14 실측).

### D5 — 영속화: `llm.yaml` 키 `llm.harness`

- **선택**: `llm.harness: claude|codex|both` — `llm.yaml`은 LLM 설정의 자연 소재, plain yaml이라 템플릿 편집이 단순, update 3-way 병합 대상이라 선택 생존이 공짜다(research.md §2.5).
- init은 **항상** 해석값을 쓴다(기본 claude 포함) — "키 없음=claude" 추론은 기존 프로젝트 호환용 폴백으로만 남긴다. defaults.go 시딩 `"claude"`.
- 소비자: update 재배포(D6), doctor 판정(D7). 위저드·플래그는 이 키를 읽지 않는다(init 시점 선택은 해석 지점이 소유).
- 기각: `.moai/manifest.json`에 하니스 기록 — manifest는 파일 추적자이지 설정 저장소가 아니고, 사용자가 manifest를 지우면 하니스가 증발한다. 기각.
- 기각: 존재 게이트(`.codex/` 유무 추론) — 사용자가 `.claude/`를 지운 claude 프로젝트와 codex 단독을 구분할 수 없다(research.md §2.5). 기각.

### D6 — update: 하니스 인지 재배포

- `runCleanReinstall` 재배포 경로에서 config loader로 `llm.harness`를 읽고(없으면 claude), codex면 D3의 harnessFS로 deployer를 만든다. claude·both는 오늘 경로.
- CleanMoaiManagedPaths는 부재 경로 스킵이 동작이므로 codex 단독 프로젝트에서 무해 — 본문 수정 불필요(판정은 run 단계 확인 과제, AC-IH-009).
- codex 배선 리프레시(`update.go:494-502`)는 하니스 무관 유지 — codex 단독에서도 갱신 대상.
- `moai update -c` reconfigure는 문항 불변 — 하니스 변경 표면은 이 SPEC 범위 밖(Out of Scope — t589 경로).

### D7 — doctor: 하니스 조건부 검사

- doctor 진입에서 `llm.harness`를 한 번 읽고, claude 표면 검사(settings.json·hooks .sh·statusline·rules·미러→`.claude` 링크 검사)를 codex 단독에서 skip 또는 INFO 강등 — 판정 선택은 "조용한 skip보다 명시적 INFO(예: `claude surface: not deployed (harness=codex)`)가 낫다"는 원칙으로 **INFO 강등**을 기본으로 제안한다. codex 검사(hooks 신뢰·config.toml·[[skills.config]] 등록)는 그대로.
- 미러 드리프트 검사(`doctor_codex_mirror_drift`)는 codex 단독에서 "미러 없음"이 정상 — 카탈로그 원본 위치가 `.agents/skills`로 옮겨갔음을 하니스에서 읽어 판정.
- golden 테스트(`doctor_golden_test.go`)에 codex 단독 시나리오 추가.

### D8 — 공개 문서와 AGENTS.md 보강 (HARD 제약의 이행)

- **AGENTS.md**(REQ-IH-008): 기존 결속표에 동결 목록 미커버 5건을 보강한다 — Agent 소환, 스킬 로더 미제공, output style, 슬래시 명령, Workflow 스크립트. 형태는 기존 표에 행 추가(3행→8행 전후)가 제일 싸고 읽기 쉽다. 상한 여유 ~7.5KB로 충분(research.md §4).
- **문서**(REQ-IH-009): README init 절 + docs-site init 가이드에 3-way 선택 절 추가 — codex 단독의 제약(위 동결 목록 요약)을 명시. 4-locale 동기 의무는 oss-docs 규칙(`.moai/docs/docs-site-i18n-rules.md`)이 지배 — sync 단계 실행.
- 이 보강은 **claude 선택 배포본에도 같이 실린다**(AGENTS.md는 보편 표면) — 문제 삼지 않는다: 결속표는 원래 "어느 하네스가 무엇을 못 하는지"의 계약이고 claude 사용자에게 무해.

### D9 — 테스트 전략 (무엇을 실행으로 증명하나)

- **codex 단독 배포 파일집**: 홈 지문 슬롯 규약(t583 REQ-IQW-014/016 준용 — 선언 후 실행 전후 실제 홈 8항목 지문, stderr 분리) 안에서 `--llm codex` 실행 → 부정 단정(`.claude/**` 0, CLAUDE.md·.mcp.json·.claudeignore·status_line.sh 0) + 긍정 단정(AGENTS.md·.codex 3종+TOML 11·게시 16+재매핑 34·config sections).
- **claude·both 보존(파일집·로직)**: claude·both 실행 테스트는 기존 것을 무변경 통과시키는 게 증거다 — 기존 테스트 본문을 고치면 보존 증거가 소멸한다(AC-IH-005/006의 왼쪽 절). 보존 단정의 범위는 배포 파일집과 배포 로직이며, REQ-IH-002(`llm.harness` 기록)·REQ-IH-008(AGENTS.md 내용)의 갱신은 각자의 AC가 잰다.
- **재매핑 무결성**: 재매핑된 `.agents/skills/<name>`이 실디렉터인지(`Lstat` — 심볼릭 아님), 링크 0인지(`find -type l` 성격의 단정), SKILL.md가 읽히는지.
- **update 부활 방지**: codex 단독 고정 프로젝트에 update 실행 → `.claude/` 여전히 0 + codex 표면 갱신 + `llm.harness` 생존.
- **RED 우선**: 각 마일스퀀스의 첫 커밋은 RED 관측(§E.2 슬롯 지문) — 특히 AC-IH-002의 RED는 오늘 나무에서 `--llm codex`가 `.claude/`를 심는 것을 보여주는 첫 증거다.

## §3 위험

| 위험 | 완화 |
|---|---|
| 재매핑이 R-011 게시 보호와 얽힘 | 게시 경로 보호는 대상 경로 기준으로 유지 — 재매핑은 원본 경로(`.claude/skills`)만 건드린다. design.md D4 표에서 축 분리 확인 |
| 하니스 값과 실제 디스크 불일치(사용자가 `.claude/`를 손으로 추가) | doctor가 불일치를 INFO로 보고 — 강제 삭제는 하지 않는다(사용자 파일 존중) |
| `llm.harness` 키를 모르는 오래된 바이너리가 codex 단독 프로젝트를 update | 오래된 바이너리는 전면 재배포로 `.claude/`를 부활시킨다 — 키-버전 결합 결함. 완화: CHANGELOG 공지 + doctor가 "키 없음+`.claude/` 부재" 조합을 감지해 INFO. 근본 수리는 범위 밖(기록) |
| 3-way 병합이 신규 키를 유실 | `RestoreMoaiConfigRetained`는 section yaml 전체 병합 — 키 추가가 아니라 기존 파일 내 키라 생존 경로가 동일. AC-IH-015로 실행 증명 |

## §4 결정 요약

| 결정 | 선택 | 핵심 이유 |
|---|---|---|
| 질문 | 기존 `agent_wiring` 재사용, Label/Desc만 갱신 | t393+t583이 이미 심어둠 — 4문항 프레임 보존 |
| 플래그 | `--llm codex` 의미 변경 → codex 단독 | 운영자 결정 정합, 파괴면 2주값, 3옵션 유지 |
| 필터 위치 | FS 래퍼(harnessFS) — slimFS 선례 | deployer 본문 무변경, 재매핑이 자연스러운 층 |
| 영속화 | `llm.yaml llm.harness` 항상 기록 | 3-way 병합 생존, 명시 > 추론 |
| update/doctor | 키 읽어 조건 분기 + INFO 강등 | 부활 방지·거짓 경고 제거, 강제 삭제 없음 |
| 공개 | AGENTS.md 표 보강 + README/docs-site | HARD 제약 — 숨김 없음 |
