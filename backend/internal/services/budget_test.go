package services

import (
	"testing"

	"expense-tracker/backend/internal/models"

	"github.com/shopspring/decimal"
)

func dec(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}

func TestComputeBudget(t *testing.T) {
	tests := []struct {
		name          string
		income        string
		expenses      string
		targetType    models.SavingsTargetType
		targetValue   string
		wantTarget    string
		wantSavings   string
		wantRemaining string
		wantMet       bool
	}{
		{
			name:       "percent target met",
			income:     "50000", expenses: "30000",
			targetType: models.SavingsPercent, targetValue: "20",
			wantTarget: "10000", wantSavings: "20000", wantRemaining: "10000", wantMet: true,
		},
		{
			name:       "percent target not met",
			income:     "50000", expenses: "45000",
			targetType: models.SavingsPercent, targetValue: "20",
			wantTarget: "10000", wantSavings: "5000", wantRemaining: "-5000", wantMet: false,
		},
		{
			name:       "fixed target met exactly",
			income:     "40000", expenses: "32000",
			targetType: models.SavingsFixed, targetValue: "8000",
			wantTarget: "8000", wantSavings: "8000", wantRemaining: "0", wantMet: true,
		},
		{
			name:       "fixed target overspent",
			income:     "40000", expenses: "39000",
			targetType: models.SavingsFixed, targetValue: "8000",
			wantTarget: "8000", wantSavings: "1000", wantRemaining: "-7000", wantMet: false,
		},
		{
			name:       "zero income",
			income:     "0", expenses: "0",
			targetType: models.SavingsPercent, targetValue: "20",
			wantTarget: "0", wantSavings: "0", wantRemaining: "0", wantMet: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setting := &models.AppSetting{
				SavingsTargetType:  tc.targetType,
				SavingsTargetValue: dec(tc.targetValue),
			}
			got := ComputeBudget(dec(tc.income), dec(tc.expenses), setting)

			if !got.SavingsTarget.Equal(dec(tc.wantTarget)) {
				t.Errorf("SavingsTarget = %s, want %s", got.SavingsTarget, tc.wantTarget)
			}
			if !got.ActualSavings.Equal(dec(tc.wantSavings)) {
				t.Errorf("ActualSavings = %s, want %s", got.ActualSavings, tc.wantSavings)
			}
			if !got.RemainingBudget.Equal(dec(tc.wantRemaining)) {
				t.Errorf("RemainingBudget = %s, want %s", got.RemainingBudget, tc.wantRemaining)
			}
			if got.TargetMet != tc.wantMet {
				t.Errorf("TargetMet = %v, want %v", got.TargetMet, tc.wantMet)
			}
		})
	}
}
