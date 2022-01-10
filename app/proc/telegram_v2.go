package proc

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/html"

	log "github.com/go-pkgz/lgr"
	"github.com/microcosm-cc/bluemonday"
	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/umputun/feed-master/app/feed"
)

// TelegramClientV2 client
type TelegramClientV2 struct {
	Bot *tb.Bot
}

// NewTelegramClient init telegram client
func NewTelegramV2Client(token, apiURL string, timeout time.Duration) (*TelegramClientV2, error) {
	if timeout == 0 {
		timeout = time.Second * 60
	}

	if token == "" {
		return nil, errors.New("empty telegram token")
	}

	bot, err := tb.NewBot(tb.Settings{
		URL:    apiURL,
		Token:  token,
		Poller: &tb.LongPoller{Timeout: timeout},
	})

	if err != nil {
		return nil, err
	}

	result := TelegramClientV2{
		Bot: bot,
	}
	return &result, err
}

func (client TelegramClientV2) sendText(channelID string, item feed.Item) (*tb.Message, error) {
	message, err := client.Bot.Send(
		recipient{chatID: channelID},
		client.getMessageHTML(item, true),
		tb.ModeHTML,
		tb.NoPreview,
	)

	return message, err
}

// https://core.telegram.org/bots/api#html-style
func (client TelegramClientV2) tagLinkOnlySupport(htmlText string) string {
	p := bluemonday.NewPolicy()
	p.AllowAttrs("href").OnElements("a")
	return html.UnescapeString(p.Sanitize(htmlText))
}

// getMessageHTML generates HTML message from provided feed.Item
func (client TelegramClientV2) getMessageHTML(item feed.Item, withMp3Link bool) string {
	description := string(item.Description)

	description = strings.TrimPrefix(description, "<![CDATA[")
	description = strings.TrimSuffix(description, "]]>")

	// apparently bluemonday doesn't remove escaped HTML tags
	description = client.tagLinkOnlySupport(html.UnescapeString(description))
	description = strings.TrimSpace(description)

	messageHTML := description

	title := strings.TrimSpace(item.Title)
	if title != "" {
		switch {
		case item.Link == "":
			messageHTML = fmt.Sprintf("%s\n\n", title) + messageHTML
		case item.Link != "":
			messageHTML = fmt.Sprintf("<a href=\"%s\">%s</a>\n\n", item.Link, title) + messageHTML
		}
	}

	if withMp3Link {
		messageHTML += fmt.Sprintf("\n\n%s", item.Enclosure.URL)
	}

	return messageHTML
}

type recipient struct {
	chatID string
}

func (r recipient) Recipient() string {
	if !strings.HasPrefix(r.chatID, "@") {
		return "@" + r.chatID
	}

	return r.chatID
}

const (
	commandHello  = "/hello"
	commandHelp   = "/help"
	commandStop   = "/stop"
	commandStart  = "/start"
	commandImport = "/import"
	commandExport = "/export"

	msgStart = `Welcome!
Use commands:
/import - for load OPML-file
/stop - for stop send updates
`

	msgHelp = `Use commands:
/import - for load OPML-file
/stop - for stop send updates
`
)

func (client TelegramClientV2) Start() {

	var (
		menu      = &tb.ReplyMarkup{ResizeReplyKeyboard: true}
		btnImport = menu.Text(commandImport)
		btnExport = menu.Text(commandExport)
		btnStop   = menu.Text(commandStop)
		btnHelp   = menu.Text(commandHelp)
	)

	menu.Reply(
		menu.Row(btnHelp),
		menu.Row(btnImport),
		menu.Row(btnExport),
		menu.Row(btnStop),
	)

	client.Bot.Handle("/hello", func(m *tb.Message) {
		client.Bot.Send(m.Sender, "Hello World!", menu)
	})

	// Command: /start
	client.Bot.Handle(commandStart, func(m *tb.Message) {
		if !m.Private() {
			// TODO: send "Bot works in private mode only"
			return
		}

		client.Bot.Send(m.Sender, msgStart, menu)

		logCommand(commandStart, m.Chat.ID, m.Payload)
	})

	// Command: /help
	client.Bot.Handle(commandHelp, func(m *tb.Message) {
		if !m.Private() {
			return
		}

		client.Bot.Send(m.Sender, msgHelp, menu)

		logCommand(commandHelp, m.Chat.ID, m.Payload)
	})

	// Command: /stop
	client.Bot.Handle(commandStop, func(m *tb.Message) {
		if !m.Private() {
			return
		}

		logCommand(commandStop, m.Chat.ID, m.Payload)
	})

	// Command: /import
	client.Bot.Handle(commandImport, func(m *tb.Message) {
		if !m.Private() {
			return
		}

		// client.Bot.Send(m.Sender, "")
		logCommand(commandImport, m.Chat.ID, m.Payload)
	})

	// Command: /export
	client.Bot.Handle(commandExport, func(m *tb.Message) {
		if !m.Private() {
			return
		}

		logCommand(commandExport, m.Chat.ID, m.Payload)
	})

	client.Bot.Handle(tb.OnDocument, func(m *tb.Message) {
		log.Printf("[DEBUG] telegram message receive document: '%s', with size: '%d'", m.Document.FileName, m.Document.FileSize)
	})

	client.Bot.Handle(tb.OnText, func(m *tb.Message) {
		log.Printf("[DEBUG] telegram receive unknown text: \n%s", m.Text)
	})

	log.Print("[INFO] telegram bot started")
	client.Bot.Start()
}

func logCommand(command string, chatID int64, payload string) {
	log.Printf("[DEBUG] telegram receive command: '%s' in chat: '%d'\n%s", command, chatID, payload)
}
