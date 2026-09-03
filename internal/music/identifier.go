package music

import (
	"strings"

	"github.com/disgoorg/disgolink/v3/lavalink"
)

// isURL reports whether identifier is a link to load as given rather than a
// phrase to search for.
func isURL(identifier string) bool {
	return strings.HasPrefix(identifier, "http://") || strings.HasPrefix(identifier, "https://")
}

// resolveIdentifier turns a raw /play option into something Lavalink can load,
// prefixing a search type for anything that is not a URL.
func resolveIdentifier(identifier string) string {
	if isURL(identifier) {
		return identifier
	}
	return lavalink.SearchTypeYouTubeMusic.Apply(identifier)
}
