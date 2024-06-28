package testdata

import (
	"context"
	"fmt"
	"log"

	"github.com/wvell/messages"
)

func other(ctx context.Context) {
	tr, err := messages.FromDir("dir")
	if err != nil {
		log.Fatal(err)
	}

	message := tr.Translate(ctx, "zipcode", nil)
	fmt.Println(message)
}

func translate(ctx context.Context) {
	tr, err := messages.FromDir("dir")
	if err != nil {
		log.Fatal(err)
	}

	message := tr.Translate(ctx, "login.welcome", map[string]any{"user": "john"})
	fmt.Println(message)

	// Use zipcode twice.
	tr.Translate(ctx, "zipcode", nil)
}
