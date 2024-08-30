package messages

import (
	"context"
	"fmt"
	"regexp"

	"golang.org/x/text/language"
)

var (
	languageKey = ctxKey("locale")
	langRe      = regexp.MustCompile(`(?i)([a-z]{2,8})([-_][a-z]{4})?([-_][a-z]{2}|\d{3})?`)
)

// WithLanguage sets the language in the ctx.
// Language is parsed to retrieve the language and region.
// If the region can not reliably be parsed, it is set to an empty string.
// An error is returned if the language can not be parsed.
func WithLanguage(ctx context.Context, lang string) (context.Context, error) {
	id, err := ParseLanguage(lang)
	if err != nil {
		return nil, err
	}

	return toCtx(ctx, id), nil
}

// ToCtx is comparable to WithLanguage, but does not return an error when parsing the language fails.
// On failure it will return the context as is.
func ToCtx(ctx context.Context, lang string) context.Context {
	_ctx, err := WithLanguage(ctx, lang)
	if err != nil {
		return ctx
	}

	return _ctx
}

// LanguageFromCtx returns the language from the ctx.
func FromCtx(ctx context.Context) LanguageID {
	l, ok := ctx.Value(languageKey).(LanguageID)
	if ok {
		return l
	}

	return LanguageID{}
}

// ParseLanguage parses the language string into a LanguageID.
func ParseLanguage(lang string) (LanguageID, error) {
	match := langRe.FindString(lang)
	if match == "" {
		return LanguageID{}, fmt.Errorf("invalid language: %s", lang)
	}
	lang = match

	tag, err := language.Parse(lang)
	if err != nil {
		return LanguageID{}, fmt.Errorf("error parsing %s: %w", lang, err)
	}

	var id LanguageID

	base, baseconf := tag.Base()
	if baseconf != language.Exact {
		return LanguageID{}, fmt.Errorf("error parsing %s: could not parse base language", lang)
	}

	id.Language = base.String()

	region, regionconf := tag.Region()
	if regionconf == language.Exact {
		id.Region = region.String()
	}

	return id, nil
}

func toCtx(ctx context.Context, id LanguageID) context.Context {
	return context.WithValue(ctx, languageKey, id)
}

// LanguageID holds the language and an optional region.
type LanguageID struct {
	Language string
	Region   string
}

func (l LanguageID) String() string {
	if l.Region != "" {
		return l.Language + "-" + l.Region
	}

	return l.Language
}

func (l LanguageID) Empty() bool {
	return l.Language == "" && l.Region == ""
}

type ctxKey string
