// role_rules.go delivers the role-gated rule core to role sessions at
// SessionStart (SPEC-ALWAYS-LOADED-BUDGET-001 REQ-ALB-007..011).
//
// The role-gated rules (the factory dispatch rule and the cross-session
// messaging rule) leave the always-loaded surface through the M2/M3 arc; a
// factory leader or lane session still needs their binding blocks, so the
// SessionStart hook injects the role core — the regions the neutral markers
// enclose in the DEPLOYED rule files — into the session's additional
// context. The core is built from the deployed files, with the binary's
// embedded template copy as the stale-copy fallback (REQ-ALB-023 keeps the
// binding ledger fixture out of both sources).
//
// Three properties shape the flow:
//
//   - Fail-visible for what the binary cannot replace (REQ-ALB-009): a role
//     session whose rule files are absent, unreadable, empty, or malformed
//     gets an operator-visible warning AND an agent-facing read directive,
//     never a silent start. A STALE deployed copy — unmarked, or stamped
//     behind the binary (t1729) — does NOT fail: the core is served from the
//     binary's embedded template copy, with a one-line notice naming the
//     cause and the refresh path (a worktree branched from a main with an
//     uncommitted `moai update` cannot be fixed by running `moai update`
//     inside the worktree).
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
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
	"gopkg.in/yaml.v3"
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
	// Context is the agent-facing core text ("" when nothing is injected).
	Context string
	// RecoveryHead is a recovery directive that must open the FINAL
	// composite additionalContext — placed ahead of every earlier producer's
	// text by the assembler, because a runtime side-channel cut (save
	// failure → first 10,000 characters delivered) keeps the composite HEAD
	// but would drop a directive appended after 10,000+ characters of prior
	// context ("" when none).
	RecoveryHead string
	// OperatorNotice is the operator-facing warning for systemMessage
	// ("" when the session needs none).
	OperatorNotice string
}

// assembleInjectionComposite builds the final additionalContext from the
// accumulated context and an injection decision: the recovery directive
// (when present) opens the composite, the earlier producers' text follows,
// and the core comes last. The order is the save-failure guarantee — see
// RecoveryHead.
func assembleInjectionComposite(existing string, inj roleRuleInjection) string {
	if inj.RecoveryHead == "" {
		if existing == "" {
			return inj.Context
		}
		if inj.Context == "" {
			return existing
		}
		return existing + "\n\n" + inj.Context
	}
	var sb strings.Builder
	sb.WriteString(inj.RecoveryHead)
	if existing != "" {
		sb.WriteString("\n\n")
		sb.WriteString(existing)
	}
	if inj.Context != "" {
		sb.WriteString("\n\n")
		sb.WriteString(inj.Context)
	}
	return sb.String()
}

// roleRulePoliciesRel projects a Claude-layout rule path onto the
// harness-neutral deployment surface the Codex-only install uses (.moai/policies
// — the deploy-side projection of .claude/rules/moai in the shared deploy
// surfaces table). ok=false for paths outside the rules tree.
func roleRulePoliciesRel(rel string) (string, bool) {
	if strings.HasPrefix(rel, ".claude/rules/moai/") {
		return ".moai/policies/" + strings.TrimPrefix(rel, ".claude/rules/moai/"), true
	}
	return "", false
}

// roleRuleDeployRel resolves the path rule is actually deployed at under
// root: the Claude-layout path when present, else the Codex-only deployment
// shape (.moai/policies). The Claude form is returned when neither exists —
// the caller's failure path names the canonical path.
func roleRuleDeployRel(root string, rule roleRuleFile) string {
	path := filepath.Join(root, filepath.FromSlash(rule.Rel))
	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		return rule.Rel
	}
	if alt, ok := roleRulePoliciesRel(rule.Rel); ok {
		altPath := filepath.Join(root, filepath.FromSlash(alt))
		if fi, err := os.Stat(altPath); err == nil && !fi.IsDir() {
			return alt
		}
	}
	return rule.Rel
}

