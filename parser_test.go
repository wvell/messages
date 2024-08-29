package messages

import (
	"bytes"
	"flag"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

var genGolden = flag.Bool("gen_golden", false, "Generate golden template files")

func TestAllowEmptyTranslationFiles(t *testing.T) {
	parser := NewParser(afero.NewOsFs())

	_, err := parser.MessagesFromFile("./testdata/empty/empty.json")
	require.NoError(t, err)
}

func TestMarshalSorts(t *testing.T) {
	raw := RawMessages{
		Messages: map[string]string{
			"zero": "Zero",
			"one":  "One",
		},
		Attributes: map[string]string{
			"required": "Required",
			"email":    "Email",
		},
	}

	data, err := raw.MarshalJSON()
	require.NoError(t, err)

	if *genGolden {
		err := os.WriteFile("./testdata/marshal/expected.json", data, 0644)
		require.NoError(t, err)
	} else {
		expected, err := os.ReadFile("./testdata/marshal/expected.json")
		require.NoError(t, err)
		require.True(t, bytes.Equal(data, expected), "expected: %q\n, got: %q\n", string(expected), string(data))
	}
}
