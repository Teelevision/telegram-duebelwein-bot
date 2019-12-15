package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/Teelevision/telegram-duebelwein-bot/room"
	"github.com/gorilla/websocket"
)

// RoomProvider provides rooms by telegram chat id
type RoomProvider interface {
	Room(chatID int64) *room.Room
}

// Run starts the WebSocket api.
func Run(roomProvider RoomProvider, listenAddr string) {
	http.HandleFunc("/", server(roomProvider))
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { // TODO: remove in production
		return true
	},
}

func server(roomProvider RoomProvider) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		// overwrite close handler
		origCloseHandler := c.CloseHandler()
		closed := atomic.Value{}
		c.SetCloseHandler(func(code int, text string) error {
			closed.Store(true)
			if origCloseHandler != nil {
				return origCloseHandler(code, text)
			}
			return nil
		})

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("could not read websocket:", err)
				break
			}

			// handle next messages
			if len(message) > 5 && string(message[:5]) == "next " {
				chatID, err := strconv.Atoi(string(message[5:]))
				if err != nil {
					log.Println("could not parse chat id:", err)
					break
				}
				room := roomProvider.Room(int64(chatID))
				if room == nil {
					log.Println("room with chat id not found:", chatID)
					break
				}
				go func() {
					for {
						if isClosed, ok := closed.Load().(bool); ok && isClosed {
							return
						}
						if queue := room.Queue(); len(queue) > 0 {
							m := queue[0]
							msg := []byte(fmt.Sprintf("play %s %v", m.Provider(), m.ID()))
							err = c.WriteMessage(mt, msg)
							if err != nil {
								log.Println("could not write to websocket:", err)
								break
							}
							room.MediumPlayed(m)
							break
						}
						time.Sleep(time.Second)
					}
				}()
			} else if string(message) == "keep-alive" {
				err = c.WriteMessage(mt, message)
				if err != nil {
					log.Println("could not write to websocket:", err)
					break
				}
			}
		}
	}
}
