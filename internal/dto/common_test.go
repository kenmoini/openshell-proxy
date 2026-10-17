package dto

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUint64String_RoundTrip(t *testing.T) {
	b, err := json.Marshal(Uint64String(math.MaxUint64))
	require.NoError(t, err)
	require.Equal(t, `"18446744073709551615"`, string(b))

	var v Uint64String
	require.NoError(t, json.Unmarshal(b, &v))
	require.Equal(t, Uint64String(math.MaxUint64), v)

	require.NoError(t, json.Unmarshal([]byte(`7`), &v))
	require.Equal(t, Uint64String(7), v)
	require.NoError(t, json.Unmarshal([]byte(`"7"`), &v))
	require.Equal(t, Uint64String(7), v)

	for _, bad := range []string{`"x"`, `-1`, `"-1"`, `1.5`, `true`, `null1`} {
		require.Error(t, json.Unmarshal([]byte(bad), &v), bad)
	}
}

func TestObjectMeta_ResourceVersionOmitted(t *testing.T) {
	b, err := json.Marshal(ObjectMeta{Name: "a", ResourceVersion: 3})
	require.NoError(t, err)
	require.JSONEq(t, `{"name":"a","resourceVersion":"3"}`, string(b))
}
