#!/usr/bin/env perl
# Anchor checker for card t633 (SK-03 anchor follow-up).
# Usage: perl anchorcheck.pl <file>...   (run from the worktree root)
# Reports every "spec-workflow.md#<frag>" occurrence, in two forms:
#   MDLINK  - inside a markdown link ](<path>spec-workflow.md#frag); <path> resolved
#             against the file's own directory
#   MENTION - a bare mention (e.g. an @MX:NOTE comment); resolved against the tree
#             root's .claude/rules/moai/workflow/spec-workflow.md
# The fragment is compared against GitHub-style heading slugs of the target:
#   lowercase; drop every character that is not a letter, digit, space, "_" or "-";
#   each space becomes "-" (no collapsing). Duplicate-heading "-1" suffixes are not
#   generated (none needed for the headings checked here).
# A self-test line for the two headings in scope is printed first.
use strict;
use warnings;
use File::Basename qw(dirname);
use File::Spec;

sub slug {
    my ($h) = @_;
    $h = lc $h;
    $h =~ s/[^\p{L}\p{N} _-]//g;
    $h =~ s/ /-/g;
    return $h;
}

sub slugs {
    my ($f) = @_;
    my %s;
    open(my $fh, '<:encoding(UTF-8)', $f) or return \%s;
    my $infence = 0;
    while (my $l = <$fh>) {
        $infence = !$infence if $l =~ /^\s*```/;
        next if $infence;
        next unless $l =~ /^#{1,6}\s+(.*?)\s*$/;
        $s{slug($1)} = 1;
    }
    close $fh;
    return \%s;
}

sub tree_root {
    my ($file) = @_;
    return 'internal/template/templates' if $file =~ m{^internal/template/templates/};
    return '.';
}

print "SELFTEST\t", slug('Subcommand Classification (Pipeline vs Multi-Agent)'), "\n";
print "SELFTEST\t", slug('Mode Dispatch Cross-Reference'), "\n";

for my $file (@ARGV) {
    open(my $fh, '<:encoding(UTF-8)', $file) or do { print "$file\tMISSING-FILE\n"; next };
    my $ln = 0;
    while (my $l = <$fh>) {
        $ln++;
        while ($l =~ /(\]\(([^)\s]*?)spec-workflow\.md#([A-Za-z0-9_-]+)\))|((?<![\/\w])spec-workflow\.md#([A-Za-z0-9_-]+))/g) {
            my ($kind, $target, $frag);
            if (defined $1) {
                $kind = 'MDLINK';
                $target = File::Spec->catfile(dirname($file), $2 . 'spec-workflow.md');
                $frag = $3;
            } else {
                $kind = 'MENTION';
                $target = File::Spec->catfile(tree_root($file), '.claude/rules/moai/workflow/spec-workflow.md');
                $frag = $5;
            }
            my $norm = File::Spec->canonpath($target);
            1 while $norm =~ s{[^/]+/\.\./}{};
            my $st = -e $target ? 'FILE=OK' : 'FILE=BROKEN';
            my $anc = ($st eq 'FILE=OK' && slugs($target)->{$frag}) ? 'ANCHOR=OK' : 'ANCHOR=MISSING';
            print "$file:$ln\t$kind\t$st\t$anc\t#$frag\t-> $norm\n";
        }
    }
    close $fh;
}
