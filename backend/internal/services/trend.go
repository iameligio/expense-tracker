package services

import (
	"time"

	"github.com/shopspring/decimal"
)

// MonthPoint is one month's income/expense totals in a trend series.
type MonthPoint struct {
	Month    string          `json:"month"` // YYYY-MM
	Income   decimal.Decimal `json:"income"`
	Expenses decimal.Decimal `json:"expenses"`
	Net      decimal.Decimal `json:"net"` // income - expenses
}

// Trend is the response for GET /api/trends: one dense point per month in the
// inclusive [from, to] range (months with no activity are zero-filled).
type Trend struct {
	From   string       `json:"from"`
	To     string       `json:"to"`
	Months []MonthPoint `json:"months"`
}

// BuildTrend sums income and expenses per month across [from, to] inclusive.
// from and to are first-of-month timestamps in Local time. Gaps are filled with
// zeros so the series is dense and safe to plot directly.
func (s *DashboardService) BuildTrend(userID string, from, to time.Time) (*Trend, error) {
	points := make([]MonthPoint, 0)
	for m := from; !m.After(to); m = m.AddDate(0, 1, 0) {
		next := m.AddDate(0, 1, 0)
		income, err := s.incomes.TotalForUserMonth(userID, m, next)
		if err != nil {
			return nil, err
		}
		expenses, err := s.expenses.TotalForUserMonth(userID, m, next)
		if err != nil {
			return nil, err
		}
		points = append(points, MonthPoint{
			Month:    m.Format("2006-01"),
			Income:   income,
			Expenses: expenses,
			Net:      income.Sub(expenses),
		})
	}
	return &Trend{From: from.Format("2006-01"), To: to.Format("2006-01"), Months: points}, nil
}
