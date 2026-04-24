// Kotlin valid fixture — no rule violations.
// This file should produce 0 findings when scanned with the kotlin rule set.

val x = "hello"
val value: String? = "world"
val len = value?.length ?: 0

fun implemented(): String = "result"
