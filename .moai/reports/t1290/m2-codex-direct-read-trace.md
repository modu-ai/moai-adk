# `codex -C` 로컬 지침 수신 추적

기준: `WT-agents-local-migration`의 `9dca1215d`에 루트·템플릿 `AGENTS.md`의 직접 Codex 조항을 수정한 작업 트리. `AGENTS.local.md`는 커밋 `0ab480d9d`의 678행 파일이다. 실행 인자: `codex -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1290 exec --sandbox read-only --json -o .moai/reports/t1290/m2-codex-direct-read-last.txt '프로젝트 지침이 요구하는 세션 시작 확인을 끝낸 뒤 READY라고만 답해.'`

`m2-codex-direct-read-events.jsonl`에서 명령 실행 항목만 읽은 결과:

| 순서 | 실행 명령 | exit | 출력 문자 수 |
|---|---|---:|---:|
| 1 | `pwd; rg --files ...; git rev-parse --short HEAD; git branch --show-current` | 0 | 2,034 |
| 2 | `cat AGENTS.local.md` | 0 | 37,061 |
| 3 | `wc -l AGENTS.local.md; rg -n ... AGENTS.md AGENTS.local.md` | 0 | 1,153 |
| 4 | `sed -n '240,450p' AGENTS.local.md` | 0 | 10,420 |
| 5 | `sed -n '1,120p' AGENTS.local.md` | 0 | 4,398 |
| 6 | `sed -n '121,239p' AGENTS.local.md` | 0 | 10,640 |
| 7 | `sed -n '451,678p' AGENTS.local.md` | 0 | 11,603 |

최종 답변 파일: `READY`. `wc -l AGENTS.local.md`는 678행이다. 첫 `cat`의 출력 문자 수가 파일 문자 수 37,061과 같고, 추가 조회가 1~678행 전체를 덮었다. 실행 초기에 스킬 설명 컨텍스트 예산 초과 진단은 있었으나 도구 읽기와 최종 응답은 exit 0이었다. 원본 JSONL은 106,534바이트이고 로컬에 보관한다. 지침 전문을 중복 커밋하지 않기 위해 이 추적만 증거로 싣는다.
