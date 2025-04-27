package handlers

import (
	"net/http"
)

// Submissions handles GET requests for the submissions page
func (h *Handler) Submissions(w http.ResponseWriter, r *http.Request) {
	submissions, err := h.submissionRepo.GetAllSubmissions()
	if err != nil {
		http.Error(w, "Failed to fetch submissions", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:       "Submissions",
		Submissions: submissions,
	}

	renderTemplate(w, "submissions", data)
}
