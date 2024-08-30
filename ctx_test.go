package messages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguageCtx(t *testing.T) {
	cases := []struct {
		input     string
		expectErr bool
		language  string
		region    string
	}{
		{
			input:     "en",
			expectErr: false,
			language:  "en",
		},
		{
			input:     "nl",
			expectErr: false,
			language:  "nl",
		},
		{
			input:     "en-US",
			expectErr: false,
			language:  "en",
			region:    "US",
		},
		{
			input:     "en_GB",
			expectErr: false,
			language:  "en",
			region:    "GB",
		},
		{
			input:     "invalid",
			expectErr: true,
		},
		{
			// An accept header.
			input:    "en-GB,en;q=0.5",
			language: "en",
			region:   "GB",
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			ctx, err := WithLanguage(context.Background(), c.input)
			if c.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			lang := FromCtx(ctx)
			require.Equal(t, c.language, lang.Language)
			require.Equal(t, c.region, lang.Region)
		})
	}
}

func TestToCtx(t *testing.T) {
	cases := []struct {
		input     string
		expectErr bool
		language  string
		region    string
	}{
		{
			input:     "invalid",
			expectErr: true,
		},
		{
			// An accept header.
			input:    "en-GB,en;q=0.5",
			language: "en",
			region:   "GB",
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			ctx := ToCtx(context.Background(), c.input)
			if c.expectErr {
				require.True(t, FromCtx(ctx).Empty())
				return
			}

			require.False(t, FromCtx(ctx).Empty())

			lang := FromCtx(ctx)
			require.Equal(t, c.language, lang.Language)
			require.Equal(t, c.region, lang.Region)
		})
	}
}
