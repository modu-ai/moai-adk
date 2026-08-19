---
title: moai memory
weight: 18
draft: false
new: true
added_in: "v3.1.1"
---

{{< new-badge v3.1.1 >}}

교훈 저장소(Lessons Protocol memory store)의 **건강 진단과 아카이브** 도구입니다. MoAI의 교훈은 `MEMORY.md` 색인과 토픽 파일들로 이루어지는데, 세션이 시작할 때 읽는 것은 색인이지 디렉터리가 아닙니다. 그래서 색인 줄 없이 파일만 쓰인 교훈은 **저장은 됐지만 다시는 회상되지 않는** 상태가 됩니다.

{{< callout type="info" >}}
**한 줄 요약**: `moai memory doctor`가 색인↔파일 불일치(고아 파일·빈 링크)와 토픽 파일 수 상한을 진단하고, `moai memory archive`가 지명한 파일을 아카이브로 접어 색인에서 내립니다.
{{< /callout >}}

## moai memory doctor

```bash
$ moai memory doctor            # 사람이 읽는 보고
$ moai memory doctor --json     # 구조화된 출력
$ moai memory doctor --dir <경로>   # 다른 저장소 진단
$ moai memory doctor --cap 80   # 상한을 바꿔 검사
```

진단 항목:

| 항목 | 뜻 |
|------|-----|
| 고아 토픽 파일 | 색인 줄이 없는 토픽 파일 — 회상되지 않는 교훈 |
| 빈 색인 링크 | 색인 줄이 가리키는 파일이 없음 — 읽기 실패로 남는 줄 |
| 토픽 파일 수 | 프로젝트별 상한(기본 50) 대비 현재 개수 |

`--dir`과 `--cap`은 검사 대상·기준을 바꾸는 옵션일 뿐, 파일을 고치지 않습니다. doctor는 진단만 합니다.

## moai memory archive

```bash
$ moai memory archive feedback-old-lesson.md
```

지명한 토픽 파일을 `memory/_archive/`로 옮기고 색인 줄을 내립니다. **삭제가 아닙니다** — 아카이브는 감사 기록을 보존합니다. 무엇이 낡았는지는 판단의 문제이므로 대상은 운영자가 파일 이름으로 직접 지목하고, 자동 선별은 없습니다. 상한(기본 50개)을 넘어가면 초과분을 아카이브로 접어 색인을 가볍게 유지하는 것이 이 도구의 주된 용도입니다.

## 관련 문서

- [컨텍스트와 메모리](/ko/claude-code/context-memory/memory) — Claude Code 메모리의 동작 원리
- [자가 진화](/ko/advanced/self-evolving) — 교훈이 규칙 승격 제안으로 올라가는 흐름
- [결정 메모리](/ko/advanced/decision-memory) — 라우팅 결정이 쌓이는 또 하나의 기억
