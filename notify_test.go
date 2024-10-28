package notify

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNotify_Send(t *testing.T) {
	type fields struct {
		config *Config
	}
	type args struct {
		msg string
	}

	type Info struct {
		Subject string `json:"subject"`
		Content string `json:"content"`
	}

	type info struct {
		Subject string
		Content string
	}

	email := &Info{
		Subject: "Chainbot subscription successful",
		Content: `
			<div style="text-align: center; font-family: Arial, Helvetica, sans-serif;">
				<div style="text-align: center; max-width: 600px; margin: auto;">
					<img src="https://chainbot.io/images/logo.png" style="width: 180px; margin-top: 50px" />
					<h1 style="font-weight: bold; margin-top: 20px">Thanks for subscription!</h1>
					<p style="font-size: 16px; text-align: left;">Your account has successfully upgrade to pro member, if you have any question, please contact us in <a style="color: #5C00F3" href="https://discord.gg/RQF3KzsrDC">discord</a>.</p>

					<ul style="text-align: left; padding-left: 20px;">
						<li style="margin-bottom: 4px;">Subscription: 3 months</li>
						<li style="margin-bottom: 4px;">Type: Pro member</li>
						<li style="margin-bottom: 4px;">Amount received: 30 USDT</li>
						<li style="margin-bottom: 4px;">Transaction ID: 0x00123123</li>
						<li style="margin-bottom: 4px;">Valid until: 2023-10-19</li>
					</ul>

					<p style="font-size: 16px; text-align: left;">Click <a style="color: #5C00F3; text-align: left;" href="https://chainbot.io/my/subs" target="_blank">here</a> to view your dashboard and create new monitors, thanks for choosing Chainbot.</p>
				</div>
			</div>
		`,
	}

	// stringify email
	email_bytes, err := json.Marshal(email)
	if err != nil {
		t.Error(err)
	}

	email_string := string(email_bytes)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"test pushover notify",
			fields{config: &Config{
				Platform: Platform("pushover"),
				Token:    os.Getenv("PUSHOVER_TOKEN"),
				Channel:  os.Getenv("PUSHOVER_CHANNEL"),
			}},
			args{msg: "test case"},
		},
		{
			"test slack notify",
			fields{config: &Config{
				Platform: Platform("slack"),
				Token:    os.Getenv("SLACK_TOKEN"),
				Channel:  os.Getenv("SLACK_CHANNEL"),
			}},
			args{msg: "test case"},
		},
		{
			"test pagerduty severity is null",
			fields{config: &Config{
				Platform: Platform("pagerduty"),
				Token:    os.Getenv("PAGERDUTY_TOKEN"),
				Source:   "api-test",
				Severity: "",
			}},
			args{msg: "test pagerduty"},
		},
		{
			"test pagerduty severity is error",
			fields{config: &Config{
				Platform: Platform("pagerduty"),
				Token:    os.Getenv("PAGERDUTY_TOKEN"),
				Source:   "api-test",
				Severity: "error",
			}},
			args{msg: "test pagerduty is error"},
		},
		{
			"test discord notify",
			fields{
				config: &Config{
					Platform: PlatformDiscord,
					Token:    os.Getenv("DISCORD_TOKEN"),
					Channel:  os.Getenv("DISCORD_CHANNEL"),
				},
			},
			args{msg: "test case"},
		},
		{
			name: "test telegram notify",
			fields: fields{
				config: &Config{
					Platform: PlatformTelegram,
					Token:    os.Getenv("TELEGRAM_TOKEN"),
					Channel:  os.Getenv("TELEGRAM_CHANNEL"),
				},
			},
			args: args{
				msg: "test case",
			},
		},
		{
			"test dingtalk notify",
			fields{config: &Config{
				Platform: PlatformDingTalk,
				Token:    os.Getenv("DingTalk_TOKEN"),
				Channel:  os.Getenv("DingTalk_CHANNEL"),
			}},
			args{msg: "test case"},
		},
		// {
		// 	"test email notify",
		// 	fields{config: &Config{
		// 		Platform: PlatformEmail,
		// 		Token:    os.Getenv("Email_Token"),
		// 		User:     os.Getenv("Email_User"),
		// 		Password: os.Getenv("Email_Password"),
		// 		Host:     os.Getenv("Email_Host"),
		// 	}},
		// 	args{msg: "test case"},
		// },
		{
			"test ses notify",
			fields{config: &Config{
				Platform: PlatformEmail,
				Token:    os.Getenv("Email_To"),
				Sender:   os.Getenv("Email_SENDER"),
				Key:      os.Getenv("IAM_KEY"),
				Secret:   os.Getenv("IAM_SECRET"),
				Area:     os.Getenv("IAM_AREA"),
			}},
			args{msg: email_string},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Notify{
				config: tt.fields.config,
			}
			err := n.Send(tt.args.msg)
			if err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}
