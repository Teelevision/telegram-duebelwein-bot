package playlist

import (
	"sort"
	"sync"
	"time"
)

// Playlist is a playlist that is ordered bei the score of the entries. The
// higher the score the further to the top an entry is. Tied entries are ordered
// by the order they were added.
type Playlist struct {
	l       sync.RWMutex
	entries []*Entry
}

// Add adds a medium to the playlist.
func (p *Playlist) Add(medium interface{}) *Entry {
	e := &Entry{
		playlist: p,
		addedAt:  time.Now(),
		medium:   medium,
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.entries = append(p.entries, e)
	return e
}

// Order returns the order of all entries.
func (p *Playlist) Order() []*Entry {
	p.l.Lock()
	defer p.l.Unlock()
	p.sort()
	return append([]*Entry(nil), p.entries...) // TODO: why copy?
}

func (p *Playlist) sort() {
	var s sorter
	s = append(s, p.entries...)
	sort.Sort(s)
	p.entries = s
}

func (p *Playlist) remove(entry *Entry) {
	p.l.Lock()
	defer p.l.Unlock()
	for i, e := range p.entries {
		if entry == e {
			p.entries = append(p.entries[:i], p.entries[i+1:]...)
			return
		}
	}
}

type sorter []*Entry

func (s sorter) Len() int {
	return len(s)
}

func (s sorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sorter) Less(i, j int) bool {
	// compare score
	if is, js := s[i].score, s[j].score; is != js {
		return is > js
	}
	// compare age
	return s[i].addedAt.Before(s[j].addedAt)
}

// Entry is an entry in a playlist.
type Entry struct {
	playlist *Playlist
	score    int
	addedAt  time.Time
	medium   interface{}
}

// Score returns the current score of the entry.
func (e *Entry) Score() int {
	e.playlist.l.RLock()
	defer e.playlist.l.RUnlock()
	return e.score
}

func (e *Entry) SetScore(score int) {
	e.playlist.l.Lock()
	defer e.playlist.l.Unlock()
	e.score = score
}

func (e *Entry) Medium() interface{} {
	e.playlist.l.RLock()
	defer e.playlist.l.RUnlock()
	return e.medium
}

func (e *Entry) Remove() {
	e.playlist.remove(e)
}
