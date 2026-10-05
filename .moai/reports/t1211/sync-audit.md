# Sync-Audit 판정서 — SPEC-POWERSHELL-DENY-PARITY-001 (card t1211)

- 감사 대상: 워크트리 `.claude/worktrees/t1211`, 브랜치 `WT-powershell-deny`, HEAD `7017309f0` (base `8337fdbbd`, BASE `5ac030965`)
- 감사자: sync-auditor (읽기 전용 — 코드·SPEC 무수정, 변이 실험은 스크래치 사본 `git archive HEAD` 에서만 수행)
- 평가 프로필: 내장 기본값 (Functionality·Security must-pass)
- 교차 모델: `audit_multi` → claude PASS(P2 1건: YAML↔템플릿 드리프트 무방비), codex `inconclusive`(사용량 한도), glm `inconclusive`(응답 텍스트 없음) — fail-open

## 종합 판정: **PASS-WITH-DEBT**

차단(blocking) 결함 0건. 출하된 36개 규칙·가드·문서는 모두 재측정으로 확인됐다. 다만 REQ-PSD-013 의 「SSOT 에서 생성기로 전파」는 문언대로 충족되지 않았고, YAML↔템플릿 사이의 드리프트는 **지금 이 시점에만** 일치할 뿐 기계적으로 지켜지지 않는다(F1, 실측으로 재현). 이 부채는 이 카드에서 테스트 한 개로 닫거나 후속 카드로 등록해야 한다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 88 | PASS | AC 13건(PSD-001..013) 재측정 PASS, AC-PSD-014 N/A(branch P). REQ-PSD-013 문언 이탈은 F1 |
| Security (25%) | 85 | PASS | 벤인 코퍼스 74건 중 오차단 0, 변이 6종 전부 적색. Critical/High 없음. 잔여: 훅 matcher(t1224), pwsh 하의 네이티브 `rm`, 별칭·`.exe` 정규화 미측정 |
| Craft (20%) | 82 | PASS | 가드가 스스로 변이를 품고 있고 lint·vet·gofmt 깨끗. 약점: 소스 층(YAML) 가드가 PowerShell 을 보지 않음(F1·F2) |
| Consistency (15%) | 86 | PASS | 기존 toolpolicy·template 테스트 관례 준수, 4개 로케일 내용 일치, 템플릿 중립성 유지. CHANGELOG 수치 표기 오류(F4) |

가중 평균 85.8, 조화 평균 85.2. must-pass 두 차원 모두 통과.

## 질의별 결론

### 1. REQ-PSD-013 이탈 — 의도가 충족됐는가, 향후 드리프트가 잡히는가

- **종료 시점 일치: 충족.** 독립 파서로 비교한 결과 템플릿 `PowerShell(` 집합 == YAML `tool: "PowerShell"` 집합 == 로컬 `.claude/settings.json` 집합(36).
- **지속적 무드리프트: 미충족.** 스크래치 사본에서 YAML 과 로컬을 함께 고치고(생성기 `--local-only` 흐름과 동일한 결과) 템플릿은 그대로 둔 세 변이가 모두 초록으로 통과했다.
  - g) YAML 에 `PowerShell(wipefs:*)` 추가 → toolpolicy exit 0, template exit 0
  - h) YAML 에서 `PowerShell(git push -f:*)` 제거 → 둘 다 exit 0 (로컬 사본이 짝을 잃어도 무검출)
  - i) YAML 에 `PowerShell(rm -rf C\:/:*)` 추가 → 둘 다 exit 0 (t1210 escape 가드가 `Tool == "Bash"` 만 봄)
  - 양성 대조: YAML 만 고치고 로컬을 두면 `deny only-in-yaml: PowerShell(wipefs:*)` 로 exit 1 — 변이 자체는 유효하다.
- 원인: `make tool-policy-drift-check` 는 YAML↔로컬만, 폐쇄형 가드는 템플릿 안의 Bash↔PowerShell 대응만 본다. YAML↔템플릿을 잇는 검사가 없다(이 결함 구조는 Bash 행에서도 이전부터 있었다).

### 2. 폐쇄형 가드 변이 실험 (스크래치 사본)

6종 모두 적색, 원복 후 초록. 출력은 아래 증거 절에 원문 그대로.

### 3. 규칙 정확성·오차단

