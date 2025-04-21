package handlers

import (
	"html/template"
	"net/http"
	"online-judge/internal/repository"
	"strings"
)

type PageData struct {
	Title     string
	Error     string
	User      *repository.User
	Questions []Question
	Question  *Question
}

type Question struct {
	ID          int
	Title       string
	Description string
	Difficulty  string // "easy", "medium", "hard"
	CreatedAt   string
	CreatedBy   string
	TestCases   []TestCase
}

type TestCase struct {
	Input  string
	Output string
}

type Handler struct {
	userRepo *repository.UserRepository
}

func NewHandler(userRepo *repository.UserRepository) *Handler {
	return &Handler{
		userRepo: userRepo,
	}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
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

		user, err := h.userRepo.GetUserByUsername(username)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if user == nil || user.Password != password {
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

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:  "username",
			Value: username,
			Path:  "/",
		})

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
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
		email := r.FormValue("email")
		fullName := r.FormValue("full_name")

		// Validation
		if len(username) < 3 {
			renderError(w, "Username must be at least 3 characters long")
			return
		}

		if len(password) < 8 {
			renderError(w, "Password must be at least 8 characters long")
			return
		}

		if password != confirmPassword {
			renderError(w, "Passwords do not match")
			return
		}

		if !isValidEmail(email) {
			renderError(w, "Invalid email format")
			return
		}

		// Check if username exists
		exists, err := h.userRepo.UsernameExists(username)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if exists {
			renderError(w, "Username already exists")
			return
		}

		// Check if email exists
		exists, err = h.userRepo.EmailExists(email)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if exists {
			renderError(w, "Email already exists")
			return
		}

		// Create user
		user := &repository.User{
			Username: username,
			Password: password,
			Email:    email,
			FullName: fullName,
			Role:     "user",
		}

		err = h.userRepo.CreateUser(user)
		if err != nil {
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func renderError(w http.ResponseWriter, errorMessage string) {
	data := PageData{
		Title: "Create Account",
		Error: errorMessage,
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
}

func isValidEmail(email string) bool {
	// Simple email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user data from session
	user, err := h.userRepo.GetUserBySession(session.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Render dashboard template
	tmpl, err := template.ParseFiles("templates/dashboard.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		User *repository.User
	}{
		User: user,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Questions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user data from session
	_, err = h.userRepo.GetUserBySession(session.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// TODO: Get questions from repository
	questions := []struct {
		ID          int
		Title       string
		Description string
		Difficulty  string
	}{
		{1, "Two Sum", "Given an array of integers...", "Easy"},
		{2, "Add Two Numbers", "You are given two non-empty linked lists...", "Medium"},
	}

	// Render questions template
	tmpl, err := template.ParseFiles("templates/questions.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Questions []struct {
			ID          int
			Title       string
			Description string
			Difficulty  string
		}
	}{
		Questions: questions,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//// Check if user is authenticated
	//session, err := r.Cookie("session")
	//if err != nil {
	//	http.Redirect(w, r, "/login", http.StatusSeeOther)
	//	return
	//}

	//// Get user data from session
	//user, err := h.userRepo.GetUserBySession(session.Value)
	//if err != nil {
	//	http.Redirect(w, r, "/login", http.StatusSeeOther)
	//	return
	//}

	// Check if user is admin
	//if !user.IsAdmin {
	//	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	//	return
	//}

	// Render create question template
	tmpl, err := template.ParseFiles("templates/create_question.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) SubmitQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	//session, err := r.Cookie("session")
	//if err != nil {
	//	http.Redirect(w, r, "/login", http.StatusSeeOther)
	//	return
	//}

	// Get user data from session
	//user, err := h.userRepo.GetUserBySession(session.Value)
	//if err != nil {
	//	http.Redirect(w, r, "/login", http.StatusSeeOther)
	//	return
	//}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// TODO: Save submission to repository
	// TODO: Run code against test cases
	// TODO: Return result

	http.Redirect(w, r, "/submissions", http.StatusSeeOther)
}

func (h *Handler) Submissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//// Check if user is authenticated
	//session, err := r.Cookie("session")
	//if err != nil {
	//	http.Redirect(w, r, "/login", http.StatusSeeOther)
	//	return
	//}

	// Get user data from session
	//user, err := h.userRepo.GetUserBySession(session.Value)
	//if err != nil {
	//	http.Redirect(w, r, "/login", http.StatusSeeOther)
	//	return
	//}

	// TODO: Get user's submissions from repository
	submissions := []struct {
		ID         int
		QuestionID int
		Status     string
		Language   string
		Time       string
	}{
		{1, 1, "Accepted", "Go", "2024-03-20 10:00:00"},
		{2, 2, "Wrong Answer", "Python", "2024-03-20 11:00:00"},
	}

	// Render submissions template
	tmpl, err := template.ParseFiles("templates/submissions.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Submissions []struct {
			ID         int
			QuestionID int
			Status     string
			Language   string
			Time       string
		}
	}{
		Submissions: submissions,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user data from session
	user, err := h.userRepo.GetUserBySession(session.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Render profile template
	tmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		User *repository.User
	}{
		User: user,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
