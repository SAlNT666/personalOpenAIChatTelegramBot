package processing

import (
	"context"
	"fmt"
	botApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
	"strings"
)

type OpenAIClient interface {
	CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (response openai.ChatCompletionResponse, err error)
}

// MemorizedMessages number of messages stored in the context
const MemorizedMessages = 6
const userName = "user"
const assistantName = "AI assistant"

type Source struct {
	TelegramBot  *botApi.BotAPI
	OpenaiClient *openai.Client
}

// NewSource constructor
func NewSource(telegramBot *botApi.BotAPI, openaiClient *openai.Client) *Source {
	return &Source{TelegramBot: telegramBot, OpenaiClient: openaiClient}
}

// ProcessRequest request processing logic
func (s *Source) ProcessRequest(ctx context.Context, update botApi.Update, chatContext *[]string, errChan chan<- error) {
	requestText := s.getContext(&update, update.Message.Text, chatContext)

	response, err := s.getResponse(ctx, requestText)
	if err != nil {
		errChan <- fmt.Errorf("request was not processed: %w", err)
	}

	err = s.sendResponse(&update, response)
	if err != nil {
		errChan <- fmt.Errorf("response was not sent: %w", err)
	}
	addMessage(assistantName, response, chatContext)
}

// addMessage add a message to the chat context
func addMessage(author string, text string, chatContext *[]string) {
	message := fmt.Sprintf("%s: %s", author, text)

	// Append the message to the chat context and remove any messages over the MemorizedMessages limit
	*chatContext = append(*chatContext, message)
	if len(*chatContext) > MemorizedMessages {
		*chatContext = (*chatContext)[len(*chatContext)-MemorizedMessages:]
	}
}

// getContext get the chat context for the request
func (s *Source) getContext(update *botApi.Update, userMessage string, chatContext *[]string) string {
	if update.Message.ReplyToMessage == nil {
		// If this message is not reply, reset the chat context to an empty slice and return the user message
		*chatContext = []string{}
		return userMessage
	}

	addMessage(userName, userMessage, chatContext)
	userMessage = assembleContext(chatContext)

	return userMessage
}

// getResponse get a response from OpenAI for the request text
func (s *Source) getResponse(ctx context.Context, requestText string) (string, error) {
	response, err := s.generateResponse(ctx, requestText)
	if err != nil {
		return "OpenAi Chat did not respond.", fmt.Errorf("error: response was  not generated: %w", err)
	}

	return response, nil
}

// generateResponse generate a ChatCompletionRequest for the request text
func (s *Source) generateResponse(ctx context.Context, request string) (string, error) {
	requestStruct := newRequestStruct(request)
	resp, err := s.OpenaiClient.CreateChatCompletion(ctx, requestStruct)
	return resp.Choices[0].Message.Content, err
}

// newRequestStruct create a ChatCompletionRequest struct for the request text
func newRequestStruct(request string) openai.ChatCompletionRequest {
	requestStruct := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: request,
			},
		},
	}
	return requestStruct
}

// assembleContext assemble the context string from the chat context slice
func assembleContext(chatContext *[]string) string {
	return strings.Join(*chatContext, "\n\n")
}
