package bot

import (
	"context"
	"fmt"
	"log"
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
		tg.WithCallbackQueryDataHandler("game_", tg.MatchTypePrefix, m.callbackHandler),
		tg.WithMiddlewares(m.checkUserExists),
		tg.WithDebug(),
		tg.WithWorkers(4),
	}

	token := os.Getenv("TELEGRAM_BOT_KEY")
	bot, err := tg.New(token, opts...)
	if err != nil {
		return nil, err
	}

	bot.RegisterHandler(tg.HandlerTypeMessageText, "start", tg.MatchTypeCommand, m.startHandler)

	m.bot = bot
	return m, nil
}

func (m *TgBotManager) Start(ctx context.Context) {
	m.bot.Start(ctx)
}

func (m *TgBotManager) checkUserExists(next tg.HandlerFunc) tg.HandlerFunc {
	return func(ctx context.Context, b *tg.Bot, update *tgm.Update) {
		if update.Message != nil {
			log.Printf("%d say: %s", update.Message.From.ID, update.Message.Text)
		}
		if update.CallbackQuery != nil {
			log.Printf("%d say: %s", update.CallbackQuery.From.ID, update.CallbackQuery.Data)
		}
		next(ctx, b, update)
	}
}

func (m *TgBotManager) callbackHandler(ctx context.Context, b *tg.Bot, update *tgm.Update) {
	b.AnswerCallbackQuery(ctx, &tg.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   "Welcome to the game!!!",
	})
}

func (m *TgBotManager) startHandler(ctx context.Context, b *tg.Bot, update *tgm.Update) {
	kb := tgm.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgm.InlineKeyboardButton{
			{
				{Text: "Join Game", CallbackData: "game_join"},
			},
		},
	}

	b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Welcome to the Shorts Battle Game!!!",
		ReplyMarkup: kb,
	})
}

func (m *TgBotManager) defaultHandler(ctx context.Context, b *tg.Bot, update *tgm.Update) {
	if update.Message != nil && len(update.Message.Entities) > 0 {
		for _, e := range update.Message.Entities {
			if e.Type == tgm.MessageEntityTypeURL {
				url := update.Message.Text[e.Offset:e.Length]
				m.handleUrl(ctx, url, b, update)
				return
			}
		}
	}

	b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "It's Shorts Battle Game!!!",
	})
}

func (m *TgBotManager) handleUrl(ctx context.Context, url string, b *tg.Bot, update *tgm.Update) {
	b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Found url: %s", url),
	})
}
