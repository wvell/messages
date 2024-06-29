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
		transformers: rawMessages.Transformers,
	}

	for key, value := range rawMessages.Messages {
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
func RawTranslationsFromFile(filename string) (*RawMessages, error) {
	// Open the translations file.
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	rawMessages := &RawMessages{
		Messages:     make(map[string]string),
		Transformers: make(map[string]map[string]string),
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting file info: %w", err)
	}

	// If the file is empty, return an empty map.
	if stat.Size() == 0 {
		return rawMessages, nil
	}

	err = json.NewDecoder(f).Decode(&rawMessages)
	if err != nil {
		return nil, fmt.Errorf("decoding file: %w", err)
	}

	return rawMessages, nil
}

type RawMessages struct {
	Messages     map[string]string
	Transformers map[string]map[string]string
}

func (r *RawMessages) UnmarshalJSON(data []byte) error {
	var temp map[string]any
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	r.Messages = make(map[string]string)
	r.Transformers = make(map[string]map[string]string)

	for key, value := range temp {
		if key == "@transform" {
			transformers, ok := value.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid format for @transform")
			}

			for tKey, tValue := range transformers {
				if transformerMap, ok := tValue.(map[string]interface{}); ok {
					r.Transformers[tKey] = make(map[string]string)
					for subKey, subValue := range transformerMap {
						if subStr, ok := subValue.(string); ok {
							r.Transformers[tKey][subKey] = subStr
						} else {
							return fmt.Errorf("invalid format for transformer value: %s", subKey)
						}
					}
				} else {
					return fmt.Errorf("invalid format for transformer: %s", tKey)
				}
			}
		} else {
			if strValue, ok := value.(string); ok {
				r.Messages[key] = strValue
			} else {
				return fmt.Errorf("invalid format for message value: %s expected string", key)
			}
		}
	}
	return nil
}
