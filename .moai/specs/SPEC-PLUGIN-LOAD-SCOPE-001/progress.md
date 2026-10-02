# Progress — SPEC-PLUGIN-LOAD-SCOPE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-02 (iteration 3)
tier: M
artifacts: spec.md, plan.md, acceptance.md (progress.md not counted)
budget: 16 requirements, 15 acceptance criteria (Tier M ceilings 16/16)
plan_audit_iteration: 3 (harness.yaml plan_audit_tier_ceilings.M is 2; Tier M plan-audit ceiling (2) exceeded with leader approval, 2026-10-02; one extra iteration and a scope trim, then one final plan-audit with no further iteration)
answers: .moai/reports/t1434/plan-audit-iter2.md (iteration 2, FAIL 0.73, threshold 0.80, findings F1-F17); iteration 1 was .moai/reports/t1434/plan-audit.md (FAIL 0.69, defects D1-D21). Both reports are local-only (.gitignore:235).
run_start_sha: 207ee936e
run_start_sha_note: set by the orchestrator to the commit that carries the final plan-phase revision of spec.md, plan.md and acceptance.md (207ee936e); the commit that records this value changes only this file. AC-001 reads its base from this line, because the third form of AC-001 lists plan.md, spec.md and acceptance.md for any older base.

Iteration 3 notes: leader scope trim applied (composite fixture, the old AC-006, the evidence-mode mutants and
the negative-controls self-mutant deleted; R05, R12, R13 become static rows; 10 runtime rows, run cap 60;
14 fixtures; 15 criteria, ids renumbered); the eight-edit minimum change set of the iteration-2 report is
applied (control session allowed, STATIC-LINE tied to a row token, mutants for every check-verdict.sh counter,
AC-001 base pinned, verb lists and live-name re-enumeration in the checker, manager-develop named as the writer
of the tracked block, blocker and LEAK contracts, R08 final arguments, RED-now ledger re-pinned to 6d0d75af3
and re-measured). The iteration-1 commit subject on 676293144 says "4 artifacts" while the Tier M count is 3
(progress.md is not counted) and cannot be amended here (no git write is run).

## §E.2 Run-phase Evidence

