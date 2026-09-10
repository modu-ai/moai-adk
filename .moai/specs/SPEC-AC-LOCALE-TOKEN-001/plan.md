# Plan — SPEC-AC-LOCALE-TOKEN-001

## §A Context

- 카드: t573 (lane-12, Factory Mode). 브랜치 `WT-ascii-token-criterion`, base `3ac58b5a1` (origin/develop).
- 원인 SPEC: SPEC-DOCS-LOCALE-PARITY-REPAIR-001 (status: completed) — t538. 그 acceptance.md 의 AC-004 가 zh 문서의 쓰기 방식을 왜곡했다.
- 계열 뿌리: "기준이 말하는 대상을 계측기가 재지 못한다" — AC-008(행 vs 출현), AC-011(무한 도메인 vs 고정 기준선)에 이은 세 번째 형(ASCII 토큰 vs 비(非)Latin 로케일).
- 아티팩트: spec.md / plan.md / acceptance.md / progress.md + `.moai/reports/t573/{plan-research.md, census.md}`.
- 측정 근거: `.moai/reports/t573/plan-research.md` (트리 `3ac58b5a1`, 2026-09-08).

## §B Known Issues

- B-측정: 셸 `grep` 은 ugrep 래퍼 — `/usr/bin/grep` 사용 (t538 REQ-012).
- B-불일치: 리드 진술("전면 되돌려도 4 건 잔여")은 실측으로 반증됐다(전면 2, 라벨만 5). run phase 에서 수치를 인용할 때는 본 트리 재측정값만 쓴다.
- B-규율: closed SPEC 인 SPEC-DOCS-LOCALE-PARITY-REPAIR-001 의 AC 수정은 t538 의 HISTORY 규율(날짜 + before/after 명령 + 실측값 + 승인 귀속)을 따른다. 본 카드의 AC 수리는 리드 승인(카드 발행 dispatch)이 근거다.
- B-RED-now: 모든 신규/재작성 AC 는 관측된 실패(RED-now) + 녹색 경로를 쌍으로 채택한다 (`.claude/rules/moai/development/verification-completeness.md` §2).

## §C Pre-flight

```bash
git rev-parse --show-toplevel      # .claude/worktrees/t573
git rev-parse --short HEAD         # 3ac58b5a1
moai spec lint .moai/specs/SPEC-AC-LOCALE-TOKEN-001/spec.md   # ✓ No findings, exit 0 (이 워크트리에서 직접 실행 — plan-audit D8 흡수, 2026-09-08)
```

- **측정 기준 트리**: 이 워크트리 작업 트리 (base `3ac58b5a1` + 미커밋 plan-phase 아티팩트). census 명령은 `SPEC-AC-LOCALE-TOKEN-001` 디렉터리 자체를 **제외**한다 — 자기 아티팩트가 착지하면 같은 명령의 계수가 움직인다(t538 AC-011 "계수가 움직였다" 동형). 아래 수치는 모두 이 기준의 plan-시 스냅샷이며, M1 시작 시점에 같은 명령으로 **재측정**한다 — 스냅샷값을 기준선으로 인용하지 않는다.

## §D Constraints

- PRESERVE: docs-site 전체 (M3 미실행), SPEC-DOCS-LOCALE-PARITY-REPAIR-001 의 spec.md/plan.md/progress.md (acceptance.md 의 AC 행만 M2 대상), census 에서 benign 판정을 받은 모든 기준.
- 금지: `grep -c` 를 출현 수 의도로 사용, 셸 `grep` 사용, 타 SPEC 아티팩트 무단 수정, Go 코드 변경.
- 커밋: 레인이 수행 (본 plan-phase 에서는 커밋 없음).

## §E Self-Verification

- E1: AC PASS/FAIL 매트릭스 — 각 행 명령+실측 출력+트리 SHA.
- E2: census 장부의 swept-set 계수가 명시돼 있는가 (빈 집합 위의 통과 없음).
- E3: 각 재작성 기준이 REQ-001 네 요소를 충족하는가.
- E4: 뮤턴트 probe — 재작성된 zh 기준이 ASCII 보강을 제거한 사본(/tmp, sed 치환)에서도 통과하는가. `/usr/bin/grep -c "原生桌面" <보강-제거 사본>` ≥ 기준선 (실측: 전면 제거 후 6).
- E5: `moai spec lint .moai/specs/SPEC-AC-LOCALE-TOKEN-001/spec.md` → `✓ No findings — all SPEC documents are valid`, exit 0 (이 워크트리에서 직접 실행, 2026-09-08 — plan-audit D8 흡수). 주석: 이 린터의 REQ 수집 패턴은 1-세그먼트 `REQ-001` 형 ID 를 못 잡으므로 `No findings` 는 REQ 레이어 0건 판정이 아니다 — GEARS 형식 판정은 plan-audit 이 수동 수행한다.

