package protostructure

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	type nested struct {
		Value string `json:"value"`
	}

	type test struct {
		Value     string            `json:"v"`
		Map       map[string]string `json:"map"`
		Nested    nested            `json:"nested"`
		NestedPtr *nested           `json:"nested_ptr"`
		Slice     []int             `json:"slice"`
		Array     [3]int            `json:"array"`
	}

	cases := []struct {
		Name  string      // name of the test
		Value interface{} // value to encode
		Err   string      // err to expect if any
		JSON  string      // json value to test unmarshal/marshal equality (can be blank)
	}{
		{
			"major field test",
			&test{},
			"",
			`
{
	"v": "hello",
	"map": { "key": "value" },
	"nested": { "value": "direct" },
	"nested_ptr": { "value": "ptr" },
	"slice": [1, 4, 8],
	"array": [1, 2, 3]
}`,
		},

		{
			"direct struct (not pointer)",
			test{},
			"",
			"",
		},

		{
			"not a struct",
			12,
			"got int",
			"",
		},
	}

	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			require := require.New(t)

			s, err := Encode(tt.Value)
			if tt.Err != "" {
				require.Error(err)
				require.Nil(s)
				require.Contains(err.Error(), tt.Err)
				return
			}
			require.NoError(err)
			require.NotNil(s)

			if tt.JSON == "" {
				return
			}

			// Unmarhal into real
			require.NoError(json.Unmarshal([]byte(tt.JSON), tt.Value))

			// Unmarshal into dynamic
			v, err := New(s)
			require.NoError(err)
			require.NoError(json.Unmarshal([]byte(tt.JSON), v))

			// Remarshal both and compare results
			marshalReal, err := json.Marshal(tt.Value)
			require.NoError(err)
			marshalDynamic, err := json.Marshal(v)
			require.NoError(err)
			require.Equal(string(marshalReal), string(marshalDynamic))
		})
	}
}
