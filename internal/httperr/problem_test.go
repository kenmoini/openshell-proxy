package httperr

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProblem_JSONShape(t *testing.T) {
	b, err := json.Marshal(New(http.StatusNotFound, ""))
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "Not Found", m["title"])
	assert.EqualValues(t, 404, m["status"])
	assert.Len(t, m, 2)
}

func TestWrite(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("X-Request-ID", "rid-1")
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	Write(w, r, New(http.StatusTeapot, "short"))
	assert.Equal(t, http.StatusTeapot, w.Code)
	assert.Equal(t, "application/problem+json", w.Header().Get("Content-Type"))
	var m map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &m))
	assert.Equal(t, "rid-1", m["requestId"])
	assert.Equal(t, "short", m["detail"])
}

func TestInstall_HumaValidationUsesProblem(t *testing.T) {
	Install()
	se := huma.Error422UnprocessableEntity("x", &huma.ErrorDetail{Message: "bad", Location: "body.name", Value: 1})
	p, ok := se.(*Problem)
	require.True(t, ok)
	assert.Equal(t, 422, p.Status)
	assert.Equal(t, "x", p.Detail)
	require.Len(t, p.Errors, 1)
	assert.Equal(t, "body.name", p.Errors[0].Location)
}

func TestInstall_RedactsBodyValues(t *testing.T) {
	Install()
	se := huma.Error422UnprocessableEntity("x",
		&huma.ErrorDetail{Message: "bad", Location: "body.spec.credentials.TOKEN", Value: "s3cret"},
		&huma.ErrorDetail{Message: "bad", Location: "query.pageSize", Value: 5})
	p := se.(*Problem)
	require.Len(t, p.Errors, 2)
	assert.Nil(t, p.Errors[0].Value)
	assert.Equal(t, 5, p.Errors[1].Value)
}

func TestInstall_PreservesPlainErrors(t *testing.T) {
	Install()
	se := huma.Error422UnprocessableEntity("x", errors.New("plain"), nil)
	p := se.(*Problem)
	require.Len(t, p.Errors, 1)
	assert.Equal(t, "plain", p.Errors[0].Message)
}

func TestProblem_WithHeader(t *testing.T) {
	p := New(http.StatusUnauthorized, "x").WithHeader("WWW-Authenticate", "Bearer")
	assert.Equal(t, "Bearer", p.GetHeaders().Get("WWW-Authenticate"))
	rec := httptest.NewRecorder()
	Write(rec, httptest.NewRequest(http.MethodGet, "/", nil), p)
	assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"))
}
