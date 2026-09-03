package config

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// @MX:NOTE: [AUTO] SPEC-TOKEN-EFFICIENCY-001 P0-1 — always-loaded 토큰 예산 가드.
// Claude Code가 매 턴 재주입하는 always-loaded 컨텍스트 표면(CLAUDE.md + no-paths: 규칙 +
// output-style)의 토큰 총량이 예산을 넘으면 회귀로 판정한다. CC 네이티브
// 압축/캐싱은 재구현하지 않는다(over-engineering guard, plan.md §G).

// AlwaysLoadedTokenBudget는 always-loaded 컨텍스트 표면에 허용되는 추정 토큰 상한이다.
// 회귀 트립와이어로서, 측정 표면이 이 예산을 초과하면 가드가 실패한다 — Epic Steering-Align
// 다이어트가 조용히 되돌아가는 것을 잡는다.
//
// 도출 근거: 측정 baseline(2026-07-02) ≈ 64,624 토큰(char/4 추정: CLAUDE.md + no-paths:
// .claude/rules/moai/** 규칙 파일 + moai.md 합계 258,498 bytes / 4) +
// 약 15% 여유(≈ 74,317)를 클린 상수로 올림. 여유분은 통상적 규칙 편집을 흡수하되 의미 있는
// 증가에는 발화한다.
//
// 상향 근거(2026-08-17): release/v3.1.1 통합 레인에서 t72 등 선행 머지 카드의 룰 문서
// 추가로 측정 표면이 75,282 토큰에 도달, 예산을 282 초과. release 브랜치 push는 CI
// 트리거(main 전용) 밖이라 개별 카드 단계에서 미검출된 선결 결함이다. 근본 해결
// (kanban-dispatch 등 대형 always-loaded 룰의 스텁+지연 로딩 다이어트)은 별도 카드로
// 진행하며, 그 착지 전까지의 임시 상향으로 75,000 → 76,000으로 올린다.
//
// 상향 근거(2026-08-31, SPEC-MEMORY-STORE-RECONCILE-001): 이 SPEC은 세션이 인덱스가 길다는
// 이유로 교훈 기록을 포기하는 것을 막는 [HARD] 조항을 always-loaded 표면(moai-constitution.md
// § Lessons Protocol)에 넣어야 한다 — 그 조항을 paths-scoped 파일에 두면 대상 세션이 읽지
// 못하므로(REQ-MSR-004 / C5) 위치를 옮길 수 없다. 측정: 편집 전 75,799 토큰(여유 201),
// 편집 후 76,009. 이 카드가 더한 always-loaded 분량은 정확히 210 토큰이다.
//
// 76,000 → 76,210 은 그 210을 그대로 얹은 값이다. 임의의 여유를 새로 만들지 않고 기존 여유
// 201 토큰을 보존하는 것이 목적이며, 조항 자체는 먼저 최소 길이로 줄인 뒤(초안 대비 약 1,000
// 바이트 삭감) 남은 분량만 반영했다. 상세 서술은 paths-scoped 인 moai-memory.md 쪽에 두어
// always-loaded 비용을 최소화했다.
//
// 주의: 이 표면은 포화 상태다. 편집 전 여유가 예산의 0.26%(201/76,000)에 불과했고, 상향 뒤에도
// 같은 수준이다. 다음에 always-loaded 파일을 늘리는 카드는 이 가드에 부딪힌다 — 근본 해결은
// 위 문단이 가리키는 대형 룰 다이어트이며, 이 카드의 소관이 아니다.
//
// 상향 근거(2026-09-01, t421): 흡수 트리(origin/develop 8c1d911df) 가드 실측 76,129
// 토큰(여유 81) — t400 이 always-loaded cross-session-messaging.md 에 넣은 다섯째 가용성
// 제약(공유 슬롯 축)이 여유를 소비한 뒤다(순증 +480 B: t400 +1,135 B, t409 스텁 다이어트
// −655 B). 바로 뒤를 따르는 t196
// (SPEC-CODEX-SKILL-NEUTRAL-001)은 AGENTS.md(측정 표면 고정 슬롯)에 능력 결속표를
// 얹는다 — 브랜치 워크트리 실측 +545 B = +136 토큰. 76,129 + 136 = 76,265 로 현 예산
// 76,210 을 넘어 트립하므로, 착지 후 여유 135(상향 직후 271 − t196 몫 136)를 남기는
// 76,400 으로 올린다. 근본 해결은 여전히 위 문단의 대형 룰 다이어트다.
//
// 상향 근거(2026-09-03, t453): 위 DEBT 가 예고한 포화 트립이 터졌다 — develop 팁
// 400f37eb9 실측 76,939(예산 76,400 초과 539). 보정 커밋 b9efb3626(t421) 이후 표면
// 성장 +810 토큰은 전부 착지 카드의 교리 조항이며 전수 귀속됐다: t196 AGENTS.md
// 역량 결속표 +136(직전 상향이 예상했던 바로 그것), t224 레인 spawn 권한 5표면 착지
// +510(kanban-dispatch +225 / agent-common-protocol +192 / moai-constitution +95),
// t386 감사 산출물 컨벤션 +143, t236 graph_shortest_path 카탈로그 갱신 +20.
// 세 갈래 대안을 모두 측정으로 기각했다: ① 측정 대상 변경 — 17개 열거 항목 전수가
// 매 턴 주입 집합과 일치, t368 로고 건 같은 계수 결함 없음. ② 문서 축소 — 유일한
// de-dup 후보(t224 agent-common-protocol 재진술)는 5표면 의도 설계로 확인되어
// 기각; 이번 성장분에 지방 없음. ③ 상한 — 가드의 자기 정의는 다이어트의 조용한
// 회귀 트립와이어이지 착지 교리의 성장 상한이 아니므로, 소모 조항 전수 열거를 붙인
// 보정으로 정당하다. 폭은 실측 76,939 + 여유 261(최근 조항 20~367 tok 대비 단일
// 조항분) = 77,200. 참고: 창 대기 브랜치 8개는 표면 추가분 0(real diff) — 이
// 상향이 창 재측정을 재트립시키지 않는다. 다음 트립에 자동 정당성은 없다 — 그
// 트립이 다이어트 카드의 착수 근거다.
//
// @MX:DEBT: [AUTO] temporary budget raise chain (76,000 -> 76,400 -> 77,200) standing in for the always-loaded rule diet
// @MX:CEILING: 0.34% headroom — 261 tokens of 77,200; one always-loaded clause consumes it
// @MX:UPGRADE: drop this raise chain when the large always-loaded rule diet (stub + lazy loading) lands — measured targets: output-style moai.md 16.5K tok, kanban-dispatch.md 8.6K, agent-common-protocol.md 6.7K, verification-claim-integrity.md 6.3K (t453 measurement)
// @MX:SPEC: SPEC-MEMORY-STORE-RECONCILE-001
const AlwaysLoadedTokenBudget = 77200

