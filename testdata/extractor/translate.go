package extractor

import (
	"context"
	"fmt"

	"github.com/wvell/messages"
)

const (
	usedConst                = "used.const"
	unusedConst messages.Key = "unused.const"
)

var (
	usedVar                = "used.var"
	unusedVar messages.Key = "unused.var"

	tr *messages.Translator
)

func UseMessagesTranslate(ctx context.Context) {
	message := tr.Translate(context.Background(), "login.welcome", map[string]any{"user": "john"})
	fmt.Println(message)

	// Use zipcode twice.
	tr.Translate(context.Background(), "zipcode", map[string]any{"user": "john"})
	tr.Translate(context.Background(), "zipcode", map[string]any{"user": "john"})
	fmt.Println(unusedVar)
}

func UseFunc(ctx context.Context) {
	Translate("use.func", nil)
}

func UseFuncWithConst(ctx context.Context) {
	Translate(usedConst, nil)
}

func UseFuncWithVar(ctx context.Context) {
	Translate(messages.Key(usedVar), nil)
}

func UseFuncWithInlineVar(ctx context.Context) {
	var translation messages.Key = "inline.var"

	Translate(translation, nil)
}

func Translate(key messages.Key, replacements map[string]interface{}) string {
	return string(key)
}

func SameSignature(key string, replacements map[string]interface{}) string {
	return key
}
