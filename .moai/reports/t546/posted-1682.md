bjw202님, 제출해 주신 가설이 맞습니다 — 그리고 문서로 수용했습니다.

세션 간 메시징 채널의 가용성 게이트는 `~/.claude.json`의 `cachedGrowthBookFeatures.tengu_harbor_kite` 슬롯 하나를 읽는데, 이 슬롯은 머신 전역에서 last-writer-wins으로 공유됩니다. 1st-party 세션만 이 슬롯을 쓰기 때문에, 서드파티 백엔드 세션(`moai glm`, `moai cg`의 GLM 패널)은 직전 1st-party 세션이 남겨둔 값을 물려받게 되고 — 제보하신 대로 — 채널이 세션 도중에 사라졌다 돌아오는 일이 일어날 수 있습니다.

가설과 진단·기전·수동 탈출구를 정리해 개발 라인 문서에 반영했습니다(develop 기준, `cross-session-messaging.md`의 가용성 제약 「The shared flag slot」 절과 `cross-session-messaging-detail.md`의 상세 절). 세션에서 채널이 의심되면 `/list-agents`(별칭 `/peers`)가 인식되는지로 가장 빨리 가려볼 수 있습니다 — 인식되지 않으면 채널이 없는 상태입니다.

추가 관측이 있으시면 이 이슈에 남겨 주세요. 릴리스는 아직 커팅되지 않았습니다 — 나오는 대로 이 이슈에 알려드리겠습니다.
