package messages

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTranslationKeysFromSourceCode(t *testing.T) {
	translations, err := TranslationKeysFromSourceCode("./testdata/extractor")
	require.NoError(t, err)

	require.Len(t, translations, 9)

	for _, find := range []string{"login.welcome", "zipcode", "use.func", "used.const", "unused.const", "used.var", "unused.var", "inline.var"} {
		require.Contains(t, translations, find)
	}
}

func TestTranslationKeysFromSourceCodeInvalid(t *testing.T) {
	_, err := TranslationKeysFromSourceCode("./testdata/extractor-invalid")
	require.ErrorIs(t, err, ErrInvalidTranslationKey)
}
