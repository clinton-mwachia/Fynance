package helpers

import (
	"fmt"
	"math"
)

func FormatAmount(amount float64) string {
	absAmount := math.Abs(amount)

	switch {
	case absAmount >= 1_000_000_000_000:
		return fmt.Sprintf("%.1fT", amount/1_000_000_000_000)
	case absAmount >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", amount/1_000_000_000)
	case absAmount >= 1_000_000:
		return fmt.Sprintf("%.1fM", amount/1_000_000)
	case absAmount >= 1_000:
		return fmt.Sprintf("%.1fK", amount/1_000)
	default:
		return fmt.Sprintf("%.2f", amount)
	}
}
