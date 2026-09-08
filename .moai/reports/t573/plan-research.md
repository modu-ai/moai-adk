# Plan Research — card t573 (SPEC-AC-LOCALE-TOKEN-001)

> 측정 트리: `3ac58b5a1` (worktree `WT-ascii-token-criterion`, base = origin/develop) · 날짜: 2026-09-08
> 모든 grep 은 `/usr/bin/grep`. 명령과 출력은 그대로 옮겼다.

## 1. 기원과 결함

t538 (SPEC-DOCS-LOCALE-PARITY-REPAIR-001) sync-audit F2 → 리드가 카드로 승격. AC-004(`.moai/specs/SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md` :42, 검증식 §D.4 :72 `grep -ci "desktop-native"`)가 zh 문서에 ASCII 보강 표기를 유도했다. 계열 뿌리는 같은 파일의 AC-008(행 vs 출현)·AC-011(무한 도메인) — "기준이 말하는 대상을 계측기가 재지 못한다".

## 2. 실측 (本 트리)

| 측정 | 명령 | 출력 |
|---|---|---|
| zh `desktop-native` 출현(대소문자 무시) | `/usr/bin/grep -oi "desktop-native" docs-site/content/zh/utility-commands/moai-e2e.md \| wc -l` | `8` |
| zh 출현(대소문자 구분) | `/usr/bin/grep -o "desktop-native" … \| wc -l` | `8` |
| zh 행 수 | `/usr/bin/grep -c "desktop-native" …` | `7` |
| 히트 행 | `/usr/bin/grep -n "desktop-native" …` | :58(×2, 플래그 표 — 코드 토큰), :78-80(×3, 매트릭스 라벨 ASCII 보강), :84·:98·:176(×3, 산문 보강) |
| ja 출현 | `… ja … \| wc -l` | `2` (플래그 행뿐, 라벨은 순수 원어 `デスクトップネイティブ`) |
| ko 출현 | `… ko … \| wc -l` | `2` |
| en 출현 — **대소문자 구분 `-o`** | `/usr/bin/grep -o "desktop-native" …en… \| wc -l` | `4` (자연 산출) |
| en 출현 — 대소문자 무시 `-oi` | `/usr/bin/grep -oi "desktop-native" …en… \| wc -l` | `8` (`Desktop-native` 대문자 변이 :78-80·:98 — plan-audit D7 에서 보완 기록) |
| zh 원어 토큰 `原生桌面` | `/usr/bin/grep -c "原生桌面" …zh…` / `-o … \| wc -l` | `6`행 / `7`출현 |
| ja 원어 토큰 `デスクトップネイティブ` | `/usr/bin/grep -c "デスクトップネイティブ" …ja…` | `6`행 |

→ 왜곡은 zh 에만 격리. ASCII 계수 기준이 겨냥한 유일 로케일만 쓰기 방식이 바뀌었다.

## 3. 불일치 재측정 — 리드 진술 vs 레인 수치

| 시나리오 | 명령 | 출현 수 | 리드 주장(4) 판정 |
|---|---|---|---|
| 현재 | 실측(§2) | 8 | — |
| 전면 보강 제거 | `sed 's/, desktop-native)/)/g; s/ (desktop-native)//g' … > /tmp/t573-zh-reverted.md && /usr/bin/grep -o "desktop-native" /tmp/t573-zh-reverted.md \| wc -l` | **2** (플래그 행뿐 — 옛 기준 `≥4` 미달) | **거짓** |
| 라벨만 제거 | `sed 's/, desktop-native)/)/g' … > /tmp/t573-zh-labelonly.md && …` | **5** | — |

레인의 재측정 수치(2 / 5)가 맞다. 산술 근거: 8 출현의 분포가 :58(×2)·:78·:79·:80·:84·:98·:176 이므로 전면 제거 시 2, 라벨만 제거 시 5.

## 4. 재작성 후보 기준의 왜곡 불가성 (뮤턴트 probe)

- `原生桌面` 은 전면 보강 제거 사본에서도 `/usr/bin/grep -c "原生桌面" /tmp/t573-zh-reverted.md` → **6**행 유지. → 원어 토큰 기준은 ASCII 보강 여부와 무관하게 통과 — 왜곡 압력이 없다.
- 구조 기준(매트릭스 행 `^| \*\*原生桌面` = 3)도 동일.

## 5. Census 설계 (M1 판별식) — **plan-audit iter1 에서 갱신됨 (§7 참조)**

> 아래 최초 기록(A∩B=17)은 B 명령이 기록되지 않아 재현 불가로 폐기됐다(plan-audit D1). 유효한 설계는 §7 의 재실행 funnel 이다 — 여기는 경과 기록으로 남긴다.

