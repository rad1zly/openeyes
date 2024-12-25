package handlers

import (
    "crypto/rand"
	"math/big"
	"fmt"
	"net/http"
	"strings"
	"time"
	"openeyes/database"
	"openeyes/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("50128793515ef24d06d59540b64f57df2866b169b3eceeebff5bcee488b446935fe0d60103e83054ddc196a31045907cc9cf2a20914dd68f68a1fc25cba26f71eb1b59999ce8292720f96e5be88920cdc50dcc2da0965f20dbeda59058a1c4e8429346aa180bb3b36fb1cd4eedd9d675c7cf701cc3fe075a536021618d5ce2cb93128894df7abaee518043e1fe43e79e8918bb4e3dfdb4b0799e53e2364e3540aab2e7f59d3979b1a1aeb3d8fa63aa3121ad68a5cd491fd0535639b9c31dc3dfbd05e27c02937801cf8351ba0165968d25c975b1bcfac7bcd543a61e385a91248049de9adae3049c0c5153a4c42130a5606fde89dbd484a9ca251f16ed015b98")

func LoginHandler(c *gin.Context) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var user models.User
	err := db.QueryRow("SELECT id, username, password, role FROM users WHERE username = ?", loginData.Username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token := generateToken(user)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func LogoutHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func CreateUserHandler(c *gin.Context) {
	user, err := authenticate(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if user.Role != "superadmin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var newUser models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	newUser.Password = string(hashedPassword)
	newUser.Role = "user"

	db := database.GetDB()
	_, err = db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", newUser.Username, newUser.Password, newUser.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

func ResetPasswordHandler(c *gin.Context) {
	user, err := authenticate(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if user.Role != "superadmin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var resetData struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&resetData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tempPassword := generateRandomPassword()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)

	db := database.GetDB()
	_, err = db.Exec("UPDATE users SET password = ? WHERE username = ?", string(hashedPassword), resetData.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully", "tempPassword": tempPassword})
}

func ChangePasswordHandler(c *gin.Context) {
	user, err := authenticate(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var passwordData struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&passwordData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var currentPassword string
	err = db.QueryRow("SELECT password FROM users WHERE id = ?", user.ID).Scan(&currentPassword)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(passwordData.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid old password"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(passwordData.NewPassword), bcrypt.DefaultCost)
	_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", string(hashedPassword), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func DeleteUserHandler(c *gin.Context) {
	user, err := authenticate(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if user.Role != "superadmin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var deleteData struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&deleteData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	_, err = db.Exec("DELETE FROM users WHERE username = ?", deleteData.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func authenticate(c *gin.Context) (models.User, error) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		return models.User{}, fmt.Errorf("No token provided")
	}

	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
	claims, err := verifyToken(tokenString)
	if err != nil {
		return models.User{}, err
	}

	userID := uint(claims.(jwt.MapClaims)["id"].(float64))
	var user models.User
	db := database.GetDB()
	err = db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Username, &user.Role)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func generateToken(user models.User) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString
}

func verifyToken(tokenString string) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	return token.Claims, nil
}

func generateRandomPassword() string {
    length := 12
    chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"
    password := make([]byte, length)

    for i := 0; i < length; i++ {
        charIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
        password[i] = chars[charIndex.Int64()]
    }

    return string(password)
}

func GetAllUsersHandler(c *gin.Context) {
    // Cek autentikasi
    user, err := authenticate(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    // Cek role superadmin
    if user.Role != "superadmin" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Get database connection
    db := database.GetDB()

    // Get all users from database
    rows, err := db.Query("SELECT id, username, role FROM users")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
        return
    }
    defer rows.Close()

    var users []models.User
    for rows.Next() {
        var user models.User
        if err := rows.Scan(&user.ID, &user.Username, &user.Role); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user data"})
            return
        }
        users = append(users, user)
    }

    // Check for errors from iterating over rows
    if err = rows.Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating users"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Users retrieved successfully",
        "data": users,
    })
}