package extractor_sub

import (
	"context"
	"fmt"

	"github.com/wvell/messages"
)

var (
	tr      *messages.Translator
	message = tr.Translate(context.Background(), "sub.translation", nil)
)

func init() {
	fmt.Println(message)
}
