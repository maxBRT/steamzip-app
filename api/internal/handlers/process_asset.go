package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/maxbrt/steamzip-app/internal/models"
	"github.com/maxbrt/steamzip-app/internal/utils"
)

type Point struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}

func (p Point) Validate() error {
	if p.X < 0 || p.X > 1 || p.Y < 0 || p.Y > 1 {
		return fmt.Errorf("coordinates must be in [0, 1]")
	}
	if p.Zoom <= 0 {
		return fmt.Errorf("zoom must be positive")
	}
	return nil
}

type SteamAssetFocalPoints struct {
	HeaderCapsule        Point `json:"header_capsule"`
	SmallCapsule         Point `json:"small_capsule"`
	MainCapsule          Point `json:"main_capsule"`
	VerticalCapsule      Point `json:"vertical_capsule"`
	Screenshots          Point `json:"screenshots"`
	PageBackground       Point `json:"page_background"`
	ShortcutIcon         Point `json:"shortcut_icon"`
	AppIcon              Point `json:"app_icon"`
	LibraryCapsule       Point `json:"library_capsule"`
	LibraryHero          Point `json:"library_hero"`
	LibraryLogo          Point `json:"library_logo"`
	LibraryHeaderCapsule Point `json:"library_header_capsule"`
	EventCover           Point `json:"event_cover"`
	EventHeader          Point `json:"event_header"`
}

func (f SteamAssetFocalPoints) ToMap() map[string]Point {
	return map[string]Point{
		"header_capsule":         f.HeaderCapsule,
		"small_capsule":          f.SmallCapsule,
		"main_capsule":           f.MainCapsule,
		"vertical_capsule":       f.VerticalCapsule,
		"screenshots":            f.Screenshots,
		"page_background":        f.PageBackground,
		"shortcut_icon":          f.ShortcutIcon,
		"app_icon":               f.AppIcon,
		"library_capsule":        f.LibraryCapsule,
		"library_hero":           f.LibraryHero,
		"library_logo":           f.LibraryLogo,
		"library_header_capsule": f.LibraryHeaderCapsule,
		"event_cover":            f.EventCover,
		"event_header":           f.EventHeader,
	}
}

type ProcessingRequest struct {
	FocalPoints SteamAssetFocalPoints `json:"focalPoints"`
}

type ProcessingResponse struct {
	EstimatedTimeSeconds int `json:"estimatedTimeSeconds"`
}

func (h *Handler) HandleProcessAsset(w http.ResponseWriter, r *http.Request) {
	session, ok := h.loadSession(w, r)
	if !ok {
		return
	}

	if session.Status != models.SessionStatusPaid {
		http.Error(w, "Payment required", http.StatusPaymentRequired)
		return
	}

	if session.Asset == nil || session.Asset.Status != models.AssetStatusUploaded {
		http.Error(w, "Asset is not ready for processing", http.StatusConflict)
		return
	}

	var body ProcessingRequest
	if err := utils.ReadJSON(r, &body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, p := range body.FocalPoints.ToMap() {
		if err := p.Validate(); err != nil {
			http.Error(w, "Invalid focal point: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	if err := h.Database.Model(&session.Asset).Update("status", models.AssetStatusProcessing).Error; err != nil {
		log.Println(err)
		http.Error(w, "Error updating asset status", http.StatusInternalServerError)
		return
	}

	asset := *session.Asset
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic in processing job for asset %s: %v", asset.ID, rec)
				h.Database.Model(&asset).Update("status", models.AssetStatusFailed)
			}
		}()
		h.runProcessingJob(asset, body.FocalPoints.ToMap())
	}()

	if err := utils.WriteJSON(w, http.StatusAccepted, ProcessingResponse{
		EstimatedTimeSeconds: 30,
	}); err != nil {
		log.Println(err)
	}
}
