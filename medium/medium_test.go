package medium_test

import (
	"errors"
	"testing"

	. "github.com/Teelevision/telegram-duebelwein-bot/medium"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		desc   string
		rawurl string
		err    error
		medium Medium
	}{
		{
			desc:   "not a url",
			rawurl: "foobar",
			err:    ErrNotSupported,
			medium: nil,
		}, {
			desc:   "youtube url",
			rawurl: "https://www.youtube.com/watch?v=cNtZAbq2Ig4",
			err:    nil,
			medium: youTubeVideo("cNtZAbq2Ig4"),
		}, {
			desc:   "youtube url without video",
			rawurl: "https://www.youtube.com/",
			err:    ErrInvalidURL,
			medium: nil,
		}, {
			desc:   "youtube short url",
			rawurl: "https://youtu.be/YgGzAKP_HuM",
			err:    nil,
			medium: youTubeVideo("YgGzAKP_HuM"),
		}, {
			desc:   "youtube playlist url",
			rawurl: "https://www.youtube.com/watch?v=jZya02M_caU&list=PL81aLNZD3wMVLx-weUkf_un7MOHFnD08D",
			err:    nil,
			medium: youTubeVideo("jZya02M_caU"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			m, err := New(tC.rawurl)
			if !errors.Is(err, tC.err) {
				t.Fatalf("got error %q, expected %q", err, tC.err)
			}
			if !Identical(m, tC.medium) {
				t.Errorf("expected medium %#v (%s), got %#v (%s).", tC.medium.ID(), tC.medium.Provider(), m.ID(), m.Provider())
			}
		})
	}
}

func TestIdentical(t *testing.T) {
	ytVid := youTubeVideo("foobar")

	testCases := []struct {
		desc   string
		a, b   Medium
		result bool
	}{
		{
			desc:   "same object",
			a:      ytVid,
			b:      ytVid,
			result: true,
		}, {
			desc:   "different media",
			a:      &someMedium{fooProvider{}, 1234},
			b:      &someMedium{barProvider{}, "abcd"},
			result: false,
		}, {
			desc:   "different media of same provider",
			a:      &someMedium{fooProvider{}, 1234},
			b:      &someMedium{fooProvider{}, 9999},
			result: false,
		}, {
			desc:   "different media of with same id",
			a:      &someMedium{fooProvider{}, 1111},
			b:      &someMedium{barProvider{}, 1111},
			result: false,
		}, {
			desc:   "different instances of same media",
			a:      &someMedium{fooProvider{}, 1111},
			b:      &someMedium{fooProvider{}, 1111},
			result: true,
		}, {
			desc:   "one nil",
			a:      &someMedium{fooProvider{}, 1111},
			b:      nil,
			result: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			if i := Identical(tC.a, tC.b); i != tC.result {
				t.Errorf("expected Identical(%#v [%s], %#v [%s]) to be %t, was %t.", tC.a.ID(), tC.a.Provider(), tC.b.ID(), tC.b.Provider(), tC.result, i)
			}
		})
	}
}

func youTubeVideo(videoID string) Medium {
	m, err := NewYouTubeVideo(videoID)
	if err != nil {
		panic(err)
	}
	return m
}

type fooProvider struct{}

func (p fooProvider) String() string {
	return "foo"
}

type barProvider struct{}

func (p barProvider) String() string {
	return "bar"
}

type someMedium struct {
	provider Provider
	id       interface{}
}

func (m *someMedium) Provider() Provider {
	return m.provider
}

func (m *someMedium) ID() interface{} {
	return m.id
}
