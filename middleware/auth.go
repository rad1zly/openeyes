package middleware

import (
    "net/http"
    "strings"

    "github.com/dgrijalva/jwt-go"
    "github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip middleware untuk endpoint login
        if c.FullPath() == "/api/login" {
            c.Next()
            return
        }

        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token not provided"})
            c.Abort()
            return
        }

        tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
        claims, err := verifyToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        userID := uint(claims.(jwt.MapClaims)["id"].(float64))
        c.Set("userID", userID)
        c.Next()
    }
}

func verifyToken(tokenString string) (jwt.Claims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte("abclimadasarapakahkamutahu"), nil
    })

    if err != nil {
        return nil, err
    }

    return token.Claims, nil
}