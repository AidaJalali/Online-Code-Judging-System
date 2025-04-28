package handlers

import (
	"io"
	"log"
	"net/http"
	"online-judge/internal/judge"
	"online-judge/internal/models"
	"strconv"
	"time"
)

// Submissions handles GET requests for the submissions page
func (h *Handler) Submissions(w http.ResponseWriter, r *http.Request) {
	log.Printf("Submissions handler called: %s %s", r.Method, r.URL.Path)

	// Get username from cookie
	cookie, err := r.Cookie("username")
	if err != nil {
		log.Printf("No username cookie found: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := h.userRepo.GetUserByUsername(cookie.Value)
	if err != nil || user == nil {
		log.Printf("Failed to get user: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get all submissions for this user, join with questions for title
	submissions, err := h.submissionRepo.GetUserSubmissionsWithQuestionTitle(user.ID)
	if err != nil {
		log.Printf("Failed to fetch submissions: %v", err)
		http.Error(w, "Failed to fetch submissions", http.StatusInternalServerError)
		return
	}

	success := ""
	if r.URL.Query().Get("success") == "true" {
		success = "Submission successful! Your code has been judged."
	}

	data := PageData{
		Title:       "Submissions",
		User:        user,
		Submissions: submissions,
		Success:     success,
	}

	log.Printf("Rendering submissions page with %d submissions", len(submissions))
	renderTemplate(w, "submissions", data)
}

// HandleCodeSubmission handles POST requests for code submissions
func (h *Handler) HandleCodeSubmission(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleCodeSubmission called: %s %s", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get username from cookie
	cookie, err := r.Cookie("username")
	if err != nil {
		log.Printf("No username cookie found: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := h.userRepo.GetUserByUsername(cookie.Value)
	if err != nil || user == nil {
		log.Printf("Failed to get user: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse form data
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		log.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form data
	questionID := r.FormValue("question_id")
	language := r.FormValue("language")
	log.Printf("Processing submission for question %s in %s", questionID, language)

	// Validate language
	if language == "" {
		log.Printf("Language field is empty")
		http.Error(w, "Please select a programming language", http.StatusBadRequest)
		return
	}

	// Validate supported languages
	supportedLanguages := map[string]bool{
		"python": true,
		"java":   true,
		"cpp":    true,
	}
	if !supportedLanguages[language] {
		log.Printf("Unsupported language: %s", language)
		http.Error(w, "Unsupported programming language", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, _, err := r.FormFile("code")
	if err != nil {
		log.Printf("Failed to read uploaded file: %v", err)
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file content
	code, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read file content: %v", err)
		http.Error(w, "Failed to read file content", http.StatusInternalServerError)
		return
	}

	// Get question and test cases
	question, err := h.questionRepo.GetQuestionByID(questionID)
	if err != nil {
		log.Printf("Failed to get question: %v", err)
		http.Error(w, "Failed to get question", http.StatusInternalServerError)
		return
	}

	// Create submission
	submission := judge.Submission{
		ID:        questionID,
		Code:      string(code),
		Language:  language,
		TestCases: question.GetTestCases(),
	}

	// Judge the submission
	log.Printf("Judging submission for question %s", questionID)
	result, err := judge.Judge(submission)
	if err != nil {
		log.Printf("Failed to judge submission: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert questionID to int64
	questionIDInt, err := strconv.ParseInt(questionID, 10, 64)
	if err != nil {
		log.Printf("Failed to parse question ID: %v", err)
		http.Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	// Create submission record
	submissionRecord := models.Submission{
		UserID:     user.ID,
		QuestionID: questionIDInt,
		Code:       string(code),
		Language:   language,
		Status:     result.Status,
		Message:    result.Message,
		TimeTaken:  result.TimeTaken.Milliseconds(),
		MemoryUsed: result.MemoryUsed,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	// Save the submission to the database
	err = h.submissionRepo.CreateSubmission(&submissionRecord)
	if err != nil {
		log.Printf("Failed to save submission: %v", err)
		http.Error(w, "Failed to save submission", http.StatusInternalServerError)
		return
	}

	// Set the submission ID in the result
	result.SubmissionID = strconv.FormatInt(submissionRecord.ID, 10)

	// Update the submission with the judge result
	err = h.submissionRepo.SaveSubmissionResult(result)
	if err != nil {
		log.Printf("Failed to save submission result: %v", err)
		http.Error(w, "Failed to save submission result", http.StatusInternalServerError)
		return
	}

	// Redirect to submissions page with success message
	log.Printf("Redirecting to submissions page")
	http.Redirect(w, r, "/submissions?success=true", http.StatusSeeOther)
	return
}
