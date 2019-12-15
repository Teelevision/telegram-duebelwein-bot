package room

import (
	"sort"
	"sync"
	"time"

	"github.com/Teelevision/telegram-duebelwein-bot/medium"
)

// Room is a room where media is played.
type Room struct {
	l     sync.RWMutex
	users map[interface{}]*userInfo
	media map[medium.Medium]*mediumInfo
}

// New creates a new room.
func New() *Room {
	return &Room{
		users: make(map[interface{}]*userInfo),
		media: make(map[medium.Medium]*mediumInfo),
	}
}

// UserJoins adds a user to the room.
func (r *Room) UserJoins(user interface{}) {
	r.l.Lock()
	defer r.l.Unlock()
	if _, exists := r.users[user]; !exists {
		r.users[user] = &userInfo{}
	}
}

// UserLeaves removes a user from the room.
func (r *Room) UserLeaves(user interface{}) {
	r.l.Lock()
	defer r.l.Unlock()
	delete(r.users, user)
	// remove all media and votes of that user
	for m, info := range r.media {
		if info.user == user {
			// inform that the medium was removed
			info.played <- ErrMediumUnknown
			delete(r.media, m)
			continue
		}
		// remove vote
		if _, ok := info.votes[user]; ok {
			info.vote(user, 0)
		}
	}
}

// UserQueuesMedium adds a medium to the room.
func (r *Room) UserQueuesMedium(user interface{}, m medium.Medium) (<-chan error, error) {
	r.l.Lock()
	defer r.l.Unlock()
	// get user info
	if _, ok := r.users[user]; !ok {
		return nil, ErrUserUnknown
	}
	// check if duplicate
	for existing := range r.media {
		if medium.Identical(m, existing) {
			return nil, ErrMediumAlreadyExists
		}
	}
	// add medium
	info := &mediumInfo{
		user:    user,
		addedAt: time.Now(),
		votes:   make(map[interface{}]int),
		played:  make(chan error, 1),
	}
	r.media[m] = info
	return info.played, nil
}

// MediumPlayed removes the medium from the room.
func (r *Room) MediumPlayed(m medium.Medium) {
	r.l.Lock()
	defer r.l.Unlock()
	// inform that the medium was played
	if info := r.media[m]; info != nil {
		info.played <- nil
	}
	// remove medium
	delete(r.media, m)
}

// UserVotesMedium counts a vote that a user casts. The gravity is +1 for an
// upvote and -1 for a downvote. 0 resets the vote.
func (r *Room) UserVotesMedium(user interface{}, m medium.Medium, gravity int) error {
	r.l.Lock()
	defer r.l.Unlock()
	// get user and medium info
	if _, ok := r.users[user]; !ok {
		return ErrUserUnknown
	}
	mediumInfo, ok := r.media[m]
	if !ok {
		return ErrMediumUnknown
	}
	// apply vote
	mediumInfo.vote(user, gravity)
	return nil
}

// GetMediumScore returns the score of the medium and whether the medium exists.
func (r *Room) GetMediumScore(m medium.Medium) (int, bool) {
	r.l.RLock()
	defer r.l.RUnlock()
	// get medium info
	mediumInfo, ok := r.media[m]
	if !ok {
		return 0, false
	}
	return mediumInfo.score, true
}

// Queue returns the queue of media in the order they are supposed to be played.
func (r *Room) Queue() []medium.Medium {
	r.l.RLock()
	defer r.l.RUnlock()
	q := make(mediaQueue, 0, len(r.media))
	for m, info := range r.media {
		q = append(q, mediaItem{m, info})
	}
	sort.Sort(q)
	mq := make([]medium.Medium, len(q))
	for i, item := range q {
		mq[i] = item.m
	}
	return mq
}

type mediaItem struct {
	m    medium.Medium
	info *mediumInfo
}

type mediaQueue []mediaItem

func (q mediaQueue) Len() int {
	return len(q)
}

func (q mediaQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q mediaQueue) Less(i, j int) bool {
	// compare score
	if is, js := q[i].info.score, q[j].info.score; is != js {
		return is > js
	}
	// compare age
	return q[i].info.addedAt.Before(q[j].info.addedAt)
}

type userInfo struct{}

type mediumInfo struct {
	user    interface{}
	addedAt time.Time
	votes   map[interface{}]int
	score   int

	// sending nil if medium was played or ErrMediumUnknown if it was removed
	played chan error
}

func (m *mediumInfo) vote(user interface{}, gravity int) {
	gravity = clamp(gravity, -1, +1)
	m.score += gravity - m.votes[user]
	if gravity == 0 {
		delete(m.votes, user)
	} else {
		m.votes[user] = gravity
	}
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
