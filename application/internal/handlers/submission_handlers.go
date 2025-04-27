package handlers

import (
	"net/http"
)

// Submissions handles GET requests for the submissions page
func (h *Handler) Submissions(w http.ResponseWriter, r *http.Request) {
	// Get username from cookie
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := h.userRepo.GetUserByUsername(cookie.Value)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get all submissions for this user, join with questions for title
	submissions, err := h.submissionRepo.GetUserSubmissionsWithQuestionTitle(user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch submissions", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:       "Submissions",
		User:        user,
		Submissions: submissions,
	}

	renderTemplate(w, "submissions", data)
}
