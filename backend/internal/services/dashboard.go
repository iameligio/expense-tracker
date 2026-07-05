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
	incomes  *repository.IncomeRepository
	settings *repository.SettingRepository
}

// NewDashboardService builds a DashboardService.
func NewDashboardService(u *repository.UserRepository, e *repository.ExpenseRepository, i *repository.IncomeRepository, s *repository.SettingRepository) *DashboardService {
	return &DashboardService{users: u, expenses: e, incomes: i, settings: s}
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

// IncomeSlice aggregates income by source.
type IncomeSlice struct {
	Source models.IncomeSource `json:"source"`
	Total  decimal.Decimal     `json:"total"`
}

// Dashboard is the full response for GET /api/dashboard.
type Dashboard struct {
	Month             string          `json:"month"`
	Summary           BudgetSummary   `json:"summary"`
	CategoryBreakdown []CategorySlice `json:"categoryBreakdown"`
	TypeBreakdown     []TypeSlice     `json:"typeBreakdown"`
	IncomeBreakdown   []IncomeSlice   `json:"incomeBreakdown"`
}

// Build computes the dashboard for a user and month range [from, to).
func (s *DashboardService) Build(userID, month string, from, to time.Time) (*Dashboard, error) {
	setting, err := s.settings.Get()
	if err != nil {
		return nil, err
	}
	total, err := s.expenses.TotalForUserMonth(userID, from, to)
	if err != nil {
		return nil, err
	}
	income, err := s.incomes.TotalForUserMonth(userID, from, to)
	if err != nil {
		return nil, err
	}
	rows, err := s.expenses.CategoryBreakdown(userID, from, to)
	if err != nil {
		return nil, err
	}
	incomeRows, err := s.incomes.SourceBreakdown(userID, from, to)
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

	incomeSlices := make([]IncomeSlice, 0, len(incomeRows))
	for _, r := range incomeRows {
		incomeSlices = append(incomeSlices, IncomeSlice{Source: r.Source, Total: r.Total})
	}

	return &Dashboard{
		Month:             month,
		Summary:           ComputeBudget(income, total, setting),
		CategoryBreakdown: catSlices,
		TypeBreakdown:     typeSlices,
		IncomeBreakdown:   incomeSlices,
	}, nil
}
