// role_rules.go delivers the role-gated rule core to role sessions at
// SessionStart (SPEC-ALWAYS-LOADED-BUDGET-001 REQ-ALB-007..011).
//
// The role-gated rules (the factory dispatch rule and the cross-session
// messaging rule) leave the always-loaded surface through the M2/M3 arc; a
// factory leader or lane session still needs their binding blocks, so the
// SessionStart hook injects the role core — the regions the neutral markers
// enclose in the DEPLOYED rule files — into the session's additional
// context. The core is built solely from the deployed files (REQ-ALB-023):
// the binding ledger fixture is a test artifact this package never reads.
//
// Three properties shape the flow:
//
//   - Fail-visible, not fail-open (REQ-ALB-009): a role session whose rule
//     files are absent, unreadable, empty, or unmarked gets an
//     operator-visible warning AND an agent-facing read directive, never a
//     silent start.
//   - No truncation (REQ-ALB-010): when the final additional context
//     exceeds the runtime's documented 10,000-character per-string delivery
//     cap, the role core goes out INTACT — the runtime saves oversized
//     output to a session-directory file and passes the path plus a
//     2,000-character preview, and that file is the delivery channel. The
//     agent-facing directive names it; the operator warning reports it.
//   - Registry-derived (REQ-ALB-011): which sessions count as role
//     sessions is the config role-marker registry, not a hand-written list.
package hook

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// roleRulesContextLimit is the documented delivery cap per
// additionalContext string (decision-index Q4: 10,000 characters,
// unraisable; oversized output is saved by the runtime to a session file
// with a 2,000-character preview — CC 2.1.89).
const roleRulesContextLimit = 10000

// roleRuleFile is one deployed role-gated rule the injection delivers.
type roleRuleFile struct {
	// Rel is the project-root-relative deployed path (the path an agent
	// reads it back by).
	Rel string
	// Name is the file's bare name for notices and the citation guard.
	Name string
}

// roleRuleFiles lists the role-gated rules in injection order. The dispatch
// rule carries the role core; the messaging rule binds every session, so its
// marker pair is empty and the injection carries its pointer only.
var roleRuleFiles = []roleRuleFile{
	{Rel: ".claude/rules/moai/workflow/factory-dispatch.md", Name: "factory-dispatch.md"},
	{Rel: ".claude/rules/moai/workflow/cross-session-messaging.md", Name: "cross-session-messaging.md"},
}

// roleRulesInjectSources are the SessionStart sources that receive the role
// core: a genuinely new session (startup), and the two re-entry sources that
// discard the previous injection (clear empties the conversation, compact
// replaces it with a summary). resume restores the previous transcript —
// including the earlier injection — and receives nothing (REQ-ALB-008).
var roleRulesInjectSources = map[string]bool{"startup": true, "clear": true, "compact": true}

// Seams (tests pin or stub them; production never assigns).
var (
	// roleMarkerRegistry is the registry seam: the guarded marker set and
	// detection derive from it, so a fixture registry entry is automatically
	// covered (REQ-ALB-011).
	roleMarkerRegistry = config.RoleMarkerRegistry
	// roleRulesOverflowDelivery models whether the runtime's oversized-
	// output file delivery is available. It is a variable so tests can
	// simulate the unavailable/truncating fallback (REQ-ALB-010's REQ-ALB-009
	// retreat); production always reports available.
	roleRulesOverflowDelivery = func() bool { return true }
)

// roleRuleInjection is the SessionStart injection decision for one event.
type roleRuleInjection struct {
	// Context is the agent-facing text to append to additionalContext
	// ("" when nothing is injected).
	Context string
	// OperatorNotice is the operator-facing warning for systemMessage
	// ("" when the session needs none).
	OperatorNotice string
}

