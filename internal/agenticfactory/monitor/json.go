package monitor

import "encoding/json"

// MarshalView renders rows as the data API the web screen consumes. Every field keeps its source,
// read time and confirmation flag; Row text is already masked by the collector.
func MarshalView(rows []Row) ([]byte, error) {
	return json.Marshal(struct {
		Rows []Row `json:"rows"`
	}{rows})
}
