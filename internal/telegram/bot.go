package telegram

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"k8s.io/klog"
)

type Update = tgbotapi.Update
type CommandHandler = func(update Update) error

type Bot struct {
  token string
  chatid string
  api *tgbotapi.BotAPI
  handlers map[string]CommandHandler
}

type BotOptions struct {
  Token string
  ChatId string
  Debug bool
}

func New(opts BotOptions) (Bot, error) {
  bot, err := tgbotapi.NewBotAPI(opts.Token)
  if err != nil {
    return Bot{}, err
  }

  if opts.Debug == true {
    bot.Debug = true
  }

  return Bot{token:opts.Token,chatid:opts.ChatId,api:bot,handlers:make(map[string]CommandHandler)}, nil
}

func (b *Bot) Start() error {
  updateconfig := tgbotapi.NewUpdate(0)
  updateconfig.Timeout = 30

  updates := b.api.GetUpdatesChan(updateconfig)

  for update := range updates {
    if (update.Message == nil) {
      continue
    }

    inputCommand := strings.Split(update.Message.Text, " ")[0]

    for command, handler := range b.handlers {
      if inputCommand == command {
        if err := handler(update); err != nil {
          klog.Errorln("Handler ended with an error", err)
          b.Send(err.Error())
        }
      }
    }
  }

  return nil
}

func (b *Bot) Send(text string) error {
  parsedChatId, err := strconv.ParseInt(b.chatid, 10, 64)
  if err != nil {
    return err
  }

  msg := tgbotapi.NewMessage(parsedChatId, text)

  if _, err := b.api.Send(msg); err != nil {
    return err
  }

  return nil
}

func (b *Bot) AddHandler(command string, handler CommandHandler) error {
  b.handlers[command] = handler

  return nil
}
