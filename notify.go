package notify

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ChainbotAI/go-notify/dingtalk"
	"github.com/ChainbotAI/go-notify/discord"
	"github.com/ChainbotAI/go-notify/email"
	"github.com/ChainbotAI/go-notify/lark"
	"github.com/ChainbotAI/go-notify/pagerduty"
	"github.com/ChainbotAI/go-notify/pushover"
	"github.com/ChainbotAI/go-notify/ses"
	"github.com/ChainbotAI/go-notify/slack"
	"github.com/ChainbotAI/go-notify/telegram"
	"github.com/sirupsen/logrus"
	tb "gopkg.in/telebot.v3"
)

type Platform string

const (
	PlatformSlack     Platform = "Slack"
	PlatformPushover           = "Pushover"
	PlatformPagerduty          = "Pagerduty"
	PlatformDiscord            = "Discord"
	PlatformTelegram           = "Telegram"
	PlatformDingTalk           = "DingTalk"
	PlatformEmail              = "Email"
	PlatformSes                = "AwsEmail"
	PlatformLark               = "Lark"
	PlatformArgus              = "Argus"
)

type ChannelType string

const (
	ChannelTypeTgBot ChannelType = "TgBot"
)

type Notify struct {
	config *Config
}

type Config struct {
	Platform Platform

	ToEmail string
	Key     string
	Secret  string
	Area    string
	Sender  string

	Token       string
	Channel     string
	ChannelType ChannelType
	Source      string
	Severity    string
	User        string
	Password    string
	Host        string
	Priority    int
	Others      map[string]string
	ChatIDs     []int64
}

func NewNotify(config *Config) *Notify {
	return &Notify{
		config: config,
	}
}

func (n *Notify) Send(msg string) error {
	switch n.config.Platform {
	case PlatformPushover:
		return n.sendPushOverNotify(msg)
	case PlatformSlack:
		return n.sendSlackNotify(msg)
	case PlatformPagerduty:
		return n.sendPagerdutyNotify(msg)
	case PlatformDiscord:
		return n.sendDiscordNotify(msg)
	case PlatformTelegram:
		return n.sendTelegramNotify(msg)
	case PlatformDingTalk:
		return n.sendDingTalkNotify(msg)
	case PlatformEmail:
		// change to ses
		return n.sendSesNotify(msg)
	case PlatformSes:
		return n.sendSesNotify(msg)
	case PlatformLark:
		return n.sendLarkNotify(msg)
	default:
		return errors.New("not supported notify platform")
	}
	return nil
}

func (n *Notify) sendPushOverNotify(msg string) error {
	options := pushover.Options{
		Token:    n.config.Token,
		User:     n.config.Channel,
		Priority: n.config.Priority,
	}
	if retry, exist := n.config.Others["retryInterval"]; exist {
		options.Retry, _ = strconv.ParseFloat(retry, 64)
	}
	if expire, exist := n.config.Others["retryExpire"]; exist {
		options.Expire, _ = strconv.ParseFloat(expire, 64)
	}
	app := pushover.New(options)
	err := app.Send(msg)
	return err
}

func (n *Notify) sendSlackNotify(msg string) error {
	app := slack.New(slack.Options{
		Token:   n.config.Token,
		Channel: n.config.Channel,
	})
	err := app.Send(msg)
	return err
}

func (n *Notify) sendPagerdutyNotify(msg string) error {
	app := pagerduty.New(pagerduty.Options{
		Token:    n.config.Token,
		Source:   n.config.Source,
		Severity: n.config.Severity,
	})
	err := app.Send(msg)
	return err
}

func (n *Notify) sendDiscordNotify(msg string) error {
	app := discord.New(discord.Options{
		Token:   n.config.Token,
		Channel: n.config.Channel,
	})
	err := app.Send(msg)
	return err
}

func (n *Notify) sendTelegramNotify(msg string) error {
	if n.config.ChannelType == ChannelTypeTgBot {
		return n.sendTelegramBotNotify(msg)
	}

	var _channel int64
	var _chatName string
	var _topicId int64
	if strings.Contains(n.config.Channel, "@") {
		_channel = 0
		_chatName = n.config.Channel
	} else {
		_channel, _ = strconv.ParseInt(n.config.Channel, 10, 64)
		_chatName = ""
	}

	_topicId, _ = strconv.ParseInt(n.config.Others["topic_id"], 10, 64)
	// 新版本中用户输入@开头的 name 时，会查询对应的 chat_id 并存入 others["chat_id"]
	if n.config.Others["chat_id"] != "" {
		chatId, _ := strconv.ParseInt(n.config.Others["chat_id"], 10, 64)
		if chatId != 0 {
			_channel = chatId
		}
	}

	app := telegram.New(telegram.Options{
		Token:    n.config.Token,
		Channel:  _channel,
		ChatName: _chatName,
		TopicId:  int(_topicId),
	})
	err := app.Send(msg)
	return err
}

func (n *Notify) sendTelegramBotNotify(msg string) error {
	botToken := n.config.Token
	bot, _ := tb.NewBot(tb.Settings{
		Token:   botToken,
		Offline: true,
	})
	for _, chatID := range n.config.ChatIDs {
		if _, err := bot.Send(tb.ChatID(chatID), msg); err != nil {
			logrus.Errorf("fail to send tg bot msg, err: %v", err)
		}
	}
	return nil
}

func (n *Notify) sendDingTalkNotify(msg string) error {
	app := dingtalk.New(dingtalk.Options{
		WebhookUrl: n.config.Channel,
		Secret:     n.config.Token,
	})
	err := app.Send(msg)
	return err
}

func (n *Notify) sendEmailNotify(msg string) error {
	app := email.New(email.Options{
		ToEmail:  n.config.Token,
		User:     n.config.User,
		Password: n.config.Password,
		Host:     n.config.Host,
	})
	err := app.Send(msg)
	return err
}

func (n *Notify) sendSesNotify(msg string) error {
	app := ses.New(ses.Options{
		ToEmail: n.config.Token,
		Key:     n.config.Key,
		Secret:  n.config.Secret,
		Area:    n.config.Area,
		Sender:  n.config.Sender,
	})
	err := app.Send(msg)
	return err
}

func (n *Notify) sendLarkNotify(msg string) error {
	app := lark.New(lark.Options{
		Token: n.config.Token,
	})
	err := app.Send(msg)
	return err
}
