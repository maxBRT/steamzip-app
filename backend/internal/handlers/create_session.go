package handlers

import (
	"log"
	"net/http"

	"github.com/maxbrt/steamzip-app/internal/models"
	"github.com/maxbrt/steamzip-app/internal/utils"
)

type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

func (h *Handler) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	newSession := models.Session{}

	if err := h.Database.Create(&newSession).Error; err != nil {
		log.Println(err)
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	response := CreateSessionResponse{
		SessionID: newSession.ID.String(),
		Status:    newSession.Status,
	}

	if err := utils.WriteJSON(w, http.StatusCreated, response); err != nil {
		log.Println(err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
	}
}
