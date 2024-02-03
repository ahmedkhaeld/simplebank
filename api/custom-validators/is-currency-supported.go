package customValidator

import "github.com/go-playground/validator/v10"

const (
	USD = "USD"
	EUR = "EUR"
	EGP = "EGP"
	GBP = "GBP"
)

// validateCurrency custom validator function returns true if the currency is valid for the given currency
var ValidateCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	//get the valu of the field
	value, ok := fieldLevel.Field().Interface().(string)
	if !ok {
		return false
	}

	return isSupportedCurrency(value)

}

func isSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, EGP, GBP:
		return true
	}
	return false
}
