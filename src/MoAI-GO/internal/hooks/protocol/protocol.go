package protocol

// Timeout configuration for different hook events
// Matches Python timeout specifications
var HookTimeouts = map[HookEvent]int{
	EventSessionStart: 5,  // 5 seconds
	EventPreTool:      5,  // 5 seconds
	EventPostTool:     30, // 30 seconds (format)
	EventSessionEnd:   5,  // 5 seconds
	EventPreCompact:   3,  // 3 seconds
	EventStop:         5,  // 5 seconds
	EventNotification: 5,  // 5 seconds
	EventQualityGate:  10, // 10 seconds
	EventCommit:       5,  // 5 seconds
	EventPush:         5,  // 5 seconds
	EventCompact:      5,  // 5 seconds
}

// GetTimeout returns the timeout in seconds for a given event
func GetTimeout(event HookEvent) int {
	if timeout, ok := HookTimeouts[event]; ok {
		return timeout
	}
	return 5 // Default timeout
}
