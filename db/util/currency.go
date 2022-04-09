package util

const (
	USD = "USD"
	EUR = "EUR"
	RMB = "RMB"
)

//判断是否支持该货币
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, RMB:
		return true
	}
	return false
}
