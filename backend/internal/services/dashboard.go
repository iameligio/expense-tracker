package services

import (
	"time"

	"expense-tracker/backend/internal/models"
	"expense-tracker/backend/internal/repository"

	"github.com/shopspring/decimal"
)

// DashboardService assembles the dashboard payload from the repositories.
type DashboardService struct {
	users    *repository.UserRepository
	expenses *repository.ExpenseRepository
	settings *repository.SettingRepository
}

// NewDashboardService builds a DashboardService.
func NewDashboardService(u *repository.UserRepository, e *repository.ExpenseRepository, s *repository.SettingRepository) *DashboardService {
	return &DashboardService{users: u, expenses: e, settings: s}
}

// CategorySlice is one pie-chart segment.
type CategorySlice struct {
	CategoryID string              `json:"categoryId"`
	Name       string              `json:"name"`
	Type       models.CategoryType `json:"type"`
	Total      decimal.Decimal     `json:"total"`
}

// TypeSlice aggregates spend by the four budgeting buckets.
type TypeSlice struct {
	Type  models.CategoryType `json:"type"`
	Total decimal.Decimal     `json:"total"`
}

// Dashboard is the full response for GET /api/dashboard.
type Dashboard struct {
	Month             string          `json:"month"`
	Summary           BudgetSummary   `json:"summary"`
	CategoryBreakdown []CategorySlice `json:"categoryBreakdown"`
	TypeBreakdown     []TypeSlice     `json:"typeBreakdown"`
}

// Build computes the dashboard for a user and month range [from, to).
func (s *DashboardService) Build(userID, month string, from, to time.Time) (*Dashboard, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	setting, err := s.settings.Get()
	if err != nil {
		return nil, err
	}
	total, err := s.expenses.TotalForUserMonth(userID, from, to)
	if err != nil {
		return nil, err
	}
	rows, err := s.expenses.CategoryBreakdown(userID, from, to)
	if err != nil {
		return nil, err
	}

	catSlices := make([]CategorySlice, 0, len(rows))
	typeTotals := map[models.CategoryType]decimal.Decimal{}
	for _, r := range rows {
		catSlices = append(catSlices, CategorySlice{
			CategoryID: r.CategoryID, Name: r.Name, Type: r.Type, Total: r.Total,
		})
		typeTotals[r.Type] = typeTotals[r.Type].Add(r.Total)
	}

	typeSlices := make([]TypeSlice, 0, len(typeTotals))
	for _, t := range []models.CategoryType{models.CategoryFixed, models.CategoryVariable, models.CategoryWants, models.CategoryDebts} {
		if v, ok := typeTotals[t]; ok {
			typeSlices = append(typeSlices, TypeSlice{Type: t, Total: v})
		}
	}

	return &Dashboard{
		Month:             month,
		Summary:           ComputeBudget(user.MonthlyIncome, total, setting),
		CategoryBreakdown: catSlices,
		TypeBreakdown:     typeSlices,
	}, nil
}