// buildRoleCore reads the deployed role-gated rule file under root and
// returns its role core: the marker-enclosed regions joined with blank
// lines. An empty marker pair yields an empty core (a legitimate state —
// the caller emits the pointer only). Every failure names its cause; the
// binding ledger is never read (REQ-ALB-023).

// errNotRegularFile is the shared non-regular rejection: the platform reader
// (role_rules_read_{unix,windows}.go) returns it when the opened handle is
// not a regular file, and the caller wraps it into the REQ-ALB-009 failure
// path with the rule path named.
var errNotRegularFile = fmt.Errorf("role rule file is not a regular file")

// errRoleRuleUnmarked is the no-markers rejection, a SENTINEL so the caller
// can distinguish "stale deployed copy" (t1729: the embedded-fallback case)
// from the other REQ-ALB-009 failures, which keep failing the session.
var errRoleRuleUnmarked = fmt.Errorf("role rule file carries no %s markers", config.RoleCoreMarkerStart)

// roleRuleEmbeddedRule is the embedded-template seam (t1729): it returns the
// binary's embedded template copy of one deployed rule file, addressed by
// the deployment-relative path (EmbeddedTemplates strips the "templates/"
// prefix, so the embedded paths match the deployment targets 1:1). ok=false
// is a miss — the caller keeps the original failure. Production reads the
// real embed; tests stub it with fixture content.
var roleRuleEmbeddedRule = func(rel string) (string, bool) {
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		return "", false
	}
	data, err := fs.ReadFile(fsys, rel)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func buildRoleCore(root string, rule roleRuleFile) (string, error) {
	rel := roleRuleDeployRel(root, rule)
	path := filepath.Join(root, filepath.FromSlash(rel))
	// Only regular files are read, judged on the OPENED HANDLE (POSIX
	// TOCTOU guard — role_rules_read_unix.go): a FIFO (or any special file)
	// at the rule path — including one swapped in after a path Stat — would
	// block a plain read forever waiting for a writer, hanging SessionStart
	// with no rules, warning, or directive. Non-regular opens take the
	// REQ-ALB-009 failure path instead.
	data, err := readRuleFileBytes(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("role rule file absent: %s", rel)
		}
		if errors.Is(err, errNotRegularFile) {
			return "", fmt.Errorf("%w: %s", errNotRegularFile, rel)
		}
		return "", fmt.Errorf("role rule file unreadable: %s: %w", rel, err)
	}
	return roleCoreFromContent(string(data), rel)
}

// roleCoreFromContent validates one rule file's content and returns its role
// core: the marker-enclosed regions joined with blank lines. label names the
// source in error messages (the deployed rel for a session-tree read, the
// template rel for an embedded-copy read). An empty marker pair yields an
// empty core (a legitimate state — the caller emits the pointer only).
// Every failure names its cause; the binding ledger is never read
// (REQ-ALB-023).
func roleCoreFromContent(content, label string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("role rule file empty: %s", label)
	}
	regions, marked := config.ExtractRoleCoreRegions(content)
	if !marked {
		// Markers PRESENT but unusable (audit round-1 P2, card t1729): a
		// file carrying only an end marker also reads marked=false, and the
		// fallback would silently rescue corrupted bytes. The fallback is
		// for a STALE GENERATION — a file with NEITHER marker. Any marker
		// presence without a usable region is a named failure.
		if strings.Contains(content, config.RoleCoreMarkerStart) || strings.Contains(content, config.RoleCoreMarkerEnd) {
			return "", fmt.Errorf("role rule file has %s markers but no usable region (corrupted marker shape): %s", config.RoleCoreMarkerStart, label)
		}
		return "", fmt.Errorf("%w: %s", errRoleRuleUnmarked, label)
	}
	// A start marker without its closing pair is a malformed file, not an
	// empty core: the required rules would silently vanish from the session.
	// Balanced marker counts keep the legitimate empty-pair case (adjacent
	// start+end) on the empty-core path while sending every unclosed region
	// to the REQ-ALB-009 failure path.
	if strings.Count(content, config.RoleCoreMarkerStart) != strings.Count(content, config.RoleCoreMarkerEnd) {
		return "", fmt.Errorf("role rule file has an unclosed %s region (unbalanced marker counts): %s", config.RoleCoreMarkerStart, label)
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
		return "", fmt.Errorf("role rule file carries a %s marker before any %s marker (malformed marker order): %s", config.RoleCoreMarkerEnd, config.RoleCoreMarkerStart, label)
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
					return "", fmt.Errorf("role rule file has a %s marker before its enclosing region is closed: %s", config.RoleCoreMarkerStart, label)
				}
				depth = 1
				rest = rest[si+len(config.RoleCoreMarkerStart):]
				continue
			}
			depth = 0
			rest = rest[ei+len(config.RoleCoreMarkerEnd):]
		}
		if depth != 0 {
			return "", fmt.Errorf("role rule file has an unclosed %s region (marker sequence ends inside a region): %s", config.RoleCoreMarkerStart, label)
		}
	}
	return strings.Join(regions, "\n\n"), nil
}

