package dto

import "time"

// LogLine is one sandbox or gateway log entry.
type LogLine struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level,omitempty"`
	Target    string            `json:"target,omitempty"`
	Message   string            `json:"message"`
	Source    string            `json:"source,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// LogResult is a bounded batch of log lines.
type LogResult struct {
	Lines       []LogLine `json:"lines"`
	BufferTotal uint32    `json:"bufferTotal"`
}
