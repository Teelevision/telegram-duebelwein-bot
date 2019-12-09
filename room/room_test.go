package room_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/Teelevision/telegram-duebelwein-bot/medium"
	. "github.com/Teelevision/telegram-duebelwein-bot/room"
)

type testRoom struct {
	*Room
}

func (r testRoom) UserQueuesMedium(user interface{}, m medium.Medium) {
	if err := r.Room.UserQueuesMedium(user, m); err != nil {
		log.Fatalf("did not expect error when adding %q, got %q", m.ID(), err)
	}
}

func (r testRoom) UserVotesMedium(user interface{}, m medium.Medium, gravity int) {
	if err := r.Room.UserVotesMedium(user, m, gravity); err != nil {
		log.Fatalf("did not expect error when voting %q, got %q", m.ID(), err)
	}
}

var (
	songBySerj             = &someMedium{"song by serj"}
	failCompilation        = &someMedium{"fail compilation"}
	anotherFailCompilation = &someMedium{"another fail compilation"}
	nightWitchesBySabaton  = &someMedium{"night witches by sabaton"}
	cowsCowsCows           = &someMedium{"cows cows cows"}
	wodkaByDaTweekaz       = &someMedium{"wodka by da tweekaz"}
)

func ExampleRoom_Queue() {
	room := testRoom{New()}
	room.UserJoins("Marius")
	room.UserQueuesMedium("Marius", failCompilation)
	room.MediumPlayed(failCompilation)
	room.UserJoins("Max")
	room.UserJoins("Andy")
	room.UserQueuesMedium("Andy", songBySerj)
	room.UserQueuesMedium("Marius", anotherFailCompilation)
	room.UserVotesMedium("Marius", songBySerj, +1)
	room.UserQueuesMedium("Max", nightWitchesBySabaton)
	room.UserVotesMedium("Marius", nightWitchesBySabaton, +1)
	room.UserVotesMedium("Andy", songBySerj, +1)
	room.UserQueuesMedium("Andy", cowsCowsCows)
	room.UserVotesMedium("Andy", nightWitchesBySabaton, -1)
	room.UserLeaves("Andy") // is physically kicked for adding cows cows cows
	room.UserQueuesMedium("Marius", wodkaByDaTweekaz)
	for i, medium := range room.Queue() {
		score, _ := room.GetMediumScore(medium)
		fmt.Printf("#%d %s (Score: %d)\n", i+1, medium.ID(), score)
	}

	// The song by serj has the most upvotes, but it was removed because Andy
	// was kicked. Night witches by sabaton has one upvote which is counted, the
	// downvote by andy was removed when he left, giving it a total of +1 and
	// thus the next position in the queue. Both other songs have not votes and
	// therefore are played in the order they were added.

	// Output:
	// #1 night witches by sabaton (Score: 1)
	// #2 another fail compilation (Score: 0)
	// #3 wodka by da tweekaz (Score: 0)
}

func TestRoom_UserQueuesMedium(t *testing.T) {
	t.Run("does not allow duplicates", func(t *testing.T) {
		room := New()
		room.UserJoins(1)
		room.UserJoins(2)
		if err := room.UserQueuesMedium(1, &someMedium{"dog video"}); err != nil {
			t.Fatalf("did not expect error when adding dog video the first time, got %q", err)
		}
		if err := room.UserQueuesMedium(2, &someMedium{"dog video"}); err != ErrMediumAlreadyExists {
			t.Fatalf("did expect error %q when adding dog video a second time, got %q", ErrMediumAlreadyExists, err)
		}
	})
}

type someProvider struct{}

func (p someProvider) String() string {
	return "foobar"
}

type someMedium struct {
	string
}

func (m *someMedium) Provider() medium.Provider {
	return someProvider{}
}

func (m *someMedium) ID() interface{} {
	return m.string
}