// buildRoleCore reads the deployed role-gated rule file under root and
// returns its role core: the marker-enclosed regions joined with blank
// lines. An empty marker pair yields an empty core (a legitimate state —
// the caller emits the pointer only). Every failure names its cause; the
// binding ledger is never read (REQ-ALB-023).
func buildRoleCore(root string, rule roleRuleFile) (string, error) {
	path := root + string(os.PathSeparator) + rule.Rel
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("role rule file absent: %s", rule.Rel)
		}
		return "", fmt.Errorf("role rule file unreadable: %s: %w", rule.Rel, err)
	}
	content := string(data)
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("role rule file empty: %s", rule.Rel)
	}
	regions, marked := config.ExtractRoleCoreRegions(content)
	if !marked {
		return "", fmt.Errorf("role rule file carries no %s markers: %s", config.RoleCoreMarkerStart, rule.Rel)
	}
	// A start marker without its closing pair is a malformed file, not an
	// empty core: the required rules would silently vanish from the session.
	// Balanced marker counts keep the legitimate empty-pair case (adjacent
	// start+end) on the empty-core path while sending every unclosed region
	// to the REQ-ALB-009 failure path.
	if strings.Count(content, config.RoleCoreMarkerStart) != strings.Count(content, config.RoleCoreMarkerEnd) {
		return "", fmt.Errorf("role rule file has an unclosed %s region (unbalanced marker counts): %s", config.RoleCoreMarkerStart, rule.Rel)
	}
	// Balanced counts alone still pass an END-before-START file: the
	// extractor scans forward from the FIRST start marker, so a file whose
	// first marker occurrence is the end marker yields an empty (or
	// tail-only) core with err=nil, and the caller would treat that empty
	// core as the legitimate empty-pair state — the required rules vanish
	// silently. Marker order is therefore validated too: the first marker
	// occurrence in the content must be a start marker. Balanced counts
	// guarantee both indexes exist at this point.
	if strings.Index(content, config.RoleCoreMarkerEnd) < strings.Index(content, config.RoleCoreMarkerStart) {
		return "", fmt.Errorf("role rule file carries a %s marker before any %s marker (malformed marker order): %s", config.RoleCoreMarkerEnd, config.RoleCoreMarkerStart, rule.Rel)
	}
	// Full sequence validation: every start marker must be closed by its own
	// end marker BEFORE the next start marker. Count balance plus first-
	// marker order still pass a START … END END … START file (two of each,
	// first marker a start), yet its second start region is never closed and
	// the extractor silently discards it — the injected core would be
	// missing a region without any warning. Any open-at-next-start or
	// close-without-open shape takes the REQ-ALB-009 failure path.
	{
		depth := 0
		rest := content
		for {
			si := strings.Index(rest, config.RoleCoreMarkerStart)
			ei := strings.Index(rest, config.RoleCoreMarkerEnd)
			if si < 0 && ei < 0 {
				break
			}
			if si >= 0 && (ei < 0 || si < ei) {
				if depth > 0 {
					return "", fmt.Errorf("role rule file has a %s marker before its enclosing region is closed: %s", config.RoleCoreMarkerStart, rule.Rel)
				}
				depth = 1
				rest = rest[si+len(config.RoleCoreMarkerStart):]
				continue
			}
			depth = 0
			rest = rest[ei+len(config.RoleCoreMarkerEnd):]
		}
		if depth != 0 {
			return "", fmt.Errorf("role rule file has an unclosed %s region (marker sequence ends inside a region): %s", config.RoleCoreMarkerStart, rule.Rel)
		}
	}
	return strings.Join(regions, "\n\n"), nil
}

// roleRuleLocaleTable carries the operator-facing warnings by locale. The
// operator warning is user-facing output, so it renders in the settings
// conversation language the factory notices already use; an unknown locale
// resolves to English (fail-open, per the shared locale-helper contract).
type roleRuleLocaleTable struct {
	// InjectionFailed renders the REQ-ALB-009 failure warning: the session
	// role name and the joined failure causes.
	InjectionFailed func(session, detail string) string
	// OverflowUnavailable renders the overflow-delivery-unavailable retreat
	// warning: the session role name, the measured total, and the cap.
	OverflowUnavailable func(session string, total, limit int) string
	// Overflow renders the deliberate overflow-file delivery warning.
	Overflow func(session string, total, limit int) string
}

