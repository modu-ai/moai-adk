# iter-3 감사 재현 하네스 — 증거, 프로젝트 코드 아님

`plan-audit-iter3.md` 의 N2 검증(54 / 53 / 52)을 **repair-scratch 와 독립적으로** 재현한 자산이다.
배포되지 않고, 어디서도 import 되지 않으며, 빌드·테스트 대상이 아니다.

## 왜 fixture 를 새로 만들었나

`repair-scratch/` 의 `inside.md` · `outside.md` · `eol.md` 는 **수리 쪽이 만든 입력**이다.
그 입력을 그대로 쓰면 계수기만 독립적이고 입력은 공유되므로, 재현이 절반만 독립적이다.
그래서 iter-3 은 원본(`SPEC-AGENT-PARALLEL-OPT-001/acceptance.md`, `completed`)을 다시 복사해
248행 치환을 손으로 다시 만들었다. 계수기는 `iter2-scratch/counter.py`(감사 쪽 구현)를 썼다.

| 파일 | 무엇 |
|---|---|
| `orig.md` | 원본 사본 — 스윕 54 |
| `inside.md` | 248행 `` `AC-DCP-010` `` → `` `AC-DCP-010 [REF]` `` (토큰이 코드 스팬 **안**) |
| `outside.md` | 248행 `` `AC-DCP-010` `` → `` `AC-DCP-010` [REF] `` (백틱 개입) |
| `eol.md` | 248행 끝에 `[REF]` — AC-ACD-001 의 행 끝 뮤턴트 |

## 관측값

```
sweep(orig.md)          54
inside.md   adj         COUNT 53   (live=53 excluded=1 ambiguous=0)
outside.md  adj         COUNT 54   (live=54 excluded=0 ambiguous=0)
eol.md      adj         COUNT 54   (live=54 excluded=0 ambiguous=0)
eol.md      line        COUNT 52   (live=52 excluded=2 ambiguous=0)
inside.md   line        COUNT 52   (live=52 excluded=2 ambiguous=0)
```

원본은 편집하지 않았다 — 사본에만 치환했고, `git status --short` 에 추적 파일 변경이 없다.

## 린트 게이트가 여기 걸리면 무시한다

`iter2-scratch/README.md` 와 같은 이유다. 재현 하네스는 감사가 돌던 그 시점의 형태를
보존해야 값이 있고, 고치는 순간 "그때 무엇을 돌렸는가" 의 기록이 아니게 된다.
