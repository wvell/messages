package sub

import (
	"context"

	"github.com/wvell/messages"
)

var (
	tr      *messages.Translator
	message = tr.Translate(context.Background(), "sub.translation", nil)

	key          = "sub.translation"
	notCollected = tr.Translate(context.Background(), key, nil)
)
