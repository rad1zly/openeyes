package handlers

import (
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/dgrijalva/jwt-go"
    "golang.org/x/crypto/bcrypt"
    "openeyes/database"
    "openeyes/models"
)

var jwtSecret = []byte("your_secret_key")

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var loginData struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    json.NewDecoder(r.Body).Decode(&loginData)

    db := database.GetDB()
    var user models.User
    err := db.QueryRow("SELECT id, username, password, role FROM users WHERE username = ?", loginData.Username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    token := generateToken(user)
    json.NewEncoder(w).Encode(map[string]string{
        "token": token,
    })
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Logged out successfully",
    })
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    user, err := authenticate(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    if user.Role != "superadmin" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var newUser models.User
    json.NewDecoder(r.Body).Decode(&newUser)

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
    newUser.Password = string(hashedPassword)
    newUser.Role = "user"

    db := database.GetDB()
    _, err = db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", newUser.Username, newUser.Password, newUser.Role)
    if err != nil {
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "message": "User created successfully",
    })
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
    user, err := authenticate(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    if user.Role != "superadmin" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var resetData struct {
        Username string `json:"username"`
    }
    json.NewDecoder(r.Body).Decode(&resetData)

    tempPassword := generateRandomPassword()
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)

    db := database.GetDB()
    _, err = db.Exec("UPDATE users SET password = ? WHERE username = ?", string(hashedPassword), resetData.Username)
    if err != nil {
        http.Error(w, "Failed to reset password", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "message":       "Password reset successfully",
        "tempPassword": tempPassword,
    })
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
    user, err := authenticate(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    var passwordData struct {
        OldPassword string `json:"oldPassword"`
        NewPassword string `json:"newPassword"`
    }
    json.NewDecoder(r.Body).Decode(&passwordData)

    db := database.GetDB()
    var currentPassword string
    err = db.QueryRow("SELECT password FROM users WHERE id = ?", user.ID).Scan(&currentPassword)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(passwordData.OldPassword)); err != nil {
        http.Error(w, "Invalid old password", http.StatusBadRequest)
        return
    }

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(passwordData.NewPassword), bcrypt.DefaultCost)

    _, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", string(hashedPassword), user.ID)
    if err != nil {
        http.Error(w, "Failed to change password", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "message": "Password changed successfully",
    })
}

func authenticate(r *http.Request) (models.User, error) {
    tokenString := r.Header.Get("Authorization")
    if !strings.HasPrefix(tokenString, "Bearer ") {
        return models.User{}, fmt.Errorf("Invalid token")
    }
    tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("Invalid token")
        }
        return jwtSecret, nil
    })
    if err != nil {
        return models.User{}, err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        username := claims["username"].(string)
        db := database.GetDB()
        var user models.User
        err := db.QueryRow("SELECT id, username, role FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Role)
        if err != nil {
            return models.User{}, fmt.Errorf("Invalid token")
        }
        return user, nil
    }

    return models.User{}, fmt.Errorf("Invalid token")
}

func generateToken(user models.User) string {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": user.Username,
        "role":     user.Role,
        "exp":      time.Now().Add(time.Hour * 24).Unix(),
    })
    tokenString, _ := token.SignedString(jwtSecret)
    return tokenString
}

func generateRandomPassword() string {
    b := make([]byte, 8)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}