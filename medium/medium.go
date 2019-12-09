package medium

import (
	"log"
	"net/url"
)

// Medium is a medium, like on YouTube or Soundcloud.
type Medium interface {
	// Provider returns the provider of the medium.
	Provider() Provider
	// ID returns the identifier of the medium. The syntax highly depends on the
	// provider of the medium.
	ID() interface{}
}

// New returns a new medium or an error if it's not a supported medium.
func New(rawurl string) (Medium, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return nil, ErrNotSupported
	}

	switch url.Host {
	case "youtube.com", "youtu.be", "www.youtube.com":
		return NewYouTubeVideoFromURL(url)
	}
	log.Println(url.Host, url.Path)

	return nil, ErrNotSupported
}

// Identical returns whether both media are the same.
func Identical(a, b Medium) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Provider() == b.Provider() && a.ID() == b.ID()
}
