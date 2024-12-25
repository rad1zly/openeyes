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

var jwtSecret = []byte("c46a2ead356271394b6fbe6b07de017156675d44b391fd910ad3c24fc24103196a29ad5eeda5d7086fc1129941cc1b5fe72df13c5dfb25f1974bd054637c1d0b4f9914507914ea3a229a2b7a4baea7d71489473c25a97d4720b2a8dee5ab7c328adb05bd3e268120aeadd4d4e1f90912a05db358606ca33a3e415aa17d5ca2a083d019001907e423e343dac155d0aae187a15f5d50d469f60b4a4c92fd24303eebc45178d93157597f9859566008f39ed597d3a0ff1d8c4f8b20c77b0fdc0b87d8e0cd04a7b4fafaf45d4b32be42bd4cffd1ec78f0f9da1011b70c3384cf6b778ee5ebcf94acaef9b52fb0e93bc45f9db3c3ae54baa6fbda58f4bf5b4db67eef074d8831a6742a3cd475f5e59f0eb49ac3d82883d3d013ae8a3d3426e0419d695e315c33f87ca0f06c4b3027e3eae4586cd6aed777aacd51d274c2637b577bb429127478b15a27e87fd40a51b6d6c1d30290224d4e795c7fd4f0c34a2ef6eed64337dcafca53e59103c8fd3e8379af99e3cbec064af50f53815f121d5912410128e380021656f624a1feac289699d7f8113b943879f2df9f853d5760dbffff5617bd9ce1de31191932edc3eebbc446fe9c3f900f7343f1be6db6eea4640f311bc3fc29b6654daca7112f16cf1dc1c68bd8b1051e672afdbb87dbd97a7253a9ee7996ed1224fc471c2da72e311d73b7d4a2e0da1ef950d23375594b9abce4a0fd")

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