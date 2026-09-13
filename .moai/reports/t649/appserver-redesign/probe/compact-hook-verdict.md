# AS4: 네이티브 압축 hook 신호 실증

## 판정

| 여정 | 상태 | 시간 | 산출물 |
|---|---|---:|---:|
| seed → PreCompact → 압축 HTTP → SessionStart(compact) → PostCompact | PASS | 3.19초 | 4 |

## 공식 근거

공식 hooks reference https://code.claude.com/docs/en/hooks 에서 PreCompact와 PostCompact의 manual/auto matcher 및 SessionStart의 compact source를 확인했다. 내려받은 공식 문서는 compact-hook-official-docs.md에 보관했다. 현재 문서는 PreCompact가 압축 전에, PostCompact가 압축 완료 후 호출된다고 설명한다. PostCompact 출력은 압축 결과를 변경하지 못한다. 문서 확인은 지원 계약의 근거이고 실제 설치본 동작은 아래 실증으로 확인했다.

## 실행과 관측

```sh
python3 .moai/reports/t649/appserver-redesign/probe/compact-hook.py
```

실제 Claude Code 2.1.269를 두 번 실행했다. 첫 실행은 짧은 합성 대화, 두 번째는 같은 합성 session_id로 --resume /compact였다. 두 번 모두 exit 0이고 실제 모델 서버 호출은 없었다.

동일 시계의 time_ns로 정렬한 관측 순서:

```text
1 seed HTTP request
2 PreCompact(trigger=manual)
3 compact HTTP request
4 SessionStart(source=compact)
5 PostCompact(trigger=manual)
```

세 hook의 session_id는 동일했다. transcript_path 필드 존재는 true였지만 파일 내용과 경로는 저장하지 않았다. PostCompact의 compact_summary 역시 읽어 기록하지 않았다. compact 실행의 CLI 출력에서는 system/compact_boundary 이벤트도 확인했다.

## 격리

임시 HOME/CLAUDE_CONFIG_DIR/cwd, 합성 API 키, loopback 모의 서버를 사용했다. 빈 프로필에 --settings로 시험 hook 세 개만 등록했다. hook 명령은 해당 시험 스크립트를 --hook으로 실행하며 허용 필드만 임시 파일에 추가한다. 사용자 hooks와 인증 파일을 사용하지 않았다. CLI 실행 상한은 각각 45초, hook 상한은 5초, 종료 시 프로세스 그룹을 정리했다. 오류는 없었다.

## 설계 의미와 한계

프롬프트 문구를 추측하지 않고 압축 HTTP 이전에 명시적인 네이티브 신호를 받을 수 있음이 확인됐다. 다만 hook payload의 session_id만으로 외부 브리지 호출을 인증할 수는 없다. 제품 연결 시 세션별 토큰/IPC 권한, 재전송 방지, 만료, 부모·서브에이전트 세션 구별을 별도 검증해야 한다.

이 실증은 print/resume 모드의 manual compact다. 실제 interactive UI, 자동 압축, subagent 압축, 실패·취소·재시도는 미검증이다. CLI 이벤트는 실행 종료 후 파싱했으므로 PostCompact와 compact_boundary 출력의 정밀 선후관계까지 비교하지 않았다. after-HTTP hook 순서만 측정했다. 구현을 변경하거나 압축을 인터셉트하지 않았다.

산출물: compact-hook.py, compact-hook-result.json, compact-hook-official-docs.md, compact-hook-verdict.md.
