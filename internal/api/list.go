package api

import (
	"context"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// listPage fetches exactly one page from pager and converts its items. A nil
// page yields an empty (non-nil) Items slice so JSON renders [] not null.
// Errors are returned unwrapped; callers map them with sdkErr.
func listPage[T, D any](ctx context.Context, pager *v1.Pager[T], conv func(T) D) (*dto.Page[D], error) {
	p, err := pager.NextPage(ctx)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return &dto.Page[D]{Items: []D{}}, nil
	}
	items := make([]D, 0, len(p.Items))
	for _, it := range p.Items {
		items = append(items, conv(it))
	}
	return &dto.Page[D]{Items: items, NextPageToken: p.NextPageToken}, nil
}
