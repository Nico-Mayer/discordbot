package music

import (
	"strings"
	"testing"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/stretchr/testify/require"
)

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
		{name: "scheme uppercased", identifier: "HTTPS://example.com", want: false},
		{name: "leading space", identifier: " https://example.com", want: false},
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
