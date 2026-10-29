package monitor

import "encoding/json"

// MarshalView renders the view as the data API the web screen consumes. Every field keeps its source,
// read time and confirmation flag; Row text is already masked by the collector.
func MarshalView(v View) ([]byte, error) { return json.Marshal(v) }
