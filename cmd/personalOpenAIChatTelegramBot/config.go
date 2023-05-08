package main

type Config struct {
	BotToken   string  `json:"bot_token"`
	ApiKey     string  `json:"api_key"`
	AllowedIds []int64 `json:"allowed_ids"`
}
