package services

import (
	"expense-tracker/backend/internal/models"

	"github.com/shopspring/decimal"
)

// BudgetSummary holds the computed dashboard KPIs.
type BudgetSummary struct {
	Income          decimal.Decimal `json:"income"`
	TotalExpenses   decimal.Decimal `json:"totalExpenses"`
	SavingsTarget   decimal.Decimal `json:"savingsTarget"`
	ActualSavings   decimal.Decimal `json:"actualSavings"`
	RemainingBudget decimal.Decimal `json:"remainingBudget"`
	TargetMet       bool            `json:"targetMet"`
}

var oneHundred = decimal.NewFromInt(100)

// ComputeBudget derives the savings/budget KPIs from income, spend, and the
// global savings-target policy. Pure function — unit-tested directly.
func ComputeBudget(income, totalExpenses decimal.Decimal, setting *models.AppSetting) BudgetSummary {
	var target decimal.Decimal
	switch setting.SavingsTargetType {
	case models.SavingsFixed:
		target = setting.SavingsTargetValue
	default: // percent
		target = income.Mul(setting.SavingsTargetValue).Div(oneHundred)
	}

	actualSavings := income.Sub(totalExpenses)
	// Remaining budget = income minus the amount we intend to save minus what we've spent.
	remaining := income.Sub(target).Sub(totalExpenses)

	return BudgetSummary{
		Income:          income,
		TotalExpenses:   totalExpenses,
		SavingsTarget:   target,
		ActualSavings:   actualSavings,
		RemainingBudget: remaining,
		TargetMet:       actualSavings.GreaterThanOrEqual(target),
	}
}
