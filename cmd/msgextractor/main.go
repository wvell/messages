package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wvell/messages"
	"golang.org/x/tools/go/packages"
)

type matcher struct {
	fset *token.FileSet
}

func main() {
	var srcDir, translationDir, defaultLang string
	flag.StringVar(&srcDir, "src", ".", "The directory that contains the go source files where the translations are used. The search is recursive and includes all subdirectories.")
	flag.StringVar(&translationDir, "dst", "", "The directory that contains the translation files.")
	flag.StringVar(&defaultLang, "default-lang", "", "Provide a default language to use when adding new translations. If not provided, new translations will be added as empty strings.")
	flag.Usage = func() {
		fmt.Print(`Usage: msgextractor -src ./ -dst ./translations

Only files that exist in the translation directory will be updated. If there are no files in the translation directory nothing will be updated.
Add en empty translation file to the translation directory to add new translations.

touch ./translations/en.json

Flags:
`)

		flag.PrintDefaults()
	}

	flag.Parse()

	activeTranslations, err := collectTranslationsRecursive(srcDir)
	if err != nil {
		log.Fatalf("error collecting translations: %v", err)
	}

	err = combineTranslations(translationDir, activeTranslations, defaultLang)
	if err != nil {
		log.Fatalf("error combining translations: %v", err)
	}
}

// combineTranslations reads all translations from the files in the translationsDir.
// It performs the following actions to synchronize the active translations used in code to the translation files:
//  1. Removes translations from files if they are not present in the activeTranslations.
//  2. Adds translations to files if they are present in the activeTranslations but missing from the file.
//     If defaultLanguage is given a new translation will not be added as empty string but will take the current value of the default language.
//  3. Sorts the translations alphabetically within each file.
//
// This ensures that each translation file in the directory contains an up-to-date, consistent,
// and alphabetically sorted set of translations based activeTranslations.
//
// Note: activeTranslations is expected to be deduplicated and sorted.
func combineTranslations(translationsDir string, activeTranslations []string, defaultLanguage string) error {
	files, err := messages.TranslationFilesFromDir(translationsDir)
	if err != nil {
		return fmt.Errorf("getting translation files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("there are no translation files in dir %s, create an empty file to write translations", translationsDir)
	}

	type langTranslations struct {
		filename        string
		rawTranslations messages.RawMessages
	}

	translations := make(map[string]langTranslations)
	for languageID, file := range files {
		rawTranslations, err := messages.RawTranslationsFromFile(file)
		if err != nil {
			return fmt.Errorf("reading file %s: %w", file, err)
		}

		translations[languageID] = langTranslations{
			filename:        file,
			rawTranslations: *rawTranslations,
		}
	}

	var defaultTranslation *langTranslations
	if defaultLanguage != "" {
		translation, ok := translations[defaultLanguage]
		if !ok {
			return fmt.Errorf("default language %s not found in translation files %q", defaultLanguage, keys(translations))
		}

		defaultTranslation = &translation
	}

	for _, rawTranslations := range translations {
		newTranslations := &bytes.Buffer{}
		_, err = newTranslations.WriteString("{\n")
		if err != nil {
			return fmt.Errorf("error writing buffer: %w", err)
		}

		for _, active := range activeTranslations {
			val, ok := rawTranslations.rawTranslations.Messages[active]
			if !ok {
				if defaultTranslation != nil {
					val = defaultTranslation.rawTranslations.Messages[active]
				} else {
					val = ""
				}
			}

			var line string = fmt.Sprintf("\t\"%s\": \"%s\",\n", active, val)

			_, err := newTranslations.WriteString(line)
			if err != nil {
				return fmt.Errorf("error writing buffer: %w", err)
			}
		}

		// Write the transformers to json.
		transformerData, err := json.MarshalIndent(rawTranslations.rawTranslations.Transformers, "", "    ")
		if err != nil {
			return fmt.Errorf("error marshalling transformers: %w", err)
		}

		_, err = newTranslations.WriteString("\t\"@transform\": ")
		if err != nil {
			return fmt.Errorf("error writing buffer: %w", err)
		}
		_, err = newTranslations.Write(transformerData)
		if err != nil {
			return fmt.Errorf("error writing buffer: %w", err)
		}

		_, err = newTranslations.WriteString("}")
		if err != nil {
			return fmt.Errorf("error writing buffer: %w", err)
		}

		os.WriteFile(rawTranslations.filename, newTranslations.Bytes(), os.ModePerm)
	}

	return nil
}

func collectTranslationsRecursive(srcDir string) ([]string, error) {
	dirs, err := findSubdirectories(srcDir)
	if err != nil {
		log.Fatal(err)
	}

	activeTranslations := []string{}
	for _, dir := range dirs {
		dirTranslations, err := collectTranslations(dir)
		if err != nil {
			return nil, fmt.Errorf("error collection translations from dir %s: %w", dir, err)
		}

		activeTranslations = append(activeTranslations, dirTranslations...)
	}

	// Deduplicate the translations.
	// This is necessary because the same translation key can be used multiple times.
	activeTranslations = removeDuplicates(activeTranslations)

	// Sort the translations.
	slices.Sort(activeTranslations)

	return activeTranslations, nil
}

func collectTranslations(dir string) ([]string, error) {
	fset := token.NewFileSet()

	mode := packages.NeedName | packages.NeedSyntax |
		packages.NeedTypes | packages.NeedTypesInfo | packages.NeedCompiledGoFiles

	cfg := &packages.Config{
		Mode:  mode,
		Dir:   dir,
		Fset:  fset,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg)
	if err != nil {
		return nil, fmt.Errorf("loading package: %w", err)
	}

	pkgsErrs := ""
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			pkgsErrs += err.Error() + "\n"
		}
	})
	if pkgsErrs != "" {
		return nil, fmt.Errorf("package load error: %s", pkgsErrs)
	}

	var translations []string
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			trs, err := walkNode(file, pkg.TypesInfo)
			if err != nil {
				return nil, fmt.Errorf("walking file: %w", err)
			}

			translations = append(translations, trs...)
		}
	}

	return translations, nil
}