## §F Milestones

### M1 CENSUS (Priority High)

> 아래 계수는 plan-시 스냅샷(트리 `3ac58b5a1` 작업 트리, 자기 SPEC 디렉터리 제외)이다. M1 첫 단계에서 같은 명령을 재실행하고, census.md 에는 **재측정값**을 기록한다 — 본 문서의 스냅샷값이 아니다 (plan-audit D1·D6).

1. **코퍼스** — 결함 보유 후보가 실제로 존재하는 파일 종류 전부:

   ```bash
   find .moai/specs -name SPEC-AC-LOCALE-TOKEN-001 -prune -o \
     \( -name acceptance.md -o -name spec.md -o -name plan.md \) -print | wc -l
   # → 2276 (plan-시 스냅샷)
   ```

   **경계 선언(1줄)**: `progress.md` 는 동기-단계 증거 기록으로서 구속력 있는 검증 기준이 아니므로 코퍼스에서 제외한다 — acceptance.md·spec.md·plan.md 만이 기준을 진술하는 표면이다.

2. **필터 A — 로케일 문서 참조 (3철자 병합)**:

   ```bash
   cat <코퍼스 목록> | xargs /usr/bin/grep -l 'docs-site/content/\(ko\|ja\|zh\)'   # A1 슬래시 철자 → 55
   cat <코퍼스 목록> | xargs /usr/bin/grep -lE 'content/\{[ekjz]'                  # A2 중괄호 철자 → 98
   cat <코퍼스 목록> | xargs /usr/bin/grep -lE 'README\.(ko|ja|zh)'                # A3 README 로케일 → 62
   sort -u A1 A2 A3                                                                # A (합집합) → 161
   ```

   A2 가 D3 의 brace-expansion 탈락(`{en,ko,ja,zh}` — SUBCOMMAND-RETIRE AC-SCR-011 등)을 잡는다. 표본 검증 완료: A2 매치는 실제 로케일 참조다(plan-research §6).

3. **필터 B — grep 계수식 보유 (정확한 명령, D1)**:

   ```bash
   cat <A 목록> | xargs /usr/bin/grep -l 'grep -[co]'    # A∩B → 66 (plan-시 스냅샷)
   cat <A 목록> | xargs /usr/bin/grep -L 'grep -[co]'    # A∖B → 95
   ```

   과거 기록값 17 은 B 명령이 기록되지 않아 산출 근거가 없었으므로 폐기한다. A∩B=66 의 각 파일이 REQ-002 필드 스키마의 판정 대상이다.

4. **필터 C — 사람 판정 (A∩B 66건 각각)**: (1) 계수 대상 토큰이 ASCII/Latin 인가 (2) 대상 문서가 비(非)Latin 로케일인가 (3) 토큰이 코드 식별자(CLI 플래그 값·파일명·SPEC-ID)**이면서 계수식이 코드 위치(코드 블록/플래그 표)로 한정**되는가 → benign. 산문/라벨 자리의 ASCII 토큰 계수는 토큰이 코드 식별자여도 왜곡형이다 — 유형 사례가 `desktop-native` 그 자체다(플래그 값이면서 산문 보강을 유발). (4) 계수 의미가 행인지 출현인지 함께 기록.

5. **A∖B 전수 사람 판독 (D4, 신설)** — A∖B 95파일을 전부 읽고 파일별 판정을 census.md 에 기록한다: `benign` / `비-grep 계수 관찰(awk·rg·wc·python 등 — 도구명과 대상)` / `왜곡형`. 이 단계는 패턴으로 걸러지지 않는 은신처를 닫는다 — grep 이 아닌 계수 도구는 B 필터에 안 잡힌다.

