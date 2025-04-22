package handlers

import (
	"html/template"
	"net/http"
	"online-judge/internal/logger"
	"online-judge/internal/repository"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
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
		logger.Info("Login page accessed")
		data := PageData{
			Title: "Sign In",
		}

		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/login.html",
		)
		if err != nil {
			logger.Error("Failed to parse login template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			logger.Error("Failed to execute login template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		logger.Info("Login attempt for user: %s", username)

		// Get user from database
		user, err := h.userRepo.GetUserByUsername(username)
		if err != nil {
			logger.Error("Database error during login for user %s: %v", username, err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check if user exists
		if user == nil {
			logger.Info("Failed login attempt for user %s: User not found", username)
			data := PageData{
				Title: "Sign In",
				Error: "Invalid username or password",
			}

			tmpl, err := template.ParseFiles(
				"templates/base.html",
				"templates/login.html",
			)
			if err != nil {
				logger.Error("Failed to parse login template: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
				logger.Error("Failed to execute login template: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Verify password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			logger.Info("Failed login attempt for user %s: Invalid password", username)
			data := PageData{
				Title: "Sign In",
				Error: "Invalid username or password",
			}

			tmpl, err := template.ParseFiles(
				"templates/base.html",
				"templates/login.html",
			)
			if err != nil {
				logger.Error("Failed to parse login template: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
				logger.Error("Failed to execute login template: %v", err)
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

		logger.Info("Successful login for user: %s with role: %s", username, user.Role)

		// Redirect based on role
		if user.Role == "admin" {
			http.Redirect(w, r, "/admin-dashboard", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/user-dashboard", http.StatusSeeOther)
		}
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		logger.Info("Register page accessed")
		data := PageData{
			Title: "Create Account",
		}
		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/register.html",
		)
		if err != nil {
			logger.Error("Failed to parse register template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			logger.Error("Failed to execute register template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")
		roleStr := r.FormValue("role")

		logger.Info("Registration attempt for user: %s with role: %s", username, roleStr)

		// Convert role string to proper enum value
		var role string
		switch roleStr {
		case "regular":
			role = "regular"
		case "admin":
			role = "admin"
		default:
			logger.Error("Invalid role value: %s", roleStr)
			renderError(w, "Invalid role selected")
			return
		}

		// Password validation
		if len(password) < 6 {
			logger.Info("Registration failed for user %s: Password too short", username)
			renderError(w, "Password must be at least 6 characters long")
			return
		}

		hasDigit := false
		hasLowercase := false
		for _, char := range password {
			if unicode.IsDigit(char) {
				hasDigit = true
			}
			if unicode.IsLower(char) {
				hasLowercase = true
			}
		}

		if !hasDigit {
			logger.Info("Registration failed for user %s: Password missing digit", username)
			renderError(w, "Password must contain at least one digit")
			return
		}

		if !hasLowercase {
			logger.Info("Registration failed for user %s: Password missing lowercase", username)
			renderError(w, "Password must contain at least one lowercase letter")
			return
		}

		if password != confirmPassword {
			logger.Info("Registration failed for user %s: Passwords do not match", username)
			renderError(w, "Passwords do not match")
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("Failed to hash password for user %s: %v", username, err)
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		// Create user
		user := &repository.User{
			Username: username,
			Password: string(hashedPassword),
			Role:     role,
		}

		err = h.userRepo.CreateUser(user)
		if err != nil {
			logger.Error("Failed to create user %s: %v", username, err)
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}

		logger.Info("Successfully registered user: %s with role: %s", username, role)
		// Redirect based on role
		if role == "admin" {
			http.Redirect(w, r, "/admin-dashboard", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/user-dashboard", http.StatusSeeOther)
		}
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

func (h *Handler) UserDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	session, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user data
	user, err := h.userRepo.GetUserByUsername(session.Value)
	if err != nil {
		logger.Error("Failed to get user data: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Render user dashboard template
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/user-dashboard/user-dashboard.html",
	)
	if err != nil {
		logger.Error("Failed to parse user dashboard template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: "User Dashboard",
		User:  user,
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		logger.Error("Failed to execute user dashboard template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	session, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user data
	user, err := h.userRepo.GetUserByUsername(session.Value)
	if err != nil {
		logger.Error("Failed to get user data: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if user is admin
	if user.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Render admin dashboard template
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/user-dashboard/admin-dashboard.html",
	)
	if err != nil {
		logger.Error("Failed to parse admin dashboard template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: "Admin Dashboard",
		User:  user,
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		logger.Error("Failed to execute admin dashboard template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
