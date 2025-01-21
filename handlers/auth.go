package handlers

import (
    "database/sql"
    "html/template"
    "net/http"
    "time"
    "golang.org/x/crypto/bcrypt"
    "crypto/rand"
    "encoding/base64"
)

type User struct {
    ID       int
    Email    string
    Username string
}

func Register(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" {
            tmpl.ExecuteTemplate(w, "register.html", nil)
            return
        }

        email := r.FormValue("email")
        username := r.FormValue("username")
        password := r.FormValue("password")

        // Hash password
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if err != nil {
            http.Error(w, "Server error", http.StatusInternalServerError)
            return
        }

        // Check if email exists
        var exists bool
        err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        if exists {
            http.Error(w, "Email already registered", http.StatusBadRequest)
            return
        }

        // Insert user
        _, err = db.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)",
            email, username, string(hashedPassword))
        if err != nil {
            http.Error(w, "Failed to create user", http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, "/login", http.StatusSeeOther)
    }
}

func Login(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" {
            tmpl.ExecuteTemplate(w, "login.html", nil)
            return
        }

        email := r.FormValue("email")
        password := r.FormValue("password")

        var user User
        var hashedPassword string
        err := db.QueryRow("SELECT id, email, username, password FROM users WHERE email = ?",
            email).Scan(&user.ID, &user.Email, &user.Username, &hashedPassword)
        if err == sql.ErrNoRows {
            http.Error(w, "Invalid credentials", http.StatusUnauthorized)
            return
        }

        if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
            http.Error(w, "Invalid credentials", http.StatusUnauthorized)
            return
        }

        // Create session
        sessionToken := generateSessionToken()
        expiresAt := time.Now().Add(24 * time.Hour)

        _, err = db.Exec("INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)",
            user.ID, sessionToken, expiresAt)
        if err != nil {
            http.Error(w, "Server error", http.StatusInternalServerError)
            return
        }

        http.SetCookie(w, &http.Cookie{
            Name:     "session_token",
            Value:    sessionToken,
            Expires:  expiresAt,
            HttpOnly: true,
            Path:     "/",
        })

        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

func Logout(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        c, err := r.Cookie("session_token")
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // Delete session from database
        _, err = db.Exec("DELETE FROM sessions WHERE session_token = ?", c.Value)
        if err != nil {
            http.Error(w, "Server error", http.StatusInternalServerError)
            return
        }

        // Delete cookie
        http.SetCookie(w, &http.Cookie{
            Name:     "session_token",
            Value:    "",
            Expires:  time.Now(),
            HttpOnly: true,
            Path:     "/",
        })

        http.Redirect(w, r, "/login", http.StatusSeeOther)
    }
}

func generateSessionToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
} 