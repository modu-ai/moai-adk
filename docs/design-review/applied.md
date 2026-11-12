# Space 적용 기록 (EP-G #2159)

설계 페이지는 ChatGPT Space에 있고 에이전트는 Space를 편집하지 않는다. 적용은 리더가 ego-browser로 하고, 페이지마다 적용 전에 사용자 확인을 받는다(결정 01:2x·23:xx(11-09), `docs/decisions-2026-11-08.md`). 이 파일은 그 적용의 체크표(AC-1)와 sha 기록(AC-2)이다.

- 상태 어휘: `not_applied`(아직 적용 안 함) · `applied`(적용하고 사본을 다시 받아 sha256을 적음).
- `applied` 행은 사용자 확인 근거(Issue 댓글 id)와 적용 뒤 다시 받은 Space 사본의 sha256(64자리)을 함께 적는다. `not_applied` 행은 두 칸을 `—`로 둔다.
- `수정안 sha256`은 적용 기준이 되는 수정안 파일의 sha256이다. 수정안을 고치면 이 값도 함께 고친다(`tests/space_applied_record_test.py`가 대조한다). 이미 `applied`인 행의 수정안이 바뀌면 그 페이지는 다시 적용 대상이다.

## G2 #2207 — 하위 13편

수정안 기준: 2026-11-12 tip `325cc2bf03` 재확인본(각 수정안의 "현재 tip 재확인" 절). tip `263a025cbb` 병합 때 PR #2534로 낡은 Codex config 문장 2곳(group3-487fa473 6번, group3-83f162ef 2번)을 고쳤고, 바뀐 3편은 "tip `263a025cbb` 병합 재확인" 절을 따로 둔다.

| 수정안 | Space 페이지 | 항목 | 수정안 sha256 | 사용자 확인 | 적용 사본 sha256 | 상태 |
| --- | --- | --- | --- | --- | --- | --- |
| `fixes/group2-0259e59572808191bdb8fd302e86bd54.md` | `page_0259e59572808191bdb8fd302e86bd54` 협업 실행 소유권과 복구 | 7 | `4b6b96665f23c58fe88a4c38cb9b5b99e94eafa7c36a0c5ed6b1541155260549` | — | — | not_applied |
| `fixes/group2-0d4538a27fb88191a19ba261bd5a3442.md` | `page_0d4538a27fb88191a19ba261bd5a3442` SYNC와 Issues 보고 | 8 | `d9910d7c687f41f633b4033c0d24912cec6d61b066a65f59d746b86bd0932c5c` | — | — | not_applied |
| `fixes/group2-136c8203bf2881919eb1f543f72f87ea.md` | `page_136c8203bf2881919eb1f543f72f87ea` Core Brain 상세 설계 | 8 | `69af57c04009d0f67fdb41e40e8f577b1b5c32c113d487a3b7c3b9b7e7c8037e` | — | — | not_applied |
| `fixes/group2-c4f994be97548191b41945af0713ecc9.md` | `page_c4f994be97548191b41945af0713ecc9` 검증 감사와 보안 | 7 | `501ca5122eff5d17213e9df044471f82503d7e3e2e1d1bf46c60ac82b55702e0` | — | — | not_applied |
| `fixes/group2-e3612b0b07888191ae58c58d66bdff4a.md` | `page_e3612b0b07888191ae58c58d66bdff4a` 판단 backend와 이미지 생성 | 7 | `5c1290b67fb9755bd7d28b36bb28489c29e303758d2f07375a74c0103e6cf48a` | — | — | not_applied |
| `fixes/group3-25fa9b23.md` | `page_25fa9b23c690819180a1351d6da22e23` 릴리스와 Desktop | 7 | `e6a7661a9733b129b13305a6a9fb0bb61e966dd47d1e8502a62ba5685338e6b7` | — | — | not_applied |
| `fixes/group3-3d917728.md` | `page_3d91772843448191a46eedf202be99cc` RUN과 TDD DDD | 6 | `767d4dd258663430730b3c9e42b065cb4dc9e8fa531477a0f22b8517b73ff4c8` | — | — | not_applied |
| `fixes/group3-487fa473.md` | `page_487fa473b0308191b572f7dbc8640ee0` 파일 소유권과 worktree 수명 | 8 | `b3110ec1a8ec34bd8bc81b804c7cf8c2fcf9f0007b4bde2169f54fa63f9e2cdd` | — | — | not_applied |
| `fixes/group3-66acadc2.md` | `page_66acadc28bdc8191875556986da72e6b` RC1 GitHub 작업 구조 | 7 | `5db46649c21d3525ad92c37df14682fda03b89f0cb9af0f968107043db6e61e5` | — | — | not_applied |
| `fixes/group3-83f162ef.md` | `page_83f162ef14d08191917913738184f167` 홈 설치와 프로젝트 생성 | 8 | `7dc71a04b2ab6f68d684e2ba5be2edd9f896b78e00d1455f9170a248b708ced3` | — | — | not_applied |
| `fixes/group3-dff088ea.md` | `page_dff088ea648081918e2eb28ab4a3e77c` Host adapter와 hook 이식성 | 6 | `ea222e3f88a72439731e4310e04fb0dc834230d4db673002fa15d87d9ed1cbd8` | — | — | not_applied |
| `fixes/group3-e0874134.md` | `page_e087413484688191b4da76c7ddd96b2b` PLAN과 GEARS SPEC | 6 | `1190ac743d4cac0c075c20087180a8ce215d75df0647789e49bf7970cc83a86c` | — | — | not_applied |
| `fixes/plan-43faf222.md` | `page_43faf222414c81919de71f7d7830e26c` 온라인 우선 최종 계획 | 12 | `60d1a9984d3250e8efb9da2207374fe15d1fb3695fd5fdeb93f4a0a64bfc551e` | — | — | not_applied |

합계 97항목. 적용 순서와 페이지별 사용자 질문은 리더가 정한다.
