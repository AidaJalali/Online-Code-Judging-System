package handlers

import (
	"html/template"
	"net/http"
	"online-judge/internal/logger"
	"online-judge/internal/models"
	"strconv"
	"time"
)

type QuestionRepository interface {
	CreateQuestion(question *models.Question) error
}

func (h *Handler) CreateQuestionForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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
		"templates/create-question-form.html",
	)
	if err != nil {
		logger.Error("Failed to parse create question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		User  *models.User
		Draft *models.Question
	}{
		Title: "Create Question",
		User:  user,
		Draft: draft,
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		logger.Error("Failed to execute create question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) HandleCreateQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated and is admin
	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil || user.Role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		logger.Error("Failed to parse multipart form: %v", err)
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
	draft := &models.Question{
		Title:         r.FormValue("title"),
		Statement:     r.FormValue("statement"),
		TimeLimitMs:   parseInt(r.FormValue("timeLimit"), 1000),
		MemoryLimitMb: parseInt(r.FormValue("memoryLimit"), 128),
		Status:        models.StatusDraft,
		OwnerID:       user.ID,
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
		now := time.Now()
		draft.Status = models.StatusDraft // This will be updated to published by the repository
		draft.CreatedAt = now.Format(time.RFC3339)
		draft.UpdatedAt = now.Format(time.RFC3339)

		if err := h.questionRepo.CreateQuestion(draft); err != nil {
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
