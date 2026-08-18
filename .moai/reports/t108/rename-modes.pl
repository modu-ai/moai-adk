#!/usr/bin/perl
# t108 batch mode rename — 6-mode numeric catalog → 4-mode spawn-count catalog.
# Ordered rules: parenthetical/compound forms first, then bare numbers.
# EXCLUDED from the run list: execution-rules.md (its "Mode 1/2/3" are git
# strategy modes — a different taxonomy), the already-rewritten canonical
# orchestration-mode-selection.md, and handoff enum tokens (targeted edits).
use strict;
use warnings;

my @rules = (
    # compounds
    qr/Modes 4\/5\/6/        => 'fanout/serial/sweep',
    qr/Modes 1-6/            => 'direct/serial/fanout/sweep',
    qr/Mode 4 \+ Mode 5/     => 'fanout + serial',
    # hyphenated shapes
    qr/Mode-5-shaped/        => 'serial-shaped',
    qr/Mode-4-shaped/        => 'fanout-shaped',
    # parenthetical (order before bare)
    qr/Mode 1 \(trivial\)/       => 'direct',
    qr/Mode 2 \(background\)/    => 'background',
    qr/Mode 3 \(agent-team\)/    => 'agent-team',
    qr/Mode 3 \(Agent Teams\)/   => 'agent-team',
    qr/Mode 4 \(parallel\)/      => 'fanout',
    qr/Mode 4 \(Parallel\)/      => 'fanout',
    qr/Mode 5 \(sub-agent\)/     => 'serial',
    qr/Mode 5 \(Sub-Agent\)/     => 'serial',
    qr/Mode 6 \(workflow\)/      => 'sweep',
    qr/Mode 6 \(Workflow\)/      => 'sweep',
    qr/Mode 6 \(dynamic-workflow\)/ => 'sweep',
    # colon tree forms
    qr/Mode 1: TRIVIAL/      => 'direct',
    qr/Mode 2: BACKGROUND/   => 'background',
    qr/Mode 3: AGENT-TEAM/   => 'agent-team',
    qr/Mode 4: PARALLEL/     => 'fanout',
    qr/Mode 5: SUB-AGENT/    => 'serial',
    qr/Mode 6: WORKFLOW/     => 'sweep',
    # bare (last)
    qr/Mode 7/               => 'new mode',
    qr/Mode 1\b/             => 'direct',
    qr/Mode 2\b/             => 'background',
    qr/Mode 3\b/             => 'agent-team',
    qr/Mode 4\b/             => 'fanout',
    qr/Mode 5\b/             => 'serial',
    qr/Mode 6\b/             => 'sweep',
    # catalog-size phrasing
    qr/6-mode/               => '4-mode',
    qr/\b6 modes\b/          => '4 modes',
    qr/\bsix modes\b/        => 'four modes',
);

local $/; # slurp
while (my $file = shift @ARGV) {
    open my $in, '<', $file or die "$file: $!";
    my $text = <$in>;
    close $in;
    my $orig = $text;
    my $i = 0;
    while ($i < @rules) {
        my ($pat, $rep) = ($rules[$i], $rules[$i + 1]);
        $text =~ s/$pat/$rep/g;
        $i += 2;
    }
    next if $text eq $orig;
    open my $out, '>', $file or die "$file: $!";
    print {$out} $text;
    close $out;
    print "transformed: $file\n";
}
