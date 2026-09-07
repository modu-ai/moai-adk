# 재닫기 증거 사본 (F15 대응)

`progress.md` §E.4 가 인용하는 증거는 원래 `.moai/state/verify/t502-resync/` 에 있었다. 그 경로는 `.gitignore:315`(`.moai/state/`)로 **추적되지 않아**, 워크트리를 폐기하는 순간 인용이 해소되지 않는다 — 2차 감사 F15.

이 저장소 규율은 *「인용한 증거 경로가 해소되지 않는 주장은 귀속되지 않은 주장」* 이므로, 폐기 전에만 가능한 조치로 사본을 추적되는 곳에 둔다. **F15 자체를 닫는 것이 아니다** — F15 는 `.moai/state/` 를 인용하는 관행 전반에 대한 계통적 지적이고 소관 미배정으로 남는다. 여기서 하는 것은 이 카드 자신의 기록을 잃지 않는 것뿐이다.

| 파일 | 내용 | 주의 |
|---|---|---|
| `gotest.txt` | `go test ./internal/cli/... ./internal/codexwiring/...` — `GOTEST_EXIT=0`, `ok` 18줄 | `internal/cli` 만 새로 돌고 나머지 17개는 `(cached)`. 변경 패키지만 재실행된 정상 형태이며, 2차 감사가 `-count=1` 전량 재실행으로 같은 판정을 냈다 |
| `e2e-copy.txt` | `e2e-verb.sh copy` — `marker=1 → marker=0`, `result=MATCH` ×2, mode 644 유지, 재실행 후 엔트리 1개 | 실사용 `~/.codex/config.toml` sha256 전후 동일 |
| `ac-testlist.txt` | `go test -list` — AC 명명 테스트 존재 확인 | 셀렉터가 0개를 골라도 러너는 `ok` 를 찍으므로, 이 목록이 그 초록의 전제다 |
| `speclint*.txt` | `moai spec lint` 3회(편집마다 재실행) | 각 49바이트 = `No findings` 한 줄 |
| `vet.txt` | `go vet` — **0바이트** | **빈 파일이 곧 증거다.** 다만 빈 출력만으로는 「vet 이 침묵했다」와 「vet 이 못 봤다」가 구분되지 않는다. 2차 감사가 고의 결함 패키지에 같은 vet 을 돌려 `fmt.Printf format %d has arg … of wrong type string`, exit 1 을 얻어 **대조군을 세웠다** — 그 대조가 `sync-audit-round2.md` 에 있다. 이 파일만 떼어 읽으면 아무것도 주장하지 않는다 |

원본 경로와 이 사본은 내용이 같다(`cp`, 편집 없음). 판정 본문은 `../sync-audit.md`(1차)와 `../sync-audit-round2.md`(2차)에 있다.