var roleRuleLocales = map[string]roleRuleLocaleTable{
	"ko": {
		InjectionFailed: func(session, detail string) string {
			return fmt.Sprintf("역할 규칙 주입이 %s 세션에서 실패했습니다: %s. 에이전트에는 두 규칙 파일을 경로로 읽으라는 지시가 전달됐습니다.", session, detail)
		},
		OverflowUnavailable: func(session string, total, limit int) string {
			return fmt.Sprintf("역할 규칙 주입 초과(%s 세션): 조립된 맥락(%d자)이 %d자 전달 한도를 넘는데 넘침 파일 전달을 쓸 수 없어, 역할 core 대신 읽기 지시를 보냈습니다.", session, total, limit)
		},
		Overflow: func(session string, total, limit int) string {
			return fmt.Sprintf("역할 규칙 주입 초과(%s 세션): 조립된 맥락(%d자)이 세션 시작 전달 한도 %d자를 넘습니다. 역할 core는 잘리지 않고 그대로 보냈으며, 런타임이 넘친 출력을 세션 디렉터리 파일로 저장해 처음 2,000자 미리보기와 함께 경로를 전달합니다.", session, total, limit)
		},
	},
	"ja": {
		InjectionFailed: func(session, detail string) string {
			return fmt.Sprintf("ロールルールの注入が %s セッションで失敗しました: %s。エージェントには両ルールファイルをパスで読むよう指示を送りました。", session, detail)
		},
		OverflowUnavailable: func(session string, total, limit int) string {
			return fmt.Sprintf("ロールルール注入のオーバーフロー(%s セッション): 組み立てたコンテキスト(%d文字)が %d文字の配信上限を超えていますが、オーバーフローファイル配信が使えないため、ロール core の代わりに読み取り指示を送りました。", session, total, limit)
		},
		Overflow: func(session string, total, limit int) string {
			return fmt.Sprintf("ロールルール注入のオーバーフロー(%s セッション): 組み立てたコンテキスト(%d文字)がセッション開始の配信上限 %d文字を超えました。ロール core は切り詰めずそのまま送信し、ランタイムが超過出力をセッションディレクトリのファイルに保存して、先頭 2,000 文字のプレビュー付きでパスを渡します。", session, total, limit)
		},
	},
	langEnglish: {
		InjectionFailed: func(session, detail string) string {
			return fmt.Sprintf(
				"Role-rule injection failed for the %s session: %s. The agent was directed to read both rule files by path.",
				session, detail)
		},
		OverflowUnavailable: func(session string, total, limit int) string {
			return fmt.Sprintf(
				"Role-rule injection overflow for the %s session: the assembled context (%d characters) exceeds the %d-character delivery cap and overflow file delivery is unavailable; the read directive was emitted instead of the core.",
				session, total, limit)
		},
		Overflow: func(session string, total, limit int) string {
			return fmt.Sprintf(
				"Role-rule injection overflow for the %s session: the assembled context (%d characters) exceeds the %d-character session-start delivery cap. The role core was emitted intact; the runtime saves oversized output to a session-directory file and passes its path with a 2,000-character preview.",
				session, total, limit)
		},
	},
}

// roleRuleLocaleFor resolves the operator-locale table, failing open to
// English for an unknown locale.
func roleRuleLocaleFor(lang string) roleRuleLocaleTable {
	if t, ok := roleRuleLocales[lang]; ok {
		return t
	}
	return roleRuleLocales[langEnglish]
}

