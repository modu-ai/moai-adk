<?php
// PHP suppressed fixture — correctly paired ast-grep-ignore + @MX:REASON.
// checkSuppressionPairing should return 0 violations for this file.

// ast-grep-ignore
// @MX:REASON test fixture for suppression policy; null assignment intentional for cache-miss handling
$x = null;

$greeting = "hello";
echo $greeting;