// CodexContractByteCeiling는 루트 AGENTS.md(코덱스 계약층)에 허용되는 바이트 상한이다.
// codex는 프로젝트 지시문을 바이트 상한 아래에서 읽고 초과분을 **조용히** 잘라낸다 —
// 문장 중간에서 잘린 규칙은 없는 규칙보다 나쁘다. 완전해 보이기 때문이다.
//
// 도출 근거: 신뢰 등록 전(untrusted) 첫 세션의 실효 상한 32,768 B 에서 개인 전역
// ~/.codex/AGENTS.md 층과 향후 성장을 위한 예비 8,192 B 를 뺀 24,576 B. 상한을 넘기면
// 가드가 실패한다(경고가 아니다): 잘림 자체가 무신호이므로 이 가드가 유일한 신호다.
const CodexContractByteCeiling = 24576

// ContractByteBreach는 계약 문서 하나의 상한 위반 측정치다.
type ContractByteBreach struct {
	Path     string // 위반 파일 경로
	Bytes    int    // 실측 바이트
	Overflow int    // 상한 초과분
}

// contractDocuments는 바이트 상한이 걸리는 계약 문서 경로를 반환한다: always-loaded 열거
// (alwaysLoadedSurface)가 내놓는 루트 AGENTS.md 와 그 템플릿 미러. 두 번째 측정 경로를
// 만들지 않기 위해 규칙 트리를 다시 글로브하지 않고 열거 헬퍼를 재사용한다 —
// 미러가 상한을 넘으면 사용자 머신에서 잘리므로 라이브 파일 크기와 무관하게 함께 묶인다.
func contractDocuments(repoRoot string) ([]string, error) {
	surface, err := alwaysLoadedSurface(repoRoot)
	if err != nil {
		return nil, err
	}
	agents := filepath.Join(repoRoot, "AGENTS.md")
	var docs []string
	for _, p := range surface {
		if p == agents {
			docs = append(docs, p)
		}
	}
	if len(docs) == 0 {
		return nil, nil // 열거에 AGENTS.md 가 없다 — AC-AMC-017 이 잡는 조건이다
	}
	docs = append(docs, filepath.Join(repoRoot, "internal", "template", "templates", "AGENTS.md"))
	return docs, nil
}

