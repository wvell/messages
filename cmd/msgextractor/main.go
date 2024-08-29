package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/spf13/afero"
	"github.com/wvell/messages"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func main() {
	var srcDir, translationDir, defaultLang string
	var overwrite bool
	flag.StringVar(&srcDir, "src", ".", "The directory that contains the go source files where the translations are used. The search is recursive and includes all subdirectories with go files.")
	flag.StringVar(&translationDir, "dst", "", "The directory that contains the translation files.")
	flag.StringVar(&defaultLang, "default-lang", "", "Provide a default language to use when adding new translations. If not provided, new translations will be added as empty strings.")
	flag.BoolVar(&overwrite, "remove", false, "Remove will remove all translations in the translation files that have not been found in src. Transformers are never removed.")
	flag.Usage = func() {
		fmt.Print(`Usage: msgextractor -src ./ -dst ./translations

Message extractor extracts all translation keys (variables of type github.com/wvell/messages.Key) from the src directory recursively and updates the translation files in the translation directory.

Only files that exist in the translation directory will be updated. If there are no files in the translation directory nothing will be updated.
Add an empty translation file to the translation directory to add new translations.

    $ touch ./translations/en.json

Flags:
`)

		flag.PrintDefaults()
	}

	flag.Parse()

	err := processTranslations(srcDir, translationDir, defaultLang, overwrite)
	if err != nil {
		log.Fatalf("error processing translations: %v", err)
	}
}

func processTranslations(srcDir, translationsDir, defaultLang string, overwrite bool) error {
	translationKeysFromSrcDir, err := messages.TranslationKeysFromSourceCode(srcDir)
	if err != nil {
		return fmt.Errorf("error reading translations from src: %w", err)
	}

	parser := messages.NewParser(afero.NewOsFs())

	files, err := parser.TranslationFilesFromDir(translationsDir)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("there are no translation files in dir %s, create an empty file to write translations", translationsDir)
	}

	defaultTranslations := &messages.RawMessages{
		Messages:   make(map[string]string),
		Attributes: make(map[string]string),
	}
	if defaultLang != "" {
		defaultLanguageID, err := messages.ParseLanguage(defaultLang)
		if err != nil {
			log.Fatalf("error parsing default language: %v", err)
		}

		// Check if the default language is a translation file.
		if _, ok := files[defaultLanguageID.String()]; !ok {
			return fmt.Errorf("default language %s not found in translation files %q", defaultLanguageID.String(), maps.Keys(files))
		}

		defaultTranslations, err = parser.MessagesFromFile(files[defaultLanguageID.String()])
		if err != nil {
			return fmt.Errorf("reading default language file: %w", err)
		}
	}

	// Loop over all translation files and update them.
	for _, file := range files {
		existingTranslations, err := parser.MessagesFromFile(file)
		if err != nil {
			return fmt.Errorf("reading language file %s: %w", file, err)
		}

		// Remove existing translations that are not present in the source code.
		if overwrite {
			existingTranslations.Messages = make(map[string]string)
		} else {
			// Output all translations that are in the translation file but not in the source code.
			for key := range existingTranslations.Messages {
				if slices.Contains(translationKeysFromSrcDir, key) {
					continue
				}

				log.Printf("translation %q is present in file %s but not found in source code, use -remove to remove this translation", key, file)
			}
		}

		for _, key := range translationKeysFromSrcDir {
			// If the key already exists we do nothing.
			if _, ok := existingTranslations.Messages[key]; ok {
				continue
			}

			// Add the key to the existing translations and use the value from the default translation if present.
			existingTranslations.Messages[key] = defaultTranslations.Messages[key]
		}

		// If there is a default language we add the missing transformers.
		if defaultLang != "" {
			for key, transformer := range defaultTranslations.Attributes {
				if _, ok := existingTranslations.Attributes[key]; !ok {
					// If the transformer is missing completely we add it.
					existingTranslations.Attributes[key] = transformer
				}
			}
		}

		// Write the translations back to the file.
		content, err := json.MarshalIndent(existingTranslations, "", "  ")
		if err != nil {
			return fmt.Errorf("marshalling translations: %w", err)
		}

		err = os.WriteFile(file, content, os.ModePerm)
		if err != nil {
			return fmt.Errorf("writing translations: %w", err)
		}
	}

	return nil
}
