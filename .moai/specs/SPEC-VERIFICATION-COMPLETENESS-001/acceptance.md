# SPEC-VERIFICATION-COMPLETENESS-001 — acceptance

> 검증 계층. 모든 AC는 REQ-VC-007(자기적용)에 따라 **RED-현재 관측(SHA 고정 + 왜 빨간지 — 옳은 이유 — 서술) + 녹색 경로(전환 마일스톤 + 통과 출력)** 2셀을 갖는다(plan-audit D1: RED 셀은 why-red를 포함하고, "무관 파일을 누가 고치면 녹색"이 되는 경로는 기준을 실격시킨다). RED-현재 관측은 전부 plan-phase에 `32d2221fa`(본 워크트리 `WT-harness-rules`)에서 실행·관측했다. 파이프가 든 명령은 표 셀 밖 코드 펜스에 둔다.

## 계측기 (Instrument)

```text
CMD-3 (always-loaded 열거, frontmatter 한정 — 정본 계측기):
  awk 'function flush(){ if(prev!="" && !has) print prev } FNR==1{ flush(); prev=FILENAME; has=0; infm=($0=="---"); next } infm && $0=="---"{ infm=0; next } infm && /^paths:/{ has=1 } END{ flush() }' .claude/rules/moai/**/*.md | sort
CMD-B (바이트 합): CMD-3 출력 경로들에 대한 wc -c 합
CMD-N (중립성 스캔 — plan-audit D5 확대 정규식):
  grep -cE '\bt[0-9]{3,4}\b|20[0-9]{2}-[0-9]{1,2}-[0-9]{1,2}|[0-9a-f]{7,8}([^0-9a-f]|$)|SPEC-|REQ-[A-Z]+-|AC-[A-Z]+-' internal/template/templates/.claude/rules/moai/development/verification-completeness.md
  기대 0. 카드 ID(t146형 포함 전 범위)·ISO/미패딩 날짜(월 1-12)·경계 무관 SHA(백틱·행끝 포함)·SPEC/REQ/AC 토큰.
  프로브 실측(2026-08-25): 주입 4종(2026-10-05 / t146 / `f7eec06c7` / 2026-1-5) 전부 적중, 초안 어휘 6행 전부 0 — 확대 전 정규식은 주입 4종 전부 누락했었다(월 10-12·미패딩·t1xx·백틱 SHA).
  잔여 위험(수용): hex-only 7-8자 영어 단어(acceded/defaced류) 위양성 — 방향이 안전(거짓 경고지 거짓 통과 아님), 초안 어휘에서 0 확인.
보정 통제(soft-pair 보강, plan-audit D5): 본 grep과 §25.3 수동 체크리스트만이 아니다 — 상시 기계 가드 2종이 뒷받침한다(존재 실측 2026-08-25):
  .github/workflows/template-neutrality-check.yaml (3,680B, path-change 트리거) + internal/template/internal_content_leak_test.go (TestTemplateNoInternalContentLeak, 73,554B).
CMD-6 (6규칙 정착 토큰): grep -c를 각 토큰에 대해 실행 — 'observed', 'reachability', 'mutant', 'green path', 'sweep', 'pin' (각 ≥1)
CMD-7 (§A.4 VC 행 카운트 — plan-audit N1 교정형; 파이프 문자를 포함하므로 표 셀이 아닌 이 펜스에 둔다):
  grep -c '^| VC-[0-9] (' plan.md
  2026-08-25 실측 → 6. 종전 표기(파이프 앞 백슬래시)는 셸 BRE에서 교대로 읽혀 verbatim 실행 시 191(=전 행) 관측 — 백슬래시 없는 위 형태가 정본.
```

## §D. AC 매트릭스

