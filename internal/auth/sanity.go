package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	// ErrMalformed means the token is not shaped like a JWT.
	ErrMalformed = errors.New("auth: malformed token")
	// ErrExpired means the token's exp claim is in the past (beyond skew).
	ErrExpired = errors.New("auth: token expired")
)

// SanityCheck cheaply rejects tokens that are obviously unusable. It does not
// verify signatures or issuers; the gateway is the authority for that.
func SanityCheck(token string, now time.Time, skew time.Duration) error {
	segs := strings.Split(token, ".")
	if len(segs) != 3 {
		return ErrMalformed
	}
	for _, s := range segs {
		if s == "" {
			return ErrMalformed
		}
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(segs[1], "="))
	if err != nil {
		return ErrMalformed
	}
	var claims map[string]json.RawMessage
	if err := json.Unmarshal(payload, &claims); err != nil || claims == nil {
		return ErrMalformed
	}
	raw, ok := claims["exp"]
	if !ok {
		return nil
	}
	var exp float64
	if err := json.Unmarshal(raw, &exp); err != nil {
		return nil // non-numeric exp is ignored
	}
	expiry := time.Unix(int64(exp), 0)
	if expiry.Add(skew).Before(now) {
		return ErrExpired
	}
	return nil
}
