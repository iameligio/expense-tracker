package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"expense-tracker/backend/internal/auth"
	appmw "expense-tracker/backend/internal/middleware"
	"expense-tracker/backend/internal/models"
	"expense-tracker/backend/internal/repository"
	"expense-tracker/backend/internal/services"

	"github.com/glebarez/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// incomeTestEnv wires the real router, auth middleware, income handlers, and
// dashboard service over in-memory SQLite so the HTTP contract is exercised
// end-to-end (routing, JSON, validation, ownership, dashboard summing).
type incomeTestEnv struct {
	router http.Handler
	tm     *auth.TokenManager
	tokenA string
	tokenB string
	userA  string
	userB  string
}

func newIncomeHTTPEnv(t *testing.T) *incomeTestEnv {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Category{}, &models.Expense{}, &models.Income{}, &models.AppSetting{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// Seed the global savings-target setting the dashboard depends on.
	if err := db.Create(&models.AppSetting{
		ID:                 models.GlobalSettingID,
		SavingsTargetType:  models.SavingsPercent,
		SavingsTargetValue: decimal.NewFromInt(20),
	}).Error; err != nil {
		t.Fatalf("seed setting: %v", err)
	}

	userA := &models.User{Email: "a@example.com", PasswordHash: "x", Status: models.StatusActive, Role: models.RoleMember}
	userB := &models.User{Email: "b@example.com", PasswordHash: "x", Status: models.StatusActive, Role: models.RoleMember}
	if err := db.Create(userA).Error; err != nil {
		t.Fatalf("create A: %v", err)
	}
	if err := db.Create(userB).Error; err != nil {
		t.Fatalf("create B: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	expRepo := repository.NewExpenseRepository(db)
	incRepo := repository.NewIncomeRepository(db)
	setRepo := repository.NewSettingRepository(db)
	dashSvc := services.NewDashboardService(userRepo, expRepo, incRepo, setRepo)

	incH := NewIncomeHandler(incRepo)
	dashH := NewDashboardHandler(dashSvc)

	tm := auth.NewTokenManager([]byte("test-secret-key"), 15*time.Minute, 168*time.Hour)

	// Mirror the real /api wiring so chi.URLParam("id") resolves for PATCH/DELETE.
	r := chi.NewRouter()
	r.Route("/api", func(api chi.Router) {
		api.Use(appmw.RequireAuth(tm, userRepo))
		api.Get("/incomes", incH.List)
		api.Post("/incomes", incH.Create)
		api.Patch("/incomes/{id}", incH.Update)
		api.Delete("/incomes/{id}", incH.Delete)
		api.Get("/dashboard", dashH.Get)
	})

	tokenA, _ := tm.GenerateAccessToken(userA.ID, models.RoleMember)
	tokenB, _ := tm.GenerateAccessToken(userB.ID, models.RoleMember)

	return &incomeTestEnv{
		router: r, tm: tm,
		tokenA: tokenA, tokenB: tokenB,
		userA: userA.ID, userB: userB.ID,
	}
}

func (e *incomeTestEnv) do(t *testing.T, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	return rec
}

func TestIncomeHTTPFlow(t *testing.T) {
	e := newIncomeHTTPEnv(t)

	// 1. A creates a salary income.
	rec := e.do(t, "POST", "/api/incomes", e.tokenA, map[string]string{
		"amount": "52000", "source": "salary", "receivedOn": "2026-07-05",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create salary: status=%d body=%s", rec.Code, rec.Body.String())
	}
	var created models.Income
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ID == "" || created.Source != models.IncomeSalary {
		t.Fatalf("unexpected created income: %+v", created)
	}
	if !created.Amount.Equal(decimal.RequireFromString("52000")) {
		t.Fatalf("amount = %s, want 52000", created.Amount)
	}

	// 2. A creates a side-project income.
	rec = e.do(t, "POST", "/api/incomes", e.tokenA, map[string]string{
		"amount": "8000", "source": "side_project", "receivedOn": "2026-07-10",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create side_project: status=%d body=%s", rec.Code, rec.Body.String())
	}

	// 3. Invalid source is rejected.
	rec = e.do(t, "POST", "/api/incomes", e.tokenA, map[string]string{
		"amount": "100", "source": "bonus", "receivedOn": "2026-07-11",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid source: status=%d, want 400", rec.Code)
	}

	// 4. A lists July income -> 2 entries.
	rec = e.do(t, "GET", "/api/incomes?month=2026-07", e.tokenA, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: status=%d body=%s", rec.Code, rec.Body.String())
	}
	var list []models.Income
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list len = %d, want 2", len(list))
	}

	// 5. Dashboard sums income for the month and breaks it down by source.
	rec = e.do(t, "GET", "/api/dashboard?month=2026-07", e.tokenA, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard: status=%d body=%s", rec.Code, rec.Body.String())
	}
	var dash services.Dashboard
	if err := json.Unmarshal(rec.Body.Bytes(), &dash); err != nil {
		t.Fatalf("decode dashboard: %v", err)
	}
	if !dash.Summary.Income.Equal(decimal.RequireFromString("60000")) {
		t.Fatalf("dashboard income = %s, want 60000", dash.Summary.Income)
	}
	if len(dash.IncomeBreakdown) != 2 {
		t.Fatalf("incomeBreakdown len = %d, want 2", len(dash.IncomeBreakdown))
	}

	// 6. A different month is empty (proves per-month variability).
	rec = e.do(t, "GET", "/api/dashboard?month=2026-08", e.tokenA, nil)
	var augDash services.Dashboard
	_ = json.Unmarshal(rec.Body.Bytes(), &augDash)
	if !augDash.Summary.Income.Equal(decimal.Zero) {
		t.Fatalf("august income = %s, want 0", augDash.Summary.Income)
	}

	// 7. Ownership: B cannot see A's income in a listing.
	rec = e.do(t, "GET", "/api/incomes?month=2026-07", e.tokenB, nil)
	var bList []models.Income
	_ = json.Unmarshal(rec.Body.Bytes(), &bList)
	if len(bList) != 0 {
		t.Fatalf("B sees %d of A's income entries, want 0", len(bList))
	}

	// 8. Ownership: B cannot delete A's income.
	rec = e.do(t, "DELETE", "/api/incomes/"+created.ID, e.tokenB, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("B delete A's income: status=%d, want 404", rec.Code)
	}

	// 9. The owner CAN delete their own income.
	rec = e.do(t, "DELETE", "/api/incomes/"+created.ID, e.tokenA, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("A delete own income: status=%d body=%s", rec.Code, rec.Body.String())
	}

	// 10. Unauthenticated requests are rejected.
	req := httptest.NewRequest("GET", "/api/incomes?month=2026-07", nil)
	unauth := httptest.NewRecorder()
	e.router.ServeHTTP(unauth, req)
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("no-token list: status=%d, want 401", unauth.Code)
	}
}