func walkNode(node ast.Node, info *types.Info) ([]string, error) {
	translations := []string{}
	ast.Inspect(node, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.CallExpr:
			// Check if the function is a call to the Translate function
			tr, ok := v.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			if tr.Sel.Name != "Translate" {
				return true
			}

			ident, ok := tr.X.(*ast.Ident)
			if !ok {
				return true
			}

			typ := info.TypeOf(ident)
			if typ == nil {
				log.Printf("no type info for: %v", tr.X)
				return true
			}

			if typ.String() != "*github.com/wvell/messages.Translator" {
				return true
			}

			if len(v.Args) != 3 {
				return true
			}

			keyArg, ok := v.Args[1].(*ast.BasicLit)
			if !ok {
				return true
			}

			// Add the translation key to the list.
			// The arg will be a string literal, so we need to trim the quotes.
			translations = append(translations, strings.Trim(keyArg.Value, "\""))
		}

		return true
	})

	return translations, nil
}

func findSubdirectories(rootDir string) ([]string, error) {
	subdirs := []string{rootDir}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != rootDir {
			hasGoFiles := false

			entries, err := os.ReadDir(path)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				if !entry.IsDir() && filepath.Ext(entry.Name()) == ".go" {
					hasGoFiles = true
					break
				}
			}

			if hasGoFiles {
				subdirs = append(subdirs, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return subdirs, nil
}

func removeDuplicates(input []string) []string {
	// Create a map to track seen elements
	seen := make(map[string]bool)
	// Create a slice to store unique elements
	var result []string

	// Iterate over the input slice
	for _, value := range input {
		// If the value has not been seen before, add it to the result
		if !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
	}

	return result
}

// keys returns the keys of a map as a slice.
func keys[K comparable, V any](src map[K]V) []K {
	keys := make([]K, 0, len(src))
	for k := range src {
		keys = append(keys, k)
	}

	return keys
}
