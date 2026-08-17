# paste-ready: unified-board-design-20260817.md 소급 대조표 (t107)

아래 블록을 primary 체크아웃의 `.moai/reports/unified-board-design-20260817.md`
에서 `## 8. 리스크` 절 바로 앞(§7 표 다음)에 삽입한다. lane 세션이 워크트리
격리로 primary 편집이 가드 차단되어 리드 적용으로 넘김 — 내용은 아래 실측 그대로.

- 큐 실측: `moai todo` 2026-08-17 (live: t108·t113·t59)
- git 실측: `git log --oneline --grep 'merge: tNN'` (t109=2c70e7aed, t56=400dde787, t55=1ea829c76, t85=162f74d99, t58=b8a25b62f / t84=0건)

---8<--- 여기부터 ---

## Card Cross-Check (카드 대조표)

§7 마일스톤→카드 대조 (t107 소급 적용 — 연결 검증을 사람 기억에서 큐·git 실측으로). 실측 시각: 2026-08-17, 큐 조회 `moai todo` + git `merge:` grep. 기계 검출: `moai graph build && moai graph query --milestones-no-card`.

| milestone | 내용 | card | 실측 |
|---|---|---|---|
| S0 | 접점 실측 | t109 | 머지 — 2c70e7aed 실측 보고서 |
| S1 | 모드 재명명 6→4 | t108 | 큐 queued ✓ |
| S2 | review 폐지 → 3단계 | t113 | 큐 queued ✓ |
| S3 | 세션 이름 역할 고정 | t56 | 머지 — 400dde787 |
| S4 | 보드 상태 디스크 파생 | t55 | 머지 — 1ea829c76 · D2 커버리지 미확인 |
| S5 | `-k N`에 `-f` 흡수 | t85 | 머지 — 162f74d99 · '확장' 범위 미확인 |
| S6 | 공장장 세션 `--chief` | [신규 발행 필요] | 원 주장 t84 — 미발급 (git grep 0건) |
| S7 | 메시지 포맷 + 용어집 | t58 + t59 | t58 머지 — b8a25b62f · t59 큐 queued ✓ |

마일스톤 8개 → 카드 8개 (S6 제외 7개 매핑, S6 신규 필요).

---8<--- 여기까지 ---