- 36 = 잔여 38 − `kill -9`(별칭 헤드) − `TRUNCATE`(대소문자 접힘). 제외 11건(루트 삭제 9 + 2) 모두 사유와 함께 가드·YAML 주석에 기재.
- BASE 대비 `Bash(` deny 47건 집합·순서 동일, 템플릿의 비-PowerShell 변경은 마지막 행 쉼표 1건뿐, deny 블록(540–632행)에 템플릿 지시문 없음(git_mode 무관).
- 가드의 매처 모델로 흔한 PowerShell cmdlet·git·DB 클라이언트 명령 74건을 평가: 오차단 0. `Format-Table`/`Format-List`/`Format-Hex`/`ft`, `Clear-Host`/`Clear-Content`, `Stop-Process`, `git init`, `npm init`, `terraform init`, `mongodump`, `dropdb`, `Initialize-Disk` 모두 비매칭. 대소문자 접힘이 넓히는 방향(`psql -c drop`, `redis-cli flushall`, `GIT PUSH --FORCE`)은 모두 파괴적 명령이라 오차단이 아니다.
- 규칙 헤드(`format`, `dd`, `init`, `shutdown`, `mongo` 등)가 PowerShell 기본 별칭과 겹치는지는 **실측 못 함**(pwsh 호출이 워크트리 격리 가드에 거부됨) — Gap.

### 4. M1 증거 재파싱

jsonl 해시 5개가 progress 표와 일치. A: Bash 규칙만, deny 이벤트 없음, `Removing victim.txt` → 간극 확인. B: `PowerShell(...)` 규칙으로 `decision_reason_type: rule` 거부(bypassPermissions 하에서도 deny 가 적용됨을 보여 A 의 비거부가 유의미). C: 무규칙 실행. D: `model_refusal_fallback`, 도구 호출 없음. 모든 arm 의 init 도구 = `['PowerShell']`, `num_turns 2`. VOID 는 cwd 가 스크래치 루트, `fatal: not a git repository` — progress 에 공개·제외돼 있음. **Branch P 판정 확인.**

### 5. 문서·CHANGELOG·중립성

- 4개 로케일이 같은 커밋 `d601647e0` 에 함께 변경, 표 3행·권고문 내용 일치. ko 페이지는 원래 구조(튜토리얼형)가 달라 `###` 로 다른 위치에 들어갔지만 의미상 적절.
- 템플릿에 SPEC·카드 ID·날짜 없음(grep 0행), 중립성·누출 테스트 포함 `internal/template/...` 전체 통과.
- 후속 카드 t1224 는 큐에 `queued` 로 존재.

## 결함 목록 (structured defect-list)

- **F1** [Medium] [optional — 부채, 후속 필수] `.moai/config/sections/tool-policy.yaml` ↔ `internal/template/templates/.claude/settings.json.tmpl` — YAML 의 PowerShell 항목 추가·삭제가 템플릿 미반영인 채 모든 검사를 통과한다(변이 g·h 실측). REQ-PSD-013 문언(「생성기로 전파」) 이탈이며 무드리프트 의도는 종료 시점에만 성립. 필수 조치: YAML 의 `tool: "PowerShell"` deny 집합과 렌더된 템플릿의 `PowerShell(` deny 집합이 같음을 단언하는 Go 테스트 1개 추가(이 카드 안에서 권장, 아니면 후속 카드 발행).
- **F2** [Low] [optional] `internal/config/toolpolicy/escape_guard_test.go:66` — `escapedColonSpecifiers` 가 `e.Tool == "Bash"` 로 걸러 YAML 층의 PowerShell `\:` 을 놓친다(변이 i). 필수 조치: 조건을 `Bash` 또는 `PowerShell` 로 넓힌다. 템플릿 가드가 최종 방어선이라 실해는 F1 이 닫히면 함께 사라진다.
- **F3** [Low] [optional] `docs-site/content/{ko,en,ja,zh}/advanced/settings-json.md` 루트 삭제 행 — 내장 보호의 조건(cmd 검사는 v2.1.283 이상, `CLAUDE_CODE_DISABLE_POWERSHELL_CMD_RM_DENY=1` 로 해제 가능)과 macOS/Linux pwsh 에서 네이티브 `rm` 은 문서화된 내장 검사 밖이라는 점이 빠져 있다. 필수 조치: 한 문장 단서 추가(4개 로케일 동시).
- **F4** [Low] [optional] `CHANGELOG.md:12` — 「the 37 residual rows minus TRUNCATE」는 spec §A.3 의 잔여 38건과 어긋난다(37 은 `kill -9` 를 이미 뺀 D1 범위). 「declared list」도 YAML 로 오독될 여지. 필수 조치: 「38 residual rows minus kill -9 and TRUNCATE」로 고치고, 가드가 비교하는 대상이 템플릿 안의 Bash deny 목록 + 제외 목록임을 명시.
- **F5** [Low] [optional] `progress.md` §E.4 `sync_commit_sha: pending-backfill` — HEAD `7017309f0` 이 그 sync 커밋인데 아직 역기입 커밋이 없다. 필수 조치: 역기입 커밋을 병합 전에 추가.
- **F6** [Info] [optional, 범위 밖] `PowerShell(mkfs:*)` 는 `mkfs.ext4 /dev/sdb1` 을 막지 못한다 — Bash 원본 `mkfs:*` 의 의미를 그대로 거울질한 것이라 이 SPEC(거울질만, 확장 금지)의 결함이 아니다. `git.exe push --force` 같은 `.exe` 헤드 우회도 미측정 잔여 위험. 필요하면 별도 카드.

