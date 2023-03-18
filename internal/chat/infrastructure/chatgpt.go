package infrastructure

import (
	"context"
	"net/http"
	"net/url"

	"github.com/jacexh/chatgpt-bot/internal/chat/domain"
	"github.com/sashabaranov/go-openai"
)

type (
	chatgptService struct {
		client *openai.Client
	}

	ChatGPTOption struct {
		AccessToken string `json:"access_token" yaml:"access_token"`
		Proxy       string `json:"proxy" yaml:"proxy"`
		BaseURL     string `json:"base_url" yaml:"base_url"`
	}
)

var _ domain.ChatGTPService = (*chatgptService)(nil)

func NewChatGTPServer(opt ChatGPTOption) domain.ChatGTPService {
	conf := openai.DefaultConfig(opt.AccessToken)
	if opt.BaseURL != "" {
		conf.BaseURL = opt.BaseURL
	}
	if opt.Proxy != "" {
		proxyURL, err := url.Parse(opt.Proxy)
		if err != nil {
			panic(err)
		}
		conf.HTTPClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}
	return &chatgptService{client: openai.NewClientWithConfig(conf)}
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
