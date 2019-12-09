package medium

import (
	"net/url"
	"strings"
)

// ProviderYouTube is the provider for youtube videos.
var ProviderYouTube = simpleProvider("youtube")

type youTubeVideo string

func (m youTubeVideo) Provider() Provider {
	return ProviderYouTube
}

func (m youTubeVideo) ID() interface{} {
	return string(m)
}

// NewYouTubeVideo returns a new medium that is a YouTube video.
func NewYouTubeVideo(videoID string) (Medium, error) {
	return youTubeVideo(videoID), nil
}

// NewYouTubeVideoFromURL returns a new medium that is a YouTube video from a
// url.
func NewYouTubeVideoFromURL(url *url.URL) (Medium, error) {
	// if the url contains a v=videoID parameter, we'll take that
	if v := url.Query().Get("v"); v != "" {
		return youTubeVideo(v), nil
	}
	// assume video id is in path, take the first part that is long enough
	path := strings.Split(url.Path, "/")
	for _, p := range path {
		if len(p) >= 11 { // assuming id is always at least 11 chars
			return youTubeVideo(p), nil
		}
	}

	return nil, ErrInvalidURL
}
