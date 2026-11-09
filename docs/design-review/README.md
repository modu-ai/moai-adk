# docs/design-review/

2026-11-09 사용자 컨펌 종합 보고서와 설계 검토 산출물의 tracked 위치입니다
(#2137, milestone design-docs-r2). 원본은 untracked
`references/chatgpt-space/`에 보존되어 있으며 이 디렉터는 그 사본을
tracked로 둔 것입니다 — 각 파일의 sha256은 원본과 동일합니다.

## 파일 목록과 대조 기준 SHA

| 파일 | sha256 |
| --- | --- |
| `synthesis-improvement-report-2026-11-09.md` | `61376caa54b7e7841bb2187035f85d717d4490146b09a27d5aa65b367e13888a` |
| `review-page_5490c4ee-2026-11-08.md` | `57f2a0509d208469a43af571c6d53faa6f5a7ae19a7b056e4c5c76f9e8e63530` |
| `gap-analysis-rc1-vs-design-2026-11-08.md` | `69de8a9a4db4ce679291e4c65c25ce94ace6e218c12b41b0de02ab339fe261fd` |
| `review-group1-rc1-plan.md` | `afee7708c19617137ed0d52eb3247ceccee4324bd880be1abaf31fb9593f10c8` |
| `review-group2-core-brain-sync-owner.md` | `8b24a2c9c0334809ed7bd8a9ba0d5ef25774619316a0d776112dd137f8d49b73`(원본 `053be44348ac69dd5d7337bf7e4ff8ba433b9ff46f0eafd709fe355204d65cf9` — 개인 절대 경로 1건을 상대 경로로 치환, #2137 AC-2) |
| `review-group3-install-host-plan-run-release.md` | `bb67d400d73335e68e384eaa9c9b29a9b3925375e9fa8987c9192007fb44a06b` |
| `review-group4-past-designs.md` | `4e125ebd70b660d667515c5280d4757bf2c9efab249e2369eeb2c957d288b8ac` |
| `pages.tsv` | `04b50fc9bdc648ddcda621d406fc4dc66a81c9603496b83d18dcc64a0e8a00f5` |

## Space 페이지 id ↔ 제목 (`pages.tsv` 기준)

| 페이지 id | 제목 |
| --- | --- |
| `page_43faf222414c81919de71f7d7830e26c` | MoAI ADK 2026.11.0-rc.1 온라인 우선 최종 계획 |
| `page_136c8203bf2881919eb1f543f72f87ea` | MoAI ADK Core Brain 상세 설계 |
| `page_83f162ef14d08191917913738184f167` | MoAI 홈 설치와 프로젝트 생성 기능 계약 |
| `page_0d4538a27fb88191a19ba261bd5a3442` | MoAI SYNC와 Issues 보고 상세 설계 |
| `page_0259e59572808191bdb8fd302e86bd54` | MoAI 협업 실행 소유권과 복구 상세 설계 |
| `page_dff088ea648081918e2eb28ab4a3e77c` | MoAI Host adapter와 hook 이식성 |
| `page_e3612b0b07888191ae58c58d66bdff4a` | MoAI 판단 backend와 이미지 생성 상세 설계 |
| `page_25fa9b23c690819180a1351d6da22e23` | MoAI 릴리스와 Desktop 상세 설계 |
| `page_e087413484688191b4da76c7ddd96b2b` | PLAN과 GEARS SPEC 승인 계약 |
| `page_3d91772843448191a46eedf202be99cc` | RUN과 TDD DDD 구현 검증 흐름 |
| `page_c4f994be97548191b41945af0713ecc9` | MoAI 검증 감사와 보안 상세 설계 |
| `page_487fa473b0308191b572f7dbc8640ee0` | MoAI 파일 소유권과 worktree 수명 |
| `page_66acadc28bdc8191875556986da72e6b` | MoAI RC1 GitHub 작업 구조와 실행 계획 |
| `page_85f9b6259d2081918f9f5be447140a66` | MoAI Factory 생산성 중심 전체 흐름 재설계 |
| `page_08f6df9587c881919a46a8e47b3d8605` | ADK Factory와 TODO 제어 루프 재설계 검토본 |
| `page_0914dfdc33d48191b4311d8884d7f15d` | MoAI ADK 3.2 독립 실행 코어와 통합 어댑터 참고 설계 버전 5 |
| `page_5a4472e76eb88191b27bcbc8f9073381` | MoAI ADK Project Plan Run Sync와 순정 기능 활용 종합 조사 |
| `page_bb3cfe5d5ed481919f88afd757fbb57c` | MoAI ADK 코어와 제공자 및 브라우저 공식 계약 조사 |
| `page_00788de2e5808191890e1c406b549a67` | MoAI ADK 팀 개발 워크플로와 전체 실행 로직 |

## page_*.md 원문 사본 안내

`page_*.md` 원문 사본(19+1편, 31k행)과 `design-review-synthesis.html`은
이 저장소에 두지 않습니다 — ChatGPT Space 원본을 참조하세요. 원문 사본은
untracked `references/chatgpt-space/`에만 존재하며, 이 디렉터의 문서가
인용하는 행 번호(`L123`)는 그 사본 기준입니다.
