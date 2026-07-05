package repository

import (
	"errors"
	"testing"
	"time"

	"expense-tracker/backend/internal/models"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// newIncomeTestDB spins up an isolated in-memory SQLite database migrated for
// income tests. Mirrors newTestDB in expense_repo_test.go.
func newIncomeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Income{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// seedUserWithIncome creates a user and one income entry, returning both ids.
func seedUserWithIncome(t *testing.T, db *gorm.DB, email, note string) (userID, incomeID string) {
	t.Helper()
	u := &models.User{Email: email, PasswordHash: "x"}
	mustCreate(t, db, u)
	n := note
	i := &models.Income{
		Amount:     decimal.RequireFromString("52000.00"),
		Note:       &n,
		ReceivedOn: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		Source:     models.IncomeSalary,
		UserID:     u.ID,
	}
	mustCreate(t, db, i)
	return u.ID, i.ID
}

// TestIncomeOwnershipEnforced verifies User A can never read, edit, or delete
// User B's income. Every cross-user op must return ErrNotFound and leave B's
// data untouched — the same authorization guarantee proven for expenses.
func TestIncomeOwnershipEnforced(t *testing.T) {
	db := newIncomeTestDB(t)
	repo := NewIncomeRepository(db)

	userA := &models.User{Email: "attacker-income@example.com", PasswordHash: "x"}
	mustCreate(t, db, userA)

	_, incomeB := seedUserWithIncome(t, db, "victim-income@example.com", "victim note")

	t.Run("A cannot read B's income", func(t *testing.T) {
		_, err := repo.FindOwned(userA.ID, incomeB)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("FindOwned(A, B's id) err = %v, want ErrNotFound", err)
		}
	})

	t.Run("A cannot edit B's income", func(t *testing.T) {
		err := repo.Update(userA.ID, incomeB, map[string]interface{}{"note": "hacked"})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("Update(A, B's id) err = %v, want ErrNotFound", err)
		}
		var got models.Income
		if err := db.First(&got, "id = ?", incomeB).Error; err != nil {
			t.Fatalf("reload B's income: %v", err)
		}
		if got.Note == nil || *got.Note != "victim note" {
			t.Fatalf("B's note was mutated: %v", got.Note)
		}
	})

	t.Run("A cannot delete B's income", func(t *testing.T) {
		err := repo.Delete(userA.ID, incomeB)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("Delete(A, B's id) err = %v, want ErrNotFound", err)
		}
		var count int64
		db.Model(&models.Income{}).Where("id = ?", incomeB).Count(&count)
		if count != 1 {
			t.Fatalf("B's income count = %d, want 1 (must not be deleted)", count)
		}
	})
}

// TestIncomeOwnerCanManageOwn is the positive control: the real owner CAN read,
// edit, and delete their own income entry.
func TestIncomeOwnerCanManageOwn(t *testing.T) {
	db := newIncomeTestDB(t)
	repo := NewIncomeRepository(db)

	ownerID, incomeID := seedUserWithIncome(t, db, "owner-income@example.com", "original")

	if _, err := repo.FindOwned(ownerID, incomeID); err != nil {
		t.Fatalf("owner FindOwned: %v", err)
	}
	if err := repo.Update(ownerID, incomeID, map[string]interface{}{"note": "updated"}); err != nil {
		t.Fatalf("owner Update: %v", err)
	}
	if err := repo.Delete(ownerID, incomeID); err != nil {
		t.Fatalf("owner Delete: %v", err)
	}
}
