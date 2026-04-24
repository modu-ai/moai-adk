<?php
// PHP violation fixture — demonstrates patterns matched by php rules.
// This file should produce >= 1 finding when scanned with the php rule set.

// Matches php-unused-var: function returning null without type annotation
function getBadResult(): mixed {
    return null;
}

// Matches php-null-deref: json_decode with null argument
$data = json_decode(null);

// Matches php-todo-marker: stub using die()
function notImplemented(): void {
    die("TODO");
}
