package matching

const DefaultCommissionRate = 0.001 // 0.1%

// CalculateFee calculates the transaction cost.
//
// Formula:
// Fee = Price × Quantity × CommissionRate
func CalculateFee(
	price float64,
	quantity float64,
	commissionRate float64,
) float64 {

	if price <= 0 || quantity <= 0 || commissionRate <= 0 {
		return 0
	}

	return price * quantity * commissionRate
}
