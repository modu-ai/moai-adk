// Kotlin violation fixture — demonstrates patterns matched by kotlin rules.
// This file should produce >= 1 finding when scanned with the kotlin rule set.

// Matches kotlin-unused-var: val initialized to null
val x = null

// Matches kotlin-null-deref: non-null assertion operator !!
val value: String? = null
val len = value!!.length

// Matches kotlin-todo-marker: TODO() stub
fun notImplemented() = TODO()
