package handlers

import (
	"html/template"
	"net/http"
	"online-judge/internal/logger"
	"online-judge/internal/models"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := PageData{
			Title: "Login",
		}
		renderTemplate(w, "login", data)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := h.userRepo.GetUserByUsername(username)
	if err != nil {
		data := PageData{
			Title: "Login",
			Error: "Database error occurred",
		}
		renderTemplate(w, "login", data)
		return
	}

	if user == nil {
		data := PageData{
			Title: "Login",
			Error: "Invalid credentials",
		}
		renderTemplate(w, "login", data)
		return
	}

	if !h.userRepo.VerifyPassword(user.Password, password) {
		data := PageData{
			Title: "Login",
			Error: "Invalid credentials",
		}
		renderTemplate(w, "login", data)
		return
	}

	// Set username cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "username",
		Value:    username,
		Path:     "/",
		HttpOnly: true,
	})

	if user.Role == "admin" {
		http.Redirect(w, r, "/admin-dashboard", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/user-dashboard", http.StatusSeeOther)
	}
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		logger.Info("Register page accessed")
		data := PageData{
			Title: "Create Account",
		}
		renderRegisterPage(w, data)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")
		roleStr := r.FormValue("role")

		logger.Info("Registration attempt for user: %s with role: %s", username, roleStr)

		// Check if username already exists
		existingUser, err := h.userRepo.GetUserByUsername(username)
		if err != nil {
			logger.Error("Database error while checking username existence: %v", err)
			data := PageData{
				Title: "Create Account",
				Error: "An error occurred while processing your request. Please try again.",
			}
			renderRegisterPage(w, data)
			return
		}

		if existingUser != nil {
			logger.Info("Registration failed for user %s: Username already exists", username)
			data := PageData{
				Title: "Create Account",
				Error: "This username is already taken. Please choose another one.",
			}
			renderRegisterPage(w, data)
			return
		}

		// Convert role string to proper enum value
		var role string
		switch roleStr {
		case "regular":
			role = "regular"
		case "admin":
			role = "admin"
		default:
			logger.Error("Invalid role value: %s", roleStr)
			data := PageData{
				Title: "Create Account",
				Error: "Please select a valid role.",
			}
			renderRegisterPage(w, data)
			return
		}

		logger.Info("Registering user with role: %s", role)

		// Password validation
		if len(password) < 6 {
			logger.Info("Registration failed for user %s: Password too short", username)
			data := PageData{
				Title: "Create Account",
				Error: "Password must be at least 6 characters long.",
			}
			renderRegisterPage(w, data)
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
			data := PageData{
				Title: "Create Account",
				Error: "Password must contain at least one digit.",
			}
			renderRegisterPage(w, data)
			return
		}

		if !hasLowercase {
			logger.Info("Registration failed for user %s: Password missing lowercase", username)
			data := PageData{
				Title: "Create Account",
				Error: "Password must contain at least one lowercase letter.",
			}
			renderRegisterPage(w, data)
			return
		}

		if password != confirmPassword {
			logger.Info("Registration failed for user %s: Passwords do not match", username)
			data := PageData{
				Title: "Create Account",
				Error: "Passwords do not match. Please make sure both passwords are identical.",
			}
			renderRegisterPage(w, data)
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("Failed to hash password for user %s: %v", username, err)
			data := PageData{
				Title: "Create Account",
				Error: "An error occurred while processing your password. Please try again.",
			}
			renderRegisterPage(w, data)
			return
		}

		// Create user
		user := &models.User{
			Username: username,
			Password: string(hashedPassword),
			Role:     role,
		}

		err = h.userRepo.CreateUser(user)
		if err != nil {
			logger.Error("Failed to create user %s: %v", username, err)
			data := PageData{
				Title: "Create Account",
				Error: "An error occurred while creating your account. Please try again.",
			}
			renderRegisterPage(w, data)
			return
		}

		logger.Info("Successfully registered user: %s with role: %s", username, role)

		// Set username cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "username",
			Value:    username,
			Path:     "/",
			HttpOnly: true,
		})

		// Redirect based on role
		if role == "admin" {
			logger.Info("Redirecting admin user %s to admin dashboard", username)
			http.Redirect(w, r, "/admin-dashboard", http.StatusSeeOther)
		} else {
			logger.Info("Redirecting regular user %s to user dashboard", username)
			http.Redirect(w, r, "/user-dashboard", http.StatusSeeOther)
		}
	}
}

// Logout handles user logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear username cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "username",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Helper function to render the register page
func renderRegisterPage(w http.ResponseWriter, data PageData) {
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/signup/register.html",
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
}