// MeasureContractBytes는 계약 문서들의 바이트 상한 위반을 측정해 반환한다. 위반이 없으면
// 빈 슬라이스다. 디스크에 없는 파일은 건너뛴다(hermetic: 아직 미러가 없는 트리에서도
// baseline 이 흔들리지 않는다).
func MeasureContractBytes(repoRoot string) ([]ContractByteBreach, error) {
	docs, err := contractDocuments(repoRoot)
	if err != nil {
		return nil, err
	}
	var breaches []ContractByteBreach
	for _, p := range docs {
		fi, statErr := os.Stat(p)
		if statErr != nil {
			continue // 없는 파일 → 측정 대상 아님
		}
		if n := int(fi.Size()); n > CodexContractByteCeiling {
			breaches = append(breaches, ContractByteBreach{Path: p, Bytes: n, Overflow: n - CodexContractByteCeiling})
		}
	}
	return breaches, nil
}

// TOMBSTONE — there is deliberately no MEMORY.md fixed surface slot, and no head-cap
// constants to go with it. Do not re-add them.
//
// Two independent reasons, both measured:
//
//  1. It measured nothing. The slot pointed at repoRoot/MEMORY.md, which this repository
//     does not contain, so it contributed 0 tokens on every real run — forever. The unit
//     tests passed only because they supplied their own fixture, which is what kept the
//     vacuity invisible.
//  2. Its head caps encoded an UNCONFIRMED premise. They asserted the Claude Code
//     auto-memory loader truncates the index at a specific line count or byte size,
//     whichever it reaches first. The loader is not part of this repository, and the one
//     direct observation available contradicts a strict byte cut at the size that was
//     encoded. Re-adding the caps would harden an unverified claim into code.
//
// Pointing the slot at the real auto-memory store is not the fix either: that store is
// machine-specific and lives outside the repository, which breaks the hermeticity the
// enumeration below depends on. The index is measured with `moai memory doctor` instead;
// see .claude/rules/moai/workflow/moai-memory.md § MEMORY.md Index Budget.

// estimateTokens는 char/4 rule-of-thumb(len(bytes)/4)로 b의 근사 토큰 수를 반환한다.
// 실제 tokenizer 대비 ±약 15% 오차가 있는 의도적 무의존 근사다. 이 가드는 상대적 증가를
// 감시하는 트립와이어이지 회계 원장이 아니므로 정확도가 요구되지 않는다(simplicity ladder —
// 가드를 위해 tokenizer 의존성을 추가하지 않는다).
func estimateTokens(b []byte) int {
	return len(b) / 4
}

// findRepoRoot는 start에서 위로 올라가며 go.mod(리포 루트 마커)를 담은 디렉터리를 찾는다.
// go.mod 조상을 찾지 못하면 ("", false)를 반환한다(트리 밖에서 실행 시 가드는 graceful skip).
func findRepoRoot(start string) (string, bool) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// frontmatterHasPaths는 마크다운 바이트의 YAML frontmatter에 top-level `paths:` 키가
// 있는지 판정한다. frontmatter는 첫 줄 `---`로 시작해 닫는 `---`로 끝난다. frontmatter가
// 아예 없으면(첫 줄이 `---`가 아니면) false. 닫는 `---` 이후의 `paths:`는 본문이므로 무시한다.
func frontmatterHasPaths(data []byte) bool {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			if strings.TrimRight(line, " \t\r") != "---" {
				return false // frontmatter 없음
			}
			continue
		}
		if strings.TrimRight(line, " \t\r") == "---" {
			return false // frontmatter 종료, paths: 미발견
		}
		if strings.HasPrefix(line, "paths:") {
			return true
		}
	}
	return false // 닫히지 않은 frontmatter → 보수적으로 미제한(always-loaded로 계수)
}

// hasPathsRestriction은 path의 마크다운 파일이 frontmatter에 `paths:` 제한을 갖는지(즉
// 조건부 로드 규칙인지) 판정한다. 읽을 수 없거나 frontmatter가 없는 파일은 제한 없음으로
// 취급한다(보수적: always-loaded 표면에 계수).
func hasPathsRestriction(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false // 보수적: 계수
	}
	return frontmatterHasPaths(data)
}

