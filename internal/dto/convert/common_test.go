package convert

import (
	"testing"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/stretchr/testify/require"
)

func TestDeletion_Mapping(t *testing.T) {
	cases := []struct {
		in     *v1.DeletionResult
		out    string
		status int
	}{
		{&v1.DeletionResult{Outcome: v1.DeletionAccepted, SandboxID: "s1"}, "Accepted", 202},
		{&v1.DeletionResult{Outcome: v1.DeletionAlreadyAbsent}, "AlreadyAbsent", 200},
		{&v1.DeletionResult{Outcome: v1.DeletionCompleted}, "Completed", 200},
		{&v1.DeletionResult{Outcome: v1.DeletionUnspecified}, "Unspecified", 200},
		{nil, "Unspecified", 200},
	}
	for _, c := range cases {
		require.Equal(t, c.out, Deletion(c.in).Outcome)
		require.Equal(t, c.status, DeletionStatus(c.in))
	}
	require.Equal(t, "s1", Deletion(cases[0].in).SandboxID)
}

func TestTimePtr(t *testing.T) {
	require.Nil(t, TimePtr(time.Time{}))
	now := time.Now()
	require.Equal(t, now, *TimePtr(now))
}
