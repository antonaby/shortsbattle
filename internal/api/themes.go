package api

import (
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

func (api *HttpApi) createTheme(c echo.Context) error {
	request, err := bindAndValidate[models.CreateThemeRequest](c)
	if err != nil {
		return sendInvalidReq(c)
	}

	var description pgtype.Text
	if request.Description != nil {
		description.String = *request.Description
		description.Valid = true
	}

	ctx := c.Request().Context()
	theme, err := api.themes.CreateTheme(ctx, request.Title, description, request.Mode)
	if err != nil {
		if sErr, ok := isServErr(err); ok {
			if sErr.Code == common.ErrorBadData {
				return sendInvalidReq(c)
			}
		}

		return logAndSendUnknowError(c, err)
	}

	return c.JSON(http.StatusOK, theme)
}

func (api *HttpApi) getTheme(c echo.Context) error {
	themeId, err := param64(c, "id")
	if err != nil {
		return sendInvalidId(c)
	}

	ctx := c.Request().Context()
	theme, err := api.themes.GetTheme(ctx, themeId)
	if err != nil {
		if sErr, ok := isServErr(err); ok {
			if sErr.Code == common.ErrorNotFound {
				return sendNotFound(c, "theme")
			}
		}

		return logAndSendUnknowError(c, err)
	}

	return c.JSON(http.StatusOK, theme)
}

func (api *HttpApi) listAllThemes(c echo.Context) error {
	themes, err := api.themes.ListAllThemes(c.Request().Context())
	if err != nil {
		return logAndSendUnknowError(c, err)
	}

	return c.JSON(http.StatusOK, themes)
}

func (api *HttpApi) createRound(c echo.Context) error {
	themeId, err := param64(c, "id")
	if err != nil {
		return sendInvalidId(c)
	}

	request, err := bindAndValidate[models.CreateRoundRequest](c)
	if err != nil {
		return sendInvalidReq(c)
	}

	var description pgtype.Text
	if request.Description != nil {
		description.String = *request.Description
		description.Valid = true
	}

	ctx := c.Request().Context()
	vr, err := api.themes.CreateRound(ctx, qg.CreateRoundParams{
		RoundN:      request.RoundN,
		Title:       request.Title,
		Description: description,
		ThemeID:     themeId,
	})

	if err != nil {
		if sErr, ok := isServErr(err); ok {
			if sErr.Code == common.ErrorConstraintViolation {
				return sendNotFound(c, "theme")
			}
		}

		return logAndSendUnknowError(c, err)
	}

	return c.JSON(http.StatusOK, vr)
}
