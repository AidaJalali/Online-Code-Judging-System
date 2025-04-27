package handlers

import (
	"html/template"
	"net/http"
	"online-judge/internal/logger"
	"online-judge/internal/models"
	"strconv"
	"time"
)

func (h *Handler) Questions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

func (h *Handler) SubmitQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
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

		err = tmpl.ExecuteTemplate(w, "base", data)
		if err != nil {
			logger.Error("Failed to execute submit question template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	if r.Method == http.MethodPost {
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

		questionID := r.URL.Query().Get("id")
		if questionID == "" {
			http.Error(w, "Question ID is required", http.StatusBadRequest)
			return
		}

		// Parse uploaded file
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
			return
		}
		defer file.Close()
		code := make([]byte, 0)
		buf := make([]byte, 4096)
		for {
			n, err := file.Read(buf)
			if n > 0 {
				code = append(code, buf[:n]...)
			}
			if err != nil {
				break
			}
		}

		// Save submission (status is 'pending')
		qid, err := strconv.ParseInt(questionID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid question ID", http.StatusBadRequest)
			return
		}
		sub := &models.Submission{
			QuestionID: qid,
			UserID:     user.ID,
			Code:       string(code),
			Status:     "pending",
			CreatedAt:  time.Now().Format("2006-01-02 15:04:05"),
		}
		err = h.submissionRepo.CreateSubmission(sub)
		if err != nil {
			http.Error(w, "Failed to save submission", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/submissions", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
