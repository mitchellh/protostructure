package protostructure

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	require := require.New(t)

	type nested struct {
		Value string `json:"value"`
	}

	type test struct {
		Value     string            `json:"v"`
		Map       map[string]string `json:"map"`
		Nested    nested            `json:"nested"`
		NestedPtr *nested           `json:"nested_ptr"`
	}

	s, err := Encode(&test{})
	require.NoError(err)
	require.NotNil(s)

	var data = `
{
	"v": "hello",
	"map": { "key": "value" },
	"nested": { "value": "direct" },
	"nested_ptr": { "value": "ptr" }
}`

	v, err := New(s)
	require.NoError(err)
	require.NoError(json.Unmarshal([]byte(data), v))
	t.Logf("%#v", v)
}
