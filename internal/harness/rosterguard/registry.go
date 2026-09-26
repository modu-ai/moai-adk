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
		{
			ID:           "docs-truth-catalog",
			Path:         ".moai/project/codemaps/docs-truth.md",
			Axis:         AxisRetainedRoster,
			Claims:       ClaimMembership | ClaimCount,
			BlockStart:   "| # | Agent | Class | Phase scope |",
			BlockEnd:     "| `Explore` | Anthropic built-in",
			CountPattern: `consists of exactly \*\*(\d+) retained agents\*\*`,
			Note:         "§1 table repaired to 13 rows by card t1069; its stale marker was retired by card t1091.",
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
		readmeSite("readme-en", "README.md", `### The (\d+)-agent catalog`, ""),
		readmeSite("readme-ko", "README.ko.md", `### (\d+)-에이전트 카탈로그`, localizedHeadingUnreachable),
		readmeSite("readme-ja", "README.ja.md", `### (\d+) エージェント・カタログ`, localizedHeadingUnreachable),
		readmeSite("readme-zh", "README.zh.md", `### (\d+) 智能体目录`, localizedHeadingUnreachable),

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

		// ── Count-only sites reached by the NUMERAL layer (card t930) ──────
		//
		// Every row below states a roster SIZE while enumerating far fewer than
		// SweepThreshold names, so the enumeration sweep cannot reach any of
		// them — which is exactly why the numeral layer (numeral.go) exists.
		// Each carries SweepUnreachable for that reason, and a CountPattern for
		// its own claim.
		//
		// Repair is NOT done here, in the stance this file already takes: a
		// KnownStale marker keeps the staleness enumerable and self-expiring,
		// and the prose repair belongs to whoever owns the document.
		//
		// A template mirror gets its OWN row rather than deriving from its local
		// twin. Deriving would cost a reviewer one row instead of two and would
		// blind the guard to a repair landing on only one copy of a pair — the
		// case recorded a few rows above, firing.
		// The two rows below carried three declarations written against a
		// sentence that t1127/t1131/t1140 replaced, and all three had to go
		// together (card t1141). The document is the party that became CORRECT:
		// it now reads "13 rows — 12 agents plus `Explore` — × 3 columns = 39
		// cells", and 13 is what template.ProfileMatrixAgents() carries — the
		// set CheckSite actually compares against (delegationmap's
		// retainedCatalog and CLAUDE.md §4 agree, as cross-checks). The
		// registry was the stale party.
		//
		// What each dropped declaration had asserted, and why it is now false:
		//
		//   KnownStale (DeclaredCount 11) — the recorded staleness is REPAIRED.
		//   Leaving the marker is not the safe side: check.go fires on a marker
		//   whose declared value no longer matches the site, so a resolved
		//   staleness left behind reads as a fresh drift.
		//
		//   SweepUnreachable ("names no agents") — the file now mentions twelve
		//   of the thirteen roster names, so the sweep reaches it and the
		//   exemption asserts something measurably untrue.
		//
		// The names are NOT replaced by a membership assertion, which is what
		// the sweep's own message suggests. The reason is a NAMING system, not
		// an omission — the tier table's effort-baseline cell does partition the
		// whole roster, but it writes six of the thirteen in shorthand
		// (`design`, `lead`, `harness`, `e2e`, `docs`, `git`) beside seven
		// canonical names. NamesIn bounds a name by non-identifier characters,
		// and `-` is one of them, so `lead` never satisfies `manager-lead`.
		// A membership claim keyed on canonical names would therefore report
		// four agents that ARE in the table as absent, and silencing that would
		// mean either four fabricated KnownStale gaps or expanding the
		// document's shorthand to suit the test.
		//
		// NumeralUnreachable is added rather than widening the numeral layer:
		// the new sentence counts "rows" and plain "agents", neither of which is
		// in rosterNounRe's noun class, so the layer cannot reach this claim
		// without a vocabulary change that would move every other file's breadth
		// set too.
		{
			ID:                 "model-policy-profile-matrix-size",
			NumeralUnreachable: "the count noun is `rows` (and `agents` unqualified), outside the numeral layer's noun class",
			Path:               ".claude/rules/moai/development/model-policy.md",
			Axis:               AxisRetainedRoster,
			Claims:             ClaimCount,
			CountPattern:       `\((\d+) rows — \d+ agents plus`,
		},
		{
			ID:                 "model-policy-profile-matrix-size-mirror",
			NumeralUnreachable: "same noun class miss; template mirror of the row above",
			Path:               "internal/template/templates/.claude/rules/moai/development/model-policy.md",
			Axis:               AxisRetainedRoster,
			Claims:             ClaimCount,
			CountPattern:       `\((\d+) rows — \d+ agents plus`,
		},
		{
			ID:               "foundation-core-skill-catalog-size",
			SweepUnreachable: "count-only claim: a module-index sentence citing the catalog size",
			Path:             ".claude/skills/moai-foundation-core/SKILL.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `(\d+)-agent retained catalog`,
			KnownStale: &Staleness{
				Reason:        "Cites an 11-agent retained catalog (10 MoAI-custom + Explore) where the roster carries 13.",
				FollowUp:      "unassigned — reported by card t930, prose repair out of its scope",
				DeclaredCount: 11,
			},
		},
		{
			ID:               "foundation-core-skill-catalog-size-mirror",
			SweepUnreachable: "count-only claim; template mirror of the row above",
			Path:             "internal/template/templates/.claude/skills/moai-foundation-core/SKILL.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `(\d+)-agent retained catalog`,
			KnownStale: &Staleness{
				Reason:        "Template mirror of the row above, stale identically.",
				FollowUp:      "unassigned — reported by card t930, prose repair out of its scope",
				DeclaredCount: 11,
			},
		},

		// INDEX.md states the same size THREE times, in three different
		// sentences. Three rows rather than one: a CountPattern must match its
		// body exactly once, and a pattern loose enough to cover all three
		// would be ambiguous about which claim it asserts.
		{
			ID:               "foundation-core-index-catalog-size-headline",
			SweepUnreachable: "count-only claim: a module-index line",
			Path:             ".claude/skills/moai-foundation-core/modules/INDEX.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `MoAI-ADK's (\d+) retained agents`,
			KnownStale:       indexCatalogStale("Cites 11 where the retained roster carries 13."),
		},
		{
			ID:               "foundation-core-index-catalog-size-bullet",
			SweepUnreachable: "count-only claim: a module-index line",
			Path:             ".claude/skills/moai-foundation-core/modules/INDEX.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `- (\d+) retained agents \(10 MoAI-custom`,
			KnownStale:       indexCatalogStale("Cites 11 where the retained roster carries 13; the same file's third claim."),
		},
		{
			ID:               "foundation-core-index-catalog-size-table",
			SweepUnreachable: "count-only claim: a module-index table cell",
			Path:             ".claude/skills/moai-foundation-core/modules/INDEX.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `\| (\d+) retained agents, flat catalog`,
			KnownStale:       indexCatalogStale("Cites 11 where the retained roster carries 13, in the module table."),
		},
		{
			ID:               "foundation-core-index-catalog-size-headline-mirror",
			SweepUnreachable: "count-only claim; template mirror",
			Path:             "internal/template/templates/.claude/skills/moai-foundation-core/modules/INDEX.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `MoAI-ADK's (\d+) retained agents`,
			KnownStale:       indexCatalogStale("Template mirror, stale identically."),
		},
		{
			ID:               "foundation-core-index-catalog-size-bullet-mirror",
			SweepUnreachable: "count-only claim; template mirror",
			Path:             "internal/template/templates/.claude/skills/moai-foundation-core/modules/INDEX.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `- (\d+) retained agents \(10 MoAI-custom`,
			KnownStale:       indexCatalogStale("Template mirror, stale identically."),
		},
		{
			ID:               "foundation-core-index-catalog-size-table-mirror",
			SweepUnreachable: "count-only claim; template mirror",
			Path:             "internal/template/templates/.claude/skills/moai-foundation-core/modules/INDEX.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `\| (\d+) retained agents, flat catalog`,
			KnownStale:       indexCatalogStale("Template mirror, stale identically."),
		},

		// The three foundation-core modules whose header banner cites the
		// catalog size in one identical sentence.
		agentCatalogSizeSite("foundation-core-delegation-advanced-catalog-size", ".claude/skills/moai-foundation-core/modules/delegation-advanced.md"),
		agentCatalogSizeSite("foundation-core-delegation-advanced-catalog-size-mirror", "internal/template/templates/.claude/skills/moai-foundation-core/modules/delegation-advanced.md"),
		agentCatalogSizeSite("foundation-core-delegation-implementation-catalog-size", ".claude/skills/moai-foundation-core/modules/delegation-implementation.md"),
		agentCatalogSizeSite("foundation-core-delegation-implementation-catalog-size-mirror", "internal/template/templates/.claude/skills/moai-foundation-core/modules/delegation-implementation.md"),
		agentCatalogSizeSite("foundation-core-token-optimization-catalog-size", ".claude/skills/moai-foundation-core/modules/token-optimization.md"),
		agentCatalogSizeSite("foundation-core-token-optimization-catalog-size-mirror", "internal/template/templates/.claude/skills/moai-foundation-core/modules/token-optimization.md"),

		{
			ID:               "foundation-quality-skill-catalog-size",
			SweepUnreachable: "count-only claim: a cross-reference sentence citing the catalog size",
			Path:             ".claude/skills/moai-foundation-quality/SKILL.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `for the (\d+)-agent catalog`,
			KnownStale: &Staleness{
				Reason:        "Cites an 11-agent catalog where the roster carries 13.",
				FollowUp:      "unassigned — reported by card t930, prose repair out of its scope",
				DeclaredCount: 11,
			},
		},
		{
			ID:               "foundation-quality-skill-catalog-size-mirror",
			SweepUnreachable: "count-only claim; template mirror of the row above",
			Path:             "internal/template/templates/.claude/skills/moai-foundation-quality/SKILL.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `for the (\d+)-agent catalog`,
			KnownStale: &Staleness{
				Reason:        "Template mirror of the row above, stale identically.",
				FollowUp:      "unassigned — reported by card t930, prose repair out of its scope",
				DeclaredCount: 11,
			},
		},

		// manager-design cites the roster size and is CURRENTLY CORRECT. It is
		// registered for exactly that reason: an already-correct count is what a
		// guard protects, and leaving it undeclared would mean the next drift in
		// it goes unreported. Three copies — the local definition, the deployed
		// mirror, and the machine-emitted codex form — each take a row.
		{
			ID:               "manager-design-catalog-citation",
			SweepUnreachable: "count-only claim: a Context field citing the roster size, naming one agent",
			Path:             ".claude/agents/moai/manager-design.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `§ 4 \((\d+) retained agents`,
		},
		{
			ID:               "manager-design-catalog-citation-mirror",
			SweepUnreachable: "count-only claim; template mirror of the row above",
			Path:             "internal/template/templates/.claude/agents/moai/manager-design.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `§ 4 \((\d+) retained agents`,
		},
		{
			ID:               "manager-design-catalog-citation-codex",
			SweepUnreachable: "count-only claim; the machine-emitted codex form of the row above",
			Path:             "internal/template/templates/.codex/agents/moai/manager-design.toml",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `§ 4 \((\d+) retained agents`,
			Note: "Emitted from the .claude mirror by internal/template/agentemit and never hand-edited. " +
				"Registered anyway: registering is a read, not an edit, and a stale emission is still a stale claim on disk.",
		},

		// NOTICE.md carries BOTH a historical citation and a live one in the
		// same sentence: "8 retained agents at consolidation time; now 10 per
		// CLAUDE.md §4". The LIVE half is registered here; the historical half
		// needs no repair and gets none. Registering the live claim rather than
		// exempting the whole path is what keeps the stale 10 enumerable.
		{
			ID:               "notice-current-catalog-size",
			SweepUnreachable: "count-only claim: an attribution paragraph citing the current roster size",
			Path:             ".claude/rules/moai/NOTICE.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `now (\d+) per CLAUDE\.md §4`,
			KnownStale: &Staleness{
				Reason: "The live half of the sentence says the catalog is now 10; the roster carries 13. " +
					"The historical half (8 at consolidation time) is correct and is not this row's subject.",
				FollowUp:      "unassigned — reported by card t930, prose repair out of its scope",
				DeclaredCount: 10,
			},
		},
		{
			ID:               "notice-current-catalog-size-mirror",
			SweepUnreachable: "count-only claim; template mirror of the row above",
			Path:             "internal/template/templates/.claude/rules/moai/NOTICE.md",
			Axis:             AxisRetainedRoster,
			Claims:           ClaimCount,
			CountPattern:     `now (\d+) per CLAUDE\.md §4`,
			KnownStale: &Staleness{
				Reason:        "Template mirror of the row above, stale identically.",
				FollowUp:      "unassigned — reported by card t930, prose repair out of its scope",
				DeclaredCount: 10,
			},
		},

		// ── Registered count rows the NUMERAL layer cannot reach ───────────
		//
		// The three localized README headings state the roster size in their own
		// language, and the adopted noun class is English-only by decision, so
		// the layer is expected not to see them. Declaring that — rather than
		// widening the noun class or quietly dropping the equality — is what
		// keeps AC-RNA-006(b) strict instead of unsatisfiable.
	}
}

