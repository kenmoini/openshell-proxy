package api

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/stretchr/testify/require"

	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
)

func wsName(w *v1.Workspace) string { return w.Name }

func TestListPage_SingleFetch(t *testing.T) {
	pager := gatewaytest.StaticPager([]*v1.Workspace{{Name: "a"}, {Name: "b"}}, "tok2")
	page, err := listPage(context.Background(), pager, wsName)
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, page.Items)
	require.Equal(t, "tok2", page.NextPageToken)
}

func TestListPage_FetchesExactlyOnce(t *testing.T) {
	calls := 0
	pager := v1.NewPager("", func(context.Context, string) (*v1.Page[*v1.Workspace], error) {
		calls++
		return &v1.Page[*v1.Workspace]{Items: []*v1.Workspace{{Name: "a"}}, NextPageToken: "more"}, nil
	})
	page, err := listPage(context.Background(), pager, wsName)
	require.NoError(t, err)
	require.Equal(t, 1, calls)
	require.Equal(t, "more", page.NextPageToken)
}

func TestListPage_EmptyIsEmptyArray(t *testing.T) {
	page, err := listPage(context.Background(), gatewaytest.StaticPager([]*v1.Workspace{}, ""), wsName)
	require.NoError(t, err)
	b, err := json.Marshal(page)
	require.NoError(t, err)
	require.Contains(t, string(b), `"items":[]`)
	require.NotContains(t, string(b), "nextPageToken")
}

func TestListPage_NilPage(t *testing.T) {
	// An exhausted pager returns a nil page from NextPage.
	pager := gatewaytest.StaticPager([]*v1.Workspace{{Name: "a"}}, "")
	_, err := pager.NextPage(context.Background())
	require.NoError(t, err)
	page, err := listPage(context.Background(), pager, wsName)
	require.NoError(t, err)
	b, _ := json.Marshal(page)
	require.JSONEq(t, `{"items":[]}`, string(b))
}

func TestListPage_Error(t *testing.T) {
	boom := errors.New("boom")
	pager := v1.NewPager("", func(context.Context, string) (*v1.Page[*v1.Workspace], error) { return nil, boom })
	page, err := listPage(context.Background(), pager, wsName)
	require.Nil(t, page)
	require.ErrorIs(t, err, boom)
}
