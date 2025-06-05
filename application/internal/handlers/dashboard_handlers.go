package handlers

import (
	"html/template"
	"net/http"
	"online-judge/internal/logger"
	"online-judge/internal/models"
)

// Dashboard handles the main dashboard page
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
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
		User *models.User
	}{
		User: user,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// UserDashboard handles the user dashboard page
func (h *Handler) UserDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
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
		logger.Println("Failed to get user data: %v", err)
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
		logger.Println("Failed to parse user dashboard template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: "User Dashboard",
		User:  toModelsUser(user),
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		logger.Println("Failed to execute user dashboard template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// AdminDashboard handles the admin dashboard page
func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	session, err := r.Cookie("username")
	if err != nil {
		logger.Println("Unauthorized access attempt to admin dashboard: No session cookie")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user data
	user, err := h.userRepo.GetUserByUsername(session.Value)
	if err != nil {
		logger.Println("Database error while accessing admin dashboard: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		logger.Println("Unauthorized access attempt to admin dashboard: User not found")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	logger.Println("Admin dashboard access attempt by user %s with role %s", user.Username, user.Role)

	// Check if user is admin
	if user.Role != "admin" {
		logger.Println("Unauthorized access attempt to admin dashboard by user %s with role %s", user.Username, user.Role)
		http.Redirect(w, r, "/user-dashboard", http.StatusSeeOther)
		return
	}

	// Render admin dashboard template
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/user-dashboard/admin-dashboard.html",
	)
	if err != nil {
		logger.Println("Failed to parse admin dashboard template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: "Admin Dashboard",
		User:  toModelsUser(user),
	}

	logger.Println("Rendering admin dashboard for user %s", user.Username)
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		logger.Println("Failed to execute admin dashboard template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
