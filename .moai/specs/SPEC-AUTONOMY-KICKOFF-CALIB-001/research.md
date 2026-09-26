# research.md — SPEC-AUTONOMY-KICKOFF-CALIB-001 (v0.1.0)

이 카드의 research 는 세 덩어리다: 정독한 규율 원천(§1–2), 추출 술어의 실측 가능성 프루브(§3), 그 프루브가 보여준 것에서 뽑은 설계 결정 기록(§4). run-phase 측정 실행은 이 문서 어디에도 없다.

## §1. 읽은 원천과 가져간 것

| 원천 | 위치 | 가져간 것 |
|---|---|---|
| A1 0.5.2 | 본 워크트리 `.moai/specs/SPEC-AUTONOMY-CONTRACT-001/spec.md` (커밋 `25283ebf8` 계열, sync 마감 `e4ea8eb05`) | 결정자 값 집합 `human \| llm \| llm+jev`, `jev` 단독은 설정 오류 `kickoff_decider_jev_sole`(로드는 성공, 영수증 서명 경로만 거부), `workflow.autonomy.kickoff.{decider, jev_min_confidence}`(0.0–1.0, 기본 0.50), 영수증의 requested/effective decider·fallback 사유 필드 |
| A3 0.3.3 | t1236 워크트리 `.moai/specs/SPEC-AUTONOMY-GATE-REWIRE-001/spec.md` | A5 배정 문구(§A 표: 「이 SPEC 은 A5 를 기다리지 않는다. `jev_min_confidence` 값의 근거 측정만 A5 몫이다」·Out of Scope), R1–R5(REQ-GR-010), R2 의 5가지 대체 사유 중 `jev_low_confidence`가 이 값을 소비, 원칙 개정 연동(REQ-GR-013·025 — 개정·R3 해제·A1 임시 규칙 해제가 한 커밋) |
| `CLAUDE.local.md` §29 | 본 워크트리 사본 684–745행 (develop 판 — §0 정본 판별식) | 0.50 게이트는 16장 표본을 한국어 카드를 **영어로 옮겨** 잰 잠정값; 「한국어 원문 기준으로 재측정하기 전에는 이 수치를 근거로 인용하지 않는다」; fail-open 자세(키 없으면 한 줄 출력·exit 0); 3등급(되돌릴 수 없는 판정)은 Jev 호출 자체 금지 |
| `CLAUDE.local.md` §30 | 같은 사본 746–850행 — **전문 정독 완료** | t943 기각 전체: 124장 한국어 원문, 2-class 58.9% vs 상수 「항상 not-dead」 75.0%, `premise_dead` 정밀도 29.2% vs 기저율 25%, 게이트 상승은 채택률만 줄임, 오답 79%가 confidence 0.30 초과(최고 0.93), 영어 대조군 효과 ~10%p(결론 불변), [HARD] 재실험 금지·행 단독 인용 금지·기각은 「판정으로 쓰는 것」 한정(t943 §9 표시 용법은 허용) |
| t943 판정서 | primary 체크아웃 `.moai/reports/t943/verdict.md` (28.8KB, 읽기 전용) | 5섹션 판정서 형식의 실물, 라벨 출처 설계(레인 판정서 결론에서 읽음), 게이트 곡선 표 형태 |
| 리드 디스패치 | 카드 배차문(2026-09-26) | 묶음 조건 9건(모집단·한국어 원문·양성 대조·기준선·사전 등록·상한·§30 검증·코드 변경 금지·PR 부재) — 전부 REQ 로 흡수 |

## §2. 관측 — 리드 디스패치 전제와의 한 건의 갈림

리드 디스패치는 A1이 「develop 미병합」이라 서술했다. 본 워크트리 관측(2026-09-26)은 다르다:

```bash
git log --oneline -3 -- .moai/specs/SPEC-AUTONOMY-CONTRACT-001/
# e4ea8eb05 docs(SPEC-AUTONOMY-CONTRACT-001): backfill sync_commit_sha (t1234)
# f4e3d0731 docs(SPEC-AUTONOMY-CONTRACT-001): sync-phase artifacts — 3-phase close (t1234)
# 4a462733f docs(SPEC-AUTONOMY-CONTRACT-001): run-phase evidence and audit-ready signal (t1234)
ls internal/contract/   # ac_contract_test.go 등 존재, 최상위 커밋 fbd277ab7 (t1234)
```

`e4ea8eb05`는 본 브랜치(`WT-kickoff-decider-eval`, 로컬 develop 에서 분기) HEAD 에서 도달 가능하다. 어느 쪽이 맞든 run 진입 전제는 상태 서술이 아니라 `git merge-base --is-ancestor e4ea8eb05 HEAD` 명령으로 판정한다(REQ-CALIB-009) — 지금은 0이고, 실행 시점에 다시 읽는다.

