package invalid

import (
	"context"

	"github.com/wvell/messages"
)

func Invalid(ctx context.Context) {
	Translate("attributes", nil)
}

func Translate(key messages.Key, replacements map[string]interface{}) string {
	return string(key)
}
