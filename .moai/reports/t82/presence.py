"""Mechanical presence check: one distinctive marker per mapped clause.

Not a proof that the obligation is preserved (that is the trace-table read), but it
catches an accidental drop during editing. Each entry is a substring that must appear
in AGENTS.md if the clause reached the document. Run from the worktree root.
"""

text = open('AGENTS.md', encoding='utf-8').read()

markers = {
    'B036': 'MUST NOT assert a verification',
    'B037': 'Baseline-integrity attribution',
    'B038': 'Evidence-bearing report format',
    'B012': 'git rev-list --count --left-right',
    'B013': 'git commit -a',
    'B076': 'Never change branch state in the primary checkout',
    'B077': 'git branch --show-current',
    'B066': 'entered through the launcher',
    'B067': 'closes L2 trees only',
    'B068': 'only copy of the work',
    'B069': 'Start a new card in a new worktree',
    'B070': 'git branch -m WT-<slug>',
    'B071': 'Three traceability carriers',
    'B010': 'Batch independent read-only verifications',
    'B064': 'is not evidence that a review ran',
    'B072': 'Scope verification to the change',
    'B073': 'Never spawn background load',
    'B074': 'one compound invocation',
    'B028': 'Surface assumptions',
    'B029': 'Manage confusion actively',
    'B030': 'Push back when warranted',
    'B031': 'Enforce simplicity',
    'B032': 'Maintain scope discipline',
    'B033': "Verify, don't assume",
    'B002': 'conversation_language`.** Code, identifiers',
    'B003': 'User-facing output is Markdown',
    'B004': 'XML is reserved for agent-to-agent data transfer',
    'B014': 'Never use time predictions',
    'B025': 'colloquial native register',
    'B034': 'native idiom, not English mapped word-for-word',
    'B095': 'hand-authored `\\uXXXX` escapes are PROHIBITED',
    'B005': 'Maintain effectiveness without MCP servers',
    'B009': 'Follow tool usage patterns optimized for accuracy',
    'B040': 'Keep command output bounded',
    'B041': 'Prefer the quiet form of routine commands',
    'B042': 'Weigh session length as a cost axis',
}

missing = [bid for bid, m in sorted(markers.items()) if m not in text]
print('clauses checked: %d | missing: %s' % (len(markers), missing or 'none'))
raise SystemExit(1 if missing else 0)
