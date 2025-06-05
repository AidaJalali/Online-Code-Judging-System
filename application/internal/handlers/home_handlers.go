package handlers

import (
	"html/template"
	"net/http"
)

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
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
