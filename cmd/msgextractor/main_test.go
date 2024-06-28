package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wvell/messages"
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

	err = combineTranslations("./testdata/translations", translations, "")
	require.NoError(t, err)

	// Make sure the nl.json file was written with all translations.
	rawTranslations, err := messages.RawTranslationsFromFile(nlFile)
	require.NoError(t, err)
	require.Len(t, rawTranslations, 3)
	require.Contains(t, rawTranslations, "login.welcome")
	require.Contains(t, rawTranslations, "sub.translation")
	require.Contains(t, rawTranslations, "zipcode")

	// Make sure every translation has an empty string value.
	for _, value := range rawTranslations {
		require.Equal(t, "", value)
	}
}

func TestWriteTranslationFilesWithDefault(t *testing.T) {
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

	err = combineTranslations("./testdata/translations", translations, "en")
	require.NoError(t, err)

	// Make sure the nl.json file was written and has the same value for the translations as the en.json file.
	rawEnTranslations, err := messages.RawTranslationsFromFile("./testdata/translations/en.json")
	require.NoError(t, err)
	rawNlTranslations, err := messages.RawTranslationsFromFile(nlFile)
	require.NoError(t, err)

	require.Equal(t, rawEnTranslations, rawNlTranslations)
}

func TestErrorOnUnknownDefaultLanguage(t *testing.T) {
	translations, err := collectTranslationsRecursive("./testdata/src")
	require.NoError(t, err)

	err = combineTranslations("./testdata/translations", translations, "de")
	require.Error(t, err)
}
