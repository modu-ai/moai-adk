# Native compact 요약 응답과 다음 이력 구조

## Claim

합성 환경에서 native /compact의 모델 요약 응답과 PostCompact summary 해시가 같았고, 다음 요청의 이력이 그 요약을 포함한 새 구조로 바뀌었다. 이 흐름을 정상 요약 요청으로 처리한 뒤 검증된 요약을 기준으로 이력을 재설정하는 설계의 구조 근거를 확보했다. 제품 연결 완료는 아니다.

| 여정 | 결과 | 시간 | 산출물 |
|---|---|---:|---:|
| seed → resume /compact → resume 새 질의 | PASS | 14.24초 | 4 |

## Evidence

```sh
python3 .moai/reports/t649/appserver-redesign/probe/compact-history-shape.py
```

```text
seed exit=0 requests=1 result=SEED_ASSISTANT_BLUEBIRD42
compact exit=0 requests=1 compact_boundary observed
followup exit=0 requests=1 result=FOLLOWUP_DONE
success=true errors=[]
```

compact 요청은 seed의 user 메시지를 canonical JSON 비교에서 그대로 유지했다. 이력의 system 메시지는 content 배열에서 문자열로 바뀌었지만 연결한 text는 바이트상 같았다. 최상위 system 배열의 canonical JSON은 같았다. 기존 assistant text SEED_ASSISTANT_BLUEBIRD42도 유지됐으며 ephemeral cache_control이 붙었다. 마지막 user 메시지로 요약 지시가 추가됐다.

모의 API가 반환한 요약은 COMPACT_SUMMARY_BLUEBIRD42였다. PostCompact compact_summary는 길이 26이며 다음 SHA256과 정확히 같았다:

```text
6fb58f445a90cbc893e201b16fcf8c876cff5f92937e12dead91ed11ebe6463f
```

후속 요청 구조:

```text
messages[0] role=user
 content[0] attribution reminder
 content[1] continuation wrapper + exact summary + transcript reference + continuation instructions
 content[2] local-command caveat
 content[3] /compact command metadata
 content[4] compact hook completion stdout
 content[5] NEW_QUERY_PINE73: recall the fixture marker.
messages[1] role=system, current deferred-tool/environment guidance
```

기존 assistant 문자열은 후속 요청에 없었다. 요약은 독립 메시지가 아니라 user content 배열 안의 wrapper text에 포함됐다. 새 질의는 별도 text block이다. 따라서 전체 배열을 단순 prefix append로 처리할 수 없다.

## Baseline-attribution

설치 Claude Code 2.1.269. 임시 HOME/CLAUDE_CONFIG_DIR/cwd, 합성 인증과 로컬 모의 API. 사용자 프로필·인증 파일을 읽지 않았다. hooks는 임시 설정의 시험 hook만 실행했다. 각 CLI 45초 제한과 프로세스 그룹 정리를 적용했고 오류가 없었다. 유료 호출은 0회다. 원시 요청 전체 대신 합성 messages/system과 구조 필드를 저장했으며 인증 헤더·metadata.user_id 값은 제외했다.

## Gaps

PreCompact, SessionStart(compact), PostCompact에는 같은 prompt_id가 있었다. HTTP 헤더 이름에는 prompt ID가 없었고 별도 직접 상관 필드는 확인하지 못했다. PreCompact 다음 첫 HTTP라는 순서만으로 압축 요청을 확정하면 안 된다. hook 인증·세션/agent 결속·단일 활성 compact 상태·실제 요약 해시 및 후속 wrapper 검증이 별도로 필요하다.

manual print/resume 합성 경로만 검증했다. interactive/auto/subagent 압축, 동시 요청, 실패·재시도, 비정상 summary와 악의적인 동일 문자열 삽입은 미검증이다. wire 원본 바이트 전체 대신 canonical JSON과 text 바이트를 비교했으므로 JSON 공백·키 순서 보존 주장은 하지 않는다.

## Residual-risk

고정 요약 응답을 사용했으므로 실제 모델의 요약 품질과 복잡한 블록 보존은 검증되지 않았다. 내용 패턴은 이 샘플의 관측이며 다른 버전의 공개 계약으로 일반화할 수 없다. summary hash 일치는 상태 결속을 돕지만 단독 인증 또는 압축 시작 증명은 아니다.

산출물: compact-history-shape.py, compact-history-shape-result.json, compact-history-shape-comparison.json, compact-history-shape-verdict.md.