// indexCatalogStale builds the staleness marker shared by the INDEX.md rows,
// which differ only in which sentence they anchor on.
func indexCatalogStale(reason string) *Staleness {
	return &Staleness{
		Reason:        reason,
		FollowUp:      "unassigned — reported by card t930, prose repair out of its scope",
		DeclaredCount: 11,
	}
}

// agentCatalogSizeSite builds a row for the foundation-core module banner,
// which cites the catalog size in one identical sentence across three modules
// and their three mirrors.
func agentCatalogSizeSite(id, path string) Site {
	return Site{
		ID:               id,
		SweepUnreachable: "count-only claim: a module header banner citing the catalog size",
		Path:             path,
		Axis:             AxisRetainedRoster,
		Claims:           ClaimCount,
		CountPattern:     `(\d+)-agent catalog in \[agents-reference\.md\]`,
		KnownStale: &Staleness{
			Reason:        "The module banner cites an 11-agent catalog where the roster carries 13.",
			FollowUp:      "unassigned — reported by card t930, prose repair out of its scope",
			DeclaredCount: 11,
		},
	}
}

// readmeSite builds the four locale rows of the README agent-catalog table,
// which differ only in the localized heading their count claim matches.
//
// Derived rather than written out four times: four hand-copied rows is the same
// forward-only-propagation shape this package guards, one level up.
// localizedHeadingUnreachable is the reason the three non-English READMEs
// carry NumeralUnreachable: their count claim is written with a LOCALIZED
// roster noun (에이전트 카탈로그 / エージェント・カタログ / 智能体目录), and the
// numeral layer's adopted noun class is English-only by decision, so the layer
// is expected not to reach them. The English README is reached and carries no
// declaration.
const localizedHeadingUnreachable = "the count claim is a localized heading; the numeral layer's noun class is English-only by decision (D1)"

