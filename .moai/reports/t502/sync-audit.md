# t502 sync-audit 판정서 — SPEC-CODEX-SKILL-DISABLE-001

- **판정**: **FAIL** · 가중합 **89.0**
- **감사 앵커**: 커밋 `43e820663` (`WT-codex-skillconfig`) — 감사 시작·종료 시점 모두 `git status --short` 무출력
- **감사 주체**: sync-auditor (독립 — 구현 주체와 분리)
- **기록 주체**: 리드 세션. 판정 본문이 세션에만 있어 나중 독자가 advisory 를 찾을 곳이 없었으므로 디스크에 남긴다.
- **증거 캡처**: `.moai/state/verify/t502-audit/{gotest.txt,e2e-symlink.txt,e2e-copy.txt}`

> **읽는 순서 주의.** 이 판정은 `43e820663` 을 잰 것이다. 그 뒤 blocking 수리(`bc651da28`)와 F2 수리(`13ae49a05`)가 착지했으므로, 아래 F1·F2 의 상태는 § 후속 처분 절을 함께 읽어야 현재 트리를 기술한다.

## 4차원 점수

| 차원 | 가중 | 점수 | 판정 | 근거 |
|---|---|---|---|---|
| Functionality | 40% | 92 | PASS | 17 AC 전원, E2E 2형태 독립 재현, 스위트 `GOTEST_EXIT=0` |
| Security | 25% | 90 | PASS | traversal rc=1 무쓰기, 주입 가드 바이트 불변, 백업 선행, 모드 보존 |
| Craft | 20% | 78 | PASS | vet·gofmt 무경고, 그러나 커버리지 공백 5곳 + 근거 없는 주석 1건 |
| Consistency | 15% | 94 | PASS | prune 규율·기존 헬퍼 재사용, 주석 언어 규약 준수 |

must-pass 방화벽(Functionality·Security) 통과. FAIL 은 점수가 아니라 아래 blocking 1건에서 나온다.

## F1 [BLOCKING] `enabled` 발행 지점이 하나가 아니다 — **수리됨**

`internal/cli/codex_skills_disable.go:286-289` 이 `upsertCodexSkillDisable` **밖에서** `enabled = false` 를 조립하는데(`setEnabledFalseInExtent` 의 삽입 분기), 같은 파일 `:13-16` 과 `:200-201` 은 발행이 「한 함수로 좁혀져 있어 키를 누락할 수 없다」고 단언했다. 그 좁힘이 t504 F1(키 없는 엔트리 하나면 사용자 codex 전면 기동 실패, rc=1)을 막는 **명시된 기구**다.

결과로 AC-CSD-001 뮤턴트 3(「`enabled` 줄 발행 제거 시 RED」)이 append 쪽만 덮었다. 커버리지가 확증: `283.2` · `283.18` · `286.2` 전부 count=0 — 삽입 분기를 실행하는 테스트가 없었다.

**코드 자체는 옳았다.** 감사관이 블랙박스로 확인(내부 무수정, 실제 verb + `codex-cli 0.153.4`): `enabled` 없는 엔트리에 codex 가 rc=1 로 죽는 것을 재현한 뒤, verb 실행 → 키 삽입 → codex rc=0, marker=0. 즉 live bug 가 아니라 **가드 공백 + 거짓 주석**이었다.

## advisory 11건 (F2-F12)

자동 수리 라우팅 금지 — 판정 시점에 어느 것도 소관이 배정되지 않았다. 파일 기준은 `internal/cli/codex_skills_disable.go`(명시된 경우 제외), 줄번호는 `43e820663` 기준.

| id | 내용 |
|---|---|
| **F2** | `:294-302` `reshapeLike` CR 분기 미실행(`296.36` count=0) → AC-CSD-022 의 CRLF 성질이 append 경로에서만 증명됨 |
| **F3** | `:366-369` 러너의 update 경로 미실행(`367.42` count=0) — 사용자가 실제로 밟는 경로인데 러너 수준 테스트 없음 |
| **F4** | `:216-218` TOML 주입/Windows 거절 가드 무테스트(`216.48` count=0)인데 CHANGELOG 는 「Windows publication is refused」를 이름 붙여 출하 |
| **F5** | skip 이 rc=0 — Windows 에서는 매 호출이 무효인데 스크립트에는 성공으로 보인다(AC-CSD-011/012 가 일부러 비0을 내는 것과 대비) |
| **F6** | `:276` 이 `enabled` 줄의 후행 주석을 버린다 |
| **F7** | `internal/codexwiring/skills.go:131` 이 `enabled = "false"`(따옴표) 를 false 로 읽어 verb 가 「이미 비활성」이라 보고 — codex 가 문자열 값에 게이트하는지는 미측정(t451 상속, 이 diff 밖) |
| **F8** | AC-CSD-033 의 「구별되는 사유」 단언이 subtest 별 `t.TempDir()` 경로로 약해짐 — 완전 붕괴는 여전히 잡히나 Skipped/Unchanged 문구만 합치면 통과(재주입하지 않음) |
| **F9** | `:307-314` CRLF 추론이 아무 줄의 `\r` 하나로 결정됨 |
| **F10** | `.moai/reports/t502/e2e-verb.sh:45` 의 `verb exit: $?` 가 `sed` 의 종료코드를 읽어 항상 0 — 계측기 결함이며, AC-CSD-003 판정 자체는 probe 의 rc 에 걸려 있어 무사 |
| **F11** | `:378` 백업 파일명이 초 단위라 같은 초에 두 번 실행하면 앞 백업을 덮어쓴다 |
| **F12** | `<cfg>.moai-staged` 심링크 추종(권한 상승 없음, 기록만) |

## 감사관이 독립 재현한 것

