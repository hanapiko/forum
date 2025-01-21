package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
)

type Post struct {
	ID         int
	Title      string
	Content    string
	Author     string
	Categories []string
	Likes      int
	Dislikes   int
	Comments   []Comment
}

type Comment struct {
	ID       int
	Content  string
	Author   string
	Likes    int
	Dislikes int
}

func Home(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
            SELECT p.id, p.title, p.content, u.username,
                   (SELECT COUNT(*) FROM reactions WHERE target_type='post' AND target_id=p.id AND reaction_type='like') as likes,
                   (SELECT COUNT(*) FROM reactions WHERE target_type='post' AND target_id=p.id AND reaction_type='dislike') as dislikes
            FROM posts p
            JOIN users u ON p.user_id = u.id
            ORDER BY p.created_at DESC
        `)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var posts []Post
		for rows.Next() {
			var p Post
			err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.Likes, &p.Dislikes)
			if err != nil {
				continue
			}
			posts = append(posts, p)
		}

		tmpl.ExecuteTemplate(w, "home.html", posts)
	}
}

func CreatePost(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := getCurrentUser(r, db)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if r.Method == "GET" {
			var categories []string
			rows, err := db.Query("SELECT name FROM categories")
			if err != nil {
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var category string
				rows.Scan(&category)
				categories = append(categories, category)
			}

			tmpl.ExecuteTemplate(w, "create-post.html", categories)
			return
		}

		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.Form["categories"]

		result, err := db.Exec("INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)",
			user.ID, title, content)
		if err != nil {
			http.Error(w, "Failed to create post", http.StatusInternalServerError)
			return
		}

		postID, _ := result.LastInsertId()

		// Associate categories
		for _, category := range categories {
			db.Exec(`INSERT INTO post_categories (post_id, category_id) 
                     SELECT ?, id FROM categories WHERE name = ?`, postID, category)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
