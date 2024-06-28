package messages

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllowEmptyTranslationFiles(t *testing.T) {
	_, err := RawTranslationsFromFile("./testdata/empty/empty.json")
	require.NoError(t, err)
}
