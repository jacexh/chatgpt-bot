package infrastructure

import (
	"context"

	"github.com/jacexh/chatgpt-bot/internal/chat/domain"
	"github.com/sashabaranov/go-openai"
)

type chatgptService struct {
	client *openai.Client
}

var _ domain.ChatGTPService = (*chatgptService)(nil)

func NewChatGTPServer(client *openai.Client) domain.ChatGTPService {
	return &chatgptService{client: client}
}

func (gpt *chatgptService) Chat(ctx context.Context, chat *domain.Chat) (*domain.Conversation, error) {
	history := chat.PreviousConversations()
	current, err := chat.CurrentConversation()
	if err != nil {
		return nil, err
	}

	messages := make([]openai.ChatCompletionMessage, len(history)+1)
	for index, conv := range history {
		messages[index] = openai.ChatCompletionMessage{
			Content: conv.Prompt,
			Role:    openai.ChatMessageRoleUser,
		}
	}
	messages[len(history)] = openai.ChatCompletionMessage{
		Content: current.Answer,
		Role:    openai.ChatMessageRoleUser,
	}

	resp, err := gpt.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)
	if err != nil {
		return chat.Interrupt(err)
	}
	return chat.Reply(resp.Choices[0].Message.Content)
}
