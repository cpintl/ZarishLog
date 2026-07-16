package zarishlog

import "time"

type MonthlyConsumption struct {
	Year  int
	Month time.Month
	Qty   float64
}

// CalculateAMC computes Average Monthly Consumption from monthly data.
func CalculateAMC(monthlyData []MonthlyConsumption, windowMonths int) float64 {
	if len(monthlyData) == 0 {
		return 0
	}

	n := windowMonths
	if n > len(monthlyData) {
		n = len(monthlyData)
	}

	data := monthlyData[len(monthlyData)-n:]

	var total float64
	for _, d := range data {
		total += d.Qty
	}

	return total / float64(n)
}

// CalculateReorderPoint computes reorder point = AMC * leadTimeDays / 30 + safetyStock
func CalculateReorderPoint(amc float64, leadTimeDays int, safetyStock float64) float64 {
	if amc <= 0 {
		return 0
	}
	return (amc * float64(leadTimeDays) / 30) + safetyStock
}
