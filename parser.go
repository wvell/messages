package messages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/spf13/afero"
	"golang.org/x/exp/maps"
)

var (
	ErrDuplicateReplacementWithDifferentCase = fmt.Errorf("duplicate replacement with different case")
)

var messageRe = regexp.MustCompile(`:[A-Za-z]+(\.[A-Za-z]+)*`)

func NewParser(fs afero.Fs) *Parser {
	return &Parser{fs: fs}
}

type Parser struct {
	fs afero.Fs
}

// TranslationFilesFromDir returns all translation files from the given directory.
func (p *Parser) TranslationFilesFromDir(dir string) (map[string]string, error) {
	// Read all files from the directory.
	entries, err := afero.ReadDir(p.fs, dir)
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

// parseFile reads the given file and parses the translations.
func (p *Parser) parseFile(file string) (*messages, error) {
	rawMessages, err := p.MessagesFromFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	messages := &messages{
		messages:   make(map[Key]message),
		attributes: rawMessages.Attributes,
	}

	for key, value := range rawMessages.Messages {
		message := message{
			message:      value,
			replacements: make(map[string]replacement),
		}

		replacements := messageRe.FindAllString(value, -1)
		for _, replacementMatch := range replacements {
			runes := []rune(replacementMatch[1:])
			replacementKey := strings.ToLower(replacementMatch[1:])
			isUpper := unicode.IsUpper(runes[0])

			// Check if the replacement already exists with a different case.
			if existing, ok := message.replacements[replacementKey]; ok {
				if existing.isUpper != isUpper {
					return nil, fmt.Errorf("%w: message %q replacement %q", ErrDuplicateReplacementWithDifferentCase, key, replacementKey)
				}
			}

			message.replacements[replacementKey] = replacement{
				isUpper:        isUpper,
				replacementKey: replacementMatch,
			}
		}
		messages.messages[Key(key)] = message
	}

	return messages, nil
}

// RawTranslationsFromFile reads the translations from the given file and returns them as a map.
func (p *Parser) MessagesFromFile(filename string) (*RawMessages, error) {
	// Open the translations file.
	f, err := p.fs.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	rawMessages := &RawMessages{
		Messages:   make(map[string]string),
		Attributes: make(map[string]string),
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
	Messages   map[string]string
	Attributes map[string]string
}

func (r *RawMessages) UnmarshalJSON(data []byte) error {
	var temp map[string]json.RawMessage
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	r.Messages = make(map[string]string)
	r.Attributes = make(map[string]string)

	for key, value := range temp {
		if key == attributesKey {
			var attributes map[string]string
			err := json.Unmarshal(value, &attributes)
			if err != nil {
				return fmt.Errorf("invalid format for @transform: %w", err)
			}

			r.Attributes = attributes
		} else {
			var message string
			if err := json.Unmarshal(value, &message); err != nil {
				return fmt.Errorf("invalid format for message value: %s: %w", key, err)
			}

			r.Messages[key] = message
		}
	}
	return nil
}

func (r *RawMessages) MarshalJSON() ([]byte, error) {
	// First marshal all the Messages and Attributes to one map with the values as json.RawMessage.
	// We can then sort the whole map.
	var rawValues = make(map[string]json.RawMessage)
	for key, value := range r.Messages {
		data, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("marshaling message: %w", err)
		}

		rawValues[key] = json.RawMessage(data)
	}

	if r.Attributes == nil {
		r.Attributes = make(map[string]string)
	}

	attributes, err := json.Marshal(r.Attributes)
	if err != nil {
		return nil, fmt.Errorf("marshaling attributes: %w", err)
	}

	rawValues[attributesKey] = attributes

	sortedMessages, err := marshalMapToJSON(rawValues)
	if err != nil {
		return nil, fmt.Errorf("marshaling transformers: %w", err)
	}

	return json.MarshalIndent(sortedMessages, "", "  ")
}

// MarshalMapToJSON sorts the given map alphabetically by it's key and marshals it JSON and writes it to the given writer.
func marshalMapToJSON[T any](src map[string]T) (json.RawMessage, error) {
	var buf bytes.Buffer

	if len(src) == 0 {
		return []byte("{}"), nil
	}

	keys := maps.Keys(src)
	slices.Sort(keys)

	buf.Write([]byte{'{'})

	for i, key := range keys {
		value, err := json.Marshal(src[key])
		if err != nil {
			return nil, fmt.Errorf("key %s: %w", key, err)
		}

		fmt.Fprintf(&buf, "%q:", fmt.Sprintf("%v", key))
		buf.Write(value)

		if i < len(keys)-1 {
			buf.Write([]byte{','})
		}
	}

	buf.Write([]byte{'}'})

	return buf.Bytes(), nil
}