## §3. 추출 술어 실측 프루브 (본 카드 귀속, 2026-09-26)

### §3.1 방법

세션 전사본 JSONL을 파싱해 assistant `tool_use`(`name == "AskUserQuestion"`) 블록의 header·question·option label에 `kickoff`(대소문자 무관)가 나타나는 블록을 후보로 세고, 같은 파일의 일치 `tool_use_id` `tool_result`로 응답 짝을 잰다. 코퍼스 루트 둘: A `~/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go*`(매칭 디렉터리 335개), B `~/.claude/projects/-Users-goos-MoAI-moai-adk-go*`(매칭 4개). 읽기 전용, 파일 수정 없음.

### §3.2 결과 (`probe/probe_output_20260926.txt`)

| 항목 | 값 |
|---|---|
| AskUserQuestion 블록(양쪽 루트 합계) | 907 |
| kickoff 언급 후보 | **148** |
| 응답 짝(`tool_result` 존재) | **148/148 = 100%** |
| 후보를 담은 세션 파일 | 87 |
| JSON 파싱 오류 / 파일 오류 | 0 / 0 |

같은 날 첫 실행에서 906·147이 나왔고 두 번째 실행에서 907·148이 됐다 — 코퍼스가 실시간으로 움직인다는 실측이고, 모집단 스냅샷 기록이 요구가 아니라 필연인 이유다.

### §3.3 관측된 어휘(라벨 매핑 씨앗)

- 질문 헤더(상위): `Kickoff`(39)·`진행 방식`(13)·`kickoff`(9)·`진행 모드`(9)·`Kickoff 승인`(5)·`카드 발행`(4)·`착수 승인`(4)·`킥오프 승인`(3)·`킥오프`(2). 헤더만으로는 게이트 라운드를 가를 수 없다 — 2단계 술어(REQ-CALIB-001 P1–P3)가 필요한 실측 근거.
- 선택지 라벨(상위): `보류`(61)·`승인 (권장)`(29)·`승인 — run 진입 (권장)`(9)·`승인 — run 진입`(7)·`run 진입 승인 (권장)`(5)·`반자율 (권장)`(5)·`지금은 보류`(4)·`중단`(4)·`승인 · 자율 (권장)`(4)·`자율`(4)·`둘 다 보류`(3).
- 응답 형식: `Your questions have been answered: "<질문>"="<선택 라벨>"` — 운영자 선택이 기계로 파싱된다.
- 복합 응답 실물: `run 미착수 3장의 Implementation Kickoff 승인을 어떻게 낼까요?` → `3장 모두 승인 (권장)`; 부분 승인형 `t531·t538만 승인, t359 보류`.
- **보존 한계(감사 D9 인정)**: 같은 날 첫 실행(906블록·147건)의 원 출력 파일은 보존하지 못했다 — 남은 것은 저자 증언뿐이고, 이후 실행부터 원 출력을 파일로 보존한다(`probe_output_20260926.txt` 가 그 첫 사례). 코퍼스가 살아 움직인다는 결론 자체는 두 실행의 대조가 아니라 최종 스냅샷의 실측만으로도 성립한다.

### §3.4 프루브 스크립트 본문 (0.1.0 저작 시점 사본 — 재현용)

run-phase 추출기는 이 스크립트를 `.moai/reports/t1244/extract/` 로 확장해 1·2단계 술어를 전부 적용한다. 재현: 같은 내용을 저장하고 `python3 <경로>` 로 실행(수 분 소요, 읽기 전용). **정본은 실행 파일** `.moai/reports/t1244/probe/kickoff_extract_probe.py` 다 — 0.2.1 시점에 리드가 unused import 제거 등 사소한 수리를 했고 출력은 동등하므로, 아래 사본과 실행 파일의 바이트 차이는 재현에 영향을 주지 않는다(감사 D8 — 「전문」 주장을 「시점 사본」으로 정정).

