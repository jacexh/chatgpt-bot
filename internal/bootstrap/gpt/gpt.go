package gpt

import (
	"net/http"
	"net/url"

	"github.com/sashabaranov/go-openai"
)

type Option struct {
	AccessToken string `json:"access_token" yaml:"access_token"`
	Proxy       string `json:"proxy" yaml:"proxy"`
	BaseURL     string `json:"base_url" yaml:"base_url"`
}

func NewChatGPT(opt Option) *openai.Client {
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
	return openai.NewClientWithConfig(conf)
}
