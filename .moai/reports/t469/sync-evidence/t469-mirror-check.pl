#!/usr/bin/perl
# AC-HWD-015 strip-aware mirror check (SPEC-HOOK-WIRING-DRIFT-001 acceptance.md
# §I.4 A4 command), extracted byte-for-byte from the criterion for re-execution.
# Same perl program; invoked as `perl t469-mirror-check.pl <files...>` so the
# worktree guard can see a plain interpreter-plus-script invocation.
local $/;
my $rc = 0;
for my $f (@ARGV) {
    open my $L, "<", $f or die "open $f: $!";
    open my $T, "<", "internal/template/templates/$f" or die "open tmpl $f: $!";
    my ($a, $b) = (<$L>, <$T>);
    for ($a, $b) {
        s/\s*\((?:SPEC-[A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*|REQ-[A-Z][A-Z0-9]*)-[0-9]{3}(?:\s+[A-Z][0-9]{1,2})?\)//g;
        s/\s*\b(?:SPEC-[A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*|REQ-[A-Z][A-Z0-9]*)-[0-9]{3}\b//g;
        s/\s*\bt[0-9]{2,3}\b//g;
        s/\s*\b[0-9a-f]{40}\b//g;
    }
    if ($a ne $b) { print "MISMATCH $f\n"; $rc = 1; }
}
exit $rc;
