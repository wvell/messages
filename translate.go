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

const (
	// AttributesKey is the key that is used for :attribute replacements that can be dynamic.
	// This is primarily used for the field names in validation messages.
	attributesKey = "attributes"

	// attributeReplacement is the key that is used for the :attribute replacement.
	attributeReplacement = "attribute"
)

// Key is a type that represents a translation key.
// Msgextractor will look for this type in the source code to extract all keys.
type Key string

var isFile = regexp.MustCompile(`^([a-zA-Z]{2}(?:[-_][a-zA-Z]{2})?)\.json$`)

// FromDir reads all translations from the given directory and returns a new Translator.
// The directory should contain simple json files with the translations.
// The filename should be the language code, e.g. en.json.
//
// Translations should be in the format:
//
//	{
//		"validation.required": ":Attribute is required.",
//		"attributes": {
//			"addr_street" : "street"
//		}
//	}
//
// A translation file can have attributes. An attribute changes the replacement value before it is inserted into the translation.
// This can be useful for validation rules. Field names often have name like first_name or last_name.
// When using the translation example above:
//
//	translator.Translate("validation.required", map[string]any{"attribute": "addr_street"})
//	Output: Street is required.
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
	// Optional default language to use when no language is set in the context or the selected language has no matching translation.
	defaultLanguage LanguageID
}

// Opt is a functional option for the Translator.
type Opt func(*Translator)

// FromLanguage returns a new translator for the given lang or an error if lang is not correctly formed.
func (t *Translator) Translate(ctx context.Context, key Key, replacements map[string]any) string {
	messages := t.messages(ctx)
	if messages == nil {
		return string(key)
	}

	return messages.format(key, replacements)
}

// messages returns the messages for the given language in the context.
func (t *Translator) messages(ctx context.Context) *messages {
	// Get the language from the context.
	// Fallback to the defaultLanguage. If no language can be detected return the translation key.
	lang := FromCtx(ctx)
	if lang.Empty() {
		if t.defaultLanguage.Empty() {
			return nil
		}

		lang = t.defaultLanguage
	}

	// Try to find a message that matches the language and the region if provided.
	messages, ok := t.languages[lang.String()]
	if ok {
		return messages
	}

	// Check if we can find a language without a region.
	messages, ok = t.languages[lang.Language]
	if ok {
		return messages
	}

	// If a defaultLanguage is provided and it is different from the current lang we retry using the defaultLanguage.
	if !t.defaultLanguage.Empty() && t.defaultLanguage != lang {
		messages, ok := t.languages[t.defaultLanguage.String()]
		if ok {
			return messages
		}
	}

	return nil
}

// Use the given default language when the ctx has no language set or the language has no translations.
func WithDefaultLanguage(lang LanguageID) Opt {
	return func(t *Translator) {
		t.defaultLanguage = lang
	}
}

// Messages holds all messages for a specific language.
type messages struct {
	messages map[Key]message
	// Attributes can be used to transform the :attribute replacement before they are inserted into the translated message.
	// This is used for validation field names.
	attributes map[string]string
}

// Format formats the message with the given replacements.
func (m *messages) format(translationKey Key, replacements map[string]any) string {
	message, ok := m.messages[translationKey]
	if !ok {
		return string(translationKey)
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

		// Check if the replacement is :attribute.
		if replacementName == attributeReplacement {
			if value, ok := m.attributes[formattedValue]; ok {
				formattedValue = value
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

// Message represents a message for a specific language.
type message struct {
	message string
	// Replacements holds the replacement options for the message.
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