```python
#!/usr/bin/env python3
"""t1244 plan-phase probe: feasibility of the kickoff gate-round extraction predicate.

Read-only. Counts AskUserQuestion tool_use blocks mentioning kickoff and their
tool_result answers, across both corpus roots. Prints a bounded summary.

Card t1244 (SPEC-AUTONOMY-KICKOFF-CALIB-001) plan-phase evidence.
"""
import json
import glob
import os
import collections
import sys

ROOTS = [
    os.path.expanduser("~/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go*"),
    os.path.expanduser("~/.claude/projects/-Users-goos-MoAI-moai-adk-go*"),
]


def round_text(tu):
    """Joined header/question/label text of an AskUserQuestion tool_use input."""
    parts = []
    q = tu.get("input", {}) or {}
    for qq in q.get("questions", []) or []:
        parts.append(str(qq.get("header", "")))
        parts.append(str(qq.get("question", "")))
        for o in qq.get("options", []) or []:
            parts.append(str(o.get("label", "")))
    return " ".join(parts)


stats = collections.Counter()
files_with_auq = 0
files_with_kickoff_round = 0
header_counter = collections.Counter()
label_counter = collections.Counter()
samples = []

for root in ROOTS:
    dirs = sorted(glob.glob(root))
    stats["dirs"] += len(dirs)
    print("root %s -> matched dirs: %d" % (root, len(dirs)))
    for d in dirs:
        for fp in glob.glob(os.path.join(d, "*.jsonl")):
            try:
                tool_uses = {}
                n_auq_file = 0
                with open(fp, "r", encoding="utf-8", errors="replace") as f:
                    for line in f:
                        line = line.strip()
                        if not line:
                            continue
                        if '"AskUserQuestion"' not in line and '"tool_use_id"' not in line:
                            continue
                        try:
                            rec = json.loads(line)
                        except Exception:
                            stats["parse_errors"] += 1
                            continue
                        msg = rec.get("message", {}) or {}
                        content = msg.get("content")
                        if not isinstance(content, list):
                            continue
                        for c in content:
                            if not isinstance(c, dict):
                                continue
                            if c.get("type") == "tool_use" and c.get("name") == "AskUserQuestion":
                                n_auq_file += 1
                                stats["auq_total"] += 1
                                txt = round_text(c)
                                if "kickoff" in txt.lower() or "kick off" in txt.lower():
                                    stats["kickoff_rounds"] += 1
                                    tool_uses[c.get("id")] = {
                                        "answered": False,
                                        "answer": None,
                                        "text": txt[:500],
                                        "ts": rec.get("timestamp"),
                                        "file": fp,
                                    }
                                    if len(samples) < 2:
                                        samples.append(tool_uses[c.get("id")])
                                    q = c.get("input", {}) or {}
                                    for qq in q.get("questions", []) or []:
                                        header_counter[str(qq.get("header", ""))[:40]] += 1
                                        for o in qq.get("options", []) or []:
                                            label_counter[str(o.get("label", ""))[:40]] += 1
                            elif c.get("type") == "tool_result":
                                tid = c.get("tool_use_id")
                                if tid in tool_uses and not tool_uses[tid]["answered"]:
                                    tool_uses[tid]["answered"] = True
                                    tool_uses[tid]["answer"] = str(c.get("content"))[:400]
                                    stats["answers_seen"] += 1
                if n_auq_file:
                    files_with_auq += 1
                if tool_uses:
                    files_with_kickoff_round += 1
            except Exception:
                stats["file_errors"] += 1

print("== t1244 kickoff extraction probe ==")
print("AskUserQuestion tool_use blocks (both roots): %d" % stats["auq_total"])
print("blocks mentioning kickoff (header/question/label): %d" % stats["kickoff_rounds"])
print("matching tool_result answers seen: %d" % stats["answers_seen"])
print("files containing >=1 AskUserQuestion: %d" % files_with_auq)
print("files containing >=1 kickoff round: %d" % files_with_kickoff_round)
print("json parse errors (skipped lines): %d" % stats["parse_errors"])
print("file errors: %d" % stats["file_errors"])
print("-- top question headers in kickoff rounds --")
for h, n in header_counter.most_common(12):
    print("  %5d  %s" % (n, h))
print("-- top option labels in kickoff rounds --")
for h, n in label_counter.most_common(15):
    print("  %5d  %s" % (n, h))
print("-- sample rounds (first %d found) --" % len(samples))
for s in samples:
    print("  file:", s["file"])
    print("  ts:", s["ts"])
    print("  answered:", s["answered"])
    print("  question text[:500]:", s["text"].replace("\n", " ")[:500])
    print("  answer[:400]:", str(s["answer"]).replace("\n", " ")[:400])
    print("  ----")
```

### §3.5 매핑 선결 규칙의 작업 예시 (감사 D2 — 관측 실물 기반)

REQ-CALIB-002 의 선결 규칙 M1–M4 를 프루브가 관측한 실물 라벨에 적용한 결과 — 측정 전에 기대 출력을 고정해 둔다.

| 관측 라벨(verbatim) | 적용 규칙 | 질문 단위 1차 | 카드별 분해 |
|---|---|---|---|
| `승인 — run 진입 (권장)` | M1 단일 적중(승인 어휘만) | `approve` | — |
| `지금은 보류` | M1 단일 적중(보류 어휘만) | `hold` | — |
| `t538 먼저 AC 실측 후 재판단` | M1 단일 적중(조건 제시) | `modify` | — |
| `3장 모두 승인 (권장)` | M3 복합·전 세그먼트 동일 값 | `approve` (`compound: uniform`) | 귀속 가능 카드 전부 `approve`, 미귀속 `unknown` |
| `t531·t538만 승인, t359 보류` | M3 복합·세그먼트 갈림 | **`modify`** (`compound: split`) | t531=`approve`, t538=`approve`, t359=`hold` |
| 가상: `승인하되 예산 항목은 보류` | M2 동시 어휘·분절 없음 | `approve`(텍스트 순서 첫 동사) + `mixed` 표지 | — |
| 가상: 다중 선택 응답 `승인 — run 진입, 보류` | M4 다중 선택 | 첫 나열 선택의 값 `approve` + `multi_select: true` | — |

