package models

// MonthlyFinance represents the aggregated financial data
type Report struct {
	Month        string  `bson:"month"`
	TotalIncome  float64 `bson:"total_income"`
	TotalExpense float64 `bson:"total_expense"`
	Balance      float64 `bson:"balance"`
}
