package pushover

import (
	"errors"
	"github.com/imroc/req"
)

const (
	ApiURL = "https://api.pushover.net/1/messages.json"
)

// Options allows full configuration of the message sent to the Pushover API
// https://pushover.net/api#messages
type Options struct {
	Token string `json:"token"`
	// User may be either a user key or a group key.
	User     string  `json:"user"`
	Message  string  `json:"message"`
	Priority int     `json:"priority"`
	Retry    float64 `json:"retry"`
	Expire   float64 `json:"expire"`
}

type client struct {
	opt Options
}

func New(opt Options) *client {
	return &client{opt: opt}
}

type Resp struct {
	Status int      `json:"status"`
	Errors []string `json:"errors"`
}

func (c *client) Send(message string) error {
	if c.opt.Token == "" {
		return errors.New("missing token")
	}
	if c.opt.User == "" {
		return errors.New("missing user")
	}
	if message == "" {
		return errors.New("missing message")
	}
	c.opt.Message = message
	resp, err := req.Post(ApiURL, req.BodyJSON(c.opt))
	if err != nil {
		return nil
	}
	r := &Resp{}
	err = resp.ToJSON(r)
	if err != nil {
		return nil
	}
	if r.Status != 1 {
		return errors.New(r.Errors[0])
	}
	return nil
}