// embeddedRoleCore validates and extracts the embedded template copy of one
// role rule (t1729). A miss — or an embedded copy that fails the same
// validation the deployed copy gets — returns an error the caller folds into
// the REQ-ALB-009 failure path; the embedded copy is never trusted blindly.
// The error label is the DEPLOYED path (audit round-4: the embed carries the
// Claude-layout path even on a Codex-only install, where the project's rule
// lives under .moai/policies — the operator-facing name must follow the
// deployed location).
func embeddedRoleCore(root string, rule roleRuleFile) (string, error) {
	content, ok := roleRuleEmbeddedRule(rule.Rel)
	if !ok {
		return "", fmt.Errorf("embedded template copy absent: %s", roleRuleDeployRel(root, rule))
	}
	// Audit round-5 (card t1729): the embed is the Claude-layout TEXT — on a
	// Codex-only install the deployed rule (and every path reference inside
	// it) lives under .moai/policies. Project the body the same way the
	// deployer's harness normalization rewrites references at deploy time,
	// or the injected core points agents at a layout the project never
	// deployed.
	if roleRuleDeployRel(root, rule) != rule.Rel {
		content = strings.ReplaceAll(content, ".claude/rules/moai/", ".moai/policies/")
	}
	return roleCoreFromContent(content, roleRuleDeployRel(root, rule))
}

// worktreeBaseBranchFor resolves the base branch the project's git-strategy
// names (audit round-3 P2, card t1729: a project may run develop or trunk —
// the refresh advice must name the branch it actually has). Falls back to
// main, the default base.
func worktreeBaseBranchFor(root string) string {
	if b := config.LoadWorktreeBaseBranch(root); b != "" {
		return b
	}
	return "main"
}

// skewKeptNotice is the one-line notice a skewed tree carries when its
// parseable deployed core was KEPT (audit round-3, card t1729): version
// stamps cannot establish build order, so the skew is surfaced by name —
// both versions plus the refresh path — instead of acted on with a swap.
func skewKeptNotice(root string) string {
	_, stamp := roleRuleSkewInfo(root)
	return fmt.Sprintf(
		"NOTE: the deployed rule files are stamped %s while this binary is %s — version strings cannot prove which is newer, so the deployed core was kept. If this session's tree is a worktree, the usual cause is an uncommitted `moai update` on the base branch — land that commit through your project's base-branch flow, then merge the base branch into this worktree (`git merge %s` inside the worktree).",
		stamp, version.GetVersion(), worktreeBaseBranchFor(root))
}

