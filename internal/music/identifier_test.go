package music

import (
	"strings"
	"testing"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/stretchr/testify/require"
)

func TestNormalizeIdentifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		identifier string
		want       string
	}{
		{name: "plain value is left alone", identifier: "https://example.com/track", want: "https://example.com/track"},
		{name: "surrounding whitespace is removed", identifier: " \t https://example.com/track \n ", want: "https://example.com/track"},
		{name: "angle brackets are removed", identifier: "<https://example.com/track>", want: "https://example.com/track"},
		{name: "angle brackets with padding inside", identifier: "< https://example.com/track >", want: "https://example.com/track"},
		{name: "nested brackets are stripped to the same value", identifier: "<<https://example.com/track>>", want: "https://example.com/track"},
		{name: "a lone opening bracket is kept", identifier: "<never gonna give you up", want: "<never gonna give you up"},
		{name: "brackets in the middle are kept", identifier: "a <b> c", want: "a <b> c"},
		{name: "only whitespace becomes empty", identifier: "   ", want: ""},
		{name: "an empty pair becomes empty", identifier: "<>", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeIdentifier(tt.identifier)
			require.Equal(t, tt.want, got)
			require.Equal(t, got, normalizeIdentifier(got), "normalising twice must change nothing")
			require.Equal(t, got, strings.TrimSpace(got), "a normalised value carries no surrounding whitespace")
		})
	}
}

func TestIsURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		identifier string
		want       bool
	}{
		{name: "http url", identifier: "http://example.com/track", want: true},
		{name: "https url", identifier: "https://music.youtube.com/watch?v=abc", want: true},
		{name: "bare http scheme", identifier: "http://", want: true},
		{name: "bare search phrase", identifier: "never gonna give you up", want: false},
		{name: "url mid string", identifier: "listen to https://example.com/track now", want: false},
		{name: "empty string", identifier: "", want: false},
		{name: "other scheme", identifier: "ftp://example.com/track", want: false},
		{name: "an uppercased scheme is still a url", identifier: "HTTPS://example.com", want: true},
		{name: "a mixed case scheme is still a url", identifier: "Http://example.com", want: true},
		{name: "a leading space is not trimmed here", identifier: " https://example.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, isURL(tt.identifier))
		})
	}
}

func TestResolveIdentifier(t *testing.T) {
	t.Parallel()

	t.Run("url is passed through unchanged", func(t *testing.T) {
		t.Parallel()
		const url = "https://example.com/track"
		require.Equal(t, url, resolveIdentifier(url))
	})

	t.Run("search phrase gets the youtube music prefix", func(t *testing.T) {
		t.Parallel()
		got := resolveIdentifier("never gonna give you up")
		require.Equal(t, lavalink.SearchTypeYouTubeMusic.Apply("never gonna give you up"), got)
		require.Contains(t, got, "never gonna give you up")
	})
}

func FuzzNormalizeIdentifier(f *testing.F) {
	for _, seed := range []string{
		" https://example.com ",
		"<https://example.com>",
		"< https://example.com >",
		"<<https://example.com>>",
		"HTTPS://example.com",
		"never gonna give you up",
		"   ",
		"<>",
		"\x00\xff",
		strings.Repeat("<", 5000),
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, identifier string) {
		normalized := normalizeIdentifier(identifier)

		require.Equal(t, normalized, normalizeIdentifier(normalized), "normalisation must reach a fixed point")
		require.Equal(t, normalized, strings.TrimSpace(normalized))

		if normalized != "" {
			require.NotEmpty(t, resolveIdentifier(normalized), "a value the member supplied must stay loadable")
		}
	})
}

func FuzzResolveIdentifier(f *testing.F) {
	for _, seed := range []string{
		"http://example.com/track",
		"https://music.youtube.com/watch?v=abc",
		"never gonna give you up",
		"listen to https://example.com/track now",
		"",
		"http://",
		"\x00\xff",
		strings.Repeat("a", 5000),
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, identifier string) {
		resolved := resolveIdentifier(identifier)
		if isURL(identifier) {
			require.Equal(t, identifier, resolved)
			return
		}
		require.Contains(t, resolved, identifier)
	})
}
