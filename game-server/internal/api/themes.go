package api

import (
	"errors"
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	m "github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type ThemesApi struct {
	ts *services.ThemeService
}

func NewThemesApi(ts *services.ThemeService) *ThemesApi {
	return &ThemesApi{
		ts: ts,
	}
}

func (api *ThemesApi) Register(g *echo.Group) {
	v1group := g.Group("/v1")

	v1group.POST("/themes", api.createTheme)
	v1group.GET("/themes", api.listAllThemes)
}

func (api *ThemesApi) createTheme(c echo.Context) error {
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

func (api *ThemesApi) listAllThemes(c echo.Context) error {
	themes, err := api.ts.ListAllThemes(c.Request().Context())
	if err != nil {
		c.Echo().Logger.Errorf("failed to create theme: %v", err)

		var tsErr services.ThemeServiceError
		if errors.As(err, &tsErr) {
			if tsErr.Code == services.TSErrNotFound {
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