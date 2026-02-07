package defs

// BackupTimestampFormat is the Go time layout for backup directory names (YYYYMMDD_HHMMSS).
const BackupTimestampFormat = "20060102_150405"

// StatusLinePath is the relative path to the status-line shell script.
// Uses a relative path because StatusLine does not support env-var expansion.
const StatusLinePath = ".moai/status_line.sh"
