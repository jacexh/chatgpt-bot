package wechat

import (
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/officialaccount"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/sirupsen/logrus"
)

type Option struct {
	APPID          string `json:"app_id" yaml:"app_id"`
	APPSecret      string `json:"app_secret" yaml:"app_secret"`
	Token          string `json:"token" yaml:"token"`
	EncodingAESKey string `json:"encoding_aes_key" yaml:"encoding_aes_key"`
}

func NewWechatClient(opt Option) *officialaccount.OfficialAccount {
	wx := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &offConfig.Config{
		AppID:          opt.APPID,
		AppSecret:      opt.APPSecret,
		Token:          opt.Token,
		EncodingAESKey: opt.EncodingAESKey,
		Cache:          memory,
	}
	logrus.SetLevel(logrus.PanicLevel)
	return wx.GetOfficialAccount(cfg)
}
