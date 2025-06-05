package handlers

import (
	"html/template"
	"net/http"
	"online-judge/internal/logger"
	"online-judge/internal/models"
	"online-judge/internal/types"
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
		logger.Println("Failed to fetch questions: %v", err)
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
			logger.Println("Failed to fetch owner for question %d: %v", q.ID, err)
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
		"templates/public/questions.html",
	)
	if err != nil {
		logger.Println("Failed to parse questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Println("Failed to execute questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

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
		logger.Println("Failed to fetch question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Render submit_question template
	tmpl, err := template.ParseFiles("templates/base.html", "templates/user-dashboard/submit_question.html")
	if err != nil {
		logger.Println("Failed to parse submit question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:    "Submit Question",
		Question: question,
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Println("Failed to execute submit question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		logger.Println("Failed to get questions: %v", err)
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
		logger.Println("Failed to parse manage questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		logger.Println("Failed to execute manage questions template: %v", err)
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
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			logger.Println("Failed to parse form: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var testCases []types.TestCase
		testInputs := r.Form["test_input[]"]
		testOutputs := r.Form["test_output[]"]
		numCases := min(len(testInputs), len(testOutputs))
		for i := 0; i < numCases; i++ {
			if testInputs[i] != "" && testOutputs[i] != "" {
				testCases = append(testCases, types.TestCase{
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
			logger.Println("Failed to save draft: %v", err)
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
				logger.Println("Failed to create question: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			err = h.draftRepo.DeleteDraft(user.ID)
			if err != nil {
				logger.Println("Failed to delete draft: %v", err)
			}

			http.Redirect(w, r, "/manage-questions", http.StatusSeeOther)
			return
		}

		// For regular users, just save as draft
		http.Redirect(w, r, "/my-questions", http.StatusSeeOther)
		return
	}

	// GET request - show the form
	draft, err := h.draftRepo.GetDraftByUserID(user.ID)
	if err != nil {
		logger.Println("Failed to get draft: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	funcMap := template.FuncMap{"add": func(a, b int) int { return a + b }}
	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(
		"templates/base.html",
		"templates/user-dashboard/create-question-form.html",
	)
	if err != nil {
		logger.Println("Failed to parse create question template: %v", err)
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
		logger.Println("Failed to execute create question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// EditQuestion handles editing a question
func (h *Handler) EditQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil {
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
		logger.Println("Failed to fetch question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if question == nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	// Check if user is admin or the owner of the question
	if user.Role != "admin" && question.OwnerID != user.ID {
		http.Error(w, "Unauthorized: You can only edit your own questions", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			logger.Println("Failed to parse form: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Update question fields
		question.Title = r.FormValue("title")
		question.Statement = r.FormValue("statement")
		question.TimeLimitMs = parseInt(r.FormValue("timeLimit"), 1000)
		question.MemoryLimitMb = parseInt(r.FormValue("memoryLimit"), 128)
		question.UpdatedAt = time.Now().Format(time.RFC3339)

		// Update test cases
		var testCases []types.TestCase
		testInputs := r.Form["test_input[]"]
		testOutputs := r.Form["test_output[]"]
		numCases := min(len(testInputs), len(testOutputs))
		for i := 0; i < numCases; i++ {
			if testInputs[i] != "" && testOutputs[i] != "" {
				testCases = append(testCases, types.TestCase{
					Input:  testInputs[i],
					Output: testOutputs[i],
				})
			}
		}
		question.SetTestCases(testCases)

		// Save the updated question
		err = h.questionRepo.UpdateQuestion(question)
		if err != nil {
			logger.Println("Failed to update question: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Redirect back to appropriate page
		if user.Role == "admin" {
			http.Redirect(w, r, "/all-drafts", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/my-questions", http.StatusSeeOther)
		}
		return
	}

	// GET request - show the edit form
	funcMap := template.FuncMap{"add": func(a, b int) int { return a + b }}
	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(
		"templates/base.html",
		"templates/user-dashboard/edit-question-form.html",
	)
	if err != nil {
		logger.Println("Failed to parse edit question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:    "Edit Question",
		User:     user,
		Question: question,
		Error:    "",
		Success:  "",
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Println("Failed to execute edit question template: %v", err)
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
		logger.Println("Failed to fetch question: %v", err)
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
		logger.Println("Failed to parse view question template: %v", err)
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
		logger.Println("Failed to execute view question template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// MyQuestions handles the page that shows a user's own draft questions
func (h *Handler) MyQuestions(w http.ResponseWriter, r *http.Request) {
	logger.Println("MyQuestions handler called")

	if r.Method != "GET" {
		logger.Println("Invalid method for MyQuestions: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.getAuthenticatedUser(r)
	if err != nil {
		logger.Println("Failed to get authenticated user: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user == nil {
		logger.Println("No authenticated user found")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	logger.Println("User authenticated: %s (ID: %d)", user.Username, user.ID)

	// If user is admin, redirect to all-drafts page
	if user.Role == "admin" {
		http.Redirect(w, r, "/all-drafts", http.StatusSeeOther)
		return
	}

	// Get user's draft questions only
	questions, err := h.questionRepo.GetDraftsByUserID(user.ID)
	if err != nil {
		logger.Println("Failed to fetch user's draft questions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Filter out published questions
	var draftQuestions []models.Question
	for _, q := range questions {
		if q.Status == models.StatusDraft {
			draftQuestions = append(draftQuestions, q)
		}
	}

	data := PageData{
		Title:     "My Draft Questions",
		User:      user,
		Questions: draftQuestions,
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/user-dashboard/my-questions.html",
	)
	if err != nil {
		logger.Println("Failed to parse my questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Println("Failed to execute my questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// PublishedQuestions handles the page that shows all published questions
func (h *Handler) PublishedQuestions(w http.ResponseWriter, r *http.Request) {
	logger.Println("PublishedQuestions handler called")

	if r.Method != "GET" {
		logger.Println("Invalid method for PublishedQuestions: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.getAuthenticatedUser(r)
	if err != nil {
		logger.Println("Failed to get authenticated user: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user == nil {
		logger.Println("No authenticated user found")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	logger.Println("User authenticated: %s (ID: %d)", user.Username, user.ID)

	// Get all published questions
	questions, err := h.questionRepo.GetPublishedQuestions()
	if err != nil {
		logger.Println("Failed to fetch published questions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	logger.Println("Found %d published questions", len(questions))

	// Get owners for all questions
	type QuestionWithOwner struct {
		Question models.Question
		Owner    *models.User
	}
	var questionsWithOwners []QuestionWithOwner

	for _, q := range questions {
		owner, err := h.userRepo.GetUserByID(q.OwnerID)
		if err != nil {
			logger.Println("Failed to fetch owner for question %d: %v", q.ID, err)
			continue
		}
		questionsWithOwners = append(questionsWithOwners, QuestionWithOwner{
			Question: q,
			Owner:    toModelsUser(owner),
		})
	}
	logger.Println("Processed %d questions with owner information", len(questionsWithOwners))

	data := PageData{
		Title:     "Published Questions",
		User:      user,
		Questions: questions,
		Error:     "",
		Success:   "",
	}

	logger.Println("Attempting to parse templates for PublishedQuestions")
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/public/published-questions.html",
	)
	if err != nil {
		logger.Println("Failed to parse published questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Println("Attempting to execute template for PublishedQuestions")
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Println("Failed to execute published questions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	logger.Println("PublishedQuestions handler completed successfully")
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
	questions, err := h.questionRepo.GetDraftsByUserID(user.ID)
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
	if r.Method != "GET" && r.Method != "POST" {
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
		logger.Println("Failed to fetch question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if question == nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	// Validate question before publishing
	if question.Title == "" {
		http.Error(w, "Question title is required", http.StatusBadRequest)
		return
	}
	if question.Statement == "" {
		http.Error(w, "Problem statement is required", http.StatusBadRequest)
		return
	}
	if question.TimeLimitMs <= 0 {
		http.Error(w, "Time limit must be greater than 0", http.StatusBadRequest)
		return
	}
	if question.MemoryLimitMb <= 0 {
		http.Error(w, "Memory limit must be greater than 0", http.StatusBadRequest)
		return
	}
	if question.TestInput == "" || question.TestOutput == "" {
		http.Error(w, "Test cases are required", http.StatusBadRequest)
		return
	}

	// Update question status to published
	question.Status = models.StatusPublished
	question.UpdatedAt = time.Now().Format(time.RFC3339)

	// Save the updated question
	err = h.questionRepo.UpdateQuestion(question)
	if err != nil {
		logger.Println("Failed to update question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect back to manage questions page
	http.Redirect(w, r, "/manage-questions", http.StatusSeeOther)
}

// DeleteQuestion handles deleting a question
func (h *Handler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getAuthenticatedUser(r)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	questionID := r.URL.Query().Get("id")
	if questionID == "" {
		http.Error(w, "Question ID is required", http.StatusBadRequest)
		return
	}

	// Get the question to check ownership
	question, err := h.questionRepo.GetQuestionByID(questionID)
	if err != nil {
		logger.Println("Failed to fetch question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if question == nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	// Check if user is admin or the owner of the question
	if user.Role != "admin" && question.OwnerID != user.ID {
		http.Error(w, "Unauthorized: You can only delete your own questions", http.StatusUnauthorized)
		return
	}

	// Delete the question
	err = h.questionRepo.DeleteQuestion(questionID)
	if err != nil {
		logger.Println("Failed to delete question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect back to appropriate page
	if user.Role == "admin" {
		http.Redirect(w, r, "/manage-questions", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/my-questions", http.StatusSeeOther)
	}
}

// AllDrafts handles the page that shows all draft questions (admin only)
func (h *Handler) AllDrafts(w http.ResponseWriter, r *http.Request) {
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

	// Get all draft questions only
	questions, err := h.questionRepo.GetDraftQuestions()
	if err != nil {
		logger.Println("Failed to fetch draft questions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Filter out published questions
	var draftQuestions []models.Question
	for _, q := range questions {
		if q.Status == models.StatusDraft {
			draftQuestions = append(draftQuestions, q)
		}
	}

	// Get owners for all questions
	type QuestionWithOwner struct {
		Question models.Question
		Owner    *models.User
	}
	var questionsWithOwners []QuestionWithOwner

	for _, q := range draftQuestions {
		owner, err := h.userRepo.GetUserByID(q.OwnerID)
		if err != nil {
			logger.Println("Failed to fetch owner for question %d: %v", q.ID, err)
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
		Title:               "All Draft Questions",
		User:                user,
		QuestionsWithOwners: questionsWithOwners,
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/admin-dashboard/all-drafts.html",
	)
	if err != nil {
		logger.Println("Failed to parse all drafts template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Println("Failed to execute all drafts template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
