package notify

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ChainbotAI/base-gokit/datatypes/types"
	"github.com/ChainbotAI/go-notify/dingtalk"
	"github.com/ChainbotAI/go-notify/discord"
	"github.com/ChainbotAI/go-notify/email"
	"github.com/ChainbotAI/go-notify/lark"
	"github.com/ChainbotAI/go-notify/pagerduty"
	"github.com/ChainbotAI/go-notify/pushover"
	"github.com/ChainbotAI/go-notify/ses"
	"github.com/ChainbotAI/go-notify/slack"
	"github.com/ChainbotAI/go-notify/telegram"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/conc"
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

type NotifyChannelType string

const (
	NotifyChannelTypeTgGroup   NotifyChannelType = "Group"
	NotifyChannelTypeTgChannel NotifyChannelType = "Channel"
	NotifyChannelTypeTgUser    NotifyChannelType = "User"
	NotifyChannelTypeTgBot     NotifyChannelType = "Bot"
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
	ChannelType NotifyChannelType
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

type ExtendMsg struct {
	Intent *types.TxIntent
}

func (n *Notify) Send(msg string, extendMsg *ExtendMsg) error {
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
		return n.sendTelegramNotify(msg, extendMsg)
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

func (n *Notify) sendTelegramNotify(msg string, extendMsg *ExtendMsg) error {
	if n.config.ChannelType == NotifyChannelTypeTgBot {
		return n.sendTelegramBotNotify(msg, extendMsg)
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
	if app == nil {
		return errors.New("create telegram client err")
	}
	err := app.Send(msg)
	return err
}

func LinkToPage(botId, path string, data interface{}) string {
	jsonBytes, _ := json.Marshal(map[string]interface{}{
		"p": path,
		"d": data,
	})
	base64Bytes := base64.StdEncoding.EncodeToString(jsonBytes)
	endpoint := fmt.Sprintf(
		"%s?startapp=%s",
		fmt.Sprintf("https://t.me/%s/swap", botId),
		base64Bytes,
	)
	return endpoint
}

func makeTradeMenu(swapIntent *types.SwapIntent) *tb.ReplyMarkup {
	tradeSelector := &tb.ReplyMarkup{
		Selective: true,
	}

	var side string
	var tokenAddr string
	if swapIntent.BuyToken == common.HexToAddress("0x") {
		side = "SELL"
		tokenAddr = swapIntent.SellToken.Hex()
	} else if swapIntent.SellToken == common.HexToAddress("0x") {
		side = "BUY"
		tokenAddr = swapIntent.BuyToken.Hex()
	} else {
		return nil
	}

	data := map[string]string{
		"s": side,
		"t": tokenAddr,
	}

	tradeSelector.Inline(
		tradeSelector.Row(
			tradeSelector.URL("Buy", LinkToPage("official_swapbot", "/trade", data)),
		),
	)
	return tradeSelector
}

func (n *Notify) sendTelegramBotNotify(msg string, extendMsg *ExtendMsg) error {
	botToken := n.config.Token
	bot, err := tb.NewBot(tb.Settings{
		Token: botToken,
	})
	if err != nil {
		logrus.Errorf("[TgBot] init tg bot err: %v", err)
		return err
	}
	var wg conc.WaitGroup
	for _, chatID := range n.config.ChatIDs {
		chatIDObj := tb.ChatID(chatID)
		var menu *tb.ReplyMarkup
		if extendMsg != nil && extendMsg.Intent != nil && len(extendMsg.Intent.SwapIntents) == 1 {
			// 根据传入的intent去构建不同的menu
			menu = makeTradeMenu(extendMsg.Intent.SwapIntents[0])
		}

		wg.Go(func() {
			if _, err := bot.Send(chatIDObj, msg, menu); err != nil {
				logrus.Errorf("[TgBot] fail to send tg bot msg, err: %v", err)
			}
		})
	}
	wg.Wait()
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
