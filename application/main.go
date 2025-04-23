package main

import (
	"html/template"
	"log"
	"net/http"
	"online-judge/internal/config"
	"online-judge/internal/handlers"
	"online-judge/internal/logger"
	"online-judge/internal/models"
	"online-judge/internal/repository"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type PageData struct {
	Title     string
	Error     string
	User      *models.User
	Questions []Question
	Question  *Question
	Success   string
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

func main() {
	// Initialize logger
	logger.Init()
	logger.Info("Application started")

	// Initialize database connection
	db, err := config.InitDB()
	if err != nil {
		logger.Error("Failed to initialize database: %v", err)
		log.Fatal(err)
	}
	defer db.Close()

	// Create repositories
	userRepo := repository.NewUserRepository(db)

	// Create handlers
	handler := handlers.NewHandler(userRepo)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Handle routes
	mux.HandleFunc("/", handler.Home)
	mux.HandleFunc("/login", handler.Login)
	mux.HandleFunc("/register", handler.Register)
	mux.HandleFunc("/user-dashboard", handler.UserDashboard)
	mux.HandleFunc("/admin-dashboard", handler.AdminDashboard)
	mux.HandleFunc("/create-question-form", handler.CreateQuestionForm)
	mux.HandleFunc("/create-question", handler.HandleCreateQuestion)
	mux.HandleFunc("/questions", handler.Questions)
	mux.HandleFunc("/questions/submit", handler.SubmitQuestion)
	mux.HandleFunc("/submissions", handler.Submissions)
	mux.HandleFunc("/profile", handler.Profile)
	mux.HandleFunc("/manage-users", handler.ManageUsers)
	mux.HandleFunc("/update-user-role", handler.UpdateUserRole)

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
		log.Printf("Error parsing templates: %v", err) // Log the detailed error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("Error executing template: %v", err) // Log execution errors too
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

		// Initialize database connection
		db, err := config.InitDB()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		userRepo := repository.NewUserRepository(db)
		user, err := userRepo.GetUserByUsername(username)
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

		// Initialize database connection
		db, err := config.InitDB()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		userRepo := repository.NewUserRepository(db)

		// Check if username exists
		exists, err := userRepo.UsernameExists(username)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if exists {
			renderError(w, "Username already exists")
			return
		}

		// Hash password (you should use a proper password hashing library like bcrypt)
		hashedPassword, err := hashPassword(password)
		if err != nil {
			http.Error(w, "Error processing password", http.StatusInternalServerError)
			return
		}

		// Create user
		user := &models.User{
			Username: username,
			Password: hashedPassword,
			Role:     "user",
		}

		err = userRepo.CreateUser(user)
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
	// You might want to use a more robust validation library
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func hashPassword(password string) (string, error) {
	// You should use a proper password hashing library like bcrypt
	// This is just a placeholder
	return password, nil
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get username from cookie
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Initialize database connection
	db, err := config.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.GetUserByUsername(cookie.Value)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := PageData{
		Title: "Dashboard",
		User:  user,
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

	// Initialize database connection
	db, err := config.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.GetUserByUsername(cookie.Value)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// TODO: Get questions from database
	data := PageData{
		Title:     "Questions",
		User:      user,
		Questions: []Question{}, // TODO: Get questions from database
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

	// Initialize database connection
	db, err := config.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.GetUserByUsername(cookie.Value)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		data := PageData{
			Title: "Create Question",
			User:  user,
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

	// Initialize database connection
	db, err := config.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.GetUserByUsername(cookie.Value)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		data := PageData{
			Title:    "Submit Solution",
			User:     user,
			Question: &Question{}, // TODO: Get question from database
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

	// Initialize database connection
	db, err := config.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.GetUserByUsername(cookie.Value)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// TODO: Get user's submissions
	data := PageData{
		Title: "Submissions",
		User:  user,
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

	// Initialize database connection
	db, err := config.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.GetUserByUsername(cookie.Value)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		data := PageData{
			Title: "Change Password",
			User:  user,
		}

		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/user-dashboard/profile.html",
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
		currentPassword := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")
		confirmPassword := r.FormValue("confirm_password")

		// Verify current password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword))
		if err != nil {
			data := PageData{
				Title: "Change Password",
				User:  user,
				Error: "Current password is incorrect",
			}
			renderProfile(w, data)
			return
		}

		// Validate new password
		if len(newPassword) < 6 {
			data := PageData{
				Title: "Change Password",
				User:  user,
				Error: "New password must be at least 6 characters long",
			}
			renderProfile(w, data)
			return
		}

		if newPassword != confirmPassword {
			data := PageData{
				Title: "Change Password",
				User:  user,
				Error: "New passwords do not match",
			}
			renderProfile(w, data)
			return
		}

		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error processing password", http.StatusInternalServerError)
			return
		}

		// Update user password
		user.Password = string(hashedPassword)
		err = userRepo.UpdateUser(user)
		if err != nil {
			http.Error(w, "Error updating password", http.StatusInternalServerError)
			return
		}

		data := PageData{
			Title:   "Change Password",
			User:    user,
			Success: "Password successfully changed",
		}
		renderProfile(w, data)
	}
}

func renderProfile(w http.ResponseWriter, data PageData) {
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/user-dashboard/profile.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
