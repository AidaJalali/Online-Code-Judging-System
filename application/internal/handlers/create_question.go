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

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/create-question-form.html",
	)
	if err != nil {
		logger.Error("Failed to parse create question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: "Create Question",
		User:  user,
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

	if err := r.ParseForm(); err != nil {
		logger.Error("Failed to parse form: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	now := time.Now()
	question := &models.Question{
		Title:         r.FormValue("title"),
		Statement:     r.FormValue("statement"),
		TimeLimitMs:   parseInt(r.FormValue("timeLimit"), 1000),
		MemoryLimitMb: parseInt(r.FormValue("memoryLimit"), 128),
		Status:        models.StatusDraft,
		OwnerID:       user.ID, // Assuming User model has ID field
		CreatedAt:     now.Format(time.RFC3339),
		UpdatedAt:     now.Format(time.RFC3339),
	}

	if err := h.questionRepo.CreateQuestion(question); err != nil {
		logger.Error("Failed to create question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/manage-questions", http.StatusSeeOther)
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
