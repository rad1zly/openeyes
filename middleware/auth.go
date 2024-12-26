package middleware

import (
    "net/http"
    "strings"
	"openeyes/database"
	"openeyes/models"

    "github.com/dgrijalva/jwt-go"
    "github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token not provided"})
            c.Abort()
            return
        }

        tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

        // Verifikasi token JWT
        claims, err := verifyToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // Ambil user ID dari klaim token
        userID := uint(claims.(jwt.MapClaims)["id"].(float64))

        // Periksa kecocokan token JWT dengan yang ada di database
        db := database.GetDB()
        var user models.User
        err = db.QueryRow("SELECT id, username, role FROM users WHERE id = ? AND jwt_token = ?", userID, tokenString).Scan(&user.ID, &user.Username, &user.Role)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // Set informasi pengguna ke konteks Gin
        c.Set("user", user)

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