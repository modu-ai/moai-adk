# t1336 판정문 초안 — 근원 진술 (레인 최종 확정용)

## Root-cause (초안 문안)

한국어 자모 오염의 근원은 이 저장소가 통제하는 어떤 파일 생산자도 아닌 **모델 백엔드의 출력 토큰 스트림**이다. 오염은 GLM 채널 출력에서 관측됐고(입수 증거: `.moai/reports/t1336/materials/t1333-verdict-original.md`, worker-61 의 t1333 판정문), 오염 음절은 모두 유효한 조합형 음절이라 NFC/NFD 정규화로는 검출·수리되지 않는다. 오염형 중에는 정상형과 렌더링상 구별이 거의 없는 변형(`스`↔`슠`)이 있어 사람 눈 검수만으로는 잔존한다.

레인의 완결 범위는 다음 3가지다:

1. **detect** — `scripts/jamo_integrity.py check` (보고서·판정문 마감 전 자가 점검, `--stdin` 지원)
2. **restore** — `restore --dry-run` 검토 후 `--apply` (d=1 유일 후보만 자동, 복합어 가드, 코드 블록 불변)
3. **prevent** — `.moai/docs/jamo-integrity-guide.md` (레인 마감 절차 + "재타이핑은 수리가 아니다" 원칙)

근원 해결은 **운영자 보고 사항**이다 — `/moai:feedback` 으로 백엔드 출력 채널(GLM) 한국어 자모 무결성 결함을 보고한다. 레인이 백엔드를 수리할 수 없다.

## 검증 요약 (증거 경로)

| 주장 | 증거 |
|---|---|
| 오염 23건 검출 (재료 파일) | `demo-check.txt` (exit=1) |
| `--stdin` 동일 결과 | `analysis/demo-check-stdin.txt` |
| d=1 유일 8건 제안 / 15건 모호 플래그 | `demo-restore-dryrun.txt` |
| apply 실증: 8건 중 6건 정확, 2건 오수리(숴환→소환, 큜→큐) | `demo-restore-apply.txt`, `demo-restore-apply-diff.txt` |
| 자가 테스트 9/9 | `selftest.txt` (exit=0) |
| 특성화 전사 무결성 (재료↔문서 기교차) | `analysis/cross-verify.txt` (PASS) |
| 깨끗한 문서 4개 6,645어휘 스캔 → 138플래그(노이즈) 중 실제 오염 1건 포착(`처리르`) | `demo-fp-clean-files.txt` |
| 가이드 자가 점검에서 본문 실제 오염 1건 포착·수리(`고되되`→`고치되`) | `analysis/selfcheck-guide.txt` |
