# t1336 verdict — GLM 채널 한국어 자모 무결성 (Class B 직행)

## Claim

1. GLM 채널 출력의 한국어 오염은 **자모 슬롯 치환** 클래스다 — 오염 음절이 모두 유효한 조합형 음절로 남아 **NFC/NFD 정규화로는 검출도 수리도 불가능**하다(기계 입증: 재료 파일 내 자모 블록 문자 0개, NFC 재인코딩 바이트 동일 — `analysis/extract-suspects.txt`).
2. 탐지·복원 유틸 `scripts/jamo_integrity.py`(stdlib 전용) + 재사용 가이드 `.moai/docs/jamo-integrity-guide.md` 를 착지했다 — selftest 9/9, 오염 재료에서 23건 검출(레인 재실행 동일).
3. 근원은 **백엔드 출력 토큰 스트림**(관측 귀속)이다 — 레인 완결 범위는 **detect · restore · prevent(가이드)** 3축이며, 근본 해소는 운영자 보고 항목이다(§ 운영자 보고).

## Evidence (증거 — 명령 + 전사)

**패턴 특성화** — `.moai/reports/t1336/characterization.md` (14쌍 확정, jamo 슬롯 델타 표):

| 오염형 | 본래형 | 슬롯 델타 요지 |
|---|---|---|
| 채넍/채넎이/채넞 | 채널 | 종성 ㄵ/ㄶ/ㅈ→ㄹ |
| 바이넄리 | 바이너리 | 중성 ㅒ→ㅓ·종성 ㅋ→없음 |
| 숴환 | 순환 | 중성 ㅝ→ㅜ·종성 없음→ㄴ |
| 신귴/신귵 | 신규 | 종성 ㅋ/ㅌ→없음 |
| 엄트리/나먲지/브랿치 | 엔트리/나머지/브랜치 | 중성·종성 치환 |
| 엉터닐 | 엉터리 | 초성 ㄴ→ㄹ·종성 ㄹ→없음 |
| 바로잠펴 | 바로잡기 | 3슬롯 (최대 거리 관측) |

거리 분포: d=1 ×7 · d=2 ×7 · d=3 ×1 → **검출기 d≤2, 자동수리 d=1-유일로 설계**된 근거.

**검출 재현 (레인 재실행, 이번 run)**:
```
$ python3 scripts/jamo_integrity.py check .moai/reports/t1336/materials/t1333-verdict-original.md
  SUSPICIOUS: 23 finding(s)  [heuristic detector — candidates, not verdicts]
  vocab: 77499 words from git-tracked prose (targets excluded)
  check-exit=1
$ python3 scripts/jamo_integrity.py --selftest
  selftest: 9/9 assertions passed   selftest-exit=0
```
첫 후보 `인트리`(d=1 → 엔트리)는 **정당한 합성어 오탐**이다 — 검출기 출력은 후보지이지 판정이 아니며(출력 첫 행 명시), 마감 판정은 사람이 후보를 읽는다.

**복원 실측** — `demo-restore-dryrun.txt`(자동수리 제안 8건 + 모호 15건), `demo-restore-apply-diff.txt`(/tmp 사본 apply): 8건 중 **6건 정확·2건 오수리**(`숴환→소환`[본래 순환], `큐`[본래 및]) — **dry-run 검토 없는 `--apply` 금지의 실측 근거**(가이드 명문화).

**부가 발견 (탐지기가 실오염 2건 추가 포착)**:
- `git-workflow-doctrine.md:237` `처리르`(본래 `처리를`) — **추적 문서 내 실오염**. 2026-05-17 사용자 지시문 **인용부**라 수리 여부는 인용 정합성 판정이 선행된다 → 리드 처분 항목.
- `jamo-integrity-guide.md:31` `고되되`→`고치되` — 위임 중 본 출력 채널의 실오염을 가이드 자가 점검(`selfcheck-guide*.txt`)이 포착·수리. **검출기가 같은 채널의 산출물을 잡아낸 실증.**

**오탐 실험** — 클린 추적 한국어 4파일(README.ko.md 65K 등) 6,645 어휘 → 138플래그(2.1%, 코퍼스 부재 노이즈 위주): `demo-fp-clean-files.txt`. exit 1 을 무차별 게이트로 쓰지 말 것 — 용도는 레인 산출물 마감 전 자가 점검.

## Baseline-attribution (baseline 귀속)

전 측정은 이번 run, 트리 `.moai/worktrees/t1336`(base `145c3d98c`, branch `WT-jamo-integrity`) 대면. 구현 커밋 `14495710e`(scripts + guide 2파일). 오염 재료는 `materials/t1333-verdict-original.md` 바이트 사본(2,503B) — **프롬프트·출력 채널 재입력 금지 규율**로 파일 직독만을 증거로 채택했다. `python3 -m py_compile` 통과.

## Gaps (명시적 미검증)

- `--apply` 실증은 /tmp 사본에서만(원본 증거 보존).
- 어휘는 런타임 수집(git-tracked 산문) — 시점·브랜치에 따라 후보 빈도 변동.
- 근원은 관측 귀속까지 — 백엔드 내부 메커니즘 규명은 레인 권한 밖.
- 실행 검증은 py_compile + selftest + 데모가 전부(별도 테스트 스위트 없음 — Class B 카드 지시 범위).

## Residual-risk (잔여 위험)

- d=1-유일 자동수리도 짧은 단어에서 오수리(실측 2건) — 무검토 apply 는 정상 단어를 갈 수 있다.
- 클린 문서에서도 2.1% 플래그 — 무차별 게이트화 시 거짓 경보.
- 신조어·고유명은 계속 플래그; 오염형이 코퍼스 빈도에 편입되면 자기정합화 위험(대상 파일 vocab 제외로 완화).
- **근원은 살아 있다** — 본 유틸은 완화 장치이지 치료가 아니다.

## 운영자 보고 항목 (근본 해소 — 레인 범위 밖)

1. **근본**: GLM 채널 출력 토큰 스트림의 자모 슬롯 치환 → `/moai:feedback` 보고 권고. 재현: GLM 세션에서 한국어 산문 출력 후 `python3 scripts/jamo_integrity.py check` .
2. **추적 문서 실오염 1건**(git-workflow-doctrine.md:237 `처리르`) — 사용자 지시문 인용부라 원문 정합성 판정 후 수리 여부 결정 필요(리드 처분).
   → **처분 완료(리드 승인 2026-09-29)**: 채널 오염 변형으로 분류(처리를 이 유일 d=1 근접형, 오염 채널 특성화와 동류) → 본 브랜치에서 기계 수리(python 코드포인트 치환 REU→REUL — 수정·커밋 메시지 모두 오염 채널 비경유)·수리 후 검출기 재판독으로 237행 플래그 소멸 확인. 인용 원문 정합성은 오염 채널이 원문을 훼손한 것으로 특성화된 이상 복원이 부합한다(리드 판정문).
