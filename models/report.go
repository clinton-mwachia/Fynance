package models

type Report struct {
	Period       string  `bson:"period"`
	TotalIncome  float64 `bson:"total_income"`
	TotalExpense float64 `bson:"total_expense"`
	Balance      float64 `bson:"balance"`
}
