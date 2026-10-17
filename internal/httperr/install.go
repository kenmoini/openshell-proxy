package httperr

import "github.com/danielgtaylor/huma/v2"

// Install makes huma build *Problem values for framework-generated errors
// (validation failures, parse errors), so every error shares one shape.
func Install() {
	huma.NewError = func(status int, message string, errs ...error) huma.StatusError {
		p := New(status, message)
		for _, e := range errs {
			if d, ok := e.(huma.ErrorDetailer); ok {
				p.Errors = append(p.Errors, d.ErrorDetail())
			}
		}
		return p
	}
}