| AC | 심각도 | 추적 | RED-현재 (관측됨, `32d2221fa`, 2026-08-25) | 녹색 경로 (전환 시점 + 통과 출력) | 붉어지는 입력 (돌연변이) |
|----|--------|------|------------------------------------------|-----------------------------------|---------------------------|
| AC-VC-001 룰 파일이 6규칙을 단일 축 구조로 착지 | MUST | REQ-VC-001 | `test -f .claude/rules/moai/development/verification-completeness.md` → rc=1 (관측됨; why-red: 산출물 부재 자체 — 무관 파일 결함 아님) | M1 후: 동일 명령 rc=0 **및** CMD-6 토큰 6종 각 ≥1회 (구조·내용이 실려 있음의 판별자) | 토큰 없는 일반 산문 파일은 존재 검사만 통과 — CMD-6이 붉힘 |
| AC-VC-002 paths 스코프 선언 | MUST | REQ-VC-004 | 파일 부재(rc=1, AC-VC-001과 동일 red). 스코프-판별력은 A-3 대조로 별도 입증 | M1 후: `sed -n '2,/^---$/p' <파일>` 내 `^paths:` 행 1개 (grep -c = 1) | frontmatter 없는 판 — CMD-3 열거에 파일이 등장해 AC-VC-003과 함께 붉힘 |
| AC-VC-003 always-loaded 예산 델타 0 | MUST | REQ-VC-006 | 기반선 관측: CMD-3 → 14파일(§A.1 목록), CMD-B → 179,081. 대조: askuser-protocol.md 포함 / spec-frontmatter-schema.md 제외 (모두 관측됨) | M3 후 run HEAD에서 CMD-3 재실행(재인용 금지): 신규 파일이 열거에 **없음**, 파일 수 14 유지, 바이트 합 179,081 유지 (델타 발생 시 명명된 외부 파일 귀속 기록) | paths 없는 판을 만들면 CMD-3이 15번째로 파일을 열거 — 즉시 붉힘 |
| AC-VC-004 템플릿 미러 바이트 동일 + 재임베드 | MUST | REQ-VC-005 | `test -f internal/template/templates/.claude/rules/moai/development/verification-completeness.md` → rc=1 (관측됨; why-red: 미러 산출물 부재 자체) | M2 후: `cmp <로컬> <템플릿>` rc=0; `make build` 종료 코드 0 (둘 다 출력 관측) | 로컬 판에 근거행을 추가하는 분기(§B.5) — cmp rc=1로 붉힘 |
| AC-VC-005 템플릿 중립성 0건 | MUST | REQ-VC-005 | 대상 파일 부재 — 스캔의 red-입력은 토큰 주입 자체(아래 열) | M2 후: CMD-N → `0` | 카드 ID/날짜/SHA/`SPEC-`/REQ·AC 토큰 1건 주입만으로 카운트 ≥1 — 붉힘 |
| AC-VC-006 6규칙 각각 인라인 근거(§2는 2행) | MUST | REQ-VC-002 | 파일 부재 → `> Evidence:` 행 0건 (rc=1) | M1 후: `grep -c '^> Evidence:'` = 6 — §1.1·§1.2·§2(규칙 4+돌연변이 2행)·§3·§4, 6행=6규칙 1:1(plan §A.3 마감 지침과 정렬, plan-audit D8), 각 근거행 ≥160자 (보일러플레이트 판별) | 한 줄짜리 상투 근거 — 길이 하한; 마커 6행 미만 — 카운트가 붉힘 |
| AC-VC-007 근거 행렬 (카드 단위 원천) | SHOULD | REQ-VC-003 | plan-phase 산출로 본 문서 작성 시점에 이미 녹색(설계상 — plan 산출물 검증). red-입력: 카드/반복/관측 열을 뺀 행렬 | 검증: plan.md §A.4의 VC 라벨 행 6개 — CMD-7(계측기 펜스의 교정형 명령, plan-audit N1) = 6, 2026-08-25 실측 6 — 문구와 명령이 같은 것을 잰다(D8). §A.5 예측 장부의 VC 행은 ID 뒤 괄호 라벨이 없어 CMD-7 패턴에 섞이지 않는다; 부속 2행·각주 1행은 별도 라벨로 §A.4 표에 존재 | 규칙 문구만 있고 관측 원천이 없는 행 — 열 검사가 붉힘; 매트릭스 행이 6에서 줄면 카운트가 붉힘 |
| AC-VC-008 자기적용 감사 (메타) | MUST | REQ-VC-007 | 본 §D의 델타/불변 AC(VC-003/004)가 SHA 고정 기반선 + why-red + 녹색 경로를 이미 실음 — plan-phase 관측 | M3 감사: 본 파일 재독 — SHA 고정 없는 델타 AC 0건, 녹색 경로 없는 AC 0건, why-red 서술 없는 RED 셀 0건 (감사 기록을 §E.2에) | 이후 수정에서 기반선 SHA를 떼어내면 본 AC가 붉힘 (규칙 6의 자기 집행) |
| AC-VC-009 zone-registry 비적용 결정 기록 | SHOULD | REQ-VC-008 | plan-phase 산출로 이미 녹색 — plan.md §A.2 D5가 결정과 정책 근거를 함께 실음. AC-VC-007이 확립한 구조를 따른다: plan 산출물의 문서-내부 검증이며 red-input은 실재함(아래 열) | 검증: `grep -c 'zone-registry' plan.md` ≥ 1 및 'ID Allocation Policy' 문구 존재 (plan-audit D4) | D5 행을 삭제하면('zone-registry' 0회) grep이, 근거 문구 없이 결론만 쓰면 문구 검사가 붉힘 |

