package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFromSource(t *testing.T) {
	translations, err := collectTranslationsRecursive("./testdata/src")
	require.NoError(t, err)

	// Make sure the translations are deduplicated and sorted.
	require.Len(t, translations, 3)
	require.Equal(t, "login.welcome", translations[0])
	require.Equal(t, "sub.translation", translations[1])
	require.Equal(t, "zipcode", translations[2])
}

func TestWriteTranslationFiles(t *testing.T) {
	// Add an empty nl.json file to the translations directory.
	nlFile := "./testdata/translations/nl.json"
	tmp, err := os.OpenFile(nlFile, os.O_TRUNC|os.O_CREATE, os.ModePerm)
	require.NoError(t, err)
	tmp.Close()
	t.Cleanup(func() {
		os.Remove(nlFile)
	})

	translations, err := collectTranslationsRecursive("./testdata/src")
	require.NoError(t, err)

	err = combineTranslations("./testdata/translations", translations)
	require.NoError(t, err)
}
