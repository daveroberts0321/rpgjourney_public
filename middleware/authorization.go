package middleware

import (
	"fmt"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		jwtCookie := c.Cookies("jwt")
		if jwtCookie == "" {
			fmt.Println("JWT cookie not found, redirecting to login")
			return c.Redirect("/login")
		}

		// Debug: print the JWT cookie
		//fmt.Println("JWT Cookie: ", jwtCookie)

		token, err := jwt.Parse(jwtCookie, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			secretKey := os.Getenv("SECRET_KEY")
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			fmt.Println("Invalid JWT token or expired, redirecting to login: ", err)
			return c.Redirect("/login")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fmt.Println("Invalid token claims")
			return fmt.Errorf("invalid token claims")
		}

		//fmt.Println("Claims: ", claims)

		c.Locals("userID", claims["ID"])
		c.Locals("userEmail", claims["email"])
		//fmt.Println("Middleware/User ID: ", claims["ID"])

		return c.Next()
	}
}

// Decode the JWT token
func DecodeJWT(jwtCookie string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(jwtCookie, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		secretKey := os.Getenv("SECRET_KEY")
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
