package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"online-judge/internal/judge"
	"online-judge/internal/models"
	"online-judge/internal/types"
)

// SubmissionHandler handles code submission requests (API)
type SubmissionHandler struct {
	// Add any dependencies here (e.g., database connection)
}

// NewSubmissionHandler creates a new submission handler
func NewSubmissionHandler() *SubmissionHandler {
	return &SubmissionHandler{}
}

// HandleSubmission processes a code submission (API)
func (h *SubmissionHandler) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var submission models.Submission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert to judge.Submission
	judgeSubmission := judge.Submission{
		ID:        string(submission.ID),
		Code:      submission.Code,
		Language:  submission.Language,
		TestCases: make([]types.TestCase, len(submission.TestCases)),
	}

	for i, tc := range submission.TestCases {
		judgeSubmission.TestCases[i] = types.TestCase{
			Input:  tc.Input,
			Output: tc.Output,
		}
	}

	// Judge the submission
	result, err := judge.Judge(judgeSubmission)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert result to response
	response := types.SubmissionResult{
		SubmissionID: result.SubmissionID,
		Status:       result.Status,
		Message:      result.Message,
		TimeTaken:    result.TimeTaken.Milliseconds(),
		MemoryUsed:   result.MemoryUsed,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Submissions handles GET requests for the submissions page (Web)
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

	// Process any pending submissions
	for _, submission := range submissions {
		if submission.Status == "Pending" {
			// Get full submission details including code
			fullSubmission, err := h.submissionRepo.GetSubmissionByID(submission.ID)
			if err != nil {
				log.Printf("Failed to get full submission details for submission %d: %v", submission.ID, err)
				continue
			}
			if fullSubmission == nil {
				log.Printf("Submission %d not found", submission.ID)
				continue
			}

			// Get question details
			question, err := h.questionRepo.GetQuestionByID(strconv.FormatInt(submission.QuestionID, 10))
			if err != nil {
				log.Printf("Failed to get question for submission %d: %v", submission.ID, err)
				continue
			}

			// Create temporary files for code and input
			codeFile := fmt.Sprintf("/tmp/submission_%d.go", submission.ID)
			inputFile := fmt.Sprintf("/tmp/input_%d.txt", submission.ID)
			resultFile := fmt.Sprintf("/tmp/result_%d.json", submission.ID)

			// Write code to file
			if err := ioutil.WriteFile(codeFile, []byte(fullSubmission.Code), 0644); err != nil {
				log.Printf("Failed to write code file: %v", err)
				continue
			}

			// Get test cases from question
			testCases := question.GetTestCases()
			if len(testCases) == 0 {
				log.Printf("No test cases found for question %d", question.ID)
				continue
			}

			// Write first test case input to file
			if err := ioutil.WriteFile(inputFile, []byte(testCases[0].Input), 0644); err != nil {
				log.Printf("Failed to write input file: %v", err)
				continue
			}

			// Run the code using the runner
			cmd := exec.Command("docker", "run", "--rm",
				"-v", fmt.Sprintf("%s:/code/code.go", codeFile),
				"-v", fmt.Sprintf("%s:/code/input.txt", inputFile),
				"-v", fmt.Sprintf("%s:/code/result.json", resultFile),
				"-e", fmt.Sprintf("TIMEOUT_SEC=%d", question.TimeLimitMs/1000), // Convert ms to seconds
				"code-runner")

			// Capture both stdout and stderr
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				log.Printf("Failed to run code: %v", err)
				log.Printf("Docker stdout: %s", stdout.String())
				log.Printf("Docker stderr: %s", stderr.String())
				continue
			}

			// Log successful Docker execution
			log.Printf("Docker execution successful")
			log.Printf("Docker stdout: %s", stdout.String())
			log.Printf("Docker stderr: %s", stderr.String())

			// Read result
			resultData, err := ioutil.ReadFile(resultFile)
			if err != nil {
				log.Printf("Failed to read result file: %v", err)
				continue
			}

			var result struct {
				Stdout   string `json:"stdout"`
				Stderr   string `json:"stderr"`
				ExitCode int    `json:"exit_code"`
				TimeMs   int64  `json:"time_ms"`
				Error    string `json:"error,omitempty"`
			}

			if err := json.Unmarshal(resultData, &result); err != nil {
				log.Printf("Failed to parse result: %v", err)
				continue
			}

			// Create judge result
			judgeResult := judge.Result{
				SubmissionID: strconv.FormatInt(submission.ID, 10),
				TimeTaken:    time.Duration(result.TimeMs) * time.Millisecond,
				MemoryUsed:   0, // TODO: Implement memory tracking
			}

			// Check if execution was successful
			if result.ExitCode != 0 || result.Error != "" {
				judgeResult.Status = "Compilation Error"
				judgeResult.Message = result.Error
				if result.Stderr != "" {
					judgeResult.Message = result.Stderr
				}
			} else {
				// Compare output with expected output
				expectedOutput := strings.TrimSpace(testCases[0].Output)
				actualOutput := strings.TrimSpace(result.Stdout)

				if expectedOutput == actualOutput {
					judgeResult.Status = "Accepted"
					judgeResult.Message = "Correct answer"
				} else {
					judgeResult.Status = "Wrong Answer"
					judgeResult.Message = "Wrong answer"
				}
			}

			// Save the result
			if err := h.submissionRepo.SaveSubmissionResult(judgeResult); err != nil {
				log.Printf("Failed to save submission result: %v", err)
			}

			// Clean up temporary files
			os.Remove(codeFile)
			os.Remove(inputFile)
			os.Remove(resultFile)
		}
	}

	// Refresh submissions after processing
	submissions, err = h.submissionRepo.GetUserSubmissionsWithQuestionTitle(user.ID)
	if err != nil {
		log.Printf("Failed to fetch updated submissions: %v", err)
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

// HandleCodeSubmission handles POST requests for code submissions (Web)
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
		"go":     true,
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
