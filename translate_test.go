package messages

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestInvalidTranslationFilename(t *testing.T) {
	_, err := NewTranslator(afero.NewOsFs(), "./testdata/invalid-language")
	require.Error(t, err)
}

func TestErrOnDuplicateReplacementWithDifferentCase(t *testing.T) {
	_, err := NewTranslator(afero.NewOsFs(), "./testdata/invalid-translation")
	require.ErrorIs(t, err, ErrDuplicateReplacementWithDifferentCase)
}

func TestTranslations(t *testing.T) {
	tr, err := NewTranslator(afero.NewOsFs(), "./testdata/valid")
	require.NoError(t, err)

	ctx, err := WithLanguage(context.Background(), "en_US")
	require.NoError(t, err)

	t.Run("non existing message", func(t *testing.T) {
		message := tr.Translate(ctx, "non.existing", nil)
		require.Equal(t, "non.existing", message)
	})

	t.Run("fill replacements with an empty string if not supplied", func(t *testing.T) {
		message := tr.Translate(ctx, "welcome.login", nil)
		require.Equal(t, "Welcome ", message)
	})

	t.Run("uppercase a replacement", func(t *testing.T) {
		message := tr.Translate(ctx, "welcome.login", map[string]any{"user": "john"})
		require.Equal(t, "Welcome John", message)
	})

	t.Run("multiple", func(t *testing.T) {
		message := tr.Translate(ctx, "multiple", map[string]any{"fruit": "apples", "total": 4, "more": 1.6})
		require.Equal(t, "I have 4 apples and will attempt to get 1.60 more apples.", message)
	})

	cases := []struct {
		name         string
		replacements map[string]any
		expected     string
	}{
		{
			name: "int",
			replacements: map[string]any{
				"total": int(5),
			},
			expected: "Total: 5",
		},
		{
			name: "int32",
			replacements: map[string]any{
				"total": int32(5),
			},
			expected: "Total: 5",
		},
		{
			name: "int64",
			replacements: map[string]any{
				"total": int64(5),
			},
			expected: "Total: 5",
		},
		{
			name: "float32",
			replacements: map[string]any{
				"total": float32(1) / float32(3),
			},
			expected: "Total: 0.33",
		},
		{
			name: "float64",
			replacements: map[string]any{
				"total": float64(1) / float64(3),
			},
			expected: "Total: 0.33",
		},
		{
			name: "bool",
			replacements: map[string]any{
				"total": true,
			},
			expected: "Total: true",
		},
		{
			name: "struct",
			replacements: map[string]any{
				"total": struct{ Name string }{Name: "john"},
			},
			expected: "Total: ",
		},
		{
			name: "slice",
			replacements: map[string]any{
				"total": []string{"test"},
			},
			expected: "Total: ",
		},
		{
			name: "map",
			replacements: map[string]any{
				"total": map[string]string{"test": "test"},
			},
			expected: "Total: ",
		},
	}

	for _, tc := range cases {
		t.Run("convert_"+tc.name, func(t *testing.T) {
			message := tr.Translate(ctx, "convert.case", tc.replacements)
			require.Equal(t, tc.expected, message)
		})
	}
}

func TestFallbackToLanguageWithoutRegion(t *testing.T) {
	tr, err := NewTranslator(afero.NewOsFs(), "./testdata/valid")
	require.NoError(t, err)

	ctx, err := WithLanguage(context.Background(), "nl_NL")
	require.NoError(t, err)

	message := tr.Translate(ctx, "welcome.login", map[string]any{"user": "jan"})
	require.Equal(t, "Welkom jan", message)
}

func TestFallbackToDefaultLanguage(t *testing.T) {
	en, err := ParseLanguage("en-US")
	require.NoError(t, err)

	tr, err := NewTranslator(afero.NewOsFs(), "./testdata/valid", WithDefaultLanguage(en))
	require.NoError(t, err)

	ctx, err := WithLanguage(context.Background(), "de-AT")
	require.NoError(t, err)

	message := tr.Translate(ctx, "welcome.login", map[string]any{"user": "jan"})
	require.Equal(t, "Welcome Jan", message)
}

func TestAttribute(t *testing.T) {
	tr, err := NewTranslator(afero.NewOsFs(), "./testdata/valid")
	require.NoError(t, err)

	ctx, err := WithLanguage(context.Background(), "en_US")
	require.NoError(t, err)

	message := tr.Translate(ctx, "required", map[string]any{"attribute": "first_name"})
	require.Equal(t, "First name is required", message)
}
