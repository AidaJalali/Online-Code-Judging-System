package routes

import (
	"net/http"
	"online-judge/internal/handlers"
)

func SetupRoutes(h *handlers.Handler) {
	// Public routes
	http.HandleFunc("/", h.Home)
	http.HandleFunc("/login", h.Login)
	http.HandleFunc("/register", h.Register)
	http.HandleFunc("/questions", h.Questions)
	http.HandleFunc("/published-questions", h.PublishedQuestions)

	// User dashboard routes (requires authentication)
	http.HandleFunc("/user-dashboard", h.UserDashboard)
	http.HandleFunc("/my-questions", h.MyQuestions)
	http.HandleFunc("/create-question-form", h.CreateQuestionForm)
	http.HandleFunc("/edit-question", h.EditQuestion)
	http.HandleFunc("/delete-question", h.DeleteQuestion)
	http.HandleFunc("/submit-question", h.SubmitQuestion)

	// Admin dashboard routes (requires admin role)
	http.HandleFunc("/admin-dashboard", h.AdminDashboard)
	http.HandleFunc("/all-drafts", h.AllDrafts)
	http.HandleFunc("/manage-questions", h.ManageQuestions)
	http.HandleFunc("/view-question", h.ViewQuestion)
	http.HandleFunc("/publish-question", h.PublishQuestion)
}
