package convert

import (
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/stretchr/testify/require"
)

func TestSettingValue_OnlyMatchingMember(t *testing.T) {
	cases := []v1.SettingValue{
		{Type: v1.SettingValueString, StringVal: "x"},
		{Type: v1.SettingValueBool, BoolVal: true},
		{Type: v1.SettingValueInt, IntVal: 5},
		{Type: v1.SettingValueBytes, BytesVal: []byte("b")},
	}
	for _, c := range cases {
		d := SettingValueFromSDK(c)
		n := 0
		if d.String != nil {
			n++
		}
		if d.Bool != nil {
			n++
		}
		if d.Int != nil {
			n++
		}
		if d.Bytes != nil {
			n++
		}
		require.Equal(t, 1, n, c.Type)
		require.Equal(t, c, SettingValueToSDK(d))
	}
}
