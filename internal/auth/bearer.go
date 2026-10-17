package auth

import "strings"

// ParseBearer extracts the token from an Authorization header value. The scheme
// is matched case-insensitively and exactly one non-empty token must follow.
func ParseBearer(header string) (token string, ok bool) {
	fields := strings.Fields(strings.TrimSpace(header))
	if len(fields) != 2 || !strings.EqualFold(fields[0], "bearer") {
		return "", false
	}
	return fields[1], true
}
