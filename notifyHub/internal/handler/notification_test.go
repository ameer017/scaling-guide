package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"notifyHub/internal/handler"
	"notifyHub/internal/models"
	"notifyHub/internal/service"
)

type stubAPI struct {
	createFn         func(context.Context, models.CreateNotificationRequest) (*models.Notification, error)
	getFn            func(context.Context, string) (*models.Notification, error)
	listFn           func(context.Context, int) ([]models.Notification, error)
	logsFn           func(context.Context, string) ([]models.DeliveryLog, error)
	createTemplateFn func(context.Context, models.CreateTemplateRequest) (*models.Template, error)
	listTemplatesFn  func(context.Context, int) ([]models.Template, error)
}

func (s *stubAPI) Create(ctx context.Context, req models.CreateNotificationRequest) (*models.Notification, error) {
	return s.createFn(ctx, req)
}
func (s *stubAPI) Get(ctx context.Context, id string) (*models.Notification, error) {
	return s.getFn(ctx, id)
}
func (s *stubAPI) List(ctx context.Context, limit int) ([]models.Notification, error) {
	return s.listFn(ctx, limit)
}
func (s *stubAPI) ListDeliveryLogs(ctx context.Context, id string) ([]models.DeliveryLog, error) {
	return s.logsFn(ctx, id)
}
func (s *stubAPI) CreateTemplate(ctx context.Context, req models.CreateTemplateRequest) (*models.Template, error) {
	return s.createTemplateFn(ctx, req)
}
func (s *stubAPI) ListTemplates(ctx context.Context, limit int) ([]models.Template, error) {
	return s.listTemplatesFn(ctx, limit)
}

func TestHealth(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.NewRouter(&handler.NotificationHandler{}).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

func TestCreateNotificationAccepted(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	api := &stubAPI{
		createFn: func(ctx context.Context, req models.CreateNotificationRequest) (*models.Notification, error) {
			return &models.Notification{
				ID:        "n1",
				Recipient: req.Recipient,
				Subject:   req.Subject,
				Body:      req.Body,
				Status:    models.StatusPending,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	body := `{"recipient":"a@example.com","subject":"Hi","body":"Hello"}`
	req := httptest.NewRequest(http.MethodPost, "/notifications/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.NewRouter(handler.NewNotificationHandler(api)).ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202; body=%s", rr.Code, rr.Body.String())
	}

	var got models.Notification
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != "n1" || got.Status != models.StatusPending {
		t.Fatalf("unexpected response: %+v", got)
	}
}

func TestCreateNotificationInvalidInput(t *testing.T) {
	t.Parallel()

	api := &stubAPI{
		createFn: func(ctx context.Context, req models.CreateNotificationRequest) (*models.Notification, error) {
			return nil, service.ErrInvalidInput
		},
	}

	body := `{"recipient":"","subject":"Hi","body":"Hello"}`
	req := httptest.NewRequest(http.MethodPost, "/notifications/", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	handler.NewRouter(handler.NewNotificationHandler(api)).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestGetNotificationNotFound(t *testing.T) {
	t.Parallel()

	api := &stubAPI{
		getFn: func(ctx context.Context, id string) (*models.Notification, error) {
			return nil, service.ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/notifications/missing", nil)
	rr := httptest.NewRecorder()
	handler.NewRouter(handler.NewNotificationHandler(api)).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

func TestCreateTemplateCreated(t *testing.T) {
	t.Parallel()

	api := &stubAPI{
		createTemplateFn: func(ctx context.Context, req models.CreateTemplateRequest) (*models.Template, error) {
			return &models.Template{
				ID:        "t1",
				Name:      req.Name,
				Subject:   req.Subject,
				Body:      req.Body,
				CreatedAt: time.Now().UTC(),
			}, nil
		},
	}

	body := `{"name":"welcome","subject":"Hi {{name}}","body":"Hello {{name}}"}`
	req := httptest.NewRequest(http.MethodPost, "/templates/", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	handler.NewRouter(handler.NewNotificationHandler(api)).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rr.Code, rr.Body.String())
	}
}
