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

	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get all questions if admin, only published if regular user
	var questions []models.Question
	if user.Role == "admin" {
		questions, err = h.questionRepo.GetAllQuestions()
	} else {
		questions, err = h.questionRepo.GetPublishedQuestions()
	}

	if err != nil {
		logger.Error("Failed to fetch questions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:     "Questions",
		User:      user,
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
	if s == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return val
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (h *Handler) CreateQuestionForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	draft, err := h.draftRepo.GetDraftByUserID(user.ID)
	if err != nil {
		logger.Error("Failed to get draft: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	funcMap := template.FuncMap{"add": func(a, b int) int { return a + b }}
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

	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		logger.Error("Failed to parse form: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var testCases []models.TestCase
	testInputs := r.Form["test_input[]"]
	testOutputs := r.Form["test_output[]"]
	numCases := min(len(testInputs), len(testOutputs))
	for i := 0; i < numCases; i++ {
		if testInputs[i] != "" && testOutputs[i] != "" {
			testCases = append(testCases, models.TestCase{
				Input:  testInputs[i],
				Output: testOutputs[i],
			})
		}
	}

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

	draft.SetTestCases(testCases)
	err = h.draftRepo.SaveDraft(draft)
	if err != nil {
		logger.Error("Failed to save draft: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Only admins can publish questions
	if user.Role == "admin" && draft.Title != "" && draft.Statement != "" && len(testCases) > 0 {
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

		err = h.draftRepo.DeleteDraft(user.ID)
		if err != nil {
			logger.Error("Failed to delete draft: %v", err)
		}

		http.Redirect(w, r, "/manage-questions", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/create-question-form", http.StatusSeeOther)
}

func (h *Handler) EditQuestionForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil || user.Role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	questionID := r.URL.Query().Get("id")
	if questionID == "" {
		http.Error(w, "Question ID is required", http.StatusBadRequest)
		return
	}

	question, err := h.questionRepo.GetQuestionByID(questionID)
	if err != nil {
		logger.Error("Failed to fetch question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	funcMap := template.FuncMap{"add": func(a, b int) int { return a + b }}
	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(
		"templates/base.html",
		"templates/admin-dashboard/create-question-form.html",
	)
	if err != nil {
		logger.Error("Failed to parse edit question template: %v", err)
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
		Title:   "Edit Question",
		User:    user,
		Draft:   question,
		Error:   "",
		Success: "",
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Error("Failed to execute edit question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) ViewQuestion(w http.ResponseWriter, r *http.Request) {
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

	questionID := r.URL.Query().Get("id")
	if questionID == "" {
		http.Error(w, "Question ID is required", http.StatusBadRequest)
		return
	}

	question, err := h.questionRepo.GetQuestionByID(questionID)
	if err != nil {
		logger.Error("Failed to fetch question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if question == nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	funcMap := template.FuncMap{"add": func(a, b int) int { return a + b }}
	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(
		"templates/base.html",
		"templates/admin-dashboard/view-question.html",
	)
	if err != nil {
		logger.Error("Failed to parse view question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:    question.Title,
		User:     user,
		Question: question,
		Error:    "",
		Success:  "",
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Error("Failed to execute view question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// MyQuestions handles the page that shows a user's draft questions
func (h *Handler) MyQuestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user's drafts
	drafts, err := h.draftRepo.GetDraftsByUserID(user.ID)
	if err != nil {
		logger.Error("Failed to fetch drafts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title     string
		User      *models.User
		Questions []models.Question
		IsAdmin   bool
	}{
		Title:     "My Questions",
		User:      user,
		Questions: drafts,
		IsAdmin:   user.Role == "admin",
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/user-dashboard/my-questions.html",
	)
	if err != nil {
		logger.Error("Failed to parse my questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Error("Failed to execute my questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// PublishedQuestions handles the page that shows all published questions with their owners
func (h *Handler) PublishedQuestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get all published questions
	questions, err := h.questionRepo.GetPublishedQuestions()
	if err != nil {
		logger.Error("Failed to fetch published questions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get owners for all questions
	type QuestionWithOwner struct {
		Question models.Question
		Owner    *models.User
	}
	var questionsWithOwners []QuestionWithOwner

	for _, q := range questions {
		owner, err := h.userRepo.GetUserByID(q.OwnerID)
		if err != nil {
			logger.Error("Failed to fetch owner for question %d: %v", q.ID, err)
			continue
		}
		questionsWithOwners = append(questionsWithOwners, QuestionWithOwner{
			Question: q,
			Owner:    toModelsUser(owner),
		})
	}

	data := struct {
		Title               string
		User                *models.User
		QuestionsWithOwners []QuestionWithOwner
	}{
		Title:               "Published Questions",
		User:                user,
		QuestionsWithOwners: questionsWithOwners,
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/user-dashboard/published-questions.html",
	)
	if err != nil {
		logger.Error("Failed to parse published questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Error("Failed to execute published questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Drafts handles the drafts page that shows unpublished questions
func (h *Handler) Drafts(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user data
	user, err := h.userRepo.GetUserByUsername(cookie.Value)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user's draft questions
	questions, err := h.questionRepo.GetDraftQuestionsByUser(user.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Render drafts template
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/drafts.html",
	)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:     "My Drafts",
		User:      user,
		Questions: questions,
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// PublishQuestion handles publishing a draft question
func (h *Handler) PublishQuestion(w http.ResponseWriter, r *http.Request) {
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

	questionID := r.URL.Query().Get("id")
	if questionID == "" {
		http.Error(w, "Question ID is required", http.StatusBadRequest)
		return
	}

	// Get the question
	question, err := h.questionRepo.GetQuestionByID(questionID)
	if err != nil {
		logger.Error("Failed to fetch question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if question == nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	// Update question status to published
	question.Status = models.StatusPublished
	question.UpdatedAt = time.Now().Format(time.RFC3339)

	// Save the updated question
	err = h.questionRepo.UpdateQuestion(question)
	if err != nil {
		logger.Error("Failed to update question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect back to manage questions page
	http.Redirect(w, r, "/manage-questions", http.StatusSeeOther)
}
