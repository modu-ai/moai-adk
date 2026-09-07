# t395 — 카드 전제 반증 (조사 결과)

카드: t395 · "SQLite 마이그레이션이 원본 `backlog.json` 을 지우지 않아 스테일 사본이 오답을 준다"
트리: `.claude/worktrees/t395` · 브랜치 `WT-stale-backlog-json` @ `ad272be20`
측정 시점: 2026-09-02 · 측정 대상: primary 체크아웃 `/Users/goos/MoAI/moai-adk-go/.moai/state/todo/`

## Claim

카드가 서술한 인과 — "마이그레이션이 원본을 남겼다" — 는 **거짓**이다.
`backlog.json` 은 마이그레이션이 안 지운 원본이 아니라, 마이그레이션 **4일 뒤에 새로 생성된** 파일이다.
카드가 서술한 **피해**(스테일 사본이 에러 없이 모든 질의에 답한다)는 그대로 참이다.

## Evidence

### E1 — 파일 생성 시각 (결정적)

```
$ stat -f '%N mtime=%Sm birth=%SB' -t '%Y-%m-%dT%H:%M:%S' .../backlog.json
.../backlog.json mtime=2026-08-31T21:13:05 birth=2026-08-31T21:13:05
```

birth == mtime == 2026-08-31T21:13:05. 마이그레이션 산출물 `backlog.json.migrated` 의 mtime 은
2026-08-27T23:01. 즉 **원본은 08/27 에 정상적으로 격리(rename)됐고**, 현재의 `backlog.json` 은
08/31 에 별개로 생성된 파일이다.

### E2 — 세 파일의 내용 대조

| 파일 | items | last_seq | t359 | t395 | `archived` 키 |
|---|---|---|---|---|---|
| `backlog.json` | 109 | 390 | 있음 | **없음** | **없음** |
| `backlog.json.migrated` | 97 | 329 | 없음 | 없음 | 없음 |
| `backlog.db` (정본) | 87 | 431 | 있음 | 있음 | archived 55행 |

`backlog.json` 의 last_seq(390) 가 `.migrated`(329) 보다 크다 — 마이그레이션 이후 상태의 스냅샷이다.
가장 최근 카드 added_at 은 `2026-08-31T10:36:50Z`, 최대 id 는 `t390`.

### E3 — 마이그레이션 격리 경로는 설계대로 동작한다

`internal/kanban/backlog_migrate.go:556-566` `quarantineLegacyBacklog` 는 원본을 삭제하지 않고
`.migrated` 로 **rename** 한다(REQ-TOSQ-014). 기존 `.migrated` 가 있으면 덮어쓰지 않고 그대로 둔다.
`internal/kanban/backlog_store.go:594-604` State D(db + json 공존)는 db 안의 in-flight 마커를 보고,
마커가 없으면 그 json 을 "다운그레이드용 export" 로 간주해 **손대지 않는다**. 이 두 경로 모두 설계대로다.

### E4 — 현재 바이너리의 유일한 `backlog.json` 작성자는 `export-json` 이고, 이 파일은 그 산출물이 아니다

비테스트 코드에서 canonical `backlog.json` 경로에 쓰는 곳은 `internal/cli/todo_export.go:74`
(`target := store.Path()`) 한 곳뿐이다. 그런데 `BacklogRecord.Archived` 의 태그는
`internal/kanban/backlog_store.go:192` 에서 `json:"archived"` — **omitempty 가 없다**. 따라서 현재
바이너리의 export 는 항상 `archived` 키를 쓴다. 이 파일에는 그 키가 없다.
archived 필드는 `3a0ce021c`(2026-08-28) 에 착지했으므로, 이 파일은 **08/28 이전 레코드 모양**을 쓴다.

## Baseline-attribution

전부 이번 실행에서 이 트리(primary 체크아웃의 라이브 상태 파일 + 워크트리 `ad272be20` 의 소스)에 대해 측정했다.
소스 인용은 `WT-stale-backlog-json` @ `ad272be20` 기준 file:line.

## Gaps (미검증)

- **08/31 21:13 에 이 파일을 쓴 주체를 특정하지 못했다.** 확인한 것: 현재 코드의 작성자 경로는
  `export-json` 하나뿐이고 그 출력 모양과 불일치(E4) · `.claude/hooks/` 와 `scripts/` 에 쉘 작성자 없음 ·
  `~/.moai/todo/**/backlog.json` 30개와 sha256 불일치(복사본 아님).
  남은 가설(미검증): 08/27~08/28 사이 빌드의 스테일 바이너리가 돈 export · 다른 루트의 큐 · 수작업 복사.
- t390 까지만 담고 t395(08/31 15:43 추가)를 빠뜨린 이유를 설명하지 못했다 — 파일 생성(21:13)이
  t395 추가(15:43)보다 **뒤**인데도 t395 가 없다. 라이브 db 를 읽어 쓴 것이 아니라는 뜻이다.
- 이 파일이 만들어진 뒤 라이브 db 가 오염됐는지는 검사하지 않았다(정본은 db 이고 State D 에서 db 가 이긴다).

## Residual-risk

- 주체를 특정하지 못했으므로 **재발을 막을 수 없다**. 파일을 한 번 치우는 정책(카드 범위 (1))만으로는
  같은 작성자가 다시 만들면 원위치다. 카드 범위 (2)(읽는 쪽이 스테일임을 알 수 있어야 한다)가
  주체와 무관하게 유효한 유일한 방어선이다.
- 실제 피해는 **사람/에이전트가 파일을 직접 읽을 때** 발생한다. 현재 바이너리는 State D 에서 db 를
  정본으로 삼으므로 도구 경로는 안전하다 — lane-10 사고도 도구가 아니라 직독에서 났다.
- canonical 이름 `backlog.json` 이 (a) 레거시 큐 (b) 다운그레이드 export 대상 (c) 읽는 쪽이 큐라고
  가정하는 경로 셋을 동시에 맡고 있고, 디스크 위에 어느 쪽인지 말해주는 표시가 없다.

## 소관 판단 (레인 → 리드)

카드는 DROP 대상이 아니다. 서술된 인과는 틀렸지만 피해와 범위 (2) 는 온전하다.
다만 SPEC 의 조준점이 카드 문면과 달라진다 — "마이그레이션 뒷정리"가 아니라
**"canonical 경로의 json 은 정본이 아님을 읽는 쪽에 알리기"** 가 된다. 재조준 판단을 요청한다.
