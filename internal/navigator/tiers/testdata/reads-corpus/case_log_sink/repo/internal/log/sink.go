package log

// LogSink is the interface every log destination implements.
type LogSink interface {
	Write(b []byte) (int, error)
}