// roleRuleInjectionFor decides the role-rules injection for one SessionStart
// event. root is the project root the deployed rule files resolve under;
// source is input.Source; existing is the additionalContext every earlier
// producer already assembled — the 10,000-character cap applies to the
// FINAL string, so the measurement runs over it plus the core plus the
// directive (REQ-ALB-010); lang is the operator-facing conversation locale
// the two operator warnings render in.
func roleRuleInjectionFor(root, source, existing, lang string) roleRuleInjection {
	// Role session detection through the registry (REQ-ALB-011).
	role, ok := detectRegisteredRole(os.Getenv)
	if !ok {
		return roleRuleInjection{}
	}
	// Source gate: startup/clear/compact inject, resume does not
	// (REQ-ALB-007/008).
	if !roleRulesInjectSources[source] {
		return roleRuleInjection{}
	}

	// Build every core first; any failure takes the REQ-ALB-009 path.
	var parts []string
	var failures []string
	pointer := ""
	for _, rule := range roleRuleFiles {
		core, err := buildRoleCore(root, rule)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		if core == "" {
			// A role-gated rule whose core is empty contributes its pointer
			// only (acceptance §D.2 boundary case).
			pointer = fmt.Sprintf("Role rule (no role-core region — its binding blocks are always-loaded): `%s`", rule.Rel)
			continue
		}
		parts = append(parts, core)
	}
	if len(failures) > 0 {
		loc := roleRuleLocaleFor(lang)
		return roleRuleInjection{
			Context: roleRulesReadDirective(""),
			OperatorNotice: loc.InjectionFailed(role.Name, strings.Join(failures, "; ")),
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Role-gated rules (factory %s session — injected at session start; the always-loaded surface carries the stubs):\n\n", role.Name)
	sb.WriteString(strings.Join(parts, "\n\n"))
	if pointer != "" {
		sb.WriteString("\n\n")
		sb.WriteString(pointer)
	}
	context := sb.String()

	// Size gate over the FINAL additional context (REQ-ALB-010). The unit
	// is the runtime's string-length unit — UTF-16 code units (Q4), not Go
	// bytes: CJK prose and supplementary-plane runes count 2 per rune.
	total := utf16Len(existing) + utf16Len("\n\n"+context)
	if total <= roleRulesContextLimit {
		return roleRuleInjection{Context: context}
	}

	if !roleRulesOverflowDelivery() {
		// Overflow delivery unavailable or truncating: REQ-ALB-009 retreat —
		// warning plus read directive, no core, zero truncated units.
		loc := roleRuleLocaleFor(lang)
		return roleRuleInjection{
			Context:        roleRulesReadDirective(""),
			OperatorNotice: loc.OverflowUnavailable(role.Name, total, roleRulesContextLimit),
		}
	}

	// Deliberate overflow-file delivery: the core goes out INTACT (zero
	// truncated units); the runtime saves it to a session file and passes
	// the path plus a 2,000-character preview.
	context += "\n\n" + roleRulesOverflowDirective()
	loc := roleRuleLocaleFor(lang)
	return roleRuleInjection{Context: context, OperatorNotice: loc.Overflow(role.Name, total, roleRulesContextLimit)}
}

// roleRulesRootFromCWD resolves the project root the deployed role-gated
// rule files resolve under for a session whose cwd may sit anywhere inside
// the tree: it walks the cwd's ancestors outward and returns the first
// ancestor carrying the deployed dispatch rule, and the cwd itself when no
// ancestor does — leaving the REQ-ALB-009 fail-visible path to name what is
// missing. The walk STOPS at the project boundary: a directory entry
// carrying a .git marker (a checkout directory, or a worktree's repository
// pointer file) ends the walk at that directory rather than reaching into a
// parent checkout — a lane worktree that lacks the deployed rules takes the
// REQ-ALB-009 missing path, never its parent tree's rules. Pure path
// arithmetic, no git subprocess: the SessionStart hook runs under a 5s
// budget (the same reasoning as cardIDFromPath).
func roleRulesRootFromCWD(cwd string) string {
	if strings.TrimSpace(cwd) == "" {
		return ""
	}
	probe := filepath.Join(roleRuleFiles[0].Rel)
	for dir := filepath.Clean(cwd); ; {
		if fi, err := os.Stat(filepath.Join(dir, filepath.FromSlash(probe))); err == nil && !fi.IsDir() {
			return dir
		}
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			// Project boundary: this directory is the checkout root — the
			// session's project root even when its rules are missing. No
			// reach beyond this checkout.
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return cwd
		}
		dir = parent
	}
}

// detectRegisteredRole walks the registry seam and returns the first role
// the lookup carries.
func detectRegisteredRole(lookup func(string) string) (config.RoleMarker, bool) {
	if lookup == nil {
		return config.RoleMarker{}, false
	}
	for _, m := range roleMarkerRegistry() {
		if lookup(m.EnvKey) != "" {
			return m, true
		}
	}
	return config.RoleMarker{}, false
}

// utf16Len counts UTF-16 code units of s (the JavaScript string-length unit
// the runtime cap is measured in: 1 per BMP rune, 2 per supplementary-plane
// rune).
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// roleRulesReadDirective is the agent-facing REQ-ALB-009 directive: read the
// full rule files by path before acting.
func roleRulesReadDirective(reason string) string {
	var sb strings.Builder
	if reason != "" {
		sb.WriteString(reason)
		sb.WriteString(" ")
	}
	sb.WriteString("[HARD] Read both role-gated rule files in full before your first action: `")
	sb.WriteString(roleRuleFiles[0].Rel)
	sb.WriteString("` and `")
	sb.WriteString(roleRuleFiles[1].Rel)
	sb.WriteString("`.")
	return sb.String()
}

// roleRulesOverflowDirective is the agent-facing REQ-ALB-010 directive that
// rides an intact over-cap emission: the runtime file is the delivery
// channel, and the rule files are the fallback read.
func roleRulesOverflowDirective() string {
	return fmt.Sprintf(
		"NOTE: the output above exceeds the session-start delivery cap (%d characters). The runtime saves the intact output to a file in the session directory and passes its path with a preview of the first 2,000 characters — read the role core from that file, or read the rule files by path: `%s`, `%s`.",
		roleRulesContextLimit, roleRuleFiles[0].Rel, roleRuleFiles[1].Rel)
}
