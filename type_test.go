package protostructure

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKindTypes(t *testing.T) {
	for kind, typ := range kindTypes {
		t.Run(kind.String(), func(t *testing.T) {
			require.Equal(t, kind, typ.Kind())
		})
	}
}
