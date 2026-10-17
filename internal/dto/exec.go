package dto

// ExecRequest runs a command in a sandbox and collects its output.
type ExecRequest struct {
	Command      []string          `json:"command" minItems:"1"`
	Env          map[string]string `json:"env,omitempty"`
	WorkDir      string            `json:"workDir,omitempty"`
	NoLoginShell bool              `json:"noLoginShell,omitempty"`
	// Encoding "base64" forces base64 output; invalid UTF-8 output switches
	// to base64 automatically and the response says so.
	Encoding string `json:"encoding,omitempty" enum:"utf8,base64"`
}

// ExecResult is the collected outcome of a command. A non-zero ExitCode is a
// normal result, not an error.
type ExecResult struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Encoding string `json:"encoding" enum:"utf8,base64"`
}
