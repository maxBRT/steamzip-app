package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/maxbrt/steamzip-app/internal/models"
	"github.com/maxbrt/steamzip-app/internal/utils"
	"github.com/stripe/stripe-go/v84"
	stripeSession "github.com/stripe/stripe-go/v84/checkout/session"
)

type CheckoutRequest struct {
	FocalPoints SteamAssetFocalPoints `json:"focal_points"`
}

type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
}

func (h *Handler) HandleCreateCheckout(w http.ResponseWriter, r *http.Request) {
	session, ok := h.loadSession(w, r)
	if !ok {
		return
	}

	if session.Status != models.SessionStatusPaymentPending {
		http.Error(w, "Session is not awaiting payment", http.StatusConflict)
		return
	}

	var body CheckoutRequest
	if err := utils.ReadJSON(r, &body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Persist focal points to the asset so the webhook can trigger processing
	if session.Asset != nil {
		fpJSON, err := json.Marshal(body.FocalPoints)
		if err != nil {
			http.Error(w, "Failed to encode focal points", http.StatusInternalServerError)
			return
		}
		fpStr := string(fpJSON)
		if err := h.Database.Model(session.Asset).Update("focal_points", fpStr).Error; err != nil {
			log.Printf("failed to save focal points for asset %s: %v", session.Asset.ID, err)
			http.Error(w, "Failed to save focal points", http.StatusInternalServerError)
			return
		}
	}

	stripe.Key = h.StripeSecretKey

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String("usd"),
					UnitAmount: stripe.Int64(900), // $9.00 in cents
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("SteamZip Pack"),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		AllowPromotionCodes: stripe.Bool(true),
		SuccessURL:          stripe.String(h.FrontendURL + "/processing"),
		CancelURL:           stripe.String(h.FrontendURL + "/focal-points"),
		Metadata: map[string]string{
			"session_id": session.ID.String(),
		},
	}

	cs, err := stripeSession.New(params)
	if err != nil {
		log.Printf("stripe checkout session creation failed: %v", err)
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	if err := utils.WriteJSON(w, http.StatusOK, CheckoutResponse{CheckoutURL: cs.URL}); err != nil {
		log.Println(err)
	}
}
