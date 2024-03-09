package middleware

import (
	"fmt"

	"github.com/0xturkey/jd-go/pkg/handlers"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"

	"github.com/google/uuid"
)

func JWTProtected(c *fiber.Ctx) error {
	tokenString := c.Get("Authorization")

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Make sure token's signature algorithm is what you expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return handlers.JwtSecretKey, nil
	})

	// Check for parsing errors
	if err != nil {
		var errMsg string
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				errMsg = "Malformed token"
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				errMsg = "Token is either expired or not active yet"
			} else {
				errMsg = "Couldn't handle this token"
			}
		} else {
			errMsg = "Couldn't handle this token"
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    fiber.StatusUnauthorized,
			"message": errMsg,
		})
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, err := uuid.Parse(claims["user_id"].(string))

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    401,
				"message": "Invalid or expired token",
			})
		}

		// Set the user in the context
		c.Locals("user_id", userID)
		return c.Next()
	} else {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    401,
			"message": "Invalid or expired token",
		})
	}
}