// embeddedFallbackNotice is the one-line notice the injected context carries
// when a rule core was served from the binary's embedded template copy
// (t1729): the session must be able to tell an embedded-copy core from a
// deployed-file core, and the operator-visible remediation names the real
// cause in a worktree — an uncommitted `moai update` on the base branch —
// not a bare "run `moai update` here".
func embeddedFallbackNotice(names []string, base string) string {
	return fmt.Sprintf(
		"NOTE: the deployed copy of %s predates this binary or carries no role-core markers — the core below was injected from the binary's embedded template copy. Usual cause in a worktree: an uncommitted `moai update` the base branch has not absorbed yet — land that commit through your project's base-branch flow, then merge the base branch into this worktree (`git merge %s` inside the worktree), or run `moai update` in the primary checkout.",
		strings.Join(names, ", "), base)
}

// roleRuleLocaleTable carries the operator-facing warnings by locale. The
// operator warning is user-facing output, so it renders in the settings
// conversation language the factory notices already use; an unknown locale
// resolves to English (fail-open, per the shared locale-helper contract).
type roleRuleLocaleTable struct {
	// InjectionFailed renders the REQ-ALB-009 failure warning: the session
	// role name, the joined failure causes, and the project's base branch
	// the refresh advice names (audit round-3 P2: the branch is the
	// project's configured worktree base, not a hardcoded main).
	InjectionFailed func(session, detail, base string) string
	// OverflowUnavailable renders the overflow-delivery-unavailable retreat
	// warning: the session role name, the measured total, and the cap.
	OverflowUnavailable func(session string, total, limit int) string
	// Overflow renders the deliberate overflow-file delivery warning.
	Overflow func(session string, total, limit int) string
}

