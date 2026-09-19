package rosterguard

// Registry is the declared set of roster-listing sites in this tree.
//
// Every row was measured at base 26243ea86 (card t922) — none was written from
// memory or inferred from a file name. A site the anti-vacuity sweep finds and
// this registry does not carry fails the guard with an instruction to declare
// its axis; that failure is the whole point, because it is what stops a new
// listing from being added silently and going stale unobserved.
//
// # Repair is deliberately NOT done here
//
// Several rows carry a KnownStale marker. The card that created this package
// scoped repair OUT: this guard's job is to make staleness enumerable, not to
// rewrite prose whose owning decision belongs elsewhere. A KnownStale marker is
// self-expiring rather than a mute — see Staleness — so a repair landing later
// fails the guard with "delete the marker", and further drift fails it with the
// new gap.
func Registry() []Site {
	return []Site{
		// ── The canonical source ───────────────────────────────────────────
		{
			ID:         "profile-matrix-order",
			Path:       "internal/template/profile_matrix.go",
			Axis:       AxisRetainedRoster,
			Claims:     ClaimMembership,
			BlockStart: "var profileMatrixAgentOrder = []string{",
			BlockEnd:   "}",
			Note: "The canonical roster, exposed as template.ProfileMatrixAgents(). " +
				"Asserted against the accessor rather than assumed equal to it, so a " +
				"defensive-copy or ordering change in the accessor cannot silently " +
				"decouple the two.",
		},
		{
			ID:         "profile-matrix-group-membership",
			Path:       "internal/template/profile_matrix.go",
			Axis:       AxisRetainedRoster,
			Claims:     ClaimMembership,
			BlockStart: "var agentGroupMembership = map[string]string{",
			BlockEnd:   "}",
			Note: "The agent→group SSOT in the same file as the canonical order. It is a " +
				"SECOND roster in one file, which is why a site key is (path, block) " +
				"rather than path alone.",
		},

		// ── Go rosters that have drifted ───────────────────────────────────
		{
			ID:         "config-retained-agent-names",
			Path:       "internal/config/profile.go",
			Axis:       AxisRetainedRoster,
			Claims:     ClaimMembership,
			BlockStart: "var retainedAgentNames = map[string]bool{",
			BlockEnd:   "}",
			// KnownStale deleted: the marker declared manager-lead absent and named
			// card t916 (commit 1814bf3e9) as the repair, noting that commit was not
			// yet an ancestor. It is now, and manager-lead is present in the map, so
			// the declared gap no longer exists. The marker expired the way Staleness
			// is designed to — it failed the guard with "delete the marker" rather
			// than going quietly stale, which is what makes it a record and not a
			// mute. The forward-only-propagation instance it recorded (mission-governor
			// registered here by its own creation commit 5ec516165 while manager-lead,
			// which arrived via the rename 310d75dd2, never was) is preserved in this
			// package's doc comment, where it is the motivating measurement.
		},
		{
			ID:         "delegationmap-retained-catalog",
			Path:       "internal/harness/delegationmap/types.go",
			Axis:       AxisRetainedRoster,
			Claims:     ClaimMembership | ClaimCount,
			BlockStart: "var retainedCatalog = map[string]struct{}{",
			// NOT "}": every entry in this map has a `{}` value, so a bare
			// closing-brace terminator matches the FIRST entry line and the
			// block collapses to one name.
			BlockEnd: "// IsRetainedAgent reports whether name",
			// The §4 on this line is a section number, not a count — hence a
			// capture group rather than a first-integer-on-the-line rule.
			//
			// No closing paren in the pattern: card t917 widened the citation to
			// "(the 13 retained agents — 12 MoAI-custom plus the / Anthropic
			// built-in Explore)", which wraps, so the ")" is no longer on this
			// line. Anchoring on it made the pattern match 0 times — a stale
			// anchor, not a stale count, and the guard reported exactly that.
			CountPattern: `CLAUDE\.md §4 \(the (\d+) retained agents`,
			// KnownStale deleted: card t917 landed mission-governor here
			// (origin/develop 8665f9eac), so this roster now equals the
			// canonical 13 and there is no gap left to declare. Keeping the
			// marker would have been the silent exemption this package exists
			// to prevent.
		},
		{
			ID:     "profile-matrix-test-expectations",
			Path:   "internal/template/profile_matrix_test.go",
			Axis:   AxisRetainedRoster,
			Claims: ClaimMembership,
			Note: "Whole-file block: the file is the canonical roster's own test and mentions " +
				"agent names only inside its expectation maps.",
			KnownStale: &Staleness{
				Reason: "The canonical roster's own test carries mission-governor in no expectation " +
					"map, so the cell added for it is unexercised by these tables.",
				FollowUp:     "unassigned — reported by card t922, repair not in its scope",
				MissingNames: []string{"mission-governor"},
			},
		},

		// ── Definition-file axis (12; the built-in Explore has no file) ────
		{
			ID:     "v4manifest-agent-tiers",
			Path:   "internal/harness/v4manifest/schema.go",
			Axis:   AxisDefinitionFiles,
			Claims: ClaimMembership,
			Note: "Tier assignments exist per agent DEFINITION; Explore has no definition file " +
				"and so no tier. Registering this on AxisRetainedRoster would report a " +
				"correct file as broken.",
		},
		{
			ID:     "v4manifest-tier-test",
			Path:   "internal/harness/v4manifest/tier_test.go",
			Axis:   AxisDefinitionFiles,
			Claims: ClaimMembership,
			Note:   "The expectation table pinning the tier map above; same axis for the same reason.",
		},
		{
			ID:     "agentemit-golden",
			Path:   "internal/template/agentemit/golden_test.go",
			Axis:   AxisDefinitionFiles,
			Claims: ClaimMembership,
			Note:   "The emitter operates on definition files, so its golden set is the file population.",
		},
		{
			ID:     "template-catalog",
			Path:   "internal/template/catalog.yaml",
			Axis:   AxisDefinitionFiles,
			Claims: ClaimMembership,
			Note:   "An inventory of templates/.claude/agents/moai/*.md — the file population by construction.",
		},
		{
			ID:     "web-i18n-agent-descriptions",
			Path:   "internal/web/assets/i18n.js",
			Axis:   AxisDefinitionFiles,
			Claims: ClaimMembership,
			Note:   "One agentdesc.* key per definition file, in four locales.",
		},

		// ── Retained-roster sites that are currently consistent ────────────
		{
			ID:           "claude-md-section-4",
			Path:         "CLAUDE.md",
			Axis:         AxisRetainedRoster,
			Claims:       ClaimMembership | ClaimCount,
			BlockStart:   "**Retained agents (",
			CountPattern: `consists of exactly \*\*(\d+) retained agents\*\*`,
			Note:         "Single-line block: the §4 enumeration sentence. Repaired by card t909.",
		},
		{
			ID:           "claude-md-section-4-mirror",
			Path:         "internal/template/templates/CLAUDE.md",
			Axis:         AxisRetainedRoster,
			Claims:       ClaimMembership | ClaimCount,
			BlockStart:   "**Retained agents (",
			CountPattern: `consists of exactly \*\*(\d+) retained agents\*\*`,
			Note:         "Template mirror of the row above; both copies are registered so a repair to one cannot leave the other behind.",
		},
		{
			ID:           "agent-authoring-catalog",
			Path:         ".claude/rules/moai/development/agent-authoring.md",
			Axis:         AxisRetainedRoster,
			Claims:       ClaimMembership | ClaimCount,
			BlockStart:   "### Retained MoAI-custom Agents (",
			BlockEnd:     "- Explore: Read-only codebase exploration",
			CountPattern: `consists of exactly \*\*(\d+) retained agents\*\*`,
			Note:         "The block deliberately spans into the following §Anthropic Built-in section so Explore is inside the membership region.",
		},
		{
			ID:           "agent-authoring-catalog-mirror",
			Path:         "internal/template/templates/.claude/rules/moai/development/agent-authoring.md",
			Axis:         AxisRetainedRoster,
			Claims:       ClaimMembership | ClaimCount,
			BlockStart:   "### Retained MoAI-custom Agents (",
			BlockEnd:     "- Explore: Read-only codebase exploration",
			CountPattern: `consists of exactly \*\*(\d+) retained agents\*\*`,
		},
		{
			ID:           "agent-patterns-static-file-criterion",
			Path:         ".claude/rules/moai/development/agent-patterns.md",
			Axis:         AxisDefinitionFiles,
			Claims:       ClaimMembership | ClaimCount,
			BlockStart:   "MoAI-custom retained agents (`manager-spec`",
			CountPattern: `The (\d+) MoAI-custom retained agents`,
			Note: "The sentence enumerates the MoAI-CUSTOM agents — the population that has " +
				"definition files. The coincidence with the file count is causal, not " +
				"accidental: the criterion the sentence states is exactly what earns an " +
				"agent a static file.",
		},
		{
			ID:           "agent-patterns-static-file-criterion-mirror",
			Path:         "internal/template/templates/.claude/rules/moai/development/agent-patterns.md",
			Axis:         AxisDefinitionFiles,
			Claims:       ClaimMembership | ClaimCount,
			BlockStart:   "MoAI-custom retained agents (`manager-spec`",
			CountPattern: `The (\d+) MoAI-custom retained agents`,
		},
		{
			ID:     "shipped-key-inventory",
			Path:   "internal/config/testdata/shipped_key_inventory.yaml",
			Axis:   AxisRetainedRoster,
			Claims: ClaimMembership,
			Note:   "Derived llm.profiles.* key inventory — one key pair per matrix row, so it tracks the full retained roster.",
		},
		{
			ID:     "template-llm-yaml",
			Path:   "internal/template/templates/.moai/config/sections/llm.yaml",
			Axis:   AxisRetainedRoster,
			Claims: ClaimMembership,
			Note:   "The shipped per-agent profile cells.",
		},

		// ── Count-only sites: a roster SIZE claim with no roster ───────────
		//
		// These two are the reason the sweep alone is not enough. The sweep is
		// name-enumeration-driven (>= SweepThreshold distinct names), and
		// tech.md states a retained-roster size while naming ZERO agents —
		// product.md names one. Lowering the threshold does not reach them:
		// the axis they ENUMERATE on and the axis they CLAIM on are different,
		// so no enumeration threshold can. They are registered by hand here,
		// and closing the general hole is card t930 (a numeral-adjacency layer,
		// the shape internal/web/docs_tab_contract_test.go already implements
		// for a different subject).
		{
			ID:               "product-md-profile-matrix-size",
			SweepUnreachable: "count-only claim; the sentence names one agent in passing, far below SweepThreshold",
			Path:             ".moai/project/product.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `(\d+) retained agents x 3 model tiers`,
			Note: "Count only: the sentence sizes the profile matrix and names one agent in " +
				"passing, so there is no membership to assert. Invisible to the sweep.",
			KnownStale: &Staleness{
				Reason:        "Cites 11 where the retained roster carries 13; the 33-cell figure is sized off that stale count.",
				FollowUp:      "card t930 (numeral-adjacency layer); the prose repair itself is unassigned",
				DeclaredCount: 11,
			},
		},
		{
			ID:               "tech-md-profile-matrix-size",
			SweepUnreachable: "count-only claim; this file names ZERO agents while sizing the roster",
			Path:             ".moai/project/tech.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `(\d+) retained agents x 3 model tiers`,
			Note: "Count only, and the starker case: this file names ZERO agents while " +
				"claiming a roster size. No enumeration threshold reaches it.",
			KnownStale: &Staleness{
				Reason:        "Cites 11 where the retained roster carries 13, identically to product.md.",
				FollowUp:      "card t930 (numeral-adjacency layer); the prose repair itself is unassigned",
				DeclaredCount: 11,
			},
		},

		// ── Retained-roster sites with a measured stale claim ──────────────
		{
			ID:           "docs-truth-catalog",
			Path:         ".moai/project/codemaps/docs-truth.md",
			Axis:         AxisRetainedRoster,
			Claims:       ClaimMembership | ClaimCount,
			BlockStart:   "| # | Agent | Class | Phase scope |",
			BlockEnd:     "| `Explore` | Anthropic built-in",
			CountPattern: `consists of exactly \*\*(\d+) retained agents\*\*`,
			KnownStale: &Staleness{
				Reason: "The §1 table carries 12 rows and the file records the mission-governor " +
					"gap in its own body as an unresolved drift it declines to adjudicate.",
				FollowUp:      "the catalog-owning document's decision, per that file's own note",
				MissingNames:  []string{"mission-governor"},
				DeclaredCount: 12,
			},
		},
		{
			ID:           "spec-workflow-catalog-sentence",
			Path:         ".claude/rules/moai/workflow/spec-workflow.md",
			Axis:         AxisRetainedRoster,
			Claims:       ClaimMembership | ClaimCount,
			BlockStart:   "the MoAI agent catalog consists of exactly",
			CountPattern: `consists of exactly (\d+) retained agents`,
			KnownStale: &Staleness{
				Reason:        "Claims completeness while enumerating 11; manager-lead and mission-governor are absent.",
				FollowUp:      "repair scoped OUT of card t922 by the lead; unassigned",
				MissingNames:  []string{"manager-lead", "mission-governor"},
				DeclaredCount: 11,
			},
		},
		{
			ID:           "spec-workflow-catalog-sentence-mirror",
			Path:         "internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md",
			Axis:         AxisRetainedRoster,
			Claims:       ClaimMembership | ClaimCount,
			BlockStart:   "the MoAI agent catalog consists of exactly",
			CountPattern: `consists of exactly (\d+) retained agents`,
			KnownStale: &Staleness{
				Reason:        "Template mirror of the row above, stale identically.",
				FollowUp:      "repair scoped OUT of card t922 by the lead; unassigned",
				MissingNames:  []string{"manager-lead", "mission-governor"},
				DeclaredCount: 11,
			},
		},
		{
			ID:           "agents-reference-catalog",
			Path:         ".claude/skills/moai-foundation-core/modules/agents-reference.md",
			Axis:         AxisRetainedRoster,
			Claims:       ClaimMembership | ClaimCount,
			BlockStart:   "| Agent | Phase scope |",
			BlockEnd:     "| `Explore` | Read-only codebase exploration",
			CountPattern: `\*\*(\d+) retained agents\*\*`,
			KnownStale: &Staleness{
				Reason:        "Claims completeness (\"11 retained agents: 10 MoAI-custom plus the Anthropic built-in Explore\"); manager-lead and mission-governor are absent.",
				FollowUp:      "repair scoped OUT of card t922 by the lead; unassigned",
				MissingNames:  []string{"manager-lead", "mission-governor"},
				DeclaredCount: 11,
			},
		},
		{
			ID:           "agents-reference-catalog-mirror",
			Path:         "internal/template/templates/.claude/skills/moai-foundation-core/modules/agents-reference.md",
			Axis:         AxisRetainedRoster,
			Claims:       ClaimMembership | ClaimCount,
			BlockStart:   "| Agent | Phase scope |",
			BlockEnd:     "| `Explore` | Read-only codebase exploration",
			CountPattern: `\*\*(\d+) retained agents\*\*`,
			KnownStale: &Staleness{
				Reason:        "Template mirror of the row above, stale identically.",
				FollowUp:      "repair scoped OUT of card t922 by the lead; unassigned",
				MissingNames:  []string{"manager-lead", "mission-governor"},
				DeclaredCount: 11,
			},
		},
		readmeSite("readme-en", "README.md", `### The (\d+)-agent catalog`),
		readmeSite("readme-ko", "README.ko.md", `### (\d+)-에이전트 카탈로그`),
		readmeSite("readme-ja", "README.ja.md", `### (\d+) エージェント・カタログ`),
		readmeSite("readme-zh", "README.zh.md", `### (\d+) 智能体目录`),

		// ── Subsets by design ──────────────────────────────────────────────
		{
			ID:           "delegation-map-count-citation",
			Path:         ".moai/config/sections/delegation.yaml",
			Axis:         AxisSubsetByDesign,
			Claims:       ClaimCount,
			CountPattern: `the (\d+) retained agents \(CLAUDE\.md section 4\)`,
			Note: "The CONTENT is a subset by design — the per-subcommand `agents:` lists are " +
				"partial, and manager-design, manager-lead, super-advisor and " +
				"mission-governor are all legitimately absent from every designation. Only " +
				"the header's citation of the retained-roster SIZE is asserted, and that " +
				"citation is now correct.",
			// KnownStale deleted: the header moved 11 -> 13 while this card
			// waited for its window. The content was never the stale part.
		},
		{
			ID:           "delegation-map-count-citation-mirror",
			Path:         "internal/template/templates/.moai/config/sections/delegation.yaml",
			Axis:         AxisSubsetByDesign,
			Claims:       ClaimCount,
			CountPattern: `the (\d+) retained agents \(CLAUDE\.md section 4\)`,
			// KnownStale deleted with the row above: both copies were repaired
			// together, which is what registering the mirror separately was for.
		},
		{
			ID:     "web-agentfm-display-rank",
			Path:   "internal/web/agentfm.go",
			Axis:   AxisSubsetByDesign,
			Claims: 0,
			Note: "agentGroupRank classifies 9 named agents into display buckets and routes " +
				"everything else — Explore, harness specialists, and anything added later — " +
				"into a documented trailing bucket. Asserting complete membership here would " +
				"be wrong. OBSERVED, not adjudicated: manager-lead and mission-governor " +
				"currently land in that trailing bucket, so two managers render outside the " +
				"manager group; whether that is the intended degradation or a defect is the " +
				"owning surface's call, not this guard's.",
		},
		{
			ID:     "web-agentfm-display-rank-test",
			Path:   "internal/web/agentfm_ordering_test.go",
			Axis:   AxisSubsetByDesign,
			Claims: 0,
			Note:   "The expectation table for the subset above; same boundary, same reason.",
		},
	}
}

// readmeSite builds the four locale rows of the README agent-catalog table,
// which differ only in the localized heading their count claim matches.
//
// Derived rather than written out four times: four hand-copied rows is the same
// forward-only-propagation shape this package guards, one level up.
func readmeSite(id, path, countPattern string) Site {
	return Site{
		ID:           id,
		Path:         path,
		Axis:         AxisRetainedRoster,
		Claims:       ClaimMembership | ClaimCount,
		BlockStart:   "| manager-spec |",
		BlockEnd:     "| Explore |",
		CountPattern: countPattern,
		Note: "Block-scoped to the table rather than the whole file: the README mentions " +
			"mission-governor in unrelated prose, so a whole-file assertion would pass " +
			"while the table omits it — the exact false pass this scoping exists to prevent.",
		// KnownStale deleted: the four README tables gained their
		// mission-governor row and their headings moved 12 -> 13 while this card
		// waited for its merge window. The guard reported both halves per
		// locale — the membership gap AND the count — which is why the repair
		// needed no announcement to be seen.
	}
}
