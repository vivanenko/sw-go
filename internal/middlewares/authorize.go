package middlewares

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func Authorize() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Get("claims") == nil {
				c.Response().WriteHeader(http.StatusUnauthorized)
				return nil
			}
			if err := next(c); err != nil {
				c.Error(err)
			}
			return nil
		}
	}
}
