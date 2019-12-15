package main

import (
	"github.com/Teelevision/telegram-duebelwein-bot/api"
	"github.com/Teelevision/telegram-duebelwein-bot/telegram"
	env "github.com/caarlos0/env/v6"
)

type config struct {
	TelegramBotToken  string `env:"TELEGRAM_BOT_TOKEN"`
	PlayerURLTemplate string `env:"PLAYER_URL_TEMPLATE"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	// start bot
	bot, err := telegram.NewBot(cfg.TelegramBotToken, cfg.PlayerURLTemplate)
	if err != nil {
		panic(err)
	}
	go bot.Start()

	// start api
	go api.Run(bot)

	select {} // keep running
}
