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

// newTestDB spins up an isolated in-memory SQLite database with the schema
// migrated, so repository queries run against a real SQL engine (not a mock).
// This is what makes the ownership assertions below meaningful: the protection
// lives in the WHERE clause, so only a real query can prove it holds.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Category{}, &models.Expense{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func mustCreate(t *testing.T, db *gorm.DB, v interface{}) {
	t.Helper()
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("seed %T: %v", v, err)
	}
}

// seedUserWithExpense creates a user, a category they own, and one expense,
// returning the user id and the expense id.
func seedUserWithExpense(t *testing.T, db *gorm.DB, email, note string) (userID, expenseID string) {
	t.Helper()
	u := &models.User{Email: email, PasswordHash: "x"}
	mustCreate(t, db, u)
	c := &models.Category{Name: "Groceries", Type: models.CategoryVariable, UserID: u.ID}
	mustCreate(t, db, c)
	n := note
	e := &models.Expense{
		Amount:     decimal.RequireFromString("12.50"),
		Note:       &n,
		SpentOn:    time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		UserID:     u.ID,
		CategoryID: c.ID,
	}
	mustCreate(t, db, e)
	return u.ID, e.ID
}

// TestExpenseOwnershipEnforced verifies the core authorization guarantee:
// User A can never read, edit, or delete User B's expense. Every cross-user
// operation must return ErrNotFound (which the HTTP layer maps to 404) and
// must leave B's data completely untouched. A dropped "AND user_id = ?" in any
// repository query would fail this test.
func TestExpenseOwnershipEnforced(t *testing.T) {
	db := newTestDB(t)
	repo := NewExpenseRepository(db)

	userA := &models.User{Email: "attacker@example.com", PasswordHash: "x"}
	mustCreate(t, db, userA)

	_, expenseB := seedUserWithExpense(t, db, "victim@example.com", "victim note")

	t.Run("A cannot read B's expense", func(t *testing.T) {
		_, err := repo.FindOwned(userA.ID, expenseB)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("FindOwned(A, B's id) err = %v, want ErrNotFound", err)
		}
	})

	t.Run("A cannot edit B's expense", func(t *testing.T) {
		err := repo.Update(userA.ID, expenseB, map[string]interface{}{"note": "hacked"})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("Update(A, B's id) err = %v, want ErrNotFound", err)
		}
		// B's expense must be unchanged.
		var got models.Expense
		if err := db.First(&got, "id = ?", expenseB).Error; err != nil {
			t.Fatalf("reload B's expense: %v", err)
		}
		if got.Note == nil || *got.Note != "victim note" {
			t.Fatalf("B's note was mutated: %v", got.Note)
		}
	})

	t.Run("A cannot delete B's expense", func(t *testing.T) {
		err := repo.Delete(userA.ID, expenseB)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("Delete(A, B's id) err = %v, want ErrNotFound", err)
		}
		// B's expense must still exist.
		var count int64
		db.Model(&models.Expense{}).Where("id = ?", expenseB).Count(&count)
		if count != 1 {
			t.Fatalf("B's expense count = %d, want 1 (must not be deleted)", count)
		}
	})
}

// TestExpenseOwnerCanManageOwn is the positive control: the real owner CAN read,
// edit, and delete their own expense. Without this, a repo that rejects
// everything would also pass the ownership test above.
func TestExpenseOwnerCanManageOwn(t *testing.T) {
	db := newTestDB(t)
	repo := NewExpenseRepository(db)

	ownerID, expenseID := seedUserWithExpense(t, db, "owner@example.com", "original")

	if _, err := repo.FindOwned(ownerID, expenseID); err != nil {
		t.Fatalf("owner FindOwned: %v", err)
	}
	if err := repo.Update(ownerID, expenseID, map[string]interface{}{"note": "updated"}); err != nil {
		t.Fatalf("owner Update: %v", err)
	}
	if err := repo.Delete(ownerID, expenseID); err != nil {
		t.Fatalf("owner Delete: %v", err)
	}
}
