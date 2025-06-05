package handlers

import (
	"html/template"
	"net/http"
	"online-judge/internal/logger"
	"online-judge/internal/models"
)

// Profile handles the user profile page
func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	cookie, err := r.Cookie("username")
	if err != nil {
		logger.Println("Unauthorized access attempt to profile page: No session cookie")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user data
	user, err := h.userRepo.GetUserByUsername(cookie.Value)
	if err != nil {
		logger.Println("Database error while accessing profile: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		logger.Println("Unauthorized access attempt to profile page: User not found")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Render profile template
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/user-dashboard/profile.html",
	)
	if err != nil {
		logger.Println("Failed to parse profile template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: "Profile",
		User:  toModelsUser(user),
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		logger.Println("Failed to execute profile template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Helper function to convert a user to a models.User
func toModelsUser(user *models.User) *models.User {
	if user == nil {
		return nil
	}
	return &models.User{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
		Role:     user.Role,
	}
}
