package main

import (
	"context"
	"encoding/json"
	"fmt"
	botApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
	"log"
	"os"
	"personalOpenAIChatTelegramBot/internal/processing"
)

const ConfigFile = "config.json"

// main is the entry point of the application
func main() {
	config, err := downloadConfig()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Initialize the Telegram Bot API
	telegramBot, err := botApi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Create an OpenAI client
	openaiClient := newOpenAIClient(config.ApiKey)

	// Create a new source object that handles the processing of incoming messages
	source := processing.NewSource(telegramBot, openaiClient)

	ctx := context.Background()

	telegramBot.Debug = true

	log.Printf("Authorized on account %s", telegramBot.Self.UserName)

	// Create a new update object with an update ID of 0 and a timeout of 60 seconds
	u := botApi.NewUpdate(0)
	u.Timeout = 60

	updates := telegramBot.GetUpdatesChan(u)

	// Create a new error channel to catch and log errors
	errChan := make(chan error)
	go catchError(errChan)

	// Create a map to store incoming messages for each user
	messagesChannels := make(map[int64][]botApi.Update, len(config.AllowedIds))
	for _, k := range config.AllowedIds {
		messagesChannels[k] = []botApi.Update{}
	}

	// Create a map to store the context for each user
	contextMap := make(map[int64]*[]string, len(config.AllowedIds))
	for _, k := range config.AllowedIds {
		contextMap[k] = &[]string{}
	}

	// Start listening for updates from the Telegram Bot API
	for update := range updates {
		chatId := update.Message.From.ID
		// Check if the message is from an authorized user
		if update.Message != nil && isUserInGroup(chatId, &config.AllowedIds) {
			// Add the message to the user's messages channel
			messagesChannels[chatId] = append(messagesChannels[chatId], update)
			// Process the request asynchronously
			go source.ProcessRequest(ctx, messagesChannels[chatId][0], contextMap[chatId], errChan)
			// Remove the message from the messages channel
			messagesChannels[chatId] = messagesChannels[chatId][1:]
		}
	}
}

// downloadConfig downloads the configuration from the config file
func downloadConfig() (Config, error) {
	configData, err := os.ReadFile(ConfigFile)
	if err != nil {
		return Config{}, fmt.Errorf("error read config: %w", err)
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return Config{}, fmt.Errorf("error unmashal data: %w", err)
	}

	fmt.Println(config.ApiKey)

	return config, nil
}

// newOpenAIClient creates a new instance of the OpenAI API client using the API key
func newOpenAIClient(apiKey string) *openai.Client {
	return openai.NewClient(apiKey)
}

// catchError listens for errors on the channel and logs them as they occur
func catchError(errChan <-chan error) {
	for errMsg := range errChan {
		log.Println(errMsg)
	}
}

// isUserInGroup checks whether a given user ID is in the list of IDs
func isUserInGroup(id int64, group *[]int64) bool {
	for _, allowed := range *group {
		if id == allowed {
			return true
		}
	}
	return false
}
