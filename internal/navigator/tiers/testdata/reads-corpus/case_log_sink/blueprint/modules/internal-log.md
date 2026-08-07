# internal-log

Provides structured logging primitives.

LogSink is the interface every log destination implements (file, stderr,
in-memory). Format helpers translate structured records into text.