## 증거 (원문)

**가드 테스트**
```
$ go test ./internal/template/ -run 'TestSettingsTemplate(PowerShellDeny|DenyWildcardSyntax)' -count=1 -v
--- PASS: TestSettingsTemplatePowerShellDenyParity (0.00s)
--- PASS: TestSettingsTemplatePowerShellDenyNoOverBlock (0.00s)
--- PASS: TestSettingsTemplatePowerShellDenyGuardDetectsMutations (0.00s)
--- PASS: TestSettingsTemplateDenyWildcardSyntax (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.202s
```

**범위 테스트·드리프트**
```
$ go test ./internal/template/... ./internal/config/toolpolicy/... -count=1   → exit=0
ok  	github.com/modu-ai/moai-adk/internal/template	102.848s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.497s
ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	0.271s
?   	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.248s
$ go test ./internal/cli/ -run ToolPolicy -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	0.801s
$ make tool-policy-drift-check
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.258s
drift_exit=0
```

**AC-PSD-007·수치·정적 검사**
```
$ grep -c 'PowerShell(.*\\\\:' internal/template/templates/.claude/settings.json.tmpl  → 0
$ grep -c '"PowerShell(' …settings.json.tmpl  → 36
$ grep -c 'tool: "PowerShell"' .moai/config/sections/tool-policy.yaml  → 36
$ gofmt -l …  → (empty);  go vet ./internal/template/ → vet_ok;  golangci-lint run ./internal/template/ → 0 issues.
```

**집합 비교 (AC-PSD-006·012, 독립 파서)**
```
tmpl deny 91 bash 47 ps 36 other 8
yaml deny 95 yaml ps 36
tmpl ps == yaml ps True
tmpl deny - yaml deny []
yaml deny - tmpl deny ['Read(*.png|*.jpg|*.jpeg|*.gif|*.webp|*.bmp)', 'WebFetch(*)', 'WebSearch(*)', 'Write']
local ps 36 True ps in allow 0 ask 0
base bash 47 identical bash set True order identical True
excluded 11 ['Bash(rm -rf /:*)', 'Bash(rm -rf /\\* *)', 'Bash(rm -rf ~:*)', 'Bash(rm -rf ~/\\* *)', 'Bash(rm -rf C:/:*)', 'Bash(rm -rf C:/\\* *)', 'Bash(del /S /Q C:/:*)', 'Bash(rmdir /S /Q C:/:*)', 'Bash(Remove-Item -Recurse -Force C:/:*)', 'Bash(kill -9:*)', 'Bash(TRUNCATE:*)']
$ git diff 5ac030965 HEAD -- …settings.json.tmpl | (non-PowerShell lines)
-      "Bash(mysql -e DROP:*)"
+      "Bash(mysql -e DROP:*)",
```

**템플릿 변이 (AC-PSD-010, 스크래치 사본)**
```
== a_remove_counterpart: exit=1
    Bash deny has neither a PowerShell counterpart nor an exclusion: Bash(git clean -fdx:*)
    destructive command git clean -fdx is not blocked by any PowerShell deny
== b_add_unmapped_bash: exit=1
    Bash deny has neither a PowerShell counterpart nor an exclusion: Bash(wipefs:*)
== c_escape_colon: exit=1
    Bash deny has neither a PowerShell counterpart nor an exclusion: Bash(git reset --hard:*)
    PowerShell deny escapes ':' and cannot match a real drive path: PowerShell(git reset --hard C\:/:*)
    PowerShell deny maps to no Bash deny: PowerShell(git reset --hard C\:/:*)
    shell deny rule escapes ':' and cannot match a real drive path: "PowerShell(git reset --hard C\\:/:*)"
== d_mixed_star: exit=1
    PowerShell deny maps to no Bash deny: PowerShell(git * clean -fdx:*)
    PowerShell deny mixes a non-trailing '*' with ':*': PowerShell(git * clean -fdx:*)
    shell deny rule mixes wildcard with legacy prefix syntax: "PowerShell(git * clean -fdx:*)"
== e_case_changed: exit=1
    Bash deny has neither a PowerShell counterpart nor an exclusion: Bash(git push -f:*)
    PowerShell deny maps to no Bash deny: PowerShell(GIT push -f:*)
== f_truncate_shipped: exit=1
    benign command truncate -s 0 app.log is blocked by PowerShell(TRUNCATE:*)
    excluded Bash deny also has a PowerShell counterpart: Bash(TRUNCATE:*)
== restored: exit=0 ok  	github.com/modu-ai/moai-adk/internal/template	0.378s
```

