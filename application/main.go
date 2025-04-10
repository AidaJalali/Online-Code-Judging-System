package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

type PageData struct {
	Title    string
	Error    string
	User     *User
	Questions []Question
	Question *Question
}

type User struct {
	Username string
	Password string
	Role     string // "user" or "admin"
	Email    string
	FullName string
}

type Question struct {
	ID          int
	Title       string
	Description string
	Difficulty  string // "easy", "medium", "hard"
	CreatedAt   time.Time
	CreatedBy   string
	TestCases   []TestCase
}

type TestCase struct {
	Input  string
	Output string
}

type Submission struct {
	ID         int
	QuestionID int
	UserID     string
	Code       string
	Language   string
	Status     string // "pending", "accepted", "rejected"
	CreatedAt  time.Time
}

// In-memory stores (replace with database in production)
var (
	users       = make(map[string]User)
	questions   = make(map[int]Question)
	submissions = make(map[int]Submission)
	nextID      = 1
)

// Add a default admin user
func init() {
	users["admin"] = User{
		Username: "admin",
		Password: "admin",
		Role:     "admin",
		Email:    "admin@example.com",
		FullName: "Administrator",
	}
}

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Handle routes
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/dashboard", dashboardHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/questions", questionsHandler)
	mux.HandleFunc("/questions/create", createQuestionHandler)
	mux.HandleFunc("/questions/submit", submitQuestionHandler)
	mux.HandleFunc("/submissions", submissionsHandler)
	mux.HandleFunc("/profile", profileHandler)

	// Start the server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := PageData{
		Title: "Welcome to Our Platform",
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/home.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := PageData{
			Title: "Sign In",
		}

		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/login.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, exists := users[username]
		if !exists || user.Password != password {
			data := PageData{
				Title: "Sign In",
				Error: "Invalid username or password",
			}

			tmpl, err := template.ParseFiles(
				"templates/base.html",
				"templates/login.html",
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Set session cookie (simplified version)
		http.SetCookie(w, &http.Cookie{
			Name:  "username",
			Value: username,
			Path:  "/",
		})

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := PageData{
			Title: "Create Account",
		}

		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/register.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		if password != confirmPassword {
			data := PageData{
				Title: "Create Account",
				Error: "Passwords do not match",
			}

			tmpl, err := template.ParseFiles(
				"templates/base.html",
				"templates/register.html",
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if _, exists := users[username]; exists {
			data := PageData{
				Title: "Create Account",
				Error: "Username already exists",
			}

			tmpl, err := template.ParseFiles(
				"templates/base.html",
				"templates/register.html",
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		users[username] = User{
			Username: username,
			Password: password,
			Role:     "user", // Default role for new users
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get username from cookie
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := cookie.Value
	user, exists := users[username]
	if !exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := PageData{
		Title: "Dashboard",
		User:  &user,
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/dashboard.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "username",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func questionsHandler(w http.ResponseWriter, r *http.Request) {
	// Get username from cookie
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := cookie.Value
	user, exists := users[username]
	if !exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Convert questions map to slice for template
	var questionsList []Question
	for _, q := range questions {
		questionsList = append(questionsList, q)
	}

	data := PageData{
		Title:     "Questions",
		User:      &user,
		Questions: questionsList,
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/questions.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createQuestionHandler(w http.ResponseWriter, r *http.Request) {
	// Get username from cookie
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := cookie.Value
	user, exists := users[username]
	if !exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		data := PageData{
			Title: "Create Question",
			User:  &user,
		}

		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/create_question.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		// TODO: Implement question creation
		http.Redirect(w, r, "/questions", http.StatusSeeOther)
	}
}

func submitQuestionHandler(w http.ResponseWriter, r *http.Request) {
	// Get username from cookie
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := cookie.Value
	user, exists := users[username]
	if !exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		questionID := r.URL.Query().Get("id")
		// TODO: Get question by ID
		data := PageData{
			Title:    "Submit Solution",
			User:     &user,
			Question: &Question{}, // TODO: Get actual question
		}

		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/submit_question.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		// TODO: Implement submission
		http.Redirect(w, r, "/submissions", http.StatusSeeOther)
	}
}

func submissionsHandler(w http.ResponseWriter, r *http.Request) {
	// Get username from cookie
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := cookie.Value
	user, exists := users[username]
	if !exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// TODO: Get user's submissions
	data := PageData{
		Title: "Submissions",
		User:  &user,
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/submissions.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	// Get username from cookie
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := cookie.Value
	user, exists := users[username]
	if !exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		data := PageData{
			Title: "Profile",
			User:  &user,
		}

		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/profile.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		// TODO: Implement profile update
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	}
}
