package messages

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrDuplicateReplacementWithDifferentCase = fmt.Errorf("duplicate replacement with different case")
)

var messageRe = regexp.MustCompile(`:[A-Za-z]+`)

// parseFile reads the given file and it's optional transformer file and parses the translations.
func parseFile(file string) (*messages, error) {
	rawMessages, err := RawTranslationsFromFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	messages := &messages{
		messages:     make(map[string]message),
		transformers: make(map[string]map[string]string),
	}

	transformerFile := strings.TrimSuffix(file, ".json") + ".transformer.json"
	if _, err = os.Stat(transformerFile); err == nil {
		tf, err := os.Open(transformerFile)
		if err != nil {
			return nil, fmt.Errorf("opening transformer file: %w", err)
		}
		defer tf.Close()

		err = json.NewDecoder(tf).Decode(&messages.transformers)
		if err != nil {
			return nil, fmt.Errorf("decoding transformer file: %w", err)
		}
	}

	for key, value := range rawMessages {
		message := message{
			message:      value,
			replacements: make(map[string]replacement),
		}

		matches := messageRe.FindAllString(value, -1)
		for _, match := range matches {
			runes := []rune(match[1:])
			replacementKey := strings.ToLower(match[1:])
			isUpper := unicode.IsUpper(runes[0])

			// Check if the replacement already exists with a different case.
			if existing, ok := message.replacements[replacementKey]; ok {
				if existing.isUpper != isUpper {
					return nil, fmt.Errorf("%w: message %q replacement %q", ErrDuplicateReplacementWithDifferentCase, key, replacementKey)
				}
			}

			message.replacements[replacementKey] = replacement{
				isUpper:        isUpper,
				replacementKey: match,
			}
		}
		messages.messages[key] = message
	}

	return messages, nil
}

// RawTranslationsFromFile reads the translations from the given file and returns them as a map.
func RawTranslationsFromFile(filename string) (map[string]string, error) {
	// Open the translations file.
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	var rawMessages map[string]string
	err = json.NewDecoder(f).Decode(&rawMessages)
	if err != nil {
		return nil, fmt.Errorf("decoding file: %w", err)
	}

	return rawMessages, nil
}
