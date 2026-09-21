# t484 — t410 후속 4항목(F1·F5·D4·D6) 확인 판정서

- **측정 트리**: `a825183dd` (로컬 develop 팁, 브랜치 `WT-t410-followups`, 본 카드 자체 커밋 0, 추적 파일 수정 0)
- **측정일**: 2026-09-04
- **한 줄 결론**: 라벨 넷은 서로 다른 실제 항목이며, **넷 모두 아직 산다**. 전부 SPEC 문서 층이라 수리 경로는 manager-spec 하나로 수렴한다.

## Claim

| 항목 | 판정서 등급 | 생존 여부 | 실체 |
|---|---|---|---|
| **F1** | SHOULD-FIX / blocking | **생존** | 소유권 교차 — run-phase(manager-develop)가 다른 SPEC(`SPEC-ERA-H3-NARROWING-001`)의 본문 표면(HISTORY 행 + `version:`/`updated:`)을 편집. 근본 원인은 AC-DCB-007이 금지 표면 편집을 run-phase에 직접 지시한 설계 |
| **F5** | MINOR / optional | **생존** | 모양 A(`<ID>:` 접두)는 콜론 뒤 본문을 전혀 제약하지 않는다 — close 커밋 본문의 비-close `<ID>:` 줄이 무죄 방면을 낼 수 있는 **설계된 잔여 위험**인데 SPEC 문서에는 미기재(코드 주석에만 존재) |
| **D4** | 문서 층 / 무해 | **생존** | REQ-DCB-002 문면("1차 워크가 `completed`도 terminal도 내지 못한 동안 조회")이 오류 갈래까지 주장 — 실제 구현은 오류 시 ① 블록 도달 전 `continue` |
| **D6** | 문서 층 / 무해 | **생존** | plan.md §A의 Tier 파일 수 판정에 **모집단 정의**(소스 파일 기준)와 **범위 안 초과 시 동작**이 없음 |

## Evidence — 항목별 현재 트리 재측정 (트리 `a825183dd`, 본 실행)

### F1 — 착지된 위반 기록 + 설계 원인 둘 다 현존

- 착지된 위반: `.moai/specs/SPEC-ERA-H3-NARROWING-001/spec.md:4` `version: "0.5.1"`, `:27` HISTORY 행 `| 0.5.1 | 2026-09-03 | manager-develop | ...` — t410 병합 diff(`git diff 71346d754 236fc0d1e`)로 편집 경로 확인.
- 설계 원인: `.moai/specs/SPEC-DRIFT-CLOSE-BODY-001/spec.md:171-186` AC-DCB-007이 run-phase에 직접 편집을 지시하며, **같은 AC의 :186이 인용한 규약(`spec-frontmatter-schema.md` § Non-transition frontmatter corrections)은 소유자가 manager-spec이라고 명시** — 판정서가 "정합성 결함"으로 단정한 모순이 문자 그대로 남아 있다.

### F5 — §5.2 에 보수적 방향만 기재

- `spec.md:240-244` (§5.2): 반대 방향 오류를 "두 모양 술어가 겨눈다"로 서술하나, **모양 A 꼬리 무제약** 잔여(판정서 F5의 핵심)는 없음. 코드 주석(`drift_index.go:432` 부근 "No text predicate separates…")에만 존재. 판정서 실측(코퍼스 18줄 중 실오해제 0건)도 SPEC에 미반영.

### D4 — 코드와 문면의 어긋남 현존

- `internal/spec/drift.go:217-221`: `gitStatus, err := inMemImpliedStatus(...); if err != nil { continue }` — `continue`가 ① 블록(본문 조회) **앞**에 있음(본 트리 직독).
- `spec.md:88` REQ-DCB-002: 문면이 오류 반환까지 포괄. 원장(`run-evidence.md:464-482`)의 수리 스케치: "REQ-DCB-002 문면을 '상태를 **반환했고**'로 좁히고 오류 경로를 §4에 한 줄".

### D6 — plan.md 에 모집단·초과 동작 부재 현존

- `plan.md:20-27`: 예상 4파일/~210 LOC 만 적히고, 파일 수 판정의 **모집단 정의**(원장 `:502` — 소스 파일 기준이어야 한다는 판정)와 **범위 안 초과 시 동작**(원장 `:504` — "원장에 기록하고 Tier 재판정을 리드에게 blocker")이 문서에 없음.

### 전제 무효화 확인 (리드 [HARD] 2번항)

```
git log --oneline 79c7a0e2f..a825183dd -- \
  .moai/specs/SPEC-ERA-H3-NARROWING-001/spec.md \
  .moai/specs/SPEC-DRIFT-CLOSE-BODY-001/ \
  internal/spec/drift.go internal/spec/drift_index.go
→ 빈 출력 (t410 이후 아무 커밋도 대상 표면을 건드리지 않았다)
```

D 계열이 "이후 카드가 구조를 바꿔 무효됐을" 가능성은 측정으로 소멸.

## Baseline-attribution

