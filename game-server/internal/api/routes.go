package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	m "github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type GameApi struct {
	ts *services.ThemeService
	gs *services.GameService
}

func NewGameApi(ts *services.ThemeService, gs *services.GameService) *GameApi {
	return &GameApi{
		ts: ts,
		gs: gs,
	}
}

func (api *GameApi) Register(g *echo.Group) {
	v1group := g.Group("/v1")

	v1group.POST("/themes", api.CreateTheme)
	v1group.GET("/themes", api.ListAllThemes)

	v1group.GET("/games", api.GetGames)
	v1group.POST("/games", api.CreateGame)
	v1group.PUT("/games/:gameId/players", api.AddPlayerToGame)

	v1group.GET("/players/:id", api.GetPlayer)
	v1group.POST("/players", api.CreatePlayer)

	v1group.POST("/videos", api.SubmitVideo)
	v1group.POST("/votes", api.SubmitVote)
}

func (api *GameApi) CreateTheme(c echo.Context) error {
	request := new(m.CreateThemeParams)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	var description pgtype.Text
	if request.Description != nil {
		description.String = *request.Description
		description.Valid = true
	}

	ctx := c.Request().Context()
	theme, err := api.ts.CreateTheme(ctx, db.CreateThemeParams{Name: request.Name, Description: description})
	if err != nil {
		c.Echo().Logger.Errorf("failed to create theme: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, theme)
}

func (api *GameApi) ListAllThemes(c echo.Context) error {
	themes, err := api.ts.ListAllThemes(c.Request().Context())
	if err != nil {
		c.Echo().Logger.Errorf("failed to create theme: %v", err)

		var tsErr services.ThemeServiceError
		if errors.As(err, &tsErr) {
			if tsErr.Code == services.ThemeServiceNotFoundErrorCode {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "No themes found",
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	if themes == nil {
		themes = []db.Theme{}
	}

	return c.JSON(http.StatusOK, themes)
}

func (api *GameApi) CreateGame(c echo.Context) error {
	request := new(m.CreateGame)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	ctx := c.Request().Context()
	game, err := api.gs.CreateGame(ctx, request.ThemeId)
	if err != nil {
		c.Echo().Logger.Errorf("failed to create game: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, game)
}

func (api *GameApi) GetGames(c echo.Context) error {
	ctx := c.Request().Context()
	games, err := api.gs.GetGames(ctx)
	if err != nil {
		c.Echo().Logger.Errorf("failed to get all games: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	if games == nil {
		games = []db.Game{}
	}

	return c.JSON(http.StatusOK, games)
}

func (api *GameApi) CreatePlayer(c echo.Context) error {
	request := new(db.CreatePlayerParams)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	ctx := c.Request().Context()
	player, err := api.gs.CreatePlayer(ctx, *request)
	if err != nil {
		c.Echo().Logger.Errorf("failed to create player: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, player)
}

func (api *GameApi) GetPlayer(c echo.Context) error {
	playerId, err := parseInt64(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	ctx := c.Request().Context()
	player, err := api.gs.GetPlayer(ctx, playerId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, m.ErrorResponse{
				Error: "Player not found",
			})
		}

		c.Echo().Logger.Errorf("failed to get player: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, player)
}

func (api *GameApi) AddPlayerToGame(c echo.Context) error {
	gameId, err := parseInt64(c.Param("gameId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	request := new(m.AddPlayerToGame)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	ctx := c.Request().Context()
	err = api.gs.AddPlayer(ctx, gameId, request.PlayerId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" { // foreign_key_violation
				return c.JSON(http.StatusBadRequest, m.ErrorResponse{
					Error: "Player Or Game not found",
				})
			}
		}

		c.Echo().Logger.Errorf("failed to get player: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, m.OkResponse{
		Msg: "Player has been added to the game",
	})
}

func (api *GameApi) SubmitVideo(c echo.Context) error {
	request := new(db.CreateVideoParams)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	ctx := c.Request().Context()
	video, err := api.gs.SubmitVideo(ctx, *request)
	if err != nil {
		c.Echo().Logger.Errorf("failed to submit video: %v", err)
		return getErrorResponse(c, err)
	}

	return c.JSON(http.StatusOK, video)
}

func (api *GameApi) SubmitVote(c echo.Context) error {
	request := new(db.CreateVoteParams)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	ctx := c.Request().Context()
	vote, err := api.gs.SubmitVote(ctx, *request)
	if err != nil {
		c.Echo().Logger.Errorf("failed to submit vote: %v", err)
		return getErrorResponse(c, err)
	}

	return c.JSON(http.StatusOK, vote)
}

func getErrorResponse(c echo.Context, err error) error {
	var gErr services.GameServiceError
	if errors.As(err, &gErr) {
		switch gErr.Code {
		case services.CodeDbError:
			return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
				Error: "Something went wrong",
			})
		case services.CodeWrongGameStatus:
			return c.JSON(http.StatusBadRequest, m.ErrorResponse{
				Error: "Wrong Game Status",
			})
		case services.CodeNotFound:
			return c.JSON(http.StatusBadRequest, m.ErrorResponse{
				Error: "Player is not in the game",
			})
		}
	}

	return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
		Error: "Something went wrong",
	})
}

func parseInt64(str string) (int64, error) {
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}
