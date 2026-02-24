package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/glebarez/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maxbrt/steamzip-app/internal/handlers"
	"github.com/maxbrt/steamzip-app/internal/models"
	"gorm.io/gorm"
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

// newHandler returns a Handler wired to an isolated in-memory SQLite database
// and a fake S3 server that returns 200 for every request.
func newHandler(t *testing.T) *handlers.Handler {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Session{}, &models.Asset{}); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	fakeS3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(fakeS3.Close)

	cfg := aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""),
	}
	s3c := awss3.NewFromConfig(cfg, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(fakeS3.URL)
		o.UsePathStyle = true
	})

	return handlers.NewHandler(db, s3c, "test-bucket", "", "", "http://localhost:8788")
}

// withSessionParam injects sessionId into the chi route context.
func withSessionParam(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("sessionId", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func createSession(t *testing.T, db *gorm.DB) models.Session {
	t.Helper()
	s := models.Session{ExpiresAt: time.Now().Add(24 * time.Hour)}
	if err := db.Create(&s).Error; err != nil {
		t.Fatalf("create session: %v", err)
	}
	return s
}

// validFocalPointsBody returns a JSON body with all 14 focal points set to valid values.
func validFocalPointsBody() *bytes.Buffer {
	const pt = `{"x":0.5,"y":0.5,"zoom":1}`
	return bytes.NewBufferString(`{"focalPoints":{` +
		`"header_capsule":` + pt + `,"small_capsule":` + pt + `,"main_capsule":` + pt +
		`,"vertical_capsule":` + pt + `,"screenshots":` + pt + `,"page_background":` + pt +
		`,"shortcut_icon":` + pt + `,"app_icon":` + pt + `,"library_capsule":` + pt +
		`,"library_hero":` + pt + `,"library_logo":` + pt + `,"library_header_capsule":` + pt +
		`,"event_cover":` + pt + `,"event_header":` + pt + `}}`)
}

func createAsset(t *testing.T, db *gorm.DB, sessionID uuid.UUID, status models.AssetStatus) models.Asset {
	t.Helper()
	a := models.Asset{
		SessionID:     sessionID,
		MimeType:      "image/png",
		Status:        status,
		MasterFileKey: "masters/test",
	}
	if err := db.Create(&a).Error; err != nil {
		t.Fatalf("create asset: %v", err)
	}
	return a
}

// ─── Create Session ───────────────────────────────────────────────────────────

func TestCreateSession_Returns201WithSessionID(t *testing.T) {
	h := newHandler(t)
	rec := httptest.NewRecorder()
	h.HandleCreateSession(rec, httptest.NewRequest("POST", "/", nil))

	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d, want 201", rec.Code)
	}

	var resp handlers.CreateSessionResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if _, err := uuid.Parse(resp.SessionID); err != nil {
		t.Errorf("session_id is not a valid UUID: %q", resp.SessionID)
	}
	if resp.Status != models.SessionStatusPaymentPending {
		t.Errorf("status = %q, want %q", resp.Status, models.SessionStatusPaymentPending)
	}
}

// ─── Get Session ──────────────────────────────────────────────────────────────

func TestGetSession_InvalidUUID_Returns400(t *testing.T) {
	h := newHandler(t)
	rec := httptest.NewRecorder()
	h.HandleGetSession(rec, withSessionParam(httptest.NewRequest("GET", "/", nil), "not-a-uuid"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestGetSession_NotFound_Returns404(t *testing.T) {
	h := newHandler(t)
	rec := httptest.NewRecorder()
	h.HandleGetSession(rec, withSessionParam(httptest.NewRequest("GET", "/", nil), uuid.New().String()))
	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rec.Code)
	}
}

func TestGetSession_ExpiredSession_Returns410(t *testing.T) {
	h := newHandler(t)
	s := models.Session{ExpiresAt: time.Now().Add(-time.Hour)}
	h.Database.Create(&s)

	rec := httptest.NewRecorder()
	h.HandleGetSession(rec, withSessionParam(httptest.NewRequest("GET", "/", nil), s.ID.String()))

	if rec.Code != http.StatusGone {
		t.Errorf("got %d, want 410", rec.Code)
	}
}

func TestGetSession_ValidSession_Returns200(t *testing.T) {
	h := newHandler(t)
	s := createSession(t, h.Database)

	rec := httptest.NewRecorder()
	h.HandleGetSession(rec, withSessionParam(httptest.NewRequest("GET", "/", nil), s.ID.String()))

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
}

// ─── Get Upload URL ───────────────────────────────────────────────────────────

func TestGetUploadURL_InvalidUUID_Returns400(t *testing.T) {
	h := newHandler(t)
	rec := httptest.NewRecorder()
	h.HandleGetUploadURL(rec, withSessionParam(httptest.NewRequest("POST", "/", nil), "bad-uuid"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestGetUploadURL_UnsupportedContentType_Returns400(t *testing.T) {
	h := newHandler(t)
	s := createSession(t, h.Database)
	body := bytes.NewBufferString(`{"contentType": "application/pdf"}`)
	rec := httptest.NewRecorder()
	h.HandleGetUploadURL(rec, withSessionParam(httptest.NewRequest("POST", "/", body), s.ID.String()))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestGetUploadURL_ReturnsAssetIDAndPresignedURL(t *testing.T) {
	h := newHandler(t)
	s := createSession(t, h.Database)

	body := bytes.NewBufferString(`{"contentType": "image/png"}`)
	rec := httptest.NewRecorder()
	h.HandleGetUploadURL(rec, withSessionParam(httptest.NewRequest("POST", "/", body), s.ID.String()))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}

	var resp handlers.GetUploadURLResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if _, err := uuid.Parse(resp.AssetID); err != nil {
		t.Errorf("assetId is not a valid UUID: %q", resp.AssetID)
	}
	if resp.UploadURL == "" {
		t.Error("uploadUrl is empty")
	}
}

// ─── Confirm Upload ───────────────────────────────────────────────────────────

func TestConfirmUpload_InvalidUUID_Returns400(t *testing.T) {
	h := newHandler(t)
	rec := httptest.NewRecorder()
	h.HandleConfirmUpload(rec, withSessionParam(httptest.NewRequest("POST", "/", nil), "bad-uuid"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestConfirmUpload_NoAsset_Returns409(t *testing.T) {
	h := newHandler(t)
	s := createSession(t, h.Database)

	rec := httptest.NewRecorder()
	h.HandleConfirmUpload(rec, withSessionParam(httptest.NewRequest("POST", "/", nil), s.ID.String()))

	if rec.Code != http.StatusConflict {
		t.Errorf("got %d, want 409", rec.Code)
	}
}

func TestConfirmUpload_AssetAlreadyUploaded_Returns409(t *testing.T) {
	h := newHandler(t)
	s := createSession(t, h.Database)
	createAsset(t, h.Database, s.ID, models.AssetStatusUploaded)

	rec := httptest.NewRecorder()
	h.HandleConfirmUpload(rec, withSessionParam(httptest.NewRequest("POST", "/", nil), s.ID.String()))

	if rec.Code != http.StatusConflict {
		t.Errorf("got %d, want 409", rec.Code)
	}
}

func TestConfirmUpload_Success_Returns204(t *testing.T) {
	h := newHandler(t)
	s := createSession(t, h.Database)
	createAsset(t, h.Database, s.ID, models.AssetStatusUploading)

	rec := httptest.NewRecorder()
	h.HandleConfirmUpload(rec, withSessionParam(httptest.NewRequest("POST", "/", nil), s.ID.String()))

	if rec.Code != http.StatusNoContent {
		t.Errorf("got %d, want 204", rec.Code)
	}
}

// ─── Process Asset ────────────────────────────────────────────────────────────

func TestProcessAsset_InvalidUUID_Returns400(t *testing.T) {
	h := newHandler(t)
	rec := httptest.NewRecorder()
	h.HandleProcessAsset(rec, withSessionParam(httptest.NewRequest("POST", "/", nil), "bad-uuid"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestProcessAsset_UnpaidSession_Returns402(t *testing.T) {
	h := newHandler(t)
	s := createSession(t, h.Database)
	createAsset(t, h.Database, s.ID, models.AssetStatusUploaded)

	body := bytes.NewBufferString(`{"focalPoints": {}}`)
	rec := httptest.NewRecorder()
	h.HandleProcessAsset(rec, withSessionParam(httptest.NewRequest("POST", "/", body), s.ID.String()))

	if rec.Code != http.StatusPaymentRequired {
		t.Errorf("got %d, want 402", rec.Code)
	}
}

func TestProcessAsset_NoAsset_Returns409(t *testing.T) {
	h := newHandler(t)
	s := createSession(t, h.Database)
	h.Database.Model(&s).Update("status", models.SessionStatusPaid)

	rec := httptest.NewRecorder()
	h.HandleProcessAsset(rec, withSessionParam(httptest.NewRequest("POST", "/", nil), s.ID.String()))

	if rec.Code != http.StatusConflict {
		t.Errorf("got %d, want 409", rec.Code)
	}
}

func TestProcessAsset_AssetNotUploaded_Returns409(t *testing.T) {
	h := newHandler(t)
	s := createSession(t, h.Database)
	h.Database.Model(&s).Update("status", models.SessionStatusPaid)
	createAsset(t, h.Database, s.ID, models.AssetStatusUploading) // still uploading, not yet confirmed

	rec := httptest.NewRecorder()
	h.HandleProcessAsset(rec, withSessionParam(httptest.NewRequest("POST", "/", nil), s.ID.String()))

	if rec.Code != http.StatusConflict {
		t.Errorf("got %d, want 409", rec.Code)
	}
}

func TestProcessAsset_Success_Returns202(t *testing.T) {
	h := newHandler(t)
	s := createSession(t, h.Database)
	h.Database.Model(&s).Update("status", models.SessionStatusPaid)
	createAsset(t, h.Database, s.ID, models.AssetStatusUploaded)

	rec := httptest.NewRecorder()
	h.HandleProcessAsset(rec, withSessionParam(httptest.NewRequest("POST", "/", validFocalPointsBody()), s.ID.String()))

	if rec.Code != http.StatusAccepted {
		t.Errorf("got %d, want 202", rec.Code)
	}
}
