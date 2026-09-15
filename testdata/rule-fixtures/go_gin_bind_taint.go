package fixtures

import "database/sql"

type ginFilter struct {
	AccountID string
}

type ginContext struct{}

func (c *ginContext) ShouldBindQuery(dst any) error { return nil }

func findBoundAccount(c *ginContext, db *sql.DB) {
	var input ginFilter
	_ = c.ShouldBindQuery(&input)
	query := "SELECT * FROM accounts WHERE account_id = '" + input.AccountID + "'"
	// EXPECT: finguard.go.sql-sprintf
	db.Query(query)
}
