package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/common"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/resources"
)

func JWT() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token, ok := strings.CutPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, resources.InvalidAccessToken)
			}

			userID, err := common.UserIDFromToken(token, common.TokenTypeAccess)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, resources.InvalidAccessToken)
			}

			c.Set(resources.UserIDKey, userID)
			return next(c)
		}
	}
}
