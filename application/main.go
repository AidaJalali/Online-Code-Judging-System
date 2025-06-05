package main

import (
	"log"
	"net/http"
	"online-judge/internal/config"
	"online-judge/internal/handlers"
	"online-judge/internal/logger"
	"online-judge/internal/repository"
)

func main() {
	// Initialize logger
	logger.Init()
	logger.Println("Application started")

	// Initialize database connection
	db, err := config.InitDB()
	if err != nil {
		logger.Println("Failed to initialize database: %v", err)
		log.Fatal(err)
	}
	defer db.Close()

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	draftRepo := repository.NewDraftRepository(db)
	questionRepo := repository.NewQuestionRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)

	// Create handlers
	handler := handlers.NewHandler(userRepo, questionRepo, draftRepo, submissionRepo)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Handle routes
	mux.HandleFunc("/", handler.Home)
	mux.HandleFunc("/login", handler.Login)
	mux.HandleFunc("/register", handler.Register)
	mux.HandleFunc("/user-dashboard", handler.UserDashboard)
	mux.HandleFunc("/admin-dashboard", handler.AdminDashboard)
	mux.HandleFunc("/questions", handler.Questions)
	mux.HandleFunc("/my-questions", handler.MyQuestions)
	mux.HandleFunc("/published-questions", handler.PublishedQuestions)
	mux.HandleFunc("/all-drafts", handler.AllDrafts)
	mux.HandleFunc("/create-question-form", handler.CreateQuestionForm)
	mux.HandleFunc("/manage-questions", handler.ManageQuestions)
	mux.HandleFunc("/edit-question", handler.EditQuestion)
	mux.HandleFunc("/view-question", handler.ViewQuestion)
	mux.HandleFunc("/submit-question", handler.SubmitQuestion)
	mux.HandleFunc("/submissions", handler.Submissions)
	mux.HandleFunc("/submissions/submit", handler.HandleCodeSubmission)
	mux.HandleFunc("/profile", handler.Profile)
	mux.HandleFunc("/manage-roles", handler.ManageRoles)
	mux.HandleFunc("/update-role", handler.UpdateRole)
	mux.HandleFunc("/delete-question", handler.DeleteQuestion)
	mux.HandleFunc("/publish-question", handler.PublishQuestion)
	mux.HandleFunc("/logout", handler.Logout)

	// Start the server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