func readmeSite(id, path, countPattern, numeralUnreachable string) Site {
	return Site{
		ID:                 id,
		Path:               path,
		Axis:               AxisRetainedRoster,
		Claims:             ClaimMembership | ClaimCount,
		BlockStart:         "| manager-spec |",
		BlockEnd:           "| Explore |",
		CountPattern:       countPattern,
		NumeralUnreachable: numeralUnreachable,
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

// NumeralExemptions is the declared set of paths the numeral layer reaches and
// must NOT report (card t930).
//
// Three classes live here, and each row says which one it is:
//
//   - HISTORICAL CITATION — prose describing the roster AS IT WAS ("the
//     then-8-agent catalog", "17→8 agent catalog"). Not drift, not repaired.
//   - MEASURED FALSE POSITIVE — a numeral that is not a roster count at all: a
//     section marker (§4), a best-practice number (#7), a release version
//     (2.1.172), a date fragment (2026-05-25), a SPEC-ID tail (…-001).
//   - SELF-DESCRIPTION — this package's own files, which quote roster count
//     claims because quoting them is their subject matter.
//
// Every row carries a non-empty reason by construction (an empty one produces a
// finding reporting the declaration as incomplete, never a suppression), and
// the exemption is per PATH rather than per pattern — a reviewer who disagrees
// with one disagrees with a named row, not with an inference.
//
// The cost this list imposes is real and was accepted rather than mitigated: an
// exemption list large enough becomes what a reader reviews instead of the
// roster. What prices each entry is the mandatory reason.
func NumeralExemptions() []NumeralExempt {
	const historicalConsolidation = "HISTORICAL CITATION: describes the 17→8 catalog consolidation as it was at the time. " +
		"Not drift and not repaired (the card's scope excludes prose repair). " +
		"OBSERVED, not adjudicated: the same sentence's live tail (\"since grown to 11\") is itself stale, " +
		"and the adopted noun class does not reach it — no noun follows that numeral."

	const t1171Fixture = "HISTORICAL CITATION: a captured test fixture — a copy of an emitted codex agent " +
		"definition or a recorded codex session, quoting roster counts as they stood at capture (the 17→8 " +
		"consolidation sentence; \"13 retained agents\" in the recorded session). " +
		"Only machine-local paths were rewritten to neutral placeholders; the roster wording is kept as " +
		"captured, and editing it would invalidate the capture. The live sources are exempted above on their own rows."

	return []NumeralExempt{
		// ── Historical citations ───────────────────────────────────────────
		{ID: "manager-docs-then-8", Path: ".claude/agents/moai/manager-docs.md", Reason: historicalConsolidation},
		{ID: "manager-spec-then-8", Path: ".claude/agents/moai/manager-spec.md", Reason: historicalConsolidation},
		{ID: "manager-docs-then-8-mirror", Path: "internal/template/templates/.claude/agents/moai/manager-docs.md", Reason: historicalConsolidation + " Template mirror."},
		{ID: "manager-spec-then-8-mirror", Path: "internal/template/templates/.claude/agents/moai/manager-spec.md", Reason: historicalConsolidation + " Template mirror."},
		{ID: "manager-docs-then-8-codex", Path: "internal/template/templates/.codex/agents/moai/manager-docs.toml", Reason: historicalConsolidation + " Machine-emitted codex form; never hand-edited."},
		{ID: "manager-spec-then-8-codex", Path: "internal/template/templates/.codex/agents/moai/manager-spec.toml", Reason: historicalConsolidation + " Machine-emitted codex form; never hand-edited."},
		{
			ID:   "git-workflow-doctrine-retain-matrix",
			Path: ".moai/docs/git-workflow-doctrine.md",
			Reason: "HISTORICAL CITATION: cites the retain-vs-archive matrix as it stood at consolidation " +
				"(\"8 retained agents 중 하나로 유지된 이유\") to explain why manager-git was kept. A record of a past decision.",
		},
		{
			ID:   "catalog-loader-test-consolidation-comment",
			Path: "internal/template/catalog_loader_test.go",
			Reason: "HISTORICAL CITATION: a test comment recording the 17→8 consolidation the fixture was " +
				"written against. The assertion itself is on the catalog file, not on a count.",
		},
		{
			ID:   "catalog-tier-audit-consolidation-comment",
			Path: "internal/template/catalog_tier_audit_test.go",
			Reason: "HISTORICAL CITATION: two comments — the 17→8 consolidation, and \"all 7 retained agents live " +
				"directly in moai/\" describing the folder layout after a superseded split.",
		},
		// Captured codex role/rollout fixtures (card t1171): byte-frozen copies of
		// emitted agent definitions and a recorded session, kept verbatim so the
		// role-load predicate is tested against real input.
		{ID: "t1171-fixture-roles-manager-design", Path: "internal/cli/testdata/codex-rollouts-t1171/roles/manager-design.toml", Reason: t1171Fixture},
		{ID: "t1171-fixture-roles-manager-docs", Path: "internal/cli/testdata/codex-rollouts-t1171/roles/manager-docs.toml", Reason: t1171Fixture},
		{ID: "t1171-fixture-roles-manager-spec", Path: "internal/cli/testdata/codex-rollouts-t1171/roles/manager-spec.toml", Reason: t1171Fixture},
		{ID: "t1171-fixture-roles-other-manager-design", Path: "internal/cli/testdata/codex-rollouts-t1171/roles-other-version/manager-design.toml", Reason: t1171Fixture},
		{ID: "t1171-fixture-roles-other-manager-docs", Path: "internal/cli/testdata/codex-rollouts-t1171/roles-other-version/manager-docs.toml", Reason: t1171Fixture},
		{ID: "t1171-fixture-roles-other-manager-spec", Path: "internal/cli/testdata/codex-rollouts-t1171/roles-other-version/manager-spec.toml", Reason: t1171Fixture},
		{ID: "t1171-fixture-real-rollout-8d51", Path: "internal/cli/testdata/codex-rollouts-t1171/real/rollout-2026-09-24T18-40-19-01a0d2c9-8d51-7623-a979-b364671a0205.jsonl", Reason: t1171Fixture},
		{
			ID:     "embed-catalog-test-consolidation-comment",
			Path:   "internal/template/embed_catalog_test.go",
			Reason: "HISTORICAL CITATION: the same 17→8 consolidation comment above an embed assertion.",
		},
		{
			ID:   "embed-test-flat-subfolder-floor",
			Path: "internal/template/embed_test.go",
			Reason: "MEASURED FALSE POSITIVE: \"at least 7 retained agent .md files\" is a LOWER BOUND on the " +
				"embedded file count, deliberately not a roster size — it stays true as the roster grows, " +
				"which is the property the assertion wants.",
		},
		{
			ID:     "contract-schema-spec-id-tail",
			Path:   "internal/template/contract_schema_test.go",
			Reason: "MEASURED FALSE POSITIVE: the numeral is a SPEC-ID tail (\"…-001): agent catalog\"), not a count.",
		},

		// ── Measured false positives ───────────────────────────────────────
		{
			ID:   "spec-frontmatter-schema-best-practice-number",
			Path: ".claude/rules/moai/development/spec-frontmatter-schema.md",
			Reason: "Two hits in one sentence, neither a live roster count: \"Best Practice #7\" is a " +
				"numbered practice, and \"(8 retained agents)\" names the consolidation policy as it was.",
		},
		{
			ID:     "spec-frontmatter-schema-best-practice-number-mirror",
			Path:   "internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md",
			Reason: "Template mirror of the row above; same two hits, same reason.",
		},
		{
			ID:     "skill-routing-section-marker",
			Path:   ".claude/rules/moai/workflow/skill-routing.md",
			Reason: "MEASURED FALSE POSITIVE: the numeral is the section marker in \"`CLAUDE.md` §4 — the retained agent catalog\".",
		},
		{
			ID:     "skill-routing-section-marker-mirror",
			Path:   "internal/template/templates/.claude/rules/moai/workflow/skill-routing.md",
			Reason: "Template mirror of the row above; same section marker.",
		},
		{
			ID:     "foundation-quality-reference-section-marker",
			Path:   ".claude/skills/moai-foundation-quality/references/reference.md",
			Reason: "MEASURED FALSE POSITIVE: the numeral is the section marker in \"CLAUDE.md §4 retained-agent catalog\".",
		},
		{
			ID:     "foundation-quality-reference-section-marker-mirror",
			Path:   "internal/template/templates/.claude/skills/moai-foundation-quality/references/reference.md",
			Reason: "Template mirror of the row above; same section marker.",
		},
		{
			ID:   "template-isolation-doctrine-forbidden-example",
			Path: ".moai/docs/template-internal-isolation-doctrine.md",
			Reason: "MEASURED FALSE POSITIVE: the hit is inside a table cell that QUOTES a forbidden-content " +
				"example (\"Per SPEC-… (2026-05-25), the agent catalog …\"); the numeral is a date fragment.",
		},
		{
			ID:     "agentlint-section-marker",
			Path:   "internal/cli/agentlint/agent_lint.go",
			Reason: "MEASURED FALSE POSITIVE: the numeral is the section marker in a comment citing \"CLAUDE.md §4 retained-agent catalog\".",
		},
		{
			ID:   "web-agentfm-subset-count",
			Path: "internal/web/agentfm.go",
			Reason: "\"The 9 named retained agents\" counts the SUBSET agentGroupRank names, not the retained " +
				"roster — the subset-by-design boundary the web-agentfm-display-rank row already records. " +
				"Registering it as a roster count would report a correct file as broken; this is also the " +
				"one path the rejected any-row discharge rule would have freed (decision D2).",
		},

		// ── This package describing itself ─────────────────────────────────
		//
		// A guard whose subject matter is roster count claims necessarily
		// quotes them. Each file is named individually rather than excluding
		// the directory: an exclusion would also hide a real claim written
		// here later, and these four rows make the self-reference visible.
		{ID: "rosterguard-axis-self", Path: "internal/harness/rosterguard/axis.go", Reason: "SELF-DESCRIPTION: the CountPattern doc comment quotes the delegationmap citation it exists to explain."},
		{ID: "rosterguard-numeral-self", Path: "internal/harness/rosterguard/numeral.go", Reason: "SELF-DESCRIPTION: this layer's own doc comments quote the claims it reaches."},
		{ID: "rosterguard-numeral-test-self", Path: "internal/harness/rosterguard/numeral_test.go", Reason: "SELF-DESCRIPTION: the layer's fixtures ARE roster count claims, synthetic and live-quoted."},
		{ID: "rosterguard-registry-self", Path: "internal/harness/rosterguard/registry.go", Reason: "SELF-DESCRIPTION: the registry's own comments quote the claims its rows assert."},
		{ID: "rosterguard-test-self", Path: "internal/harness/rosterguard/rosterguard_test.go", Reason: "SELF-DESCRIPTION: the control probe's deliberately-wrong input includes roster count claims."},
	}
}
