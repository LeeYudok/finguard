package fixtures

import "database/sql"

type safeGinFilter struct {
	AccountID string
}

type safeGinContext struct{}

func (c *safeGinContext) ShouldBindQuery(dst any) error { return nil }

func findBoundAccountSafely(c *safeGinContext, db *sql.DB) {
	var input safeGinFilter
	_ = c.ShouldBindQuery(&input)
	db.Query("SELECT * FROM accounts WHERE account_id = ?", input.AccountID)
}

func internalConstantHelper(db *sql.DB, accountID string) {
	db.Query("SELECT * FROM accounts WHERE account_id = '" + accountID + "'")
}

func callInternalConstant(db *sql.DB) {
	internalConstantHelper(db, "system-account")
}