`mixed`·`compound: split`·`multi_select` 항목은 밴드 분석에서 제외하지 않는다 — 질문 단위 1차 라벨이 존재하는 한 표본에 남고, 표지가 판정서의 해석 경계를 만든다. 카드 단위 보고는 보조(NC-2)다.

## §4. 설계 결정 기록

| 결정 | 내용과 근거 |
|---|---|
| D1 분석 단위 = 질문 | 게이트 호출은 승인 질문에 진행 방식 질문을 동봉하는 일이 잦다(관측 헤더 `진행 방식` 13·`진행 모드` 9). 호출을 단위로 삼으면 서로 다른 두 결정이 한 레코드에 뭉개진다 — 질문을 단위로 하고 동봉 질문은 `mode_context` 원문 보존(REQ-CALIB-002) |
| D2 라벨 공간 4값 | 운영자의 관측 가능한 결정에서 출발한다 — 승인/보류 어휘만으로는 「조건부 재판단」형(관측 `t538 먼저 AC 실측 후 재판단`)이 `hold`와 `approve` 어느 쪽으로도 거짓 매핑된다. `modify`를 독립값으로 둔다. 매핑 불가는 판사 실행 전 일괄 재결 — t943 의 사후 게이트 조정 금지 교훈과 같은 방향 |
| D3 판사 과업 = 결정 예측 | 판사에 운영자 응답을 주지 않고 라벨+confidence로 예측하게 한다 — `llm+jev` 교차 확인에서 Jev 가 할 일(사람 결정에 대한 두 번째 신호)의 축소 모형이다. 카드 문맥을 주지 않는 1차 배치는 판사에게 불리하지만, 문맥 추가는 그때그때 변할 수 있는 자유도라 사전 등록에 못 넣는다 — 못 넣는 자유도는 측정에 넣지 않는다 |
| D4 밴드 기준 수치 | n≥20(§30 M2 5건·4건 밴드의 인용 불가 실측), +10%p(상수 기준선을 「의미 있게」 넘는 최소선 — 측정 전 판단값, NC-1), wrong-automation ≤10%(이 도메인의 최악 오류는 사람이 보류했을 결정의 자동 승인이다). 세 값 모두 측정 전 고정이고 측정 후 변경 불가 |
| D5 최저 자격 밴드 채택 | 문턱의 목적은 최대 자동화 하에서의 안전이므로 기준을 넘는 밴드 중 가장 낮은 t가 후보다 — 높은 t를 고르는 것은 근거 없는 보수성이다 |
| D6 양팔 무효 규칙 | 파이프라인 대조(계측기 검증)와 판사 대조(판사 검증)는 서로를 대신하지 못한다 — 어느 팔에서도 재현되지 않는 대조는 두 팔을 가를 수 없다는 기존 교훈의 적용 |
| D7 `scripts/jev` 미추적 명기 | 2026-09-26 `git ls-files scripts/jev/` 카운트 0 — primary 체크아웃의 로컬 파일이다. run 기록에 경로·미추적 상태를 적고, 부재 시 측정 불가 판정서로 닫는 갈래를 사전에 둔다(REQ-CALIB-008) |
| D8 t943 비재실행의 근거 | 재실험 금지는 같은 실험의 반복을 막는다. 모집단(운영자 Kickoff 결정 vs 카드 전제 소멸)·과업(결정 예측 vs 전제 판정)·라벨 공간이 다르므로 A5 는 재실험이 아니다 — 반면 측정 규율(기준선 나란히·양성 대조·한국어 원문·기준 사전 고정)은 그대로 옮긴다. §30 인용에는 상시 상수 기준선 행이 따른다(REQ-CALIB-010) |

## §5. 열린 표지 (본문 결정 + 기본값 — 운영자가 측정 전에 바꿀 수 있다)

- **NC-1** 밴드 기준 수치(+10%p 등)는 측정 전 판단값이다 — 근거는 D4. 기본값 유지. 바꾸면 HISTORY 에 기록하고 측정 전에만 가능하다.
- **NC-2** 복합 응답의 카드 단위 보고는 보조다 — 질문 단위가 1차 판정 기준(D1·REQ-CALIB-002). 기본값 유지.