6. **이름 붙은 잔여 철자 판정** — 필터 A 가 못 잡는 관측된 우회 철자는 plan-audit 이 실재를 확인했으므로, 일반화하지 말고 **명시적으로 census 항목에 추가**해 판정한다: 접두사 없는 `content/zh/` (SPEC-I18N-001-ARCHIVED), `$loc` 변수 보간 경로 (SPEC-CC2178-TEAM-API-ALIGN-001:228), 접두사 없는 `ko/advanced/…` (SPEC-DOCS-V313-CATCHUP-001 spec.md:64), `.moai/docs/*.md` 대상 계수 (SPEC-VERSION-STAMP-GUARD-001 acceptance.md:74 — 한국어 문서 대상 grep -c).

7. **산출**: `.moai/reports/t573/census.md` — REQ-002 필드 스키마 + 단계별 swept-set 계수(재측정값) + A∖B 판독 표 + 2차 관찰(근접 미스) 기록.

### M2 REWRITE (Priority High — M1 결과에 종속)

1. census 확정 건 각각: REQ-001 네 요소를 충족하는 형태로 재작성. 선택 축: (i) 로케일-확정 토큰 (zh `原生桌面` — 전면 보강 제거 후 6행 실측), (ii) 구조 기준(매트릭스 `^| \*\*原生桌面` 3행, 표 행 수 패리티).
2. 최소 대상: SPEC-DOCS-LOCALE-PARITY-REPAIR-001 AC-004. zh 축 재작성 권장안: 매트릭스 3행(구조) + 산문 커버리지 `原生桌面` ≥1행(로케일-확정) — 두 명령 모두 전면 보강 제거 후에도 통과(실측 6행). en 축은 유지(en 은 Latin 로케일 — 왜곡 없음).
3. 재작성마다 t538 규율 HISTORY 항목: before/after 명령, 실측값(트리 SHA 포함), 승인 귀속("리드 승인 — 카드 t573 발행 dispatch").
4. 재작성 기준의 기준선은 반드시 재작성 시점 트리에서 새로 실측한다 — 본 plan 의 수치는 계획 근거일 뿐 기준선이 아니다.
5. 뮤턴트 probe 실행 후 채택: 보강-제거 사본에서도 통과하지 못하는 기준은 왜곡을 여전히 유발하므로 반려.

### M3 zh 표기 정규화 옵션 (DECISION-FLAGGED — 실행 금지)

- 운영자 게이트에 다음을 제시하고 끝낸다:
  - 옵션 A (유지): zh ASCII 보강을 그대로 둔다. 재작성 기준 기준값: `原生桌面` 6행 / 매트릭스 3행 (실측) — 어느 쪽이든 통과.
  - 옵션 B (ja 스타일 순수 원어로 되돌림): 재작성 기준 기준값 — 전면 되돌림 시 `原生桌面` 6행 유지(실측: `/usr/bin/grep -c "原生桌面" /tmp/t573-zh-reverted.md` → `6`), 매트릭스 라벨이 ja 형태가 되면 `^| \*\*原生桌面` 행 수는 3 유지. 즉 **재작성된 기준은 두 옵션 모두에서 통과한다** — 옵션 선택은 기준 통과가 아니라 표기 일관성(ja 대비)의 문제로 돌아간다.
  - 참고(옛 기준 기준값): 전면 되돌림 시 `desktop-native` 출현 2 (`≥4` 미달), 라벨만 되돌림 시 5 (충족) — 옛 기준은 A/B 선택 자체를 왜곡했다.

## §G Anti-Patterns

- 빈 swept-set 위의 census 통과 (기록: 필터별 계수 명시).
- 기준선을 실측 없이 "4 정도일 것이다"로 진술 (재측정 귀속 위반).
- `grep -c` 를 출현 수 의도로 재도입 (t538 AC-008 재발).
- census 를 M2 없이 끝내거나, 확정 건 없이 일괄 치환.

## §H Cross-References

- `.moai/specs/SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md` — HISTORY 8-33행(수리 규율 선례), AC-004 (§D.4, 69-72행).
- `.claude/rules/moai/development/verification-completeness.md` §2 (두-칸 채택, 뮤턴트 probe).
- `.claude/rules/moai/core/verification-claim-integrity.md` (측정 귀속).
- `.moai/reports/t573/plan-research.md` — 측정 근거 원본 + funnel 재실행 기록(§6).
- `.moai/reports/plan-audit/SPEC-AC-LOCALE-TOKEN-001-review-1.md` — plan-audit iter1 FAIL (0.825) 보고; 본 수리는 그 D1-D9 를 흡수했다.
