package telegram

import (
	"errors"
	"fmt"

	tgbotapi "github.com/ChainbotAI/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/conc"
	tb "gopkg.in/telebot.v3"
)

const (
	ApiURL = "https://api.telegram.org/bot"
)

// https://core.telegram.org/bots/api#sendmessage
// https://github.com/go-telegram-bot-api/telegram-bot-api

type NotifyChannelType string

const (
	NotifyChannelTypeTgGroup   NotifyChannelType = "Group"
	NotifyChannelTypeTgChannel NotifyChannelType = "Channel"
	NotifyChannelTypeTgUser    NotifyChannelType = "User"
	NotifyChannelTypeTgBot     NotifyChannelType = "Bot"
)

// Options allows full configuration of the message sent to the Pushover API
type Options struct {
	Token       string `json:"token"`
	ChannelType NotifyChannelType
	Channel     int64  `json:"channel"`
	ChatName    string `json:"chat_name"`
	TopicId     int    `json:"topic_id"`
	// User may be either a user key or a group key.
	ChatIDs []int64 `json:"chat_ids"`

	TgBotReplyMarkup *tb.ReplyMarkup
}

type client struct {
	opt Options
	bot *tgbotapi.BotAPI
}

func New(opt Options) *client {
	api, err := tgbotapi.NewBotAPI(opt.Token)
	if err != nil {
		logrus.Errorf("create tgbot api err: %v, token: %s", err, opt.Token)
		return nil
	}

	return &client{opt: opt, bot: api}
}

type Resp struct {
	Status int      `json:"status"`
	Errors []string `json:"errors"`
}

func (c *client) Send(message string) error {
	if c.opt.Token == "" {
		return errors.New("missing token")
	}

	if message == "" {
		return errors.New("missing message")
	}

	if c.opt.ChannelType == NotifyChannelTypeTgBot {
		return c.sendTelegramBotNotify(message)
	} else {
		return c.sendTelegramNotify(message)
	}
}

func (c *client) sendTelegramBotNotify(message string) error {
	botToken := c.opt.Token
	bot, err := tb.NewBot(tb.Settings{
		Token: botToken,
	})
	if err != nil {
		logrus.Errorf("[TgBot] init tg bot err: %v", err)
		return err
	}
	var wg conc.WaitGroup
	for _, chatID := range c.opt.ChatIDs {
		chatIDObj := tb.ChatID(chatID)
		var opts []interface{}
		if c.opt.TgBotReplyMarkup != nil {
			opts = append(opts, c.opt.TgBotReplyMarkup)
		}

		wg.Go(func() {
			if _, err := bot.Send(chatIDObj, message, opts...); err != nil {
				logrus.Errorf("[TgBot] fail to send tg bot msg, err: %v", err)
			}
		})
	}
	wg.Wait()
	return nil
}

func (c *client) sendTelegramNotify(message string) error {
	var msg tgbotapi.MessageConfig
	if c.opt.Channel != 0 {
		if c.opt.TopicId != 0 {
			msg = tgbotapi.NewTopicMessage(c.opt.Channel, c.opt.TopicId, message)
		} else {
			msg = tgbotapi.NewMessage(c.opt.Channel, message)
		}
	} else {
		msg = tgbotapi.NewMessageToChannel(c.opt.ChatName, message)
	}

	sendRes, err := c.bot.Send(msg)
	if err != nil {
		return err
	}

	fmt.Println("sendRes:", sendRes)
	return nil
}
