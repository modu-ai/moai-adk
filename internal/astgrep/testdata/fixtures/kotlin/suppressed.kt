// Kotlin suppressed fixture — correctly paired ast-grep-ignore + @MX:REASON.
// checkSuppressionPairing should return 0 violations for this file.

// ast-grep-ignore
// @MX:REASON test fixture for suppression policy; !! assertion intentional for test scaffold
val value: String? = null
val len = value!!.length
