package handlers

import (
	"html/template"
	"net/http"
	"online-judge/internal/models"
	"online-judge/internal/repository"
)

type PageData struct {
	Title       string
	Error       string
	Success     string
	User        *models.User
	Questions   []models.Question
	Question    *models.Question
	Users       []*models.User
	CurrentUser *models.User
	Submissions []repository.SubmissionWithQuestion
	Pagination  struct {
		CurrentPage  int
		TotalPages   int
		HasPrevious  bool
		HasNext      bool
		PreviousPage int
		NextPage     int
		PageSize     int
		TotalItems   int
	}
}

type Question struct {
	ID          int
	Title       string
	Description string
	Difficulty  string // "easy", "medium", "hard"
	CreatedAt   string
	CreatedBy   string
	TestCases   []TestCase
}

type TestCase struct {
	Input  string
	Output string
}

type QuestionRepository interface {
	CreateQuestion(question *models.Question) error
	GetAllQuestions() ([]models.Question, error)
	GetQuestionByID(id string) (*models.Question, error)
	GetPublishedQuestions() ([]models.Question, error)
	GetDraftQuestions() ([]models.Question, error)
	GetDraftsByUserID(userID int64) ([]models.Question, error)
	UpdateQuestion(question *models.Question) error
	DeleteQuestion(id string) error
}

type Handler struct {
	userRepo       *repository.UserRepository
	questionRepo   QuestionRepository
	draftRepo      *repository.DraftRepository
	submissionRepo *repository.SubmissionRepository
}

func NewHandler(userRepo *repository.UserRepository, draftRepo *repository.DraftRepository, submissionRepo *repository.SubmissionRepository) *Handler {
	return &Handler{
		userRepo:       userRepo,
		draftRepo:      draftRepo,
		submissionRepo: submissionRepo,
	}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Fetch all questions
	questions, err := h.questionRepo.GetAllQuestions()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:     "Welcome to Our Platform",
		Questions: questions,
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/home.html",
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func renderTemplate(w http.ResponseWriter, templateName string, data PageData) {
	var templatePath string
	if templateName == "login" {
		templatePath = "templates/signup/login.html"
	} else if templateName == "register" {
		templatePath = "templates/signup/register.html"
	} else {
		templatePath = "templates/user-dashboard/" + templateName + ".html"
	}

	// Create template functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"min": func(a, b int) int {
			if a < b {
				return a
			}
			return b
		},
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
	}

	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(
		"templates/base.html",
		templatePath,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) SetQuestionRepo(repo QuestionRepository) {
	h.questionRepo = repo
}
