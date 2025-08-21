package bot

import (
	"context"
	"fmt"
	"os"

	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	tg "github.com/go-telegram/bot"
	tgm "github.com/go-telegram/bot/models"
)

type CtxKey string

const PlayerDataKey CtxKey = "PlayerDataKey"

type TgBotManager struct {
	bot *tg.Bot
	ps  *services.PlayersService
	vs  *services.VideosService
}

func NewTgBotManager(ps *services.PlayersService, vs *services.VideosService) (*TgBotManager, error) {
	m := &TgBotManager{
		ps: ps,
		vs: vs,
	}

	opts := []tg.Option{
		tg.WithDefaultHandler(m.defaultHandler),
		tg.WithCallbackQueryDataHandler("game_", tg.MatchTypePrefix, m.callbackHandler),
		tg.WithMiddlewares(m.userDetailsMiddleware),
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

func (m *TgBotManager) userDetailsMiddleware(next tg.HandlerFunc) tg.HandlerFunc {
	senderFrom := func(upd *tgm.Update) *tgm.User {
		switch {
		case upd.Message != nil && upd.Message.From != nil:
			return upd.Message.From
		case upd.CallbackQuery != nil:
			return &upd.CallbackQuery.From
		default:
			return nil
		}
	}

	return func(ctx context.Context, b *tg.Bot, update *tgm.Update) {
		if from := senderFrom(update); from != nil {
			if player, err := m.getUser(ctx, from); err == nil {
				ctx = context.WithValue(ctx, PlayerDataKey, player)
			}
		}
		next(ctx, b, update)
	}
}

func (m *TgBotManager) getUser(ctx context.Context, from *tgm.User) (*qg.Player, error) {
	return m.ps.CheckPlayerExistsOrCreate(ctx, qg.CreatePlayerParams{
		TgID:           from.ID,
		TgUsername:     from.Username,
		TgLanguageCode: from.LanguageCode,
	})
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

// TODO: handle errors
func (m *TgBotManager) handleUrl(ctx context.Context, url string, b *tg.Bot, update *tgm.Update) {
	msg, err := b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Processing: %s", url),
	})

	if err != nil {
		return
	}

	player, ok := ctx.Value(PlayerDataKey).(*qg.Player)
	if ok {
		video, err := m.vs.CreateVideo(ctx, models.SubmitVideoRequest{
			PlayerID: player.TgID,
			VideoUrl: url,
		})

		if err != nil {
			b.EditMessageText(ctx, &tg.EditMessageTextParams{
				ChatID:    msg.Chat.ID,
				MessageID: msg.ID,
				Text:      fmt.Sprintf("Video can't be processed: %s", url),
			})
			return
		}

		b.EditMessageText(ctx, &tg.EditMessageTextParams{
			ChatID:    msg.Chat.ID,
			MessageID: msg.ID,
			Text:      fmt.Sprintf("Video has been added to your library: %s", video.VideoUrl),
		})
	}
}
