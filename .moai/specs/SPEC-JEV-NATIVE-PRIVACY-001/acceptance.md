# Acceptance — SPEC-JEV-NATIVE-PRIVACY-001

**AC-JNP-001** (REQ-JNP-001) — Given an existing Jev credential, When saved and loaded through the credential owner, Then the unchanged format round-trips at the single `~/.moai/.env.typesafe` location outside the repository, with no alternate writer or storage route.

**AC-JNP-002** (REQ-JNP-002) — Given a fresh Unix credential or an existing mode-0644 credential, When a credential is saved, Then its final mode is exactly 0600 in both cases and existing narrowing assertions remain unchanged.

**AC-JNP-003** (REQ-JNP-003) — Given a fresh Windows credential destination or one with broad/inheriting access, When replacement is created and credential bytes are written, Then independent native inspection proves the current user owns it and its protected, non-inheriting DACL has exactly one explicit owner grant before payload writing and after publication.

**AC-JNP-004** (REQ-JNP-004) — Given null, broad, inherited, unknown or unreadable Windows access policy, When save/readback cannot prove owner-only native protection, Then save returns an error and no success result is reported; Windows `Mode().Perm()` is never used as the privacy evidence.

**AC-JNP-005** (REQ-JNP-005) — Given the actual import graph, When credential-owner/helper/client boundaries are checked, Then only the Windows filesystem-security leaf uses existing x/sys/windows, the helper has no upward credential/application/model/network imports, and the unchanged Jev stdlib-only test and its positive control both pass.

**AC-JNP-006** (REQ-JNP-006) — Given existing schema, masking, short-key, model/gate and authority regression checks, When the native privacy amendment is integrated, Then they retain their existing expectations and pass, including no key disclosure at length four or less and no expansion of either completed-core authority exception.

**AC-JNP-007** (REQ-JNP-007) — Given the closed v0.5.0 core baseline, When this lifecycle repair is reviewed, Then core spec/plan/design/acceptance match both recorded baseline commits byte-for-byte, original progress/research are unchanged, the unchanged doctrine test passes, this successor remains honestly unfinished, and v0.6.0 history is retained without a new completion claim.

**AC-JNP-008** (REQ-JNP-008) — Given a request to close this successor, When its evidence is inspected, Then actual native Windows fresh/replacement credential integration and independent owner/DACL readback, unchanged Unix/disclosure regressions, latest-head code/security review, independent audit and all required CI are recorded by their real owners; absent observations leave closure pending.

## Quality gates and Definition of Done

All eight criteria are mandatory. Native credential integration cannot be replaced by helper-only output, skipped cases, emulated permission bits or weakened assertions. Existing security/audit functions, core and GKI historical evidence, and the CI baseline are retained. Structural lint or doctrine PASS alone does not close this successor. No new independent audit has been performed by this artifact's author.
