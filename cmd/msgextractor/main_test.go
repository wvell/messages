package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseMessages(t *testing.T) {
	t.Run("parse in func", func(t *testing.T) {
		translations, err := collectTranslationsRecursive("./testdata")
		require.NoError(t, err)

		// Make sure the translations are deduplicated and sorted.
		require.Len(t, translations, 3)
		require.Equal(t, "login.welcome", translations[0])
		require.Equal(t, "sub.translation", translations[1])
		require.Equal(t, "zipcode", translations[2])
	})

	t.Run("", func(t *testing.T) {
		translations, err := collectTranslationsRecursive("./testdata/sub")
		require.NoError(t, err)
		require.Equal(t, "sub.translation", translations[0])
	})
}