**심각도 기준** — MUST: 카드 산출 축 (a)/(b) 직결 또는 Template-First/중립성 위반. SHOULD: 내부 추적 품질(감점 대상, 단독 FAIL 사유 아님).

**간접 검증 (indirect)** — CMD-6 토큰은 6규칙 '존재'의 판별자이지 내용 충실도의 증명이 아니다. 내용 충실도(6규칙이 카드 본문 의도를 옮겼는지)는 plan.md §A.4 행렬과의 대조(AC-VC-007)와 plan-audit의 판독에 위임한다 — 본 파일은 이 지점을 Gaps로 명시한다.

## §F. Given-When-Then 시나리오 (검증 계층, 전 AC 이항 판정)

- **GWT-1 (AC-VC-003, 예산)** — Given 기반선이 `32d2221fa`에서 14파일/179,081바이트로 관측되어 있고, When run-phase가 M3에서 CMD-3을 run HEAD에 재실행하면, Then 열거에 `verification-completeness.md` 가 없고 파일 수는 14다 (15가 되면 FAIL; 외부 파일 추가면 귀속 기록으로 판정 보류 아닌 명시 처리).
- **GWT-2 (AC-VC-005, 중립성)** — Given M2가 템플릿 판을 착지했고, When CMD-N이 그 파일을 스캔하면, Then 카운트는 `0`이다 (≥1이면 FAIL, 토큰 종류와 무관).
- **GWT-3 (AC-VC-004, 동일성)** — Given 두 판이 존재하고, When `cmp` 가 실행되면, Then rc=0이다 (rc=1이면 FAIL — 근거행 몰래추가 등 분기).
- **GWT-4 (AC-VC-001/006, 내용)** — Given M1이 룰 파일을 착지했고, When CMD-6과 `> Evidence:` 카운트를 실행하면, Then 토큰 6종 각 ≥1이고 근거행은 6개다.

## §G. 품질 게이트 (Definition of Done)

1. §D 전 AC 판정 — MUST 전 PASS, SHOULD 감점 없음. 각 판정은 §E.2에 (명령, 관측 출력, 측정 HEAD) 증거 행으로.
2. PRESERVE 준수 — `git diff --name-status 32d2221fa -- .claude/rules internal/template/templates` 출력이 정확히 `A` 행 2개(로컬 룰 파일, 템플릿 미러)뿐 (pinned-SHA diff — 이동 ref 금지, 규칙 6).
3. `make build` 종료 코드 0 관측 (M2).
4. 규칙 파일 본문 영어 확인(토큰 스캔과 별도로 판독).
5. `moai spec lint` — 레인 환경에 CLI가 있으면 실행하여 본 SPEC 프론트매터/구조 findings 0-error 관측; CLI 부재 시 Gaps로 기록(판정 유예 사유로 갈음하지 않음).
6. 알려진 Gaps (사전 선언): CMD-6 토큰 검사의 내용-충실도 상한(§D 간접 검증 노트); `make build` 산출 바이너리의 실행 검증은 본 카드 범위 밖(CI 몫).