- **AC-CSD-003 / AC-CSD-050** — `e2e-verb.sh symlink`·`copy` 둘 다 exit 0, `verdict=exposed marker=1` → `verdict=gated marker=0`, `result=MATCH` ×2, `E2E PASS`, `post-run config mode: 644`, 재실행 후 엔트리 1개. `codex-cli 0.153.4`. 실사용 `~/.codex/config.toml` sha256 전후 동일. **카운터 생존 대조군**: ungated 에서 marker=1, 부재 토큰 0.
- **AC-CSD-040 경계** — 고정 base `bf779ecf2..HEAD` 의 `internal/` 4파일, 금지 2파일 없음. **그 grep 이 볼 수 있음을 대조군으로 증명**(같은 grep 이 `codex_skills_disable.go` 는 찾아낸다).
- **테스트 모집단** — 셀렉터 16개 실측, 무의미 셀렉터 0(양방향 대조).
- **스위트** — `GOTEST_EXIT=0`, `ok` 18줄, `FAIL` 0줄(빈 파일의 0이 아님을 `ok` 카운트로 확인), `internal/cli 375.946s`.
- **vet** — 영향 2패키지 exit 0, 출력 0줄, 그리고 vet 이 실제로 보고할 수 있음을 고의 결함 패키지로 확인.
- **보안 블랙박스** — `../../../etc/passwd` → rc=1 무쓰기; 따옴표 든 스킬 디렉터리 실제 생성 → skip, config 바이트 불변, 백업 미생성.

## DoD 뮤턴트 틱에 대한 독립 견해 (리드 요청)

sync 단계가 뮤턴트를 **재주입하지 않고** 체크한 것에 대해: **측정된 범위에서 타당하다.** 감사관은 문자열 존재 확인보다 강한 근거를 만들었다 — 다섯 뮤턴트가 **단언 구조상** 잡히는지 정적으로 추적해 다섯 다 잡힘을 확인했다. 재주입 못 한 사실을 표에 적은 처신은 옳다.

앞지르는 지점은 하나였다: 뮤턴트 3 이 「발행 제거」를 일반적으로 말하는데 테스트는 append 지점만 덮었다(= F1). 틱을 무효로 만들지는 않으나 문구를 좁히거나 테스트를 넓혀야 했다.

## 감사관이 보지 않은 것

Windows·Linux 런타임 동작(darwin 한정, 크로스빌드 미실행) · 0.153.4 외 codex 버전 · `golangci-lint` 미실행(sync 단계 기록을 재현하지 않음) · 전체 스위트(CI 몫) · 뮤턴트 재주입(`internal/` 쓰기, 위임 경계 밖 — 단언 구조 추적으로 대체) · docs-site 4로케일 · `gate-path-shape.md` 19셀 원 행렬 재측정(2셀 게이트만 재현).

## 후속 처분 (판정 이후 착지분 — 리드 기록)

| 항목 | 처분 | 커밋 |
|---|---|---|
| **F1** | **수리됨.** 삽입 지점을 고정하는 테스트 신설(`TestUpsertCodexSkillDisableInsertsMissingEnabledKey`) + 주석 정정. 증거는 RED→GREEN 이 아니라 **MUTANT-6 대조** — 새 테스트는 실패하고 기존 AC-CSD-001 테스트는 통과. 수리 중 발행 지점이 둘이 아니라 **셋**(재작성 분기 포함)임이 추가로 드러났다 | `bc651da28` |
| **F2** | **수리됨.** F1 픽스처를 CRLF 로 쓰는 것만으로 `reshapeLike` CR 분기가 덮였다(`332.36` 0→1). 줄 커버리지는 단언이 아니므로 **MUTANT-7**(CR 을 버리는 뮤턴트)로 판별력 확인 — 새 테스트만 잡고 기존 테스트는 통과 | `13ae49a05` |
| **F3** | **advisory 유지.** F1 픽스처로 떨어지지 않았다 — 러너 경로는 별도 픽스처·codex home 이 필요해 「첫 테스트 이름을 쓴 두 번째 테스트」가 된다. 새 테스트를 쓰지 않기로 판단 | — |
| **F4-F12** | **advisory 유지.** 소관 미배정 | — |
| 도달 불가 분기 | `pathLine < 0` bail 은 커버리지 0 으로 **의도적 유지** — 호출자가 만들 수 없는 입력을 고정하는 테스트는 아무것도 증명하지 못하고 죽은 분기를 살아있는 것처럼 굳힌다. 이유를 코드 위에 기록 | `bc651da28` |
| 후속 후보 | `upsertCodexSkillDisable` 사후조건 가드(키 없는 엔트리 반환 거부) — 세 지점을 한 번에 기계적으로 덮으나 런타임 동작 변경이라 최소 수리 범위 밖. **결함이 아니라 후보로 기록** | — |

## 절차 기록 — 동시 쓰기 1건

재닫기 중 트리에 두 번째 쓰기 주체가 있었다(구현 에이전트가 정리 지시 후에도 MUTANT-7 판별을 실행 중). docs 에이전트가 이를 발견하고 (1) 자기 커버리지 실행을 **귀속 불가로 폐기**(외부 편집이 컴파일 창 안에서 착지해 어느 트리로 빌드됐는지 말할 수 없음), (2) **살아있는 뮤턴트를 담은 트리에 SPEC 닫기를 거부**, (3) `pkill` 을 자기 명령줄·자기 빌드 경로에 고정해 남의 프로세스를 건드리지 않음 — 셋 다 옳은 처신이다. 창을 만든 책임은 리드에 있다(정리 확인 전에 다음 에이전트를 투입).

---

분류: 판정 기록 (sync-phase). 이 문서는 감사 시점(`43e820663`)의 판정과 그 이후 처분을 함께 담는다.
