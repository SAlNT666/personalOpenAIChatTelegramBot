package processing

import (
	"fmt"
	botApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

// sendResponse send a response to the user
func (s *Source) sendResponse(update *botApi.Update, response string) error {
	// OpenAI Chat may think that it should add his name to the beginning of the response
	msg := botApi.NewMessage(update.Message.From.ID, strings.TrimLeft(response, assistantName+": "))
	_, err := s.TelegramBot.Send(msg)
	if err != nil {
		return fmt.Errorf("error message not sent: %w", err)
	}

	return nil
}
