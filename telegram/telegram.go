package telegram

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Teelevision/telegram-duebelwein-bot/medium"
	"github.com/Teelevision/telegram-duebelwein-bot/room"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Bot is a Dübelwein Telegram bot.
type Bot struct {
	telegram *tb.Bot
	sync.RWMutex
	chats map[int64]*chat
}

type chat struct {
	*room.Room
	sync.RWMutex
	users map[int]*user
	media map[medium.Medium]*mediumContext
}

type user struct{}

type mediumContext struct {
	originalMessage *tb.Message
	cleanUp         func(why string)
}

// NewBot returns a new bot. It is not started, yet.
func NewBot(token string) (*Bot, error) {
	tbBot, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}
	return &Bot{
		telegram: tbBot,
		chats:    make(map[int64]*chat),
	}, nil
}

// Start starts the bot.
func (b *Bot) Start() {
	b.telegram.Handle(tb.OnAddedToGroup, func(msg *tb.Message) {
		if !msg.FromGroup() {
			return
		}
		b.seeChat(msg.Chat.ID)
		b.telegram.Send(msg.Chat, "🔥 Dübelweinbot is in da house! ☠️")
	})

	b.telegram.Handle(tb.OnUserJoined, func(msg *tb.Message) {
		if !msg.FromGroup() {
			return
		}
		b.seeUser(msg.Chat.ID, msg.Sender.ID)
	})

	b.telegram.Handle(tb.OnUserLeft, func(msg *tb.Message) {
		if !msg.FromGroup() {
			return
		}
		chat, user := b.seeUser(msg.Chat.ID, msg.Sender.ID)
		chat.UserLeaves(user)
		delete(chat.users, msg.Sender.ID)
	})

	// TODO: just for testing
	b.telegram.Handle("/play", func(msg *tb.Message) {
		if !msg.FromGroup() {
			return
		}
		chat := b.seeChat(msg.Chat.ID)
		queue := chat.Queue()
		if len(queue) > 0 {
			m := queue[0]
			chat.MediumPlayed(m)
			chat.RLock()
			mediumCtx, ok := chat.media[m]
			chat.RUnlock()
			if ok {
				b.telegram.Send(msg.Chat, fmt.Sprintf("%s (%s)", m.ID(), m.Provider()))
				mediumCtx.cleanUp("played")
			}
		}
	})

	b.telegram.Handle(tb.OnText, func(msg *tb.Message) {
		if !msg.FromGroup() {
			return
		}
		chat, user := b.seeUser(msg.Chat.ID, msg.Sender.ID)

		// try to get the medium, if one was sent
		url := getFirstURL(msg)
		m, err := medium.New(url)
		if err != nil {
			// reply that no medium could be found and abort
			b.telegram.Send(msg.Chat, "Wat?!", tb.Silent, &tb.SendOptions{
				ReplyTo: msg,
			})
			log.Printf("could not load medium from %q: %s", url, err)
			return
		}

		// add the medium to the room
		chat.Lock() // lock until clean up func is created
		defer chat.Unlock()
		chat.UserQueuesMedium(user, m)

		// show vote buttons
		mID := strconv.FormatInt(rand.Int63(), 36)
		upvote := tb.InlineButton{Unique: "upvote" + mID, Text: "❤️"}
		resetvote := tb.InlineButton{Unique: "resetvote" + mID, Text: "🤷"}
		downvote := tb.InlineButton{Unique: "downvote" + mID, Text: "💩"}
		sendOpt := &tb.SendOptions{
			ReplyTo: msg,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: [][]tb.InlineButton{{downvote, resetvote, upvote}},
			},
		}
		voteMsg, _ := b.telegram.Send(msg.Chat, "Queued (score: 0)", sendOpt)

		// vote logic
		vote := func(c *tb.Callback, gravity int) {
			chat, user := b.seeUser(msg.Chat.ID, c.Sender.ID)
			_ = chat.UserVotesMedium(user, m, gravity)
			score, _ := chat.GetMediumScore(m)
			b.telegram.Respond(c, &tb.CallbackResponse{Text: "Voted!"})
			b.telegram.Edit(voteMsg, fmt.Sprintf("Queued (score: %d)", score), sendOpt)
		}
		b.telegram.Handle(&upvote, func(c *tb.Callback) { vote(c, +1) })
		b.telegram.Handle(&resetvote, func(c *tb.Callback) { vote(c, 0) })
		b.telegram.Handle(&downvote, func(c *tb.Callback) { vote(c, -1) })

		// create clean up func
		chat.media[m] = &mediumContext{
			originalMessage: msg,
			cleanUp: func(why string) {
				chat.Lock()
				defer chat.Unlock()
				sendOpt.ReplyMarkup = nil
				b.telegram.Edit(voteMsg, why, sendOpt)
				// release resources so that the gc can do the rest
				b.telegram.Handle(&upvote, nil)
				b.telegram.Handle(&resetvote, nil)
				b.telegram.Handle(&downvote, nil)
				delete(chat.media, m)
			},
		}
	})

	b.telegram.Start()
}

func (b *Bot) seeChat(chatID int64) *chat {
	b.Lock()
	defer b.Unlock()
	if chat, ok := b.chats[chatID]; ok {
		return chat
	}
	b.chats[chatID] = &chat{
		Room:  room.New(),
		users: make(map[int]*user),
		media: make(map[medium.Medium]*mediumContext),
	}
	return b.chats[chatID]
}

func (b *Bot) seeUser(chatID int64, userID int) (*chat, *user) {
	chat := b.seeChat(chatID)
	chat.Lock()
	defer chat.Unlock()
	if user, ok := chat.users[userID]; ok {
		return chat, user
	}
	user := new(user)
	chat.users[userID] = user
	chat.UserJoins(user)
	return chat, user
}

func getFirstURL(m *tb.Message) string {
	if strings.HasPrefix(m.Text, "https://") || strings.HasPrefix(m.Text, "http://") {
		return m.Text
	}
	for _, entity := range m.Entities {
		if entity.URL != "" {
			return entity.URL
		}
	}
	return "http://" + m.Text
}
