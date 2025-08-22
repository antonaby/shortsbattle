package api

import (
	"errors"
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	m "github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type ThemesApi struct {
	ts *services.ThemeService
}

func NewThemesApi(ts *services.ThemeService, g *echo.Group) *ThemesApi {
	api := &ThemesApi{
		ts: ts,
	}

	api.register(g)

	return api
}

func (api *ThemesApi) register(g *echo.Group) {
	v1group := g.Group("/v1")

	v1group.POST("/themes", api.createTheme)
	v1group.GET("/themes", api.listAllThemes)
	v1group.GET("/themes/:id", api.getTheme)

	v1group.POST("/themes/:id/videos", api.createVideoRequest)
}

func (api *ThemesApi) createTheme(c echo.Context) error {
	request := new(m.CreateThemeRequest)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidRequestFormatMsg,
		})
	}

	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: RequestValidationErrorMsg,
		})
	}

	var description pgtype.Text
	if request.Description != nil {
		description.String = *request.Description
		description.Valid = true
	}

	ctx := c.Request().Context()
	theme, err := api.ts.CreateTheme(ctx, qg.CreateThemeParams{Name: request.Name, Description: description})
	if err != nil {
		c.Echo().Logger.Errorf("failed to create theme: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, theme)
}

func (api *ThemesApi) getTheme(c echo.Context) error {
	themeId, err := parseInt64(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidIdFormatMsg,
		})
	}

	ctx := c.Request().Context()
	theme, err := api.ts.GetTheme(ctx, themeId)
	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorDbNotFound {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "theme not found",
				})
			}
		}

		c.Echo().Logger.Errorf("failed to get theme: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, theme)
}

func (api *ThemesApi) listAllThemes(c echo.Context) error {
	themes, err := api.ts.ListAllThemes(c.Request().Context())
	if err != nil {
		c.Echo().Logger.Errorf("failed to list themes: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, themes)
}

func (api *ThemesApi) createVideoRequest(c echo.Context) error {
	themeId, err := parseInt64(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidIdFormatMsg,
		})
	}

	request := new(m.CreateVideoRequest)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidRequestFormatMsg,
		})
	}

	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: RequestValidationErrorMsg,
		})
	}

	ctx := c.Request().Context()
	vr, err := api.ts.CreateVideoRequest(ctx, qg.CreateVideoRequestParams{
		Request: request.Request,
		ThemeID: themeId,
	})

	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorDbConstraintViolation {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "theme not found",
				})
			}
		}

		c.Echo().Logger.Errorf("failed to create video request: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, vr)
}
