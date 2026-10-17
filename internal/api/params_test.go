package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIfMatchHeader_Version(t *testing.T) {
	ok := []struct {
		in  string
		v   uint64
		set bool
	}{
		{`"7"`, 7, true},
		{`7`, 7, true},
		{``, 0, false},
	}
	for _, c := range ok {
		v, set, err := IfMatchHeader{IfMatch: c.in}.Version()
		require.NoError(t, err, c.in)
		require.Equal(t, c.v, v, c.in)
		require.Equal(t, c.set, set, c.in)
	}
	for _, bad := range []string{`W/"7"`, `*`, `"abc"`} {
		_, set, err := IfMatchHeader{IfMatch: bad}.Version()
		require.Error(t, err, bad)
		require.False(t, set, bad)
	}
}
