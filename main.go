package main

import (
	"github.com/Teelevision/telegram-duebelwein-bot/telegram"
	env "github.com/caarlos0/env/v6"
)

type config struct {
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	bot, err := telegram.NewBot(cfg.TelegramBotToken)
	if err != nil {
		panic(err)
	}
	bot.Start()
}
