package graph

import (
	"ExpenseServer/middleware"
	"context"
	"database/sql"
	"errors"
)

func getUserID(ctx context.Context) (int64, error) {
	val := ctx.Value(middleware.UserIDKey)
	if val == nil {
		return 0, errors.New("unauthorized mapping")
	}
	return val.(int64), nil
}

func updateBalanceByNames(db *sql.DB, userID int64, accountName string, amount float64, categoryName string, dateStr string) {
	var catType string
	err := db.QueryRow("SELECT categories_type FROM categories WHERE name = ? AND user_id = ?", categoryName, userID).Scan(&catType)
	if err != nil {
		return
	}

	if catType == "EXPENSE" {
		_, _ = db.Exec("UPDATE accounts SET amount = amount - ? WHERE name = ? AND user_id = ?", amount, accountName, userID)
	} else if catType == "INCOME" {
		_, _ = db.Exec("UPDATE accounts SET amount = amount + ? WHERE name = ? AND user_id = ?", amount, accountName, userID)
	}

	monthStr := dateStr[:7]
	spending, income := 0.0, 0.0
	if catType == "EXPENSE" {
		spending = amount
	} else if catType == "INCOME" {
		income = amount
	}

	var meaID int64
	err = db.QueryRow("SELECT id FROM monthly_expense_analysis WHERE user_id = ? AND month = ?", userID, monthStr).Scan(&meaID)
	if err == sql.ErrNoRows {
		_, _ = db.Exec("INSERT INTO monthly_expense_analysis (user_id, spending, income, month) VALUES (?, ?, ?, ?)",
			userID, spending, income, monthStr)
	} else {
		_, _ = db.Exec("UPDATE monthly_expense_analysis SET spending = spending + ?, income = income + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			spending, income, meaID)
	}
}

func revertBalanceByNames(db *sql.DB, userID int64, accountName string, amount float64, categoryName string, dateStr string) {
	var catType string
	err := db.QueryRow("SELECT categories_type FROM categories WHERE name = ? AND user_id = ?", categoryName, userID).Scan(&catType)
	if err != nil {
		return
	}

	if catType == "EXPENSE" {
		_, _ = db.Exec("UPDATE accounts SET amount = amount + ? WHERE name = ? AND user_id = ?", amount, accountName, userID)
	} else if catType == "INCOME" {
		_, _ = db.Exec("UPDATE accounts SET amount = amount - ? WHERE name = ? AND user_id = ?", amount, accountName, userID)
	}

	monthStr := dateStr[:7]
	spending, income := 0.0, 0.0
	if catType == "EXPENSE" {
		spending = -amount
	} else if catType == "INCOME" {
		income = -amount
	}

	var meaID int64
	err = db.QueryRow("SELECT id FROM monthly_expense_analysis WHERE user_id = ? AND month = ?", userID, monthStr).Scan(&meaID)
	if err == nil {
		_, _ = db.Exec("UPDATE monthly_expense_analysis SET spending = spending + ?, income = income + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			spending, income, meaID)
	}
}