var roleRuleLocales = map[string]roleRuleLocaleTable{
	"ko": {
		InjectionFailed: func(session, detail, base string) string {
			return fmt.Sprintf("역할 규칙 주입이 %s 세션에서 실패했습니다: %s. 에이전트에는 두 규칙 파일을 경로로 읽으라는 지시가 전달됐습니다. 규칙 파일이 바이너리보다 낡은 버전 불일치라면, 워크트리에서는 보통 베이스 브랜치에 아직 반영되지 않은 `moai update` 가 원인입니다 — 프로젝트의 베이스 브랜치 반영 절차로 그 커밋을 반영한 뒤 워크트리 안에서 `git merge %s` 로 베이스 브랜치를 병합하고, 기본 체크아웃에서는 `moai update` 를 실행한 뒤 세션을 다시 시작하세요.", session, detail, base)
		},
		OverflowUnavailable: func(session string, total, limit int) string {
			return fmt.Sprintf("역할 규칙 주입 초과(%s 세션): 조립된 맥락(%d자)이 %d자 전달 한도를 넘는데 넘침 파일 전달을 쓸 수 없어, 역할 core 대신 읽기 지시를 보냈습니다.", session, total, limit)
		},
		Overflow: func(session string, total, limit int) string {
			return fmt.Sprintf("역할 규칙 주입 초과(%s 세션): 조립된 맥락(%d자)이 세션 시작 전달 한도 %d자를 넘습니다. 역할 core는 잘리지 않고 그대로 보냈으며, 런타임이 넘친 출력을 세션 디렉터리 파일로 저장해 처음 2,000자 미리보기와 함께 경로를 전달합니다.", session, total, limit)
		},
	},
	"ja": {
		InjectionFailed: func(session, detail, base string) string {
			return fmt.Sprintf("ロールルールの注入が %s セッションで失敗しました: %s。エージェントには両ルールファイルをパスで読むよう指示を送りました。ルールファイルがバイナリより古い場合(バージョン不整合)は、ワークツリーでは通常ベースブランチにまだ反映されていない `moai update` が原因です。プロジェクトのベースブランチ反映手順でそのコミットを反映してから、ワークツリー内で `git merge %s` によりベースブランチをマージし、プライマリチェックアウトでは `moai update` を実行してセッションを再起動してください。", session, detail, base)
		},
		OverflowUnavailable: func(session string, total, limit int) string {
			return fmt.Sprintf("ロールルール注入のオーバーフロー(%s セッション): 組み立てたコンテキスト(%d文字)が %d文字の配信上限を超えていますが、オーバーフローファイル配信が使えないため、ロール core の代わりに読み取り指示を送りました。", session, total, limit)
		},
		Overflow: func(session string, total, limit int) string {
			return fmt.Sprintf("ロールルール注入のオーバーフロー(%s セッション): 組み立てたコンテキスト(%d文字)がセッション開始の配信上限 %d文字を超えました。ロール core は切り詰めずそのまま送信し、ランタイムが超過出力をセッションディレクトリのファイルに保存して、先頭 2,000 文字のプレビュー付きでパスを渡します。", session, total, limit)
		},
	},
	"zh": {
		InjectionFailed: func(session, detail, base string) string {
			return fmt.Sprintf("角色规则注入在 %s 会话中失败:%s。已指示代理按路径完整读取两个规则文件。若规则文件比二进制旧(版本不匹配),工作树中的常见原因是基分支尚未吸收的 `moai update` — 请通过项目的基分支提交流程落地该提交,然后在工作树内用 `git merge %s` 合并基分支;在主检出中运行 `moai update` 并重启会话。", session, detail, base)
		},
		OverflowUnavailable: func(session string, total, limit int) string {
			return fmt.Sprintf("角色规则注入溢出(%s 会话):组装后的上下文(%d 字符)超出 %d 字符的传递上限,且溢出文件传递不可用,因此以读取指示代替角色 core。", session, total, limit)
		},
		Overflow: func(session string, total, limit int) string {
			return fmt.Sprintf("角色规则注入溢出(%s 会话):组装后的上下文(%d 字符)超出会话启动传递上限 %d 字符。角色 core 未截断、完整发出;运行时会把超长输出保存到会话目录文件,并附前 2,000 字符预览传递其路径。", session, total, limit)
		},
	},
	langEnglish: {
		InjectionFailed: func(session, detail, base string) string {
			return fmt.Sprintf(
				"Role-rule injection failed for the %s session: %s. The agent was directed to read both rule files by path. If the rule files are older than the binary (version skew), the usual cause in a worktree is an uncommitted `moai update` on the base branch — land that commit through your project's base-branch flow and merge the base branch into this worktree (`git merge %s` inside the worktree); in the primary checkout, run `moai update` and restart the session.",
				session, detail, base)
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

// roleRuleSystemYAMLRel is the deployed config file carrying the template
// generation stamp (rendered from system.yaml.tmpl's {{.Version}}) — the
// readable record of which binary last deployed these rule files.
const roleRuleSystemYAMLRel = ".moai/config/sections/system.yaml"

// detectRoleRuleVersionSkew compares the deployed system.yaml
// template_version stamp against the running binary version. ok=false means
// the rule-side stamp is absent or unreadable — the caller skips silently
// (fail-open: the skew notice is advisory, never a new failure path). The
// comparison is pure work over the parsed stamp — unit-tested with
// fixtures only (REQ-RIB-007; a live-repo self-assertion would be
// permanently red on a skewed installation).
//
// The stamp is read as YAML from moai.template_version (card t1619, codex review
// P2): a trailing comment is not part of the value, and a commented-out occurrence
// is not the stamp. A system.yaml that does not parse, or carries no stamp, is
// skipped.
func detectRoleRuleVersionSkew(systemYAML, binaryVersion string) (skewed bool, ruleVersion string, ok bool) {
	var doc struct {
		Moai struct {
			TemplateVersion string `yaml:"template_version"`
		} `yaml:"moai"`
	}
	if err := yaml.Unmarshal([]byte(systemYAML), &doc); err != nil {
		return false, "", false
	}
	stamp := strings.TrimSpace(doc.Moai.TemplateVersion)
	if stamp == "" || binaryVersion == "" {
		return false, "", false
	}
	return stamp != binaryVersion, stamp, true
}

// roleRuleSkewInfo reads the deployed system.yaml under root and reports
// whether its template_version stamp differs from the running binary.
// ok=false means the stamp is absent or unreadable — the caller skips
// silently (fail-open: the skew handling is advisory, never a new failure
// path). The comparison is pure work over the parsed stamp — unit-tested
// with fixtures only (REQ-RIB-007; a live-repo self-assertion would be
// permanently red on a skewed installation).
func roleRuleSkewInfo(root string) (skewed bool, stamp string) {
	// The non-blocking rule read (card t1619, codex review P2): a FIFO at system.yaml
	// is refused as a non-regular file instead of blocking the diagnostic.
	data, err := readRuleFileBytes(filepath.Join(root, filepath.FromSlash(roleRuleSystemYAMLRel)))
	if err != nil {
		return false, ""
	}
	skewed, stamp, ok := detectRoleRuleVersionSkew(string(data), version.GetVersion())
	if !ok {
		return false, ""
	}
	return skewed, stamp
}

// roleRuleSkewDetail renders the one-line skew for the REQ-ALB-009 failure
// detail, naming both versions and the remediation. t1729 corrected the
// advice: inside a worktree the cause is usually an uncommitted `moai
// update` sitting on the base branch, so the line names commit-then-sync —
// running `moai update` inside the worktree does not fix the class. Any
// read failure returns "" — the missing-markers failure path stays the sole
// error surface.
func roleRuleSkewDetail(root string) string {
	skewed, stamp := roleRuleSkewInfo(root)
	if !skewed {
		return ""
	}
	return fmt.Sprintf("version skew: rule files stamped %s, running binary %s — in a worktree the usual cause is an uncommitted `moai update` on the base branch: land it through your project's base-branch flow, then merge the base branch into this worktree (`git merge %s` inside the worktree); in the primary checkout, run `moai update` to refresh the rule files", stamp, version.GetVersion(), worktreeBaseBranchFor(root))
}

// roleRuleLocaleFor resolves the operator-locale table, failing open to
// English for an unknown locale.
func roleRuleLocaleFor(lang string) roleRuleLocaleTable {
	if t, ok := roleRuleLocales[lang]; ok {
		return t
	}
	return roleRuleLocales[langEnglish]
}

// roleRuleInjectionFor builds the role-rules injection pieces for one
// SessionStart event. root is the project root the deployed rule files
// resolve under; source is input.Source; lang is the operator-facing
// conversation locale the operator warnings render in.
//
// The 10,000-character size gate is NOT decided here: REQ-ALB-010 measures
// the FINAL additionalContext, and other SessionStart handlers (handoff,
// compact) merge their contributions after this one — a gate at build time
// would measure a prefix of the delivered string. The gate runs at
// dispatch-finalize time (roleRuleSizeGate, called from
// FinalizeSessionStartOutput) against the merged composite.
func roleRuleInjectionFor(root, source, lang string) roleRuleInjection {
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

	// Build every core first; any failure the embedded fallback cannot cover
	// takes the REQ-ALB-009 path.
	skewed, _ := roleRuleSkewInfo(root)
	var parts []string
	var failures []string
	var fallbackNames []string
	pointer := ""
	for _, rule := range roleRuleFiles {
		core, err := buildRoleCore(root, rule)
		fromEmbedded := false
		if err != nil {
			// t1729: a stale deployed copy — the worktree branched from a
			// main whose committed rule files predate this binary — is
			// exactly the shape a stale-until-committed update leaves
			// behind, and re-running `moai update` inside the worktree
			// cannot fix it. The binary itself carries a fresh copy, so an
			// UNMARKED deployed file falls back to the embedded template
			// copy instead of failing the session. Other failures (absent,
			// unreadable, empty, malformed markers) keep the REQ-ALB-009
			// path: the binary cannot reconstruct a file it did not
			// deploy.
			//
			// Audit rounds 2-3 (card t1729): the skew-swap arm the card
			// sketched — replacing a parseable deployed copy under a
			// differing system.yaml stamp — was removed twice-confirmed:
			// version stamps cannot establish build order (an explicit rc
			// VERSION reads HIGHER — version-management.md:97), so a swap
			// can regress a NEWER deployment. A parseable deployed core is
			// kept; the skew is surfaced by name instead (the note below,
			// the failure-detail line).
			if errors.Is(err, errRoleRuleUnmarked) {
				if fb, ferr := embeddedRoleCore(root, rule); ferr == nil {
					core, err = fb, nil
					fromEmbedded = true
				}
			}
		}
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		if core == "" {
			// A role-gated rule whose core is empty contributes its pointer
			// only (acceptance §D.2 boundary case). The pointer names the
			// path the rule is actually deployed at.
			pointer = fmt.Sprintf("Role rule (no role-core region — its binding blocks are always-loaded): `%s`", roleRuleDeployRel(root, rule))
			continue
		}
		parts = append(parts, core)
		if fromEmbedded {
			fallbackNames = append(fallbackNames, rule.Name)
		}
	}
	if len(failures) > 0 {
		detail := strings.Join(failures, "; ")
		if skew := roleRuleSkewDetail(root); skew != "" {
			detail += "; " + skew
		}
		loc := roleRuleLocaleFor(lang)
		return roleRuleInjection{
			RecoveryHead:   roleRulesReadDirective(root, ""),
			OperatorNotice: loc.InjectionFailed(role.Name, detail, worktreeBaseBranchFor(root)),
		}
	}

	var sb strings.Builder
	if len(fallbackNames) > 0 {
		// t1729: the one-line notice — the session must be able to tell an
		// embedded-copy core from a deployed-file core.
		sb.WriteString(embeddedFallbackNotice(fallbackNames, worktreeBaseBranchFor(root)))
		sb.WriteString("\n\n")
	} else if skewed {
		// Skew WITHOUT a swap (audit round-3): the deployed copy is kept,
		// and the skew rides one line naming both versions so the state is
		// visible without replacing rules on an unprovable premise.
		sb.WriteString(skewKeptNotice(root))
		sb.WriteString("\n\n")
	}
	fmt.Fprintf(&sb, "Role-gated rules (factory %s session — injected at session start; the always-loaded surface carries the stubs):\n\n", role.Name)
	sb.WriteString(strings.Join(parts, "\n\n"))
	if pointer != "" {
		sb.WriteString("\n\n")
		sb.WriteString(pointer)
	}
	context := sb.String()

	// The size gate is deferred to dispatch-finalize time — see the function
	// comment above.
	return roleRuleInjection{Context: context}
}

// roleRuleSizeGate applies the REQ-ALB-010 ladder to the FINAL merged
// additionalContext with the role core pending delivery. The unit is the
// runtime's string-length unit — UTF-16 code units (Q4), not Go bytes. It
// judges the ALREADY-ASSEMBLED context — assembled carries the earlier
// producers' text AND the role core exactly once (the session-start producer
// appended it) — and only decides shrink/warn/directive; it never re-appends
// the core. It returns the composite to install and the operator warning to
// append ("" when the composite fits the cap and no warning is owed):
//
//   - at or under the cap: the composite is returned unchanged, no warning.
//   - over the cap with overflow delivery available: the overflow read
//     directive OPENS the composite (the runtime's save-failure cut keeps
//     the head), the earlier producers' text and the intact core follow,
//     and the operator warning rides systemMessage.
//   - over the cap without overflow delivery: REQ-ALB-009 retreat — the
//     read directive opens the composite, the core is removed (zero
//     truncated units), and the operator warning rides systemMessage.
//
// The core parameter is the exact string the producer appended — used only
// for the retreat removal.
func roleRuleSizeGate(assembled, core, root, roleName, lang string) (composite, operator string) {
	total := utf16Len(assembled)
	if total <= roleRulesContextLimit {
		return assembled, ""
	}
	loc := roleRuleLocaleFor(lang)
	if !roleRulesOverflowDelivery() {
		retreated := strings.Replace(assembled, "\n\n"+core, "", 1)
		composite := roleRulesReadDirective(root, "") + "\n\n" + retreated
		return composite, loc.OverflowUnavailable(roleName, total, roleRulesContextLimit)
	}
	head := roleRulesOverflowDirective(root)
	return head + "\n\n" + assembled,
		loc.Overflow(roleName, total, roleRulesContextLimit)
}

// roleRulesRootFromCWD resolves the project root the deployed role-gated
// rule files resolve under for a session whose cwd may sit anywhere inside
// the tree: it walks the cwd's ancestors outward and returns the first
// ancestor carrying the deployed dispatch rule, and the cwd itself when no
// ancestor does — leaving the REQ-ALB-009 fail-visible path to name what is
// missing. The walk STOPS at the project boundary: a directory carrying a
// .git marker (a checkout directory, or a worktree's repository pointer
// file) OR its own .moai/ directory (a nested MoAI project, which may have
// no .git) ends the walk at that directory rather than reaching into a
// parent project — a lane worktree or nested project that lacks the
// deployed rules takes the REQ-ALB-009 missing path, never its parent
// tree's rules. Pure path arithmetic, no git subprocess: the SessionStart
// hook runs under a 5s budget (the same reasoning as cardIDFromPath).
func roleRulesRootFromCWD(cwd string) string {
	if strings.TrimSpace(cwd) == "" {
		return ""
	}
	deployedHere := func(dir string) bool {
		for _, rule := range roleRuleFiles {
			if rel := roleRuleDeployRel(dir, rule); rel != rule.Rel {
				return true // found at the alternate (Codex-only) surface
			}
			if fi, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rule.Rel))); err == nil && !fi.IsDir() {
				return true
			}
		}
		return false
	}
	for dir := filepath.Clean(cwd); ; {
		if deployedHere(dir) {
			return dir
		}
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			// Project boundary: this directory is the checkout root — the
			// session's project root even when its rules are missing. No
			// reach beyond this checkout.
			return dir
		}
		if fi, err := os.Stat(filepath.Join(dir, ".moai")); err == nil && fi.IsDir() {
			// MoAI project boundary: a directory carrying its own .moai/ is
			// a project root even without .git (a nested MoAI project) — the
			// walk stops here rather than climbing into the parent project
			// and injecting the parent's rules for a child that lacks its
			// own (REQ-ALB-009 missing path names what the child lacks).
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
// full rule files by path before acting. The named paths are joined to the
// resolved root — a session whose cwd sits in a subdirectory must receive
// paths that exist on disk, not root-relative fragments that resolve to
// nothing from where the session stands.
func roleRulesReadDirective(root, reason string) string {
	rel0 := roleRuleDeployRel(root, roleRuleFiles[0])
	rel1 := roleRuleDeployRel(root, roleRuleFiles[1])
	var sb strings.Builder
	if reason != "" {
		sb.WriteString(reason)
		sb.WriteString(" ")
	}
	sb.WriteString("[HARD] Read both role-gated rule files in full before your first action: `")
	sb.WriteString(filepath.Join(root, filepath.FromSlash(rel0)))
	sb.WriteString("` and `")
	sb.WriteString(filepath.Join(root, filepath.FromSlash(rel1)))
	sb.WriteString("`.")
	return sb.String()
}

// roleRulesOverflowDirective is the agent-facing REQ-ALB-010 directive that
// rides an intact over-cap emission: the runtime file is the delivery
// channel, and the rule files are the fallback read — named root-joined for
// the same reason the read directive's paths are.
func roleRulesOverflowDirective(root string) string {
	// The cap is rendered with a thousands separator (10,000) to match the
	// locale warnings; roleRulesContextLimit stays the numeric source.
	return fmt.Sprintf(
		"NOTE: the output above exceeds the session-start delivery cap (10,000 characters). The runtime saved it to a file in the session directory — its path and a 2,000-character preview follow — read the role core from that file, or read the rule files by path: `%s`, `%s`.",
		filepath.Join(root, filepath.FromSlash(roleRuleDeployRel(root, roleRuleFiles[0]))),
		filepath.Join(root, filepath.FromSlash(roleRuleDeployRel(root, roleRuleFiles[1]))))
}