**YAML 층 변이 (F1·F2, 스크래치 사본)**
```
== g_yaml_add_ps_without_template
   ./internal/config/toolpolicy/ -> exit=0 | ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.315s
   ./internal/template/ -> exit=0 | ok  	github.com/modu-ai/moai-adk/internal/template	0.189s
== h_yaml_drop_ps_without_template
   ./internal/config/toolpolicy/ -> exit=0 | ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.181s
   ./internal/template/ -> exit=0 | ok  	github.com/modu-ai/moai-adk/internal/template	0.183s
== i_yaml_ps_escaped_colon
   ./internal/config/toolpolicy/ -> exit=0 | ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.149s
   ./internal/template/ -> exit=0 | ok  	github.com/modu-ai/moai-adk/internal/template	0.174s
control yaml-only (local not regenerated): exit=1
   deny only-in-yaml: PowerShell(wipefs:*)
```

**오차단 코퍼스 (가드 매처 모델 복제)**
```
PowerShell deny rows: 36
benign corpus size: 74
benign commands matched by a PowerShell deny: 0
   git push --force origin main  -> PowerShell(git push --force:*)
   GIT PUSH --FORCE              -> PowerShell(git push --force:*)
   git -C r reset --hard HEAD    -> PowerShell(git * reset --hard*)
   format-volume -driveletter d  -> PowerShell(Format-Volume:*)
   psql -c drop table t          -> PowerShell(psql -c DROP:*)
   mkfs.ext4 /dev/sdb1           -> NOT BLOCKED
   kill -9 1                     -> NOT BLOCKED
   TRUNCATE TABLE t              -> NOT BLOCKED
```

**M1 재파싱**
```
== A  settings={"permissions":{"deny":["Bash(git clean -fdx:*)"]}}
   init tools=['PowerShell']  tool_use=[('PowerShell', 'git clean -fdx')]
   tool_result=[(False, 'Removing victim-dir/ | Removing victim.txt')]  permission_denied events=[]
   result={'subtype': 'success', 'num_turns': 2, 'terminal_reason': 'completed', 'stop_reason': 'end_turn', 'is_error': False, 'permission_denials': []}
== B  settings={"permissions":{"deny":["PowerShell(git clean -fdx:*)"]}}
   tool_result=[(True, 'Permission to use PowerShell with command git clean -fdx has been denied.')]
   permission_denied events=[{'tool_name': 'PowerShell', 'decision_reason_type': 'rule'}]
== C  settings={"permissions":{"deny":[]}}  tool_use=[('PowerShell', 'git clean -fdx .')]  tool_result=[(False, 'Removing victim-dir/ | Removing victim.txt')]
== D  tool_use=[]  system subtypes=[..., 'model_refusal_fallback']  permission_denials=[]
== VOID-stray-root  cwd_tail=…/scratchpad/t1211-m1  tool_result=[(True, 'Exit code 128 | fatal: not a git repository …')]
3b300ad8…d8d6  A.jsonl · 38efe665…1afd  B.jsonl · 5f842483…92c1  C.jsonl · fe4cda50…de55  D.jsonl · 7595eecf…d0ef  VOID-stray-root.jsonl
```

## Baseline 귀속

모든 수치는 이 실행에서 워크트리 HEAD `7017309f0`(스크래치 사본은 같은 커밋의 `git archive`)에 대해 측정했다. BASE 비교는 `git show 5ac030965:…settings.json.tmpl`.

## Gaps (관측하지 못한 것)

- PowerShell 기본 별칭 표와 규칙 헤드의 겹침 — pwsh 호출이 워크트리 격리 가드에 거부되어 미측정(Windows 별칭 집합은 더더욱 미측정).
- 오차단 판정은 가드의 매처 모델 기준. 실제 Claude Code 매처(별칭 정규화·복합 명령 분할·`.exe` 헤드 처리)는 미측정.
- arm 종료 코드와 arm D 의 `keep.txt` 존속은 레인 보고이며 jsonl 로는 확인 불가.
- docs-site hugo 빌드 미실행. 전체 테스트 스위트 미실행(CI 몫).
- codex·glm 교차 감사는 가용성 문제로 결론 없음.

## 잔여 위험

- 측정은 macOS·opt-in PowerShell 도구에서만 이뤄졌고 Windows 이전은 추론.
- F1 이 닫히기 전까지 YAML·템플릿·로컬 세 사본이 조용히 갈라질 수 있다.
- 훅 matcher 가 `Write|Edit|Bash` 라 PowerShell 명령은 훅 가드를 거치지 않는다(t1224).
- 거울질 원칙상 Bash 규칙의 기존 약점(`mkfs.*`, 절대경로·`.exe` 헤드)이 그대로 PowerShell 로 옮겨졌다.
