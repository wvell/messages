# Messages
This repository handles message translations using simple JSON files for Go (Golang) projects. The messages can be specific to a language or a combination of language and region.

```json
# Example translation file en.json
{
    "welcome.message": "Welcome :user!"
}
```

## Message extraction
Users can use the msgextractor tool to extract translation keys from your go source files. This will collect every value of type github.com/wvell/messages.Key from
the src directory.

```bash
msgextractor -dst path_to_translation_files --src path_to_go_source_files

// Use -default-language to set the default language as source for the translations. Missing translations for other languages will use this as the source.
msgextractor -dst path_to_translation_files -src path_to_go_source_files -default-language en
```

## Usage
```go
// Parse translations.
translations, err := messages.FromDir("directory with translation files")
if err != nil {
    // Handle error
    log.Fatalf("Failed to parse translations: %v", err)
}

// Setting the language context per request.
// Parse the language from a user request. This can be from a header or user settings, for example the http Accept-Language header.
ctx := context.Background()
ctx, err := messages.WithLanguage(ctx, r.Header.Get("Accept-Language"))
if err != nil {
    // Handle error...
}
// Translate the message.
// If the user requested "en-US" but you only have "en" translations available, the translator will use the "en" translations.
msg := translations.Translate(ctx, "welcome.message", map[string]any{"user": "wvell"})
fmt.Println(msg) // prints: Welcome wvell!
```

## Capitalization
You can use a capitalized replacement to to capitalize the replacement value:
```json
{
    "welcome.message": "Welcome :User!"
}
```
This will change the replacement value "john" to "John".

## Attributes
Attributes allow you to reuse placeholder values, which is particularly useful for validation messages.
The following example illutrates the required validation message. Without attributes you would have to create a translation for each field(required.first_name, required.street).

```json
# es.json
{
  "required": ":Attribute es requerido"
  "attributes": {
    "field" : {
      "first_name": "nombre",
      "street": "calle"
    }
  }
}
```
```go
tr.Translate(ctx, "required", map[string]any{"attribute": "first_name"}) // Nombre es requerido
tr.Translate(ctx, "required", map[string]any{"attribute": "street"}) // Calle es requerido
```

As you can see this also takes the title case for the translation message into account.
