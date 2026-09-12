#!/usr/bin/env perl
# Link/path resolver for the files touched by card t633.
# Usage: perl linkcheck.pl <file>...   (run from the worktree root)
# For each file it reports, per line:
#   MDLINK  - markdown link target  ](target)
#   PATH    - root-relative .claude/... path token (backticked or bare)
# Resolution rules:
#   - http(s)/mailto and pure "#anchor" links are skipped.
#   - "${CLAUDE_SKILL_DIR}/x" resolves against the file's own directory.
#   - a target starting with ".claude/" resolves against the tree root that
#     contains the file's ".claude" (repo root, or internal/template/templates).
#   - any other relative target resolves against the file's own directory.
#   - targets containing "<", "*", "{" or "$" (other than the token above) are
#     reported as SKIP (placeholders / globs), never as OK.
# Anchors: when the target is a .md file and carries "#frag", the fragment is
# compared against GitHub-style slugs of the target's headings (ANCHOR=OK|MISSING).
use strict;
use warnings;
use File::Basename qw(dirname);
use File::Spec;

sub slugs {
    my ($f) = @_;
    my %s;
    open(my $fh, '<:encoding(UTF-8)', $f) or return \%s;
    my $infence = 0;
    while (my $l = <$fh>) {
        $infence = !$infence if $l =~ /^\s*```/;
        next if $infence;
        next unless $l =~ /^#{1,6}\s+(.*?)\s*#*\s*$/;
        my $h = lc $1;
        $h =~ s/[^\p{L}\p{N}\s_-]//g;
        $h =~ s/\s/-/g;
        $s{$h} = 1;
    }
    close $fh;
    return \%s;
}

sub tree_root {
    my ($file) = @_;
    return 'internal/template/templates' if $file =~ m{^internal/template/templates/};
    return '.';
}

sub resolve {
    my ($file, $t) = @_;
    my $dir = dirname($file);
    if ($t =~ s/^\$\{CLAUDE_SKILL_DIR\}\///) {
        return (File::Spec->catfile($dir, $t), 'TOKEN');
    }
    return ('', 'SKIP') if $t =~ /[<*{\$]/;
    if ($t =~ m{^\.claude/}) {
        return (File::Spec->catfile(tree_root($file), $t), 'ROOT');
    }
    return (File::Spec->catfile($dir, $t), 'REL');
}

sub report {
    my ($file, $ln, $kind, $raw) = @_;
    my ($t, $frag) = split /#/, $raw, 2;
    return if $t eq '';
    my ($p, $mode) = resolve($file, $t);
    if ($mode eq 'SKIP') {
        print "$file:$ln\t$kind\tSKIP\t$raw\n";
        return;
    }
    my $st = -e $p ? 'OK' : 'BROKEN';
    # In the template tree a deployed file may ship as a "<name>.tmpl" source.
    if ($st eq 'BROKEN' && $file =~ m{^internal/template/templates/} && -e "$p.tmpl") {
        $st = 'OK';
        $mode .= '+TMPL';
    }
    my $anc = '';
    if (defined $frag && $st eq 'OK' && $p =~ /\.md$/) {
        $anc = slugs($p)->{lc $frag} ? "\tANCHOR=OK" : "\tANCHOR=MISSING";
    }
    my $norm = File::Spec->canonpath($p);
    1 while $norm =~ s{[^/]+/\.\./}{};
    print "$file:$ln\t$kind\t$st($mode)\t$raw\t-> $norm$anc\n";
}

for my $file (@ARGV) {
    open(my $fh, '<:encoding(UTF-8)', $file) or do { print "$file\tMISSING-FILE\n"; next };
    my $ln = 0;
    while (my $l = <$fh>) {
        $ln++;
        my %seen;
        while ($l =~ /\]\(([^)\s]+)\)/g) {
            my $t = $1;
            next if $t =~ m{^(https?:|mailto:|#)};
            $seen{$t} = 1;
            report($file, $ln, 'MDLINK', $t);
        }
        while ($l =~ m~(?<![\w/.\$\}])(\.claude/[A-Za-z0-9_./<>*\{\}\$-]*[A-Za-z0-9_/>*\}-])~g) {
            my $t = $1;
            next if $seen{$t};
            report($file, $ln, 'PATH', $t);
        }
    }
    close $fh;
}
