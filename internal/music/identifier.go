package music

import (
	"strings"

	"github.com/disgoorg/disgolink/v3/lavalink"
)

// schemePrefixLen is the longest scheme the bot accepts. A member may paste
// thousands of characters, and nothing past this offset can change whether the
// value starts with a scheme.
const schemePrefixLen = len("https://")

// normalizeIdentifier removes the packaging a pasted link arrives in: the
// whitespace around it and the angle brackets Discord's own link markup leaves
// behind. Pairs are stripped until none is left, which is what makes the result
// a fixed point: stripping just one would leave "<<url>>" as "<url>".
func normalizeIdentifier(identifier string) string {
	identifier = strings.TrimSpace(identifier)
	for len(identifier) >= 2 && identifier[0] == '<' && identifier[len(identifier)-1] == '>' {
		identifier = strings.TrimSpace(identifier[1 : len(identifier)-1])
	}
	return identifier
}

// isURL reports whether identifier is a link to load as given rather than a
// phrase to search for. It does not trim: callers pass a normalised value.
func isURL(identifier string) bool {
	scheme := identifier
	if len(scheme) > schemePrefixLen {
		scheme = scheme[:schemePrefixLen]
	}
	scheme = strings.ToLower(scheme)
	return strings.HasPrefix(scheme, "http://") || strings.HasPrefix(scheme, "https://")
}

// resolveIdentifier turns a normalised /play option into something Lavalink can
// load, prefixing a search type for anything that is not a URL.
func resolveIdentifier(identifier string) string {
	if isURL(identifier) {
		return identifier
	}
	return lavalink.SearchTypeYouTubeMusic.Apply(identifier)
}
