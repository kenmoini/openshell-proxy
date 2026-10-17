package httperr

import (
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// Install makes huma build *Problem values for framework-generated errors
// (validation failures, parse errors), so every error shares one shape.
func Install() {
	huma.NewError = func(status int, message string, errs ...error) huma.StatusError {
		p := New(status, message)
		for _, e := range errs {
			switch d := e.(type) {
			case nil:
			case huma.ErrorDetailer:
				ed := d.ErrorDetail()
				// Never echo request-body values: they can be credentials, and
				// for unknown properties huma echoes the whole parent object.
				if strings.HasPrefix(ed.Location, "body") {
					c := *ed
					c.Value = nil
					ed = &c
				}
				p.Errors = append(p.Errors, ed)
			default:
				p.Errors = append(p.Errors, &huma.ErrorDetail{Message: e.Error()})
			}
		}
		return p
	}
}
