# Claude style inheritance comparison

Source: modu-ai/moai-adk commit `4fa2aa9db2e0add0f1299e35407141b2b41c5ec0`, `.claude/output-styles/moai/`. Original files are preserved without activation in `references/legacy-output-styles/`, with the upstream Apache-2.0 license. The matching internal template files were checked against these source hashes.

| Style | Source SHA-256 | Selected retained behavior | Still missing or changed |
| --- | --- | --- | --- |
| MoAI | `8244a7bb3a53436aca882d7c46c7a59866f59fc3a12d232a5738a25a2f7b5f04` | Coding instructions; task/status banners and accessible progress | Many original response templates and localization details are not ported. Fixed specialist delegation, old approval gates, commands, and execution policy are not activated through this style. |
| MoAI-Easy | `c33a60904e7908fe6387d69d57f489b3565f66a2449b5103b931c8a526b8d050` | Coding instructions; plain language, four-step rhythm, friendly banners and progress | Original detailed templates, FAQ and localization rules are only partially adapted. Legacy execution and permission directives are not imported into presentation. |
| MoAI-Learn | `ef01dad72d0f76763209a94af38fee7392d12d1a98df44c6a38ff74d9f0c5638` | `keep-coding-instructions: false`, no code writing, Feynman learning phases, explicit coding-mode switch | Full original note workflow, detailed templates and optional Notion integration are not implemented. Research and note actions remain subject to actual capabilities and permission. |

## Links and support

Original styles reference legacy `.claude/rules`, `.moai/config/sections/language.yaml`, fixed agents, commands and optional Notion behavior. They are evidence, not resolved dependencies of the adapted preview. The adapted files do not claim those resources exist. Each selected presentation feature still needs a source-to-adapter mapping, host loading and behavior checks; no overall equivalence test has passed. Full source preservation does not prove relative links resolve in the new project.

## Selection rationale

Preserve user-visible voice and the coding versus no-code distinction. Keep old execution, permission and fixed-agent behavior out of styles because the RC contract assigns these to common workflows and actual host policy. This is an explicit partial adaptation pending the complete reviewed recipe, not a canonical merge of all legacy behavior.
