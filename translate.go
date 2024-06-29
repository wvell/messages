package messages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var isFile = regexp.MustCompile(`^([a-zA-Z]{2}(?:[-_][a-zA-Z]{2})?)\.json$`)

// FromDir reads all translations from the given directory and returns a new Translator.
// The directory should contain simple json files with the translations.
// The filename should be the language code, e.g. en.json.
//
// Translations should be in the format:
//
//	{
//		"translationname": "translation with some replacement :field",
//	}
//
// A translation file can have transformers. A transformer changes the replacement value before it is inserted into the translation.
// This can be useful for validation rules. Field names often have name like first_name or last_name.
// Example without transformers: "The :field is required." -> "The first_name is required."
// Example with transformers: "The :field is required." -> "The first name is required."
//
// Transformers use the @transform key in the translation file.
//
//	{
//		"translationname": "translation with some replacement :field",
//		"@transform": {
//			"field": {
//				"first_name" : "First name"
//	 		}
//		}
//	}
func FromDir(dir string, opts ...Opt) (*Translator, error) {
	t := newTranslator(opts...)

	files, err := TranslationFilesFromDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading translations files: %w", err)
	}

	for languageID, file := range files {
		messages, err := parseFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading file %s: %w", file, err)
		}

		t.languages[languageID] = messages
	}
	return t, nil
}

// TranslationFilesFromDir returns all translation files from the given directory.
func TranslationFilesFromDir(dir string) (map[string]string, error) {
	// Read all files from the directory.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading translations: %w", err)
	}

	files := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		match := isFile.FindStringSubmatch(entry.Name())
		if match == nil {
			return nil, fmt.Errorf("filename %s should have format en.json or en_US.json", entry.Name())
		}

		langID, err := ParseLanguage(match[1])
		if err != nil {
			return nil, fmt.Errorf("parsing language id: %w", err)
		}

		files[langID.String()] = filepath.Join(dir, entry.Name())
	}

	return files, nil
}

// NewTranslator creates a new translator with the given options.
func newTranslator(opts ...Opt) *Translator {
	t := &Translator{
		languages: make(map[string]*messages),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Translator holds translations for all Languages. Use the Translate message to look up translations.
type Translator struct {
	languages map[string]*messages
	// Optional default language to use when no language is set in the context or the language has no translations.
	defaultLanguage LanguageID
}

// Opt is a functional option for the Translator.
type Opt func(*Translator)

// Translate looks up name in the translation table with the language fetched from the context.
// The replacements are formatted to a string when replaced.
func (t *Translator) Translate(ctx context.Context, translationKey string, replacements map[string]any) string {
	// Get the language from the context.
	// Fallback to the defaultLanguage. If no language can be detected return the translation key.
	lang := FromCtx(ctx)
	if lang.Empty() {
		if t.defaultLanguage.Empty() {
			return translationKey
		}

		lang = t.defaultLanguage
	}

	// Try to find a message that matches the language and the region if provided.
	messages, ok := t.languages[lang.String()]
	if !ok {
		// Check if we can find a language without a region.
		messages, ok = t.languages[lang.Language]
		if !ok {
			// If a defaultLanguage is provided and it is different from the current lang we retry using the defaultLanguage.
			if !t.defaultLanguage.Empty() && t.defaultLanguage != lang {
				return t.Translate(toCtx(ctx, t.defaultLanguage), translationKey, replacements)
			}

			return translationKey
		}
	}

	message, ok := messages.messages[translationKey]
	if !ok {
		return translationKey
	}

	// Replace all placeholders in the message.
	translationMessage := message.message
	for replacementName, replacement := range message.replacements {
		var formattedValue string

		// Check if the replacement is given by the caller.
		value, ok := replacements[replacementName]
		if ok {
			switch v := value.(type) {
			case string:
				formattedValue = v
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				formattedValue = fmt.Sprintf("%d", v)
			case float32, float64:
				formattedValue = fmt.Sprintf("%.2f", v)
			case bool:
				formattedValue = fmt.Sprintf("%t", v)
			}
			// No match formattedValue will be empty.
		}

		// Apply the transformer for this field.
		if transformer, ok := messages.transformers[replacementName]; ok {
			if replacementValue, ok := transformer[formattedValue]; ok {
				formattedValue = replacementValue
			}
		}

		// Uppercase the replacement if the replacemente indicated this.
		if formattedValue != "" && replacement.isUpper {
			runes := []rune(formattedValue)
			runes[0] = unicode.ToUpper(runes[0])
			formattedValue = string(runes)
		}

		translationMessage = strings.ReplaceAll(translationMessage, replacement.replacementKey, formattedValue)
	}

	return translationMessage
}

// messages holds all messages for a specific language.
type messages struct {
	messages map[string]message
	// transformers can be used to transform the replacement values before they are inserted into the message.
	transformers map[string]map[string]string
}

// message represents a message for a specific language.
type message struct {
	message string
	// replacements holds the replacement options for the message.
	// true indicates the replacement should be title cased. False indicates the replacement should be left as is.
	replacements map[string]replacement
}

type replacement struct {
	// Indicates if the replacement should be title cased.
	isUpper bool
	// Contains the complete replacement key as it is defined in the translation message.
	// For the translation "Hello :User" this would be ":User".
	replacementKey string
}

// Use the given default language when the ctx has no language set or the language has no translations.
func WithDefaultLanguage(lang LanguageID) Opt {
	return func(t *Translator) {
		t.defaultLanguage = lang
	}
}