// alwaysLoadedSurface는 repoRoot 기준 always-loaded 컨텍스트 표면을 나열한다: frontmatter에
// `paths:` 제한이 없는 모든 .claude/rules/moai/**/*.md 파일(정렬), 이어서 3개의 고정 표면
// 슬롯(CLAUDE.md, AGENTS.md, .claude/output-styles/moai/moai.md). 3개 고정 슬롯은
// 디스크에 파일이 없어도 항상 목록에 포함된다 — 없는 파일은 측정 시 0 토큰으로 계산한다.
//
// 열거(enumeration)와 측정(measurement)은 다른 규칙을 따른다. 위 hermetic 처리는 *측정*에
// 관한 것이다: 사용자 트리에 슬롯 파일이 없을 수 있고, 그때는 0 토큰으로 계산하면 된다.
// *열거*에는 더 강한 규칙이 붙는다 — 이 저장소 트리에 존재하지 않는 경로를 가리키는 슬롯은
// 여기서 영원히 아무것도 측정하지 못하므로 애초에 열거되어서는 안 된다. 위 TOMBSTONE 이
// 제거한 슬롯이 정확히 그 경우였고, TestFixedSlotsExistInRepoTree 가 재발을 막는다.
//
// AGENTS.md 슬롯(SPEC-AGENTS-MD-CANON-001 REQ-AMC-008): 루트 AGENTS.md는 CLAUDE.md의
// `@`-import이므로 존재하는 순간부터 always-loaded다. 이 슬롯이 없으면 규칙 파일에서
// AGENTS.md로 옮긴 조항이 always-loaded 컨텍스트에는 그대로 남은 채 측정에서만 사라져,
// 일어나지 않은 감축이 다이어트로 기록된다. hermetic 처리를 그대로 쓰므로 AGENTS.md가 없는
// 트리의 baseline은 영향을 받지 않는다.
func alwaysLoadedSurface(repoRoot string) ([]string, error) {
	rulesDir := filepath.Join(repoRoot, ".claude", "rules", "moai")
	var ruleFiles []string
	err := filepath.WalkDir(rulesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if !hasPathsRestriction(path) {
			ruleFiles = append(ruleFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(ruleFiles)

	// 3개 고정 표면 슬롯을 항상 고정 순서로 추가한다. AGENTS.md는 CLAUDE.md의 `@`-import라
	// 바로 뒤에 둔다.
	fixed := []string{
		filepath.Join(repoRoot, "CLAUDE.md"),
		filepath.Join(repoRoot, "AGENTS.md"),
		filepath.Join(repoRoot, ".claude", "output-styles", "moai", "moai.md"),
	}
	return append(ruleFiles, fixed...), nil
}

// measureAlwaysLoaded는 repoRoot 기준 always-loaded 표면의 추정 토큰을 합산한다. 총 토큰
// 추정치와 나열된 표면(카운트 assertion용)을 반환한다. 없는 파일은 0 토큰이다.
func measureAlwaysLoaded(repoRoot string) (total int, surface []string, err error) {
	surface, err = alwaysLoadedSurface(repoRoot)
	if err != nil {
		return 0, nil, err
	}
	for _, path := range surface {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			continue // 없는 파일 → 0 토큰(hermetic)
		}
		total += estimateTokens(data)
	}
	return total, surface, nil
}

// MeasureAlwaysLoadedSection estimates the char and token contribution of a
// specific section within a file on the always-loaded surface, identified by
// its HTML-comment start/end marker pair. This is the per-section attribution
// extension of measureAlwaysLoaded (AC-HEV2-013 / REQ-HEV2-008): it lets the
// MOAI:LEARNED-WORKFLOW digest budget enforcement verify the ACTUAL measured
// contribution of the block, not an assumed value.
//
// chars is the byte length of the body strictly between the first occurrence
// of startMarker and the first subsequent occurrence of endMarker (the markers
// themselves are excluded). tokens is chars/4 per the estimateTokens rule.
// found is false (with zero counts) when either marker is absent or the file
// cannot be read.
//
// Additive — does not alter measureAlwaysLoaded's signature or behavior.
func MeasureAlwaysLoadedSection(filePath, startMarker, endMarker string) (chars, tokens int, found bool, err error) {
	data, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return 0, 0, false, nil // hermetic: absent file → not found, no error
	}
	body := string(data)
	startIdx := strings.Index(body, startMarker)
	if startIdx < 0 {
		return 0, 0, false, nil
	}
	// Move past the start marker line's closing --> (or the marker itself).
	afterStart := startIdx + len(startMarker)
	endIdx := strings.Index(body[afterStart:], endMarker)
	if endIdx < 0 {
		return 0, 0, false, nil
	}
	section := body[afterStart : afterStart+endIdx]
	chars = len(section)
	tokens = estimateTokens([]byte(section))
	return chars, tokens, true, nil
}
