package handlers

import (
	"html/template"
	"net/http"
	"online-judge/internal/logger"
)

// Questions handles the questions page
func (h *Handler) Questions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Fetch all questions from the repository
	questions, err := h.questionRepo.GetAllQuestions()
	if err != nil {
		logger.Error("Failed to fetch questions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:     "Questions",
		Questions: questions,
	}

	// Render questions template
	tmpl, err := template.ParseFiles("templates/base.html", "templates/questions.html")
	if err != nil {
		logger.Error("Failed to parse questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Error("Failed to execute questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// SubmitQuestion handles the question submission page
func (h *Handler) SubmitQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get question ID from query parameter
	questionID := r.URL.Query().Get("id")
	if questionID == "" {
		http.Error(w, "Question ID is required", http.StatusBadRequest)
		return
	}

	// Fetch question details from the repository
	question, err := h.questionRepo.GetQuestionByID(questionID)
	if err != nil {
		logger.Error("Failed to fetch question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Render submit_question template
	tmpl, err := template.ParseFiles("templates/base.html", "templates/user-dashboard/submit_question.html")
	if err != nil {
		logger.Error("Failed to parse submit question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:    "Submit Question",
		Question: question,
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		logger.Error("Failed to execute submit question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ManageQuestions handles the question management page for admins
func (h *Handler) ManageQuestions(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated and is admin
	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil || user.Role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get all questions from the database
	questions, err := h.questionRepo.GetAllQuestions()
	if err != nil {
		logger.Error("Failed to get questions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:     "Manage Questions",
		User:      user,
		Questions: questions,
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/user-dashboard/manage-questions.html",
	)
	if err != nil {
		logger.Error("Failed to parse manage questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		logger.Error("Failed to execute manage questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
