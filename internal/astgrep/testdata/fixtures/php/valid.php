<?php
// PHP valid fixture — no rule violations.
// This file should produce 0 findings when scanned with the php rule set.

$x = "hello";
$data = json_decode('{"key":"value"}');
echo $x;

function implementedMethod(): string {
    return "actual result";
}
