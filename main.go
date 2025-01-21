package main

import (
	"log"
	"net/http"
	"html/template"
	"forum/database"
	"forum/handlers"
)

func main() {
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Parse templates
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	// Static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/", handlers.Home(db, tmpl))
	http.HandleFunc("/register", handlers.Register(db, tmpl))
	http.HandleFunc("/login", handlers.Login(db, tmpl))
	http.HandleFunc("/logout", handlers.Logout(db))
	http.HandleFunc("/create-post", handlers.CreatePost(db, tmpl))
	http.HandleFunc("/reaction", handlers.HandleReaction(db))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
