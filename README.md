# Messages
This repository handles message translations using simple JSON files for Go (Golang) projects. The translations can be specific to a language or a combination of language and region.

## Usage
Users can use the msgextractor tool to extract messages from your go source files. The messages will be added/removed to/from your translation files.

```bash
msgextractor --translations path_to_translation_files --src path_to_go_source_files

// Use --default-language to set the default language as source for the translations. Missing translations for other languages will use this as the source.
msgextractor --translations path_to_translation_files --src path_to_go_source_files --default-language en
```

## How to use it in code
```go
// Create a new translator.
tr, err := messages.FromDir("dir")
if err != nil {
    // Handle error
    log.Fatalf("Failed to create translator: %v", err)
}

// Setting the language context per request.
// Parse the language from a user request. This can be from a header or user settings, for example.
ctx := context.Background()
ctx, err := messages.WithLanguage(ctx, "en")
if err != nil {
    // Handle error
    log.Fatalf("Failed to set language: %v", err)
}

// Translate the message.
// If the user requested en-US but you only have en translations available, the translator will use the en translations.
msg := tr.Translate(ctx, "welcome.message", map[string]any{"user": "wvell"})
fmt.Println(msg) // prints: Welcome wvell!
```

## Translation files
Translation files are named based on the language or language-region codes:
- en.json for English
- en_US.json for American English
- es.json for Spanish

Each translation file follows this format:
```json
{
    "welcome.message": "Welcome :user!"
}
```

You can use a capitalized replacement to to capitalize the replacement value:
```json
{
    "welcome.message": "Welcome :User!"
}
```
This will change the replacement value john to John.

## Transformer files
Transformer files allow you to replace placeholder values, which is particularly useful for validation messages. These files are named similarly to the translation files, with .transformer added.

Using transformers allows you to reuse translations. The following example illutrates the required validation. Without transformers you would have to create a translation for each field(required.first_name, required.street).

### Example
es.json
```json
{
  "required" : ":Field es requerido"
}
```

es.transformer.json
```json
{
  "field": {
      "first_name": "nombre"
      "street": "calle"
    }
}

tr.Translate(ctx, "required", map[string]any{"field": "first_name"}) // Nombre es requerido
tr.Translate(ctx, "required", map[string]any{"field": "street"}) // Calle es requerido
```
As you can see this also takes the title case for the translation message into account.
