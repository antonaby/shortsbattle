package bot

import (
	"context"
	"os"

	tg "github.com/go-telegram/bot"
	tgm "github.com/go-telegram/bot/models"
)

type TgBotManager struct {
	bot *tg.Bot
}

func NewTgBotManager() (*TgBotManager, error) {
	m := &TgBotManager{}

	opts := []tg.Option{
		tg.WithDefaultHandler(m.defaultHandler),
		tg.WithDebug(),
	}

	token := os.Getenv("TELEGRAM_BOT_KEY")
	bot, err := tg.New(token, opts...)
	if err != nil {
		return nil, err
	}

	m.bot = bot
	return m, nil
}

func (m *TgBotManager) Start(ctx context.Context) {
	m.bot.Start(ctx)
}

func (m *TgBotManager) defaultHandler(ctx context.Context, b *tg.Bot, update *tgm.Update) {
	b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Say /hello",
	})
}
