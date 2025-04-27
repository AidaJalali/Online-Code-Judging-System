package handlers

import (
	"net/http"
	"strconv"
)

// ManageRoles handles the role management page
func (h *Handler) ManageRoles(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated and is an admin
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	currentUser, err := h.userRepo.GetUserByUsername(cookie.Value)
	if err != nil || currentUser == nil || currentUser.Role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get page number from query parameter, default to 1
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Set page size
	pageSize := 10

	// Get paginated users
	users, total, err := h.userRepo.GetAllUsers(page, pageSize)
	if err != nil {
		data := PageData{
			Title: "Manage Roles",
			Error: "Failed to fetch users: " + err.Error(),
		}
		renderTemplate(w, "manage-roles", data)
		return
	}

	// Calculate total pages
	totalPages := (total + pageSize - 1) / pageSize

	data := PageData{
		Title:       "Manage Roles",
		User:        currentUser,
		Users:       users,
		CurrentUser: currentUser,
		Pagination: struct {
			CurrentPage  int
			TotalPages   int
			HasPrevious  bool
			HasNext      bool
			PreviousPage int
			NextPage     int
		}{
			CurrentPage:  page,
			TotalPages:   totalPages,
			HasPrevious:  page > 1,
			HasNext:      page < totalPages,
			PreviousPage: page - 1,
			NextPage:     page + 1,
		},
	}

	// Get any error or success messages from the URL
	if errorMsg := r.URL.Query().Get("error"); errorMsg != "" {
		data.Error = errorMsg
	}
	if successMsg := r.URL.Query().Get("success"); successMsg != "" {
		data.Success = successMsg
	}

	renderTemplate(w, "manage-roles", data)
}

// UpdateRole handles the role update action
func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated and is an admin
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	currentUser, err := h.userRepo.GetUserByUsername(cookie.Value)
	if err != nil || currentUser == nil || currentUser.Role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get form values
	username := r.FormValue("username")
	newRole := r.FormValue("new_role")

	// Validate inputs
	if username == "" || (newRole != "admin" && newRole != "user") {
		http.Redirect(w, r, "/manage-roles?error=Invalid input", http.StatusSeeOther)
		return
	}

	// Prevent changing own role
	if username == currentUser.Username {
		http.Redirect(w, r, "/manage-roles?error=Cannot change your own role", http.StatusSeeOther)
		return
	}

	// Prevent changing the admin user's role
	if username == "admin" {
		http.Redirect(w, r, "/manage-roles?error=Cannot change the admin user's role", http.StatusSeeOther)
		return
	}

	// Get the user to update
	user, err := h.userRepo.GetUserByUsername(username)
	if err != nil || user == nil {
		http.Redirect(w, r, "/manage-roles?error=User not found", http.StatusSeeOther)
		return
	}

	// Update the role
	err = h.userRepo.UpdateUserRole(username, newRole)
	if err != nil {
		http.Redirect(w, r, "/manage-roles?error=Failed to update role: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/manage-roles?success=Role updated successfully", http.StatusSeeOther)
}