- 최초 코퍼스: `find .moai/specs -name acceptance.md | wc -l` → 709 (작업 트리, 자기 SPEC 미제외; 커밋 트리 기준) — 이후 자기-포함 표류 관측으로 제외 규칙 신설(D6).
- 최초 A: 슬래시 철자만 — 21, README 16 → A∩B 기록 17 (B 명령 미기록 — 폐기 사유).
- 2차 관찰(수정 없음): 단위 불일치(t538 AC-008 형), 무한 도메인 vs 고정 기준선(t538 AC-011 형).

## 6. plan-audit iter1 수리 — funnel 재실행 (2026-09-08, 트리 `3ac58b5a1` 작업 트리)

감사 보고: `.moai/reports/plan-audit/SPEC-AC-LOCALE-TOKEN-001-review-1.md` (FAIL 0.825, D1-D9). 아래가 수리 재측정이다. 모든 명령은 `/usr/bin/grep`, 코퍼스에서 `SPEC-AC-LOCALE-TOKEN-001` 디렉터리 제외(D6), 코퍼스는 acceptance.md+spec.md+plan.md 로 확장(D2), A 는 중괄호 철자 병합(D3), B 명령 명기(D1).

```bash
L=/tmp/t573-corpus.txt
find .moai/specs -name SPEC-AC-LOCALE-TOKEN-001 -prune -o \
  \( -name acceptance.md -o -name spec.md -o -name plan.md \) -print > $L
wc -l < $L                                            # → 2276  (코퍼스)
cat $L | xargs /usr/bin/grep -l 'docs-site/content/\(ko\|ja\|zh\)' > /tmp/t573-A1.txt
wc -l < /tmp/t573-A1.txt                              # → 55    (A1 슬래시 철자)
cat $L | xargs /usr/bin/grep -lE 'content/\{[ekjz]' > /tmp/t573-A2.txt
wc -l < /tmp/t573-A2.txt                              # → 98    (A2 중괄호 철자 — D3)
cat $L | xargs /usr/bin/grep -lE 'README\.(ko|ja|zh)' > /tmp/t573-A3.txt
wc -l < /tmp/t573-A3.txt                              # → 62    (A3 README 로케일)
sort -u /tmp/t573-A1.txt /tmp/t573-A2.txt /tmp/t573-A3.txt > /tmp/t573-A.txt
wc -l < /tmp/t573-A.txt                               # → 161   (A 합집합)
cat /tmp/t573-A.txt | xargs /usr/bin/grep -l 'grep -[co]' > /tmp/t573-AB.txt
wc -l < /tmp/t573-AB.txt                              # → 66    (A∩B — B 의 정확한 명령: grep -l 'grep -[co]')
cat /tmp/t573-A.txt | xargs /usr/bin/grep -L 'grep -[co]' > /tmp/t573-AminusB.txt
wc -l < /tmp/t573-AminusB.txt                         # → 95    (A∖B — 전수 사람 판독 대상, D4)
```

- **경과 대조**: 최초 기록 17(acceptance.md 한정, B 미기록) → 감사자의 acceptance-only 커밋-트리 재현 21 → 본 재실행(코퍼스 확장+중괄호 병합+자기 제외, B 명령 명기) 66. 이제 어떤 수치든 위 명령으로 재현된다.
- **A2 표본 검증** (과잉 매치 배제): `SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/spec.md:162`, `SPEC-V3R6-WORKFLOW-DOCS-001/acceptance.md:11`, `SPEC-AGENT-PARALLEL-OPT-001/spec.md:405` 등 — 전부 실제 `docs-site/content/{en,ko,ja,zh}` 참조다.
- **lint 흡수 (D8)**: `moai spec lint .moai/specs/SPEC-AC-LOCALE-TOKEN-001/spec.md` → `✓ No findings — all SPEC documents are valid`, exit 0 (이 워크트리에서 직접 실행 — 종전 "primary 체크아웃만 본다"는 지연 근거는 감사가 반증했고 철회한다). 주석: 린터 REQ 수집 패턴은 1-세그먼트 `REQ-001` ID 를 못 잡으므로 REQ 레이어 판정은 plan-audit 수동 소관.
- **잔여 한계 (기록 목적)**: 필터 A 가 못 잡는 우회 철자 — 접두사 없는 `content/zh/`, `$loc` 보간, 접두사 없는 `ko/advanced/…`, `.moai/docs/*.md` 대상 계수 — 는 일반화하지 않고 plan.md M1.6 의 이름 붙은 항목 4건으로 census 가 판정한다.

## 7. RED-now 기록 (plan 시점, 트리 3ac58b5a1)

- census 장부 부재: `ls .moai/reports/t573/` → `No such file or directory` (exit 1).
- 옛 ASCII 계수 기준 잔존: `/usr/bin/grep -c 'grep -ci "desktop-native"' .moai/specs/SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md` → `1` (exit 0).
