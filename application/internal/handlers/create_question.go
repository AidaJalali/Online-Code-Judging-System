package handlers

import (
	"html/template"
	"net/http"
	"online-judge/internal/logger"
	"online-judge/internal/models"
	"strconv"
	"time"
)

func (h *Handler) CreateQuestionForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated and is admin
	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil || user.Role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get draft if exists
	draft, err := h.draftRepo.GetDraftByUserID(user.ID)
	if err != nil {
		logger.Error("Failed to get draft: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create template with functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(
		"templates/base.html",
		"templates/admin-dashboard/create-question-form.html",
	)
	if err != nil {
		logger.Error("Failed to parse create question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title   string
		User    *models.User
		Draft   *models.Question
		Error   string
		Success string
	}{
		Title:   "Create Question",
		User:    user,
		Draft:   draft,
		Error:   "",
		Success: "",
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Error("Failed to execute create question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) HandleCreateQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated and is admin
	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil || user.Role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		logger.Error("Failed to parse form: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Get test cases from form
	var testCases []models.TestCase
	testInputs := r.Form["test_input[]"]
	testOutputs := r.Form["test_output[]"]

	// Use the minimum length to avoid index out of range
	numCases := min(len(testInputs), len(testOutputs))
	for i := 0; i < numCases; i++ {
		if testInputs[i] != "" && testOutputs[i] != "" {
			testCases = append(testCases, models.TestCase{
				Input:  testInputs[i],
				Output: testOutputs[i],
			})
		}
	}

	// Create or update draft
	now := time.Now().Format(time.RFC3339)
	draft := &models.Question{
		Title:         r.FormValue("title"),
		Statement:     r.FormValue("statement"),
		TimeLimitMs:   parseInt(r.FormValue("timeLimit"), 1000),
		MemoryLimitMb: parseInt(r.FormValue("memoryLimit"), 128),
		Status:        models.StatusDraft,
		OwnerID:       user.ID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Set test cases
	draft.SetTestCases(testCases)

	err = h.draftRepo.SaveDraft(draft)
	if err != nil {
		logger.Error("Failed to save draft: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If form is complete, publish the question
	if draft.Title != "" && draft.Statement != "" && len(testCases) > 0 {
		// Update the status to published
		now = time.Now().Format(time.RFC3339)
		draft.Status = models.StatusPublished
		draft.CreatedAt = now
		draft.UpdatedAt = now

		err = h.questionRepo.CreateQuestion(draft)
		if err != nil {
			logger.Error("Failed to create question: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Delete the draft version
		err = h.draftRepo.DeleteDraft(user.ID)
		if err != nil {
			logger.Error("Failed to delete draft: %v", err)
		}

		http.Redirect(w, r, "/manage-questions", http.StatusSeeOther)
		return
	}

	// If form is incomplete, redirect back to form
	http.Redirect(w, r, "/create-question-form", http.StatusSeeOther)
}

func (h *Handler) getAuthenticatedUser(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("username")
	if err != nil {
		return nil, err
	}

	repoUser, err := h.userRepo.GetUserByUsername(cookie.Value)
	if err != nil {
		return nil, err
	}

	return toModelsUser(repoUser), nil
}

func parseInt(s string, defaultValue int) int {
	value, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return value
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