Run-phase evidence of card t1434 (spec REQ-001 and REQ-016): the verified 14-row verdict table and an evidence index. The evidence itself is local-only (.gitignore:235 ignores .moai/reports/*); this block is the durable carrier.

Deviation recorded: the delegation for this write named a general-purpose spawn (not manager-develop), because manager-develop auto-isolates into its own tree and cannot write the card tree; the orchestrator recorded the substitution. This block, the section E.3 block and the spec.md status line are the only tracked writes of the run phase.

stamp: claude=2.1.287 codex=0.160.0 fixture_sha256=55c65baabdd8289833863b5da15e20e9c5ef3cb9181264fde3bdd83adc9f11ab date=2026-10-02
versions: the start readings (evidence/version-start.txt, real home, CMD-IDs 3 and 4) and the end readings (evidence/version-end.txt, home=scratch, CMD-IDs 170 and 171) of claude --version and codex --version are equal.
verdict file: .moai/reports/t1434/verdict.md (local-only)
verdict_sha256: 5f287888a651c1c5566ca465d7e555c01748a7d9f8d6d33faa2d1e361b16a9e4

| ID | Component | Claude plugin | Codex plugin | Recommended home | Consequence |
|----|-----------|---------------|--------------|------------------|-------------|
| R01 | skills | PLUGIN-OK [cell:R01-claude] | PLUGIN-OK [cell:R01-codex] | plugin | marketplace: a plugin skills/ directory is listed by a live Claude session and counted by details, and Codex lists it too, in both cases under the namespaced name p-skill:probe-skill; init-shrink: skills can leave the init payload for plugin users of either tool, and their invocation names take the plugin namespace |
| R02 | agents | PLUGIN-OK [cell:R02-claude] | PARTIAL(installed-not-observed-active; cache copy only, absent from the prompt render in 3 of 3 renders and in the keyed form) [cell:R02-codex] | plugin | marketplace: a plugin agents/ file is listed namespaced (p-agent:probe-agent) by a live Claude session; Codex copies the file into its cache but no observation shows it active; init-shrink: .claude/agents can move to the plugin for Claude; the Codex agent surface stays undecided |
| R03 | commands | PLUGIN-OK [cell:R03-claude] | PARTIAL(command-migrated-to-skill; Codex generates a skill from commands/ and lists it, no command surface observed) [cell:R03-codex] | plugin | marketplace: a plugin command is listed as a namespaced slash command (p-command:probe-cmd) and details counts it under Skills; Codex turns it into a generated skill; init-shrink: .claude/commands can move to the plugin for Claude, and Codex receives a skill and not a command |
| R04 | mods | PLUGIN-OK [cell:R04-claude] | UNOBSERVED(no-equivalent-surface-found; raw=evidence/raw/codex-prompt-render-scan.txt; quote="R04 p-mod SENTINEL_R04_87279b70 count=0") [cell:R04-codex] | plugin | marketplace: a mod loaded from a plugin under claude -p wrote its marker through $.fs, but the claim is time-bound: other observers saw the hooks-modules rollout switch served off on this same profile (second-hand, evidence/raw/rollout-switch-lane12-record.txt) while this card saw it on; init-shrink: keep a project-side fallback until the switch state is confirmed for the target account |
| R05 | output styles | PARTIAL(static-only: validate passed with one author warning, install exit 0, details inventory has no output-styles category; runtime=UNOBSERVED(no-runtime-channel)) [cell:R05-claude] | UNOBSERVED(no-equivalent-surface-found; raw=evidence/raw/codex-prompt-render-scan.txt; quote="R05 p-outstyle SENTINEL_R05_6a7c6a6a count=0") [cell:R05-codex] | split(files install; activation unobserved) | marketplace: output-styles files install but no observation shows a session using them; init-shrink: keep output styles in the project scaffold until a runtime observation exists |
| R06 | MCP servers | PLUGIN-OK [cell:R06-claude] | PLUGIN-OK [cell:R06-codex] | plugin | marketplace: a plugin .mcp.json server is listed as plugin:p-mcp:<name> by Claude and by codex mcp list (the Claude status failed only because the fixture command exits at once); Claude drops the plugin server when a project server runs the same command (debug text duplicates manually-configured); init-shrink: the moai MCP entry can move to the plugin, and keeping it in both the project .mcp.json and the plugin leaves only the project copy active |
| R07 | plugin hooks | PLUGIN-OK [cell:R07-claude] | PARTIAL(installed-not-observed-active; cache copy only, no hook inventory or execution observed) [cell:R07-codex] | plugin | marketplace: a SessionStart hook in a plugin hooks/hooks.json ran in a live Claude session and wrote its marker, and validate warns when ${CLAUDE_PLUGIN_ROOT} is unquoted in a shell command; Codex installs the file but hook activity is unobserved; init-shrink: event hooks can move to the plugin for Claude, the Codex side is undecided |
| R08 | settings-hook equivalence | PLUGIN-OK [cell:R08-claude] | UNOBSERVED(needs-authenticated-session; raw=evidence/raw/codex-login-status-scratch.txt; quote="Not logged in") [cell:R08-codex] | plugin | marketplace: a byte-for-byte copy of the template SessionStart entry with only its script argument changed to ${CLAUDE_PLUGIN_ROOT}/hooks/marker.sh executed from a plugin with the argument expanded, and a 7 s command in it was cut off at the entry timeout of 5 s (plugin copy only); init-shrink: settings.json SessionStart entries can become plugin hooks when the script ships inside the plugin; ${CLAUDE_PROJECT_DIR} inside a plugin copy was not tested |
| R09 | always-loaded rules | PROJECT-ONLY [cell:R09-claude] | UNOBSERVED(no-equivalent-surface-found; raw=evidence/raw/codex-prompt-render-scan.txt; quote="R09 p-rules SENTINEL_R09_15285568 count=0") [cell:R09-codex] | project-scaffold | marketplace: a plugin rules/ directory did not reach the session context in 3 of 3 runs while the project rule control did; init-shrink: the always-loaded .claude/rules payload stays in the project scaffold, or must reach the session by another carrier |
| R10 | path-scoped rules | PROJECT-ONLY [cell:R10-claude] | UNOBSERVED(no-equivalent-surface-found; raw=evidence/raw/codex-prompt-render-scan.txt; quote="R10 p-pathrules SENTINEL_R10_1e0034c0 count=0") [cell:R10-codex] | project-scaffold | marketplace: a plugin rule with paths frontmatter stayed absent at session start and after a matching file was read, in 3 of 3 runs, while the project path-scoped control loaded after the read; init-shrink: path-scoped rules stay in the project scaffold |
| R11 | instructions file | PROJECT-ONLY [cell:R11-claude] | PROJECT-ONLY [cell:R11-codex] | project-scaffold | marketplace: a plugin-root CLAUDE.md or AGENTS.md did not reach the context in 3 of 3 Claude runs (under the profile value instructionFiles=claude-md-or-agents-md and under --setting-sources project,local) and in 3 of 3 Codex renders; validate says the plugin-root CLAUDE.md is not loaded as project context; init-shrink: CLAUDE.md and AGENTS.md stay in the project scaffold |
| R12 | plugin settings keys | PARTIAL(static-only: validate passed with one author warning, install exit 0, details inventory has no settings category; runtime=UNOBSERVED(no-runtime-channel)) [cell:R12-claude] | UNOBSERVED(no-equivalent-surface-found; raw=evidence/raw/codex-prompt-render-scan.txt; quote="R12 p-pluginsettings SENTINEL_R12_3954caff count=0") [cell:R12-codex] | split(settings.json installs; effect unobserved) | marketplace: a plugin settings.json installs without complaint but no observation shows it applied; init-shrink: keep repository settings in the project scaffold until a runtime observation exists |
| R13 | userConfig | PARTIAL(static-only: install recognises the manifest userConfig option and says so, validate passed with one author warning, details inventory has no userConfig category; runtime=UNOBSERVED(no-runtime-channel)) [cell:R13-claude] | UNOBSERVED(no-equivalent-surface-found; raw=evidence/raw/codex-prompt-render-scan.txt; quote="R13 p-userconfig SENTINEL_R13_b0285987 count=0") [cell:R13-codex] | split(option recognised at install; configure and use unobserved) | marketplace: install reports an unset userConfig option and tells the user to run configure, which this card did not run; init-shrink: a user-supplied option can ride in a plugin manifest only once configure is observed |
| R14 | manifest fields | PARTIAL(static-only: strict validate succeeded with no warning and no error for a manifest carrying author, homepage, repository, license and keywords, install exit 0, details prints name, version and description; runtime=UNOBSERVED(no-runtime-channel)) [cell:R14-claude] | PARTIAL(static-only: codex plugin add exit 0 and codex plugin list names p-manifest 0.0.1; every probe manifest was accepted, including unknown keys and a wrongly typed skills value, so acceptance does not discriminate fields; runtime=UNOBSERVED(no-runtime-channel)) [cell:R14-codex] | split(manifest accepted by both tools; per-field effect unobserved) | marketplace: Claude validate reads the author field and accepts the dual manifest, while Codex accepts any manifest, so Codex acceptance shows nothing about the fields it reads; init-shrink: a dual-manifest plugin installs under both tools |

evidence index: (sha256  path, relative to .moai/reports/t1434/; every file is local-only; recompute with the same hash tool over the path)
RAW-FILES=145
2e32bce1c390b17a4c340f3a2766a13ef378631ed8dca335cfeb1c5000774b2e  build-carrier.sh
38655b220e17578f3f0d9c155bc931a604fd93f80a318211d33a61d96a01a2df  build-mutants.sh
4b4507f5f31804eaac94252c3cafa88a710629545ae81ef5f3daa3508a1000fc  check-evidence.sh
31cda766cba00ac2ec7bf7d8527c89e53c3fdadf410459a5b2da8425190b802b  check-verdict.sh
e572204eaed1a9d9f021e2796f0b9168292d73db9ffce0cbd3ed1fd92c2235f4  evidence/00-isolation.txt
2a8b592f80b2b11f6b845010e87fa2d7ad88f9a0661b94a6c1018c42dd5f7290  evidence/cells/R01-claude.md
bf63650be1394dbc4a6ce3ce8bdbd313e05cc1389a882d7d98857e4c7d1ec426  evidence/cells/R01-codex.md
fe86698228b9a91f653b250914d586840291d278d287f835267af1abd5245bc1  evidence/cells/R02-claude.md
43c1e7c775f6db1603dc450220b23962dfa90c847e04f87544fd10f9fc36ac43  evidence/cells/R02-codex.md
a262ee0169c22c8fb017bd2dd08c30ce8dee3d4bc873c7b4450c784299cd1fbc  evidence/cells/R03-claude.md
417707bb7900de9db1a521f0afcb0fa8ee91313ebab259472d41fb1ee0d01f87  evidence/cells/R03-codex.md
0e6fc0eea5090697f208656f886e70e5db3663d4bfef300f6f8e15bbd3f631fd  evidence/cells/R04-claude.md
1fab32e817e3bc6b4d077f9070bd7e3f766154e4feee6a605f00016e3df29f7a  evidence/cells/R04-codex.md
8e2a67fd10326e5d170ccc8610bf9bd84be84488e454352465a4591069a1cb80  evidence/cells/R05-claude.md
63b96eed310794353ca1cc7607118a666b127079185f75a8c90db20de1c1bf03  evidence/cells/R05-codex.md
7f0d8ae29462e554217240226e5240bf25b0b7e1f21b223d90ed9e284f9de202  evidence/cells/R06-claude.md
8924f85243e447b873890defd1055113c33a0e3d4882aac7fcf8ed4564f8c07a  evidence/cells/R06-codex.md
e64ca5fa296b2cfad7c43e8436a2ca73a8e53b7fe10b9deacdbdec9f1462adb3  evidence/cells/R07-claude.md
fd0f3cc8d136776ecf21ab24603feaa336ca3cf29861bab771f9475c73a94c33  evidence/cells/R07-codex.md
5193376b7f9cf17ffbf3a4b3945773ed7eb15103548a5a2703db50eba402e0e8  evidence/cells/R08-claude.md
0e5c3cf303e8e0dce624562bc07580f4aa99625819d06e56c827383dab5bffbe  evidence/cells/R08-codex.md
b35a8ec716f6e6e0ca7670b3ae585c7c6144e9b2d91c4c10f86e0cafb592312c  evidence/cells/R09-claude.md
10f42ad0dbf65e116aec00f822d161ce15b10d060169fa2301d23728cc491dcb  evidence/cells/R09-codex.md
019a567d46bbb8fe82c339c63f0c3ca3880293a0f1718f47f5eec66a8a153bfc  evidence/cells/R10-claude.md
7553f5d148520d81f97a04d271b126386f21495514e4d9044255ac407605903e  evidence/cells/R10-codex.md
60a0a7cb72b0b2f052526f76bcd508ed15e0b526de04b2719488c8479b93625a  evidence/cells/R11-claude.md
e98a19b79aa0faf8d334f06e8a84861327c21ee152aa562f3b10672cec9cbb9d  evidence/cells/R11-codex.md
6877d930578d686c99857d7b529fc084ed6e72127be874aa5fef058a7ccff4a9  evidence/cells/R12-claude.md
539f518e4b3d3b8d1936aab416ee80782ae50752b3a447aea4714dd1f66e0985  evidence/cells/R12-codex.md
b74c33e321c55d780447da53bb6a5e57c7def4e9dbb3a429af6f5b9e1db26c26  evidence/cells/R13-claude.md
8cb6bc24e35e8f054e25e14137b9e5f495ba20a345a898000ad5e0caf004f7e5  evidence/cells/R13-codex.md
2bb67ebc79090713349e4649ac7e93d009be6412e8f125a08f6e0724883da4a0  evidence/cells/R14-claude.md
54f261efe8828abb009192bc1404298b1bd3f68aeeaf6b970682f6ab77946795  evidence/cells/R14-codex.md
c22131c64273265927da9ac0d9c1c35187c0c9ef83f5a7a25ce3756bb9628cae  evidence/commands.log
b8417d5d447315cc04a056795cec3dc16177719ab8a8b90577d4e17c3c0afc45  evidence/env-live-names-check.txt
d103ba82b025896e40ff86f98b01ec20c1c327a1c5fd5f86bc8ed684a943a92c  evidence/env-live-names-m3.txt
9dcf22071992cf510611e72185a1629cd751eb303c495361294e1d615f87f5c7  evidence/env-live-names-m4.txt
7564fc24d22203e8dedc4f4696e61156434411dfd59d66de0b21aeac7ff8e0ab  evidence/env-live-names.txt
9cb5a38fdb28ee1b7688ac4283f9b77d9bca44bf636893d96239c5fe39663c1a  evidence/env-scrub.txt
5418cf83f772980be17c43caa58d4839a5922340b40633c62a314de9fb0b94ca  evidence/env.txt
55c65baabdd8289833863b5da15e20e9c5ef3cb9181264fde3bdd83adc9f11ab  evidence/fixture-manifest.txt
d0b7c3b46a318b7438a52c33c2b1610ae72f2a5b79a24ae5058ff078f2e53dbb  evidence/fixtures-table.txt
5f9fad2e0426dd60887fe2f4608af21e5f64fc40c862e8a3bed39d21ac6f75fc  evidence/m3-batch-extra.log
4ea48a8b775a8a8296432ab2b3ebe50a5d9dc8f16c8465234b054c07dc98454d  evidence/m3-batch-main.log
16cd8acd3f80f736383f44dfb8eb52e560ad27146d3931182ddd1a212fab2585  evidence/m3-batch-obs.log
0d3fac9168137ec0ca1c0f1951f1f7b9ba1c2d97221f90f855a483f7cac50dcc  evidence/m3-manifest-summary.txt
e7c22eefde8e50a362cfd2d24ad1f21a8d3f0a9b2d6aadf827693d2ea2150be8  evidence/m3-observations.txt
d0a31ab389d1693ee0b854c538a468b025fb28d04a7dfc390e56fe2fa51c9861  evidence/m3-runs.log
529feb9973d3d60dd8f19f65a6ce0379cc3bbb65d7b96babe4e0e3f62418bb92  evidence/m3-transient-files.txt
13e8bbb6c1704398dac3535ebbfc4a200791f07a0d2b98dadf2032f0b02ec0cc  evidence/manifest-diff-1.txt
a9e915c9f58ff6847e72eecd5753a30d060a82f16260f653b58e6d1d138377a4  evidence/manifest-diff-100.txt
9ef6a20a0e70d91f4a665705c1c2917196276483def12847e9f334476054058f  evidence/manifest-diff-101.txt
efc591555b28edbf2c968a4a059389cfc1c502f7d10ab3420778c8e7f725e1ce  evidence/manifest-diff-102.txt
8b47b7acba34e0e00dfded660173984ed4185a828c93334e6658ae0eefb4a3aa  evidence/manifest-diff-103.txt
740267a290fcbad008939a091739a1867ca64ac7188ffab677ef30bdd9bfca0a  evidence/manifest-diff-104.txt
5e56e6cc29d07892078b6da6d14e61a88dbb921bfbcde6e6601ee3cb5689d40f  evidence/manifest-diff-105.txt
dfcb5a599bb6f7ab73ecdd5ae44629eb765b19d4c926a8df08c82984bd508709  evidence/manifest-diff-106.txt
b488ae26e06fc707460b4a4a769de38f8a16d043e11523b611e99be129167cd4  evidence/manifest-diff-2.txt
0ec6121eb0b948ac3cd439079a4fb81f327921f0b5892d85bd8275265dc67146  evidence/manifest-diff-3.txt
08647ce2a7609360baaecb3e7e09c0d8427255ad46a82a984d3b48ebe648ff98  evidence/manifest-diff-4.txt
ff844c3b773021c9bad75a3bf4015fc99c1e99a89ac2492d86dfc499cad8a891  evidence/manifest-diff-70.txt
881d905b42a945b63fde73765e7f0d5b9ab2bf474c085f5c1ff0c298988300da  evidence/manifest-diff-71.txt
932f68a90ace3b9ae38defb77e2912f26092009e6999dcfac8cb72e946a03cde  evidence/manifest-diff-72.txt
46091b674e2966b54f4c59d0ea12a52ac52034d6bd52276ab7cc9ab6fee498b1  evidence/manifest-diff-73.txt
653402bd8119d5c0df96806165021a1ade2e318cffc8f2f98521aaf4de2c1ff6  evidence/manifest-diff-74.txt
0dcc3e9205767ae815b899f2f468f2c9f6fec8f151ad0a4e85346088355a8215  evidence/manifest-diff-75.txt
d28188ab788faf8fd7d24cd4d7dcb2d938d72f8c26dedb3cf552dc3b5dfa6238  evidence/manifest-diff-76.txt
004ce09b912a8c71a0cc31f9dd5849eb32853a825d49aa02794c60e2cecfcbb9  evidence/manifest-diff-77.txt
90b3f0eefa66410af669aade36a40675d6255c6d6865c4b2877c60058fdc57a8  evidence/manifest-diff-78.txt
9fbf80fa241062516857b17b1f4c87a6409ba3dffb732ffe6651152b9f06f470  evidence/manifest-diff-79.txt
9ebf6703b1f95c42d1506b5d163974f57ae988458bc6d4aca17df9d416cbda9e  evidence/manifest-diff-80.txt
7348630bcc97f3307d17d745da4cf858496258db0d3c989ab8a29f3825544cfc  evidence/manifest-diff-81.txt
841b74bc9275e08cb0ab522773c76724f0ae30befbdadc54950ffbd0614d7a88  evidence/manifest-diff-82.txt
1527e52231d7a59a5122bdbe6ce703821990250127f58c4961036c6e7e7227e1  evidence/manifest-diff-83.txt
85a8709577a4e26b0cd26f7e67fdd74cafed9660a9d12e3911970508bc740c22  evidence/manifest-diff-84.txt
30ee60c9c58382ecf0a687a2fa0014031fcf2fe46773e9d4870d92e20f12bc4e  evidence/manifest-diff-85.txt
069cf21ec8158819f4367c7534ae155e48b19cd1403268dd6a2393315c4528eb  evidence/manifest-diff-86.txt
1f173735bddb8f0f53ec194072c9f50aab51e671023f415bca8c70e85eaa2864  evidence/manifest-diff-87.txt
7ce843f14cbf57c5f398a78c9293fe74bbaba3e173b584c366e80eed8f389365  evidence/manifest-diff-88.txt
83ce2a31f7d62e4abca18fad2bf528e3e9255b8ad258d9e18b5a09179dc3625a  evidence/manifest-diff-89.txt
55da7328bfa1fe349905f55bb164ee5380442c6d32a5ba17a1ef765c929452d6  evidence/manifest-diff-90.txt
75e7ce686b4341b3afa4bbb68258c08a3934f2f6cca1b0e716b3095c5cf0834c  evidence/manifest-diff-91.txt
18f185b1562c57c6713e832015f2d97b6612b6f841d8f9866d33c7977983b159  evidence/manifest-diff-92.txt
08e564ba4dc5f2f8d72583c352acf175d2dea3e77b2b98749d9dbe1dea81dd6b  evidence/manifest-diff-93.txt
209083db93ce807cfcbe4ebe6d7b5e995a49488ee553ed9ab74c03b8bb7a3e9e  evidence/manifest-diff-94.txt
b1ad93f8e1114dc5774e986a2e811fdc86498f127f9d36659c4dbe4b047f23c4  evidence/manifest-diff-95.txt
526b908de56d63fcc6689e60fcd422ebf82bddf16f7a32139f09bb97b3b27edd  evidence/manifest-diff-96.txt
af07078746c8316b0340d2967f590e0ffbc30233ccdb1447b944c1bec0e9d128  evidence/manifest-diff-97.txt
cde4cbd07e8056a06a21cef2136ef46d123651b7d86c144cc52b0afe45408f89  evidence/manifest-diff-98.txt
46a07f7b42670a95d5512cf0fb00223f99bd25c2b3eaf91623768a8d0d9da5cd  evidence/manifest-diff-99.txt
179e39db793258c1291b6046888d1d09ffe0aa30275c15a34b6ad35a9e6e0f41  evidence/manifest-diff-m4-window.txt
ff92753ac9efe74b4030829acf125371879c09319aeaa7668500982c58d6f6b3  evidence/manifest-diff-static-window.txt
8642dcddfb0d65fe6657295cc92ff26329cfd6b312e6e36acb63c5d9d82a93e5  evidence/negative-control-checker.txt
17760fbe3d3a00a504afda20b6f5092c54d0e1e2fc7e58286e5612e86e5ad671  evidence/negative-control-manifest.txt
a459840d2151a1d4a641494d0e37aba9c2a82a0341a250989af379a73bbdcbe6  evidence/probe-counts.txt
f7f3a80914c208523d0fc5a98626caa213bbc928a61020969ccb590e0a58805e  evidence/probe-m1-run.log
ba423b723eb0b7ae0f7e805ca94aea3e97c981c7e32073389fe6192f132be8ed  evidence/probe-m2-run.log
838f101480ca82d884d31ddae1dfa949bd06f4328ab057ebe5f1ba5a3ffbba9c  evidence/protected-excluded.txt
0e005b0616408d2ba02b84373416054270056e48ab4f4ef9cc842c214ba4b7eb  evidence/protected-final-diff.txt
fdb6e90156bacdb3405bc5013d1aa76edfd889304b7323eb1bfa22e6ed014062  evidence/protected-final-summary.txt
307be87d731c718e3795636e855627035817c69f066584a944b81152edb15d25  evidence/protected-final.sha256
9a5d2d1ea6ab4b5470baacb16221b1bf7aa45007deb250f5e505fe45f15d6f96  evidence/protected-m1-diff.txt
307be87d731c718e3795636e855627035817c69f066584a944b81152edb15d25  evidence/protected-m1-end.sha256
4b8e70510b16fec13987aeb9d447a04940979417647160f0d38a47e80b8058ee  evidence/protected-m3-final-diff.txt
26f1bc51f9c9a403f2b08a96445b1ae2b7e3b69125620b490711c142e8313349  evidence/protected-m3-final-summary.txt
b2fa5dd15afc17fc5f90713bac954e5d0ace21f16fed43fd3b6c758544c86f92  evidence/protected-m3-final-vs-t0-diff.txt
c1bfc0baa23a0ce16175d58ef94ce4452965cbc575135d2696a4ced297c3c96d  evidence/protected-m3-final.sha256
307be87d731c718e3795636e855627035817c69f066584a944b81152edb15d25  evidence/protected-m3-t0.sha256
307be87d731c718e3795636e855627035817c69f066584a944b81152edb15d25  evidence/protected-t0.sha256
b8e652e7e101cbc04506343ef68f51da661d53cc27f78363b1931b40520d21a6  evidence/raw/claude-details-R01.txt
284bc9517041bf569eeb3e37750916602f1e626998cb8d9c53081bb2bae3b011  evidence/raw/claude-details-R02.txt
66ccdf0c33552047ad6c39b52bafc551180091ad2e408490be057040364da961  evidence/raw/claude-details-R03.txt
ead2df89a9821dd4c9bce2a4f4d485e5c091c83803e7c35a473153e44c622379  evidence/raw/claude-details-R04.txt
218e1acbf0e5119fd52f7424410efb70dbf22356910a6249401b44076326a173  evidence/raw/claude-details-R05.txt
1805ddd8b80f531edc431a8d4c322b13b842a0334771ddf5ff029a5b7ad1a049  evidence/raw/claude-details-R06.txt
fd2c0c4d89578375d556b128dad7759a0a6696ba9c2ce9fac0ed6d872a08d934  evidence/raw/claude-details-R07.txt
e669ec3e418133e480c8a101b269d4535afa2aca5c1c71df21956f615c66e366  evidence/raw/claude-details-R08.txt
d2f92665ba125b8cff821fee5f80ee63d67862142e73f9e14409c882e0d490ec  evidence/raw/claude-details-R09.txt
654d4369bbe5cc3fdde07e9f62ac3f7af7b8f205e2df021e9a66efc65f469e9f  evidence/raw/claude-details-R10.txt
6a0ccfcc7d6c2c4a1f1d3483a1575e80d63d505b101681d7dfaa4ebdbcb6c853  evidence/raw/claude-details-R11.txt
cd473848170603695e1c66d60285e7338c8b745b5aca36fbb2133da7acbbdafd  evidence/raw/claude-details-R12.txt
4840d4e561dc097ad25af152efeeacb0993a8a9cd054f0165fd2a545777966d0  evidence/raw/claude-details-R13.txt
2f8d63cedbc8d37120a73b532a61e05a8fefc693a88b29e25582fe0ba4a28054  evidence/raw/claude-details-R14.txt
4cbd0dd3f1bc7d9f9c37847abd260a7e5983fba5038f29058c10c8ce2edb850f  evidence/raw/claude-details-bare-R01.txt
ee7b0fe878288e3c999a3b9d9b455713d21dcac5ab11ea81a2aa7410782b7a35  evidence/raw/claude-install-R01.txt
e02b8d5da85188d769a2b7d6c4e56a397d16f14806c07c495a790aab762ce0f3  evidence/raw/claude-install-R02.txt
31bb549f65b5004f222f14b26a7d98e967847b8ee08ac381c67b4e46ecfc9ab7  evidence/raw/claude-install-R03.txt
b4006a9f68f536b19b0bb5d395ca85b72c8bc33d48e467a56b96e83e04f15014  evidence/raw/claude-install-R04.txt
ca780f0c1c2b6b3cf518ecac46cd9b773d0b80b5be11a198f93fb321102cc0bc  evidence/raw/claude-install-R05.txt
5de1d2539c3741d15b64a9d91024de6721679c99ce4aa8c60628b874ae7f4d46  evidence/raw/claude-install-R06.txt
24a3a3ac5f60d0f0b23eeb65870c52b28a118bc1851fdbc0dfef509f40253c5e  evidence/raw/claude-install-R07.txt
fa0ab850715b1fc7a1161fc09807b19537cf0f8b1e4491c97bedbbc70f055c79  evidence/raw/claude-install-R08.txt
34ee994172b1d0d6060e2d520abc6cacfe81482bb429c79e70c7ff125775dbc9  evidence/raw/claude-install-R09.txt
f602a25b4265a5b728329d13f8690d56805d470b7ef21fe6038861f85522e45d  evidence/raw/claude-install-R10.txt
67e53da9d4590c4239378877c8213ff1fcfbbbb29af03694148437b7739cd691  evidence/raw/claude-install-R11.txt
6aeaffae6cb2fad6c5098d8d5e318dc2f26234704a128e6e291180300a1acae4  evidence/raw/claude-install-R12.txt
b4bb54ba683a37a34aef8719b6e0b16c9ce1244e6f1e71c768db420163aa1bf5  evidence/raw/claude-install-R13.txt
8ed123ac9202224986f9009f43b817532776ffff473039a932def5da76da4bd8  evidence/raw/claude-install-R14.txt
88e2e4878c6a66fce38a9c42072fbef6e03b05752c42c17f5ce3ddc9e496f645  evidence/raw/claude-isolation-probe.txt
4df78b8573cfb23f4cd4f9feec70b0d9653b0866a14b0ef0112f89c32b2a8496  evidence/raw/claude-marketplace-add.txt
f94d360499faf6bf436e1717ec7dc7183827e86aa859982dafa2c8cd6af0dda6  evidence/raw/claude-plugin-list-scratch.txt
bf1bb87e7a452a9e816e2f1565cd7ac2fc8293277c6d919ba6867ea5a1fd80f1  evidence/raw/claude-runtime-R01.txt
e72e3b3266f0bb5f7ffffc1cf56e82fb9da6fa5900f4cb36643cb67a98ad0169  evidence/raw/claude-runtime-R02.txt
e22881424540715accc0308ac37cf4f7e3985cab1a485597d6a9f58f588b6a53  evidence/raw/claude-runtime-R03.txt
70a4d496f4192de7af3adac40373f84132c151520182935a8bce43d153ad5dba  evidence/raw/claude-runtime-R04.txt
8bc3d5336f2f7337f1a5da7c67f2f366ce2a0ae82c23d8135bb858ce08dd624f  evidence/raw/claude-runtime-R06.txt
5e252db2a447bf8b31697e5c47eb5c784c2f5f45494447b59687f7bef726b6ef  evidence/raw/claude-runtime-R07.txt
e4b7dfdf25af703db9084c525696be47f5352dab1def1ec036c2ef8795b46252  evidence/raw/claude-runtime-R08.txt
408ddbb5bccac7731d7d2f7c0fd12f213c15e4105883be577f894a3540dc43fd  evidence/raw/claude-runtime-R09.txt
a2a437c26b172aa3b8ff3b842aa3697d5c312e2a822ef5ce34b0f29c57426d5a  evidence/raw/claude-runtime-R10.txt
fd1a7522e93a2213770e215bdd9c287e756e84d4186217265b533a29d63d8881  evidence/raw/claude-runtime-R11.txt
e9aef27b160b348e052187cdd1d00f9ab85972b49602a8f3fc4d0cc98498ea8b  evidence/raw/claude-test-R04.txt
e71274727ba11efcf0b1b5da847d9721c0d58a1b032f5cd9716811249b079e1f  evidence/raw/claude-test-empty-mod.txt
28c5cff7f685684b5df4c07ef2de131dd155015e978ca2e9995334b9abc4d2af  evidence/raw/claude-test-help-real.txt
4ed67f70ba51f83a7dd08dc29592c07605cbe401d9d750a029fc4528a6cb79ad  evidence/raw/claude-test-help-scratch-m3.txt
1eaf92571bdac6545500bd290be835153709456d76fc35b26f97158bd8e20a51  evidence/raw/claude-test-help-scratch.txt
01e2200b2241209fbe9dd73432640dce5b61c3aa285fc594625e0b70107ef5f3  evidence/raw/claude-validate-R01.txt
5cd5b17865eab24ca934f59d7837b9d5b0b23a33eb39e6255c58afb57ced14e1  evidence/raw/claude-validate-R02.txt
423c267987bda14b9c69d4cb0535ed03c247b875c8080507f56278a9ebd28798  evidence/raw/claude-validate-R03.txt
a16dd1bbc0379e9321cf38815871062e8f57a1d2d4885166adf47a7e03786036  evidence/raw/claude-validate-R04.txt
af380e8464c4b106f7cd190faea87223d2b8955264fc43b18d43cc189fa9cd1c  evidence/raw/claude-validate-R05.txt
331bb069beaaf272ae3a7708b749c8613acd993b8dff9e972294d7295f45439d  evidence/raw/claude-validate-R06.txt
82595e7921fb172c3a0bc2bc4594c0300bd385b1fccb321cde942163b4dc6a84  evidence/raw/claude-validate-R07.txt
7ae670b499666c6805330d65517f2c3a2ce5d75847e72b4ccc6ffbe4c5aab871  evidence/raw/claude-validate-R08.txt
f6b945c44f358480a55f31c804ba641b4d4466cdb31cae84f1a526523fe6c4b7  evidence/raw/claude-validate-R09.txt
79adc4f3d39e73a35bc988540899d64a261499822afc907bbaa845731dbd4e2a  evidence/raw/claude-validate-R10.txt
91249d89c69c0227531afacdec1ec89a92b4ef9323994ef38870e0dd1fa2230d  evidence/raw/claude-validate-R11.txt
d62abebe6d067fb72da7c4e38a92e2a21d4994af5291373cc6a52896fe768716  evidence/raw/claude-validate-R12.txt
405b78b67e061490c91ec9a02d5bd8adb98abf5f13b43d6e2b7a95ff5b461d51  evidence/raw/claude-validate-R13.txt
759c4b480b39b3bf12bafb7b903e4c35424cce41a844e7272dec8dd7faba3459  evidence/raw/claude-validate-R14.txt
faae3bc8f393a8ae5a45d08dc51b3c63015d57324ff9147922c7a014279b95ac  evidence/raw/claude-validate-strict-R01.txt
4b51be4ddf7f33ae45321d93cb0bf5bbcdce65e6f152fa50d87e8495335808bf  evidence/raw/claude-validate-strict-R02.txt
8f7350713e3ffba77efc3505123990404f8e816711387bb4c6e8e58e69d9ae9a  evidence/raw/claude-validate-strict-R03.txt
70e4d46756704902b936e9eaf8a88e21d444ea312fea376b0a69160020bea9b1  evidence/raw/claude-validate-strict-R04.txt
e8e1ca52513ae19e7fee39a78c1dde9dc95736fa684f766e665b12027df69240  evidence/raw/claude-validate-strict-R05.txt
59653eb23f1d81e0ee374a4bf961af8fd7582e3177e35e4b2c3137d9a4b50f5d  evidence/raw/claude-validate-strict-R06.txt
ee8fe411160abe5df4642b41a5f3cdbd257229f73b364c05a038031d2c3a7350  evidence/raw/claude-validate-strict-R07.txt
acf0fb2312d1bf7d2efe147cadbaac28aa27179abe57ab177efb4f80e1b8081d  evidence/raw/claude-validate-strict-R08.txt
5fcdd2883a666fe7bb0f4243c24745f30a56fa3eb8780baae6c5348ea3c920fb  evidence/raw/claude-validate-strict-R09.txt
3a5aec299a91af128522c4ff4304a0ab8ed2cf43802539c808209dada0e1bb9d  evidence/raw/claude-validate-strict-R10.txt
93d6188e6578439f58ab26876b8adc67ec0f6268af2605d3b1c1e16dc9442d13  evidence/raw/claude-validate-strict-R11.txt
7052f96851494dcb40d71c2e3d63b398f5242ef8a4613fd91841690f5682e8e8  evidence/raw/claude-validate-strict-R12.txt
9c2a7a94c9a8f825cc728c6d7526c20e13a56665e004c4f96fa5c9d6298dac59  evidence/raw/claude-validate-strict-R13.txt
c3df4cae00f93442117eb704bb1fd4b167c7764e24dbc48212c968db92fe3a1a  evidence/raw/claude-validate-strict-R14.txt
c57d86512d197f637f1cc8ebb94470cd04555dce50cf74322b4d5e1e78270c4b  evidence/raw/codex-debug-app-server-help.txt
56905ed0755f22092e574dffc9139703807e1db8796652fb80735d5c0de11c10  evidence/raw/codex-debug-help.txt
a4425f8c211772eff54e3b5fee8ae0754cff48f3d9c1b788a3afe963bd17ce76  evidence/raw/codex-debug-prompt-input-2.txt
71b68975d388361e9f5830f1998c233654b20e884e16807433420c9102ff3516  evidence/raw/codex-debug-prompt-input-help.txt
e5d2a6f3ed08a317a05ebb1f613f0135b08b74b7743c9afd331c3d3f4a65a56d  evidence/raw/codex-debug-prompt-input-project.txt
39e9a1c23901cea2dc7faabe0124c5bb4783d09d32c18b6237efd4b1ae14601f  evidence/raw/codex-debug-prompt-input.txt
22e3c60bfe2cc4899662cbed3d1a372942895ab9bfb0a8f721aa6657b0622325  evidence/raw/codex-doctor-help.txt
799f745b267669f45811a34c3b78aeda987d402f4708d9f02c254e86641010c1  evidence/raw/codex-doctor-json.txt
51c442a39ab2337b2929cef89cb8b77763d810fa1719a02561e86641fb1b720e  evidence/raw/codex-exec-help.txt
39aacd3283ce0a60edcd057b1d967af5d3765f8c700ce8c1dd920ca250f7b85f  evidence/raw/codex-features-list.txt
4ac5b58e1c150da32c123cb3b0c1748ea2f88a29981718e7db03d88674d8a07a  evidence/raw/codex-help-top.txt
262db57267cc30f6a65f0dc0e284179e0416ea2fe691ab1f9b9672ac7225cd09  evidence/raw/codex-isolation-probe.txt
233db6b354cff42547a7af66fc613ce5f185622a4f44a314b2a7ff7a270c048a  evidence/raw/codex-keyprobe-add-agents.txt
8cb4f7cfce622f36b7dace3dbe26c144d32661491331b0e8bdf8ac69eec8d410  evidence/raw/codex-keyprobe-add-apps.txt
8e771fa3813823efc1b6568ae1053915ea25bb46ad3c902cd9de49ced1846291  evidence/raw/codex-keyprobe-add-badtype.txt
e88f2218435c9c8c10355241290b2f52f87af14caca3a45054603d1c1198a230  evidence/raw/codex-keyprobe-add-claudeonly.txt
47c672a6c7f557270f41ff45aed906066c699c5df9bc5e41245ea917955d77e8  evidence/raw/codex-keyprobe-add-commands.txt
d3506dbf945af7834a36c32e51ab446762aaac0588b2cd5d03c5a951991a0767  evidence/raw/codex-keyprobe-add-hooks.txt
7e5baa0b2517a1a3b1aab5f63675204f0b5cb4aec9f55c00e4d0e28c86f49cc6  evidence/raw/codex-keyprobe-add-instructions.txt
51f1bb4a535ae976df3e94aa2bbd412f3de04141b44bb38a3bce21fe5962fbd5  evidence/raw/codex-keyprobe-add-mcpServers.txt
bde66909cf5e2a74447ef2e28172ee30d4ee10a6448c4254b898ff00a19a568b  evidence/raw/codex-keyprobe-add-modules.txt
445ac93a70cb032d9bfde02f99caaacf1fd623c4ffad737da2ce25dc0a52b5e2  evidence/raw/codex-keyprobe-add-outputStyles.txt
84fae0a101c200062fa300c26415770e65e21c82dac29459df3b5085f06c01ee  evidence/raw/codex-keyprobe-add-rules.txt
944c16567476daed96cc9487e275afc6a464ac2328e1d0d1a91d682307d4d7d1  evidence/raw/codex-keyprobe-add-settings.txt
8137df984232d17354815c65ffd6e5fbf1482d7e5a684a2aa3c9bd43aaedc275  evidence/raw/codex-keyprobe-add-skills.txt
595a370906d912c72095a769cd6bb4ab17be260b806fc22986ad6e87a583135b  evidence/raw/codex-keyprobe-add-userConfig.txt
5e17da7cf88c3b53f76a03cdb3abc21905c9bb22d6d1d8c44793439ab5b13e2f  evidence/raw/codex-keyprobe-add-zzunknown.txt
4c8e2542a90ba09284df93e5ff7ebf7600805d30a8f48b40ce5695e407f2248b  evidence/raw/codex-keyprobe-marketplace-add.txt
a62505aa88f8843361d098ee25d5aac620855679f9e6b608c21e431b4a2b5505  evidence/raw/codex-keyprobe-plugin-list.txt
aa39996915a58f8255b56bcdff6f729a3da005bf19105844cc517e494ddd0f9b  evidence/raw/codex-keyprobe2-add-agents.txt
953c1edfc2b82bfd45c4100973ac3a1a4c9152f794ccc06b0621d8b2554b0bd3  evidence/raw/codex-keyprobe2-add-commands.txt
c0eb71a58fe89cf8ade15eb6a9384746bcd3552c6ce4f5a5f5aeb351af978998  evidence/raw/codex-keyprobe2-add-hooks.txt
f72b2136818c6cb21b1151e45148c479ae790bcc7a3f7ffa441160cd013de5d8  evidence/raw/codex-keyprobe2-add-instructions.txt
5469246f7256ce50998d7415a3cfec4105de1cd7e6d8f6adddb50926d5cc97b1  evidence/raw/codex-keyprobe2-add-mcp.txt
e47054b494cb219c7808e70c2e8842740cfab89f02163913573eb5ebd4576232  evidence/raw/codex-keyprobe2-add-rules.txt
a0171a43cdd6e5002e9bb11bdc6fa4401c2853ff4e8314dcf1045a8ef1a60468  evidence/raw/codex-keyprobe2-add-skillsnokey.txt
b70564737e1dec3027cc725ae97943f76364e11a0a657da3684052a40e553cf8  evidence/raw/codex-keyprobe2-marketplace-add.txt
f22a1b530f1072cd7c842d679df644cda3f16d56d6205efe4494c69e61c1db73  evidence/raw/codex-login-status-scratch.txt
3a38cafaa0d6de73e9030024f6f75056fa6e9697d4282430785406ca1d03c9f8  evidence/raw/codex-marketplace-add.txt
38207f60ee9afe70cf4a19441f0dafd8e75465aa16aadd6ff481be9835acd150  evidence/raw/codex-marketplace-list-after.txt
e4009917b5b5c1a4798ae00cf1f9b31688b3a34879167e77317b598e0d520494  evidence/raw/codex-marketplace-list-before.txt
262db57267cc30f6a65f0dc0e284179e0416ea2fe691ab1f9b9672ac7225cd09  evidence/raw/codex-marketplace-schema-1.txt
da454d24bcb8fbd3a40a6e075b14a8c0d1facc4c65d11fbf8d8dda0cbd92ac7d  evidence/raw/codex-mcp-list-2.txt
3eaf5e80864ad6acf50513865c63be9ff9bce2e5fdbc2e6e169b306174302d6d  evidence/raw/codex-mcp-list.txt
67ade424ea68a36352933bc1c12ae2b5dba4e4b2cb9c8f4529cc13c78b5e49cc  evidence/raw/codex-plugin-add-R01.txt
10fe383df4183ce41b51debb0f3d83f3a8bfa886b3ce2584a928fcb48a8dee84  evidence/raw/codex-plugin-add-R02.txt
c9be97a9350f83b449e36223056d84ed4b98257675f9e0278f3bbb6f31711210  evidence/raw/codex-plugin-add-R03.txt
3c3495caac3f197b0940897c24dc07cb94ce344cbc1c74804d911702216fb62c  evidence/raw/codex-plugin-add-R04.txt
cb0a5ae8599934143dee824efc0818bcc8d860a826d56103b330c0bead3673c7  evidence/raw/codex-plugin-add-R05.txt
a128afe47acc508794dce2f04866ca6453666c176d91430414693147dbfc39ae  evidence/raw/codex-plugin-add-R06.txt
ba21e69d907c362cb6cb110b978131627a33f26b30b0fef39927cc28f6306d20  evidence/raw/codex-plugin-add-R07.txt
2379a56d3c6d4e19c882efc0ffcfe42ce9f179b178dfe3c9dc30101b876470f2  evidence/raw/codex-plugin-add-R08.txt
f24299ccb8322682c57796341693a17abec71bf6077fc6d8b71973f29b50a101  evidence/raw/codex-plugin-add-R09.txt
7a1be395a92befd4060dd206fd00ff79cf6ca4d09a64851b0a88ca469e7d5d58  evidence/raw/codex-plugin-add-R10.txt
63d18842c029e9d653f395ee83f7df5aa2bef10b41cc726b42aa545c978d6091  evidence/raw/codex-plugin-add-R11.txt
8c550ecd65b44218bb3421ef6b57446c63492871ed894a12c5f57f1b4c4f0929  evidence/raw/codex-plugin-add-R12.txt
023c21e61b29146fbd76400dd5a386b4e5e546fa7302ff453ba8a0289aee069d  evidence/raw/codex-plugin-add-R13.txt
65211713f239bf873ec011e2f3a14b36210fb3fed37c66ea812a529aaf587b3d  evidence/raw/codex-plugin-add-R14.txt
44446e2643c17973698a62eb584208d4cf5bcd24106dfaff05d31c1c32ca34b5  evidence/raw/codex-plugin-add-help.txt
821863ac8991d16f6face9145ea1aa2f81f06a76fbee0358e283be5b30f9c3cf  evidence/raw/codex-plugin-help.txt
3b2b27e974e833269a155833d9a34ea70ddf9d247de2c75e7d64192389b87d77  evidence/raw/codex-plugin-list-available-before.txt
4dc2cb883c01b5b4554de23e42d273dba9ca71d6e164e3416c9d767e06f5904b  evidence/raw/codex-plugin-list-help.txt
dbde2a5ecf034860e07954232b438803ebd1a6d09f640cec5909e73a8a65bdb5  evidence/raw/codex-plugin-list-json.txt
ac2a46a45d28522bd2967ba44eaef1bd2994c8d99a56fe98dcd4b1db57cdb581  evidence/raw/codex-plugin-list.txt
e11d34965c464482132caa1cb9b9c3f8b6da8a955125dd4e57a4b760b96bcdd3  evidence/raw/codex-plugin-marketplace-help.txt
f42f09f3fd9e5ea742ee2db2bdcc9a1fb4f2c1b774bef558d6873426b81498de  evidence/raw/codex-prompt-render-scan.txt
a6ac00f536cfa53b2b88d6c73cf818d35ff2dddedb4a5594fb3574f136eeb3d3  evidence/raw/r08-entry-diff.txt
404c0da68754d66048d9028cef92914e2d91463c4ff0f34298ad58508bb7670a  evidence/raw/rollout-switch-lane12-record.txt
2076d3eba9dec7fa54c7e69b747737ea907ababed5339ffbcdaf208529d34628  evidence/raw/version-end-claude.txt
7c10a281edbe62daed39eb6de9a06219db0e7c3087c700021594c92398096319  evidence/raw/version-end-codex.txt
d8bce495607812f305011a13dd45f357dac1fb82f6d778dcf9a8c87e84015800  evidence/refusals.txt
097ac3e741721f2057954f3a0a5e53314e65cb31b94c030384faf3659cb26105  evidence/rollout-switch-by-profile.txt
9292d1a8d31e328ad155e156ec8dccbba2ac32ebed40d2c076b80ce5ce50665d  evidence/routing.txt
a40afe3fa62975ff78b7f7df7e551d8a30959928d7cbc39bb526d0bedf328311  evidence/run-notes-m3.txt
3e44d03dce94924f603630727ef9375d8dfd3209760710f137978e86c7c0c850  evidence/run-notes-m4.txt
afe7cc6d435dc9fd69c5f3bfc22a5b7e0c65c166e599e7ab6532d5f47efb90ed  evidence/run-notes-m5.txt
5ac08d5e3871ecf13575e1962b585f81d2ce0b4323915123ba15535fc7a50764  evidence/run-notes.txt
6d00bee37447298fbc765f304b960ae8ded82c51d291a6efe85e1723c219c894  evidence/scratch-home-after.txt
810b6506a5018aba022ca68f33ab32e6726c032b2e8e6f9296af25e44c55ff43  evidence/scratch-home-before.txt
a6d470ff2315fab8a6c40119eba4b8215ac1fa95f08966b36620aad9f57b48c6  evidence/scratch-home-final.txt
4310cbcd931abd4e426ee08527e2ad51294a72e823b0856da0f62445c9d26936  evidence/scratch-path.txt
71283ed3a4f95df3811d7802d8e75a1ad758ecb11d58a4ccbeffa311ba66e0eb  evidence/version-end.txt
27db28475d07c03fd03e89a688568a2915187c13ba5f25e65f4cbe73ae9c8fe7  evidence/version-start.txt
d0ebb77654447979cc3bb173dbb3f7b16fb293ecc4c146aee176c940786e2c96  probe.sh
12009162895aa5ab175ec89369682013713007f180d9d4febad4cd7929fd81dd  run-checks.sh
5f287888a651c1c5566ca465d7e555c01748a7d9f8d6d33faa2d1e361b16a9e4  verdict.md
not indexed (derived after the index or rewritten by each checker run): evidence/carrier-e2.txt, evidence/carrier-e3.txt, evidence/check-results.txt
aggregates (directories not listed file by file; each hash is over the sorted per-file listing produced by the same hash tool):
abbadcbf760b4889eec61de321bc35efbd41aa604017df963879cf796244615a  evidence/manifest/ (94 files, *.sha256 only; aggregate = sha256 of the sorted per-file listing)
3e369572ce864ff11e9a990ef9d89a252e2567ebb4227324ddb608aebc18516f  evidence/stream/ (74 files; aggregate = sha256 of the sorted per-file listing)
3c5392dcf5dd2ce1c7dcf792a936b36a7cebb0ebb0458493a6da412ef3783ce9  evidence/m3/ (9 files; aggregate = sha256 of the sorted per-file listing)
1b9e682c2f812bdd9a4fa8a4f0ab5a1947f76f151896dc6c7384720107b2f204  evidence/m4/ (27 files; aggregate = sha256 of the sorted per-file listing)
208540411695ce881caa267d60dd3fce5b5cbee5e603e301e4c996c9a05e1d55  evidence/mutants/ (427 files; aggregate = sha256 of the sorted per-file listing)

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready with open items (a checker reads RESULT=FAIL; see the list below, nothing relaxed)
run_complete_at: 2026-10-02
phase: run, milestones M1-M5 of plan.md (measurement and synthesis; no product code, no cycle_type implementation)
executor: the lane orchestrator and general-purpose agents standing in for manager-develop (see the deviation in section E.2)
verdict: .moai/reports/t1434/verdict.md (local-only), verdict_sha256: 5f287888a651c1c5566ca465d7e555c01748a7d9f8d6d33faa2d1e361b16a9e4
checker results on this tree (verbatim counters from .moai/reports/t1434/evidence/check-results.txt, produced by run-checks.sh; the exit code follows each line):
- AC-002 commands: CMDS=171 SCRUB-MATCH=171 ENV-NAMES-MATCH=1 CMD-COUNT-MATCH=1 SECRET-VALUES=0 RESULT=PASS (EXIT=0)
- AC-003 isolation: TOOLS=2 ISOLATED=2 FAILED=0 BEFORE-EMPTY=2 BEFORE-NONEMPTY=0 AFTER-NONEMPTY=2 CRED-FILES=0 RESULT=PASS (EXIT=0)
- AC-004 observe-only: REAL-CMDS=41 FORBIDDEN-REAL=0 REAL-OFF-ALLOWLIST=1 SCRATCH-NO-OVERRIDE=0 REAL-WITH-OVERRIDE=0 RESULT=FAIL (EXIT=1)
- AC-005 manifests: ROWS=16 REQUIRED-ROWS=4 PAIRS=42 REAL-CMDS=41 LEAK=0 LEAK-FORBIDDEN=0 LEAK-UNREPORTED=0 LEAK-CONTINUED=0 EXCLUDED-DECLARED=7 BUILDER-CONTROL=PASS RESULT=PASS (EXIT=0)
- AC-006 static: ROWS=14 FIXTURES=14 VALIDATE-PLAIN=14 VALIDATE-STRICT=14 INSTALL-FILES=14 INSTALL-OK=14 DETAILS=14 TEST-FIXTURES=1 TESTS-MATCH=1 RESULT=PASS (EXIT=0)
- AC-007 runtime: RUNTIME-ROWS=10 SWEPT=10 NOT-SWEPT-UNOBSERVED=0 BAD-CONTROL=0 BAD-RUNS=0 RUNS=36 RUNTIME-RUN-CAP=60 OVER-CAP=0 RESULT=PASS (EXIT=0)
- AC-008 cells: CELLS=28 CARDLESS=0 MISSING-FIELDS=0 ORPHAN-RAW=0 QUOTE-MISMATCH=0 EMPTY-REASON=0 RESULT=PASS (EXIT=0)
- AC-009 negative-controls: BASES-GREEN=7 MUTANTS=40 MUTANTS-RED=40 RESULT=PASS (EXIT=0)
- AC-010 r08: EXEC-BAD=0 EXPANSION-BAD=0 EPOCH-BAD=0 ENTRY-DIFF-BAD=0 EXPANSION-PLUGIN=expanded EXPANSION-PROJECT=expanded TIMEOUT-CUTOFF=yes SECONDS=n/a RESULT=PASS (EXIT=0)
- AC-011 rules: RUNS-BAD=0 CONTROL-BAD=0 CONFIG-SCOPE-BAD=0 VALIDATE-MSG-BAD=0 RESULT=PASS (EXIT=0)
- AC-012 r04: TESTS-EXECUTED=1 TESTS-FOUND=1 OUTCOME=plugin-ok-allowed TESTS-GAP=0 PAYLOAD-BAD=0 CELL-MISMATCH=0 RESULT=PASS (EXIT=0)
- AC-013 codex: ADDS-OK=14 LISTED=14 ADD-MISMATCH=0 VALIDATE-LINES=0 REAL-CODEX-WRITES=0 RESULT=PASS (EXIT=0)
- AC-014 table: ROWS=14 HEADER=1 VOCAB-BAD=0 HOME-BAD=0 INCONSISTENT=0 CONSEQ-BAD=0 FLOOR=7 FLOOR-MET=7 RESULT=PASS (EXIT=0)
- AC-015 stamps: RAW=145 STAMPED=145 VERSION-EQUAL=1 SHA-RECOMPUTED=1 NONGOAL=1 CARRIER-ROWS=14 CARRIER-SHA=1 RESULT=PASS (EXIT=0)
open items for the leader (read from the results above, none relaxed):
- AC-004 is red: REAL-OFF-ALLOWLIST=1, CMD-ID 70 (claude plugin test --help, home=real, a help verb outside the REQ-004 read-only list; its protected-set pair shows rows-changed=0). No forbidden verb ran against a real home (FORBIDDEN-REAL=0). The verdict records it as Gap i. Decision owner: the leader (PASS-with-debt, spec amendment of the read-only list, or re-plan).
- the check-evidence.sh modes are measurement helpers and carry no mutants; only check-verdict.sh (40 mutants over 7 bases) and the manifest-builder control were observed failing (evidence/negative-control-checker.txt, evidence/negative-control-manifest.txt).
- the one tracked write is written but not committed here: progress.md sections E.2 and E.3, and the spec.md status line draft to in-progress; the commit is the orchestrator's.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Written by the orchestrator (the lane session) before the first run-phase `Agent()` spawn, per
`orchestration-mode-selection.md` §D. progress.md is not a plan-artifact hash subject, so this section
does not change the audited hash.

Input parameters: tier M; scope = 14 fixtures, 1 probe script, 2 checker scripts, 28 evidence cards, one tracked
write; domain count 1 (measurement tooling + evidence); file language mix = shell and markdown (no Go); concurrency
benefit LOW (every real-home command shares one profile and one nested-session contamination surface; one writer
per tree); Agent Teams prerequisites not requested.

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | probe and checker authoring is not trivial |
| serial | yes | milestones M1-M5 are ordered and each reads the previous one's evidence |
| fanout | no | not multi-domain research; parallel real-profile sessions would contaminate each other |
| sweep | no | not a uniform mechanical transform |

Decision: serial

Justification: the run phase is coding-heavy measurement work with hard ordering (isolation verdict at M1 before any
runtime command, runtime observation at M3 after static evidence at M2, verdict synthesis at M5). Per the coding-task
parallelism caveat the sequential path is the safe default.

Kickoff gate (default autonomous form, `auto-semantics.md` §9.1) — all four conditions read this run:
1. independent plan-audit verdict PASS: `.moai/reports/t1434/plan-audit-iter3.md`, 0.84 against the Tier M threshold 0.80, 0 must-fix;
2. plan phase records audit-ready: §E.1 above;
3. plan-artifact hash unchanged since that verdict: `cat acceptance.md plan.md spec.md | shasum -a 256` printed `651321b5eaef6a062af0025179424a6a5477414e7ce12d6c1accb5e2a914d8fb`, identical to the audited hash;
4. no blocker open (iteration-3 findings G1-G13 are should-fix or advisory; the Tier M ceiling extension was approved by the leader, 2026-10-02).

decision record: decided_by=lane-6 orchestrator evidence_refs=.moai/reports/t1434/plan-audit-iter3.md#PASS-0.84,.moai/reports/t1434/plan-audit-iter2.md,leader-message-2026-10-02(ceiling+1,scope-trim) ladder_path=gate-row:plan-run-kickoff(§9.1 autonomous)

Gap: the home decision board file could not be located on this build (no `moai` decision subcommand; the home state
directory holds no decision file), so the record above is carried here, where the sync audit re-reads it.

Run-phase binding clarifications (from the iteration-3 audit; they refine the SPEC without changing its text, and the
delegation prompt must carry them verbatim):
- G1: label a "control fired, mod marker absent" R04 result `UNOBSERVED(modules-not-loaded; raw; quote)` (cause-neutral, not "under claude -p"); the R04 card's Baseline-attribution states the hooks-modules rollout-switch state observed under each route (the first line of `claude plugin test --help` under the scratch home at M2; under the real profile it printed "hooks modules are turned off in this process: the rollout switch served off" when the iteration-3 auditor ran it twice, so run it once there only if the verb is added to the read-only list as a recorded not-run otherwise).
- G2: an absent R08 plugin-hook line cannot separate non-load from non-expansion; the card says so in Gaps and quotes the `hook-missing.log` line when it exists; the §5 R08 focus is read as what REQ-010 measures (plugin copy uses `${CLAUDE_PLUGIN_ROOT}`, project copy `${CLAUDE_PROJECT_DIR}`).
- G3: AC-001 is green at arrival (RN-001 proves only that the instrument is not blind); BASE is the `run_start_sha` above.
- G4: record the checker's own live-name enumeration under `evidence/` at run time; kept names go in a separate section of `env-scrub.txt` excluded from the byte-equality test.
- G5: the green-path flip milestone of the checkers that read `verdict.md` is M5 (raw files exist from M3).
- G6: the delegation for the tracked write names `cycle_type: ddd` with PRESERVE = everything outside `progress.md` §E.2/§E.3 and the `spec.md` `status:`/`updated:` lines; this §F block is the orchestrator's one additional write to progress.md.
- G7: end-of-run version readings run under the scratch override (`home=scratch`) so the LEAK stop rule is never broken by `--version`.
- G8: M1 order = enumerate names, build `env-scrub.txt`, run `claude auth status` scrubbed, then write `env.txt`; M5 order = stage the carrier block, delegate the write, then run `check-evidence.sh stamps`.