모든 좌표는 본 실행에서 워크트리 `.claude/worktrees/t484` (HEAD `a825183dd`)에서 직독했다. 원장·판정서 인용분(`run-evidence.md:464-505`, `verdict.md:126-205`)은 t410 워크트리 사본(동일 커밋에 커밋된 파일)에서 읽었으며, 대상 표면이 t410 이후 불변임이 위 git log로 담보되므로 인용 유효.

## 이미 소비된 것 / 이 카드 범위 밖

- **F2**: t410 sync가 이미 소비(`progress.md:128` §E.4.1 정정 — sync-audit F2 상환). 잔여 아님.
- **index.lock 캡처**(원장 §절차 사고 기록): 이미 소비된 산출물이며 **t485(lane-15) 소관** — 이 카드에서 미접촉.
- **F3·F4·F6·F7·F8**: 판정서 재량(optional) 항목 — 이 카드 배차 밖. 잔여로 남는다(리드 참고용).

## 처분 제안 (disposition)

**권고: 단일 manager-spec 위임으로 4건을 한 번에 상환.** 전부 SPEC 문서 층이라 위임이 하나로 수렴한다.

1. **SPEC-DRIFT-CLOSE-BODY-001 — in-place amendment** (completed → in-progress, `amendment_of: self`, 스키마의 amendment 기계 장치 경유 — D4·F5·F1-②가 REQ/AC/잔여 **산문**을 고치는 편집이라 HISTORY 행만으로 가능한 비전이 정정 면제 범위를 넘는다):
   - **D4**: REQ-DCB-002를 "상태를 반환했고"로 좁힘 + 오류 경로 한 줄(원장 스케치). HISTORY D1 행의 "REQ 2개+§4" 대비 원장 스케치는 REQ-DCB-002+§4만 명명 — 제2 REQ(추정 REQ-DCB-006) 포함 여부는 manager-spec이 편집 확정.
   - **F5**: §5.2에 모양 A 꼬리 무제약 잔여 1줄 추가(판정서:205 "한 줄로 족하다" 그대로).
   - **F1-②**: REQ-DCB-007·AC-DCB-007을 "manager-spec 재위임을 거쳐 수행"으로 개정 — 재발 방지(판정서 Residual-risk #3이 가리키는 부하 수정).
   - HISTORY(`## Amendments`)에 카드 t484 + 본 판정서 경로 기록.
   - 이 카드의 sync에서 재종결(implemented → completed).
2. **SPEC-ERA-H3-NARROWING-001 — 비전이 정정** (HISTORY 행 + `version:`/`updated:`만, manager-spec 소유):
   - **F1-①**: 0.5.1 행의 채널 귀속 정정. **권고 방식: 0.5.2 정정 행 추가**(manager-spec 명의로 "0.5.1 편집은 manager-develop 채널로 수행됐으나 소유권 스키마상 manager-spec 표면 — 본 행으로 채널 귀속을 정정"). 0.5.1 행의 author 셀을 manager-spec으로 고치는 방식은 실제 편집자를 거짓 기록하게 되므로 비권장.
3. **원장 정합화 rider**: t410의 `run-evidence.md`는 역사 증거로 수정하지 않고, amendment HISTORY 행에 "D4·D6 보류 사유는 금지 독법이 옳았고 본 amendment가 약속된 manager-spec 재위임"을 기록해 정합화.
4. **래퍼 SPEC 불요**: 수정 자체가 두 대상 SPEC의 자체 기록 장치(amendment/HISTORY)로 추적되므로 t484 전용 SPEC은 만들지 않는다. 단순성 사다리 1·2단 판단. 카드 추적성은 HISTORY 문안의 `t484` 명시 + 본 판정서 경로가 담당.

**F1 ①/② 판정서가 "둘 중 하나"로 열어둔 것에 대한 권고: 둘 다.** ②만 하면 착지된 ERA-H3 기록의 귀속 결함이 남고, ①만 하면 AC가 다음 run에 같은 교차를 다시 지시한다(판정서 Residual-risk #3). 각각 한 줄·한 행 규모라 병행 비용이 미미하다.

## Gaps

- 4항목 외에 미확인한 것 없음(4/4 생존 판정에 필요한 표면 전부 직독).
- run-phase는 **미착수** — dispatch가 "확인 결과에 따라 run"을 지시했으나, (a) F1 ①/②는 판정서가 열어둔 결정이고 (b) amendment가 제2의 완료 SPEC(`SPEC-ERA-H3-NARROWING-001`) 상태를 건드리므로, 리드 확인 후 진입한다.
- 재량 F3·F4·F6·F7·F8은 측정도 제안도 하지 않았다(배차 밖).

## Residual-risk

- manager-spec 위임이 원장 스케치와 다르게 문면을 좁히면 D4의 제2 REQ 대상이 달라질 수 있다 — 원장 스케치(REQ-DCB-002+§4)와 HISTORY D1 행("REQ 2개+§4")이 이미 서로 다르게 읽히므로, 위임 프롬프트에서 이 간극을 명시해 소진해야 한다.
- amendment 중(상태 in-progress) 해당 SPEC은 drift 탐지가 재개된다 — sync 재종결 전까지 DRIFT 행이 보이면 정상이다.
- 본 판정서는 워크트리 내 증거로, 창 때까지 primary 미반출 상태다.
