package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Authentication(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString := c.Request().Header.Get("Authorization")
			if tokenString != "" {
				tokenString = tokenString[len("Bearer "):]
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					return secret, nil
				})
				if err == nil {
					if token.Valid {
						c.Set("claims", token.Claims)
					}
				}
			}

			if err := next(c); err != nil {
				c.Error(err)
			}
			return nil
		}
	}
}
