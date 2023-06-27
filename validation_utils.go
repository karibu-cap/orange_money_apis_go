package orange_money_apis

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// An orange money phone number regex. that do not authorize the country prefix.
const omNumberRegex = "^(69\\d{7}|65[5-9]\\d{6})$"

// A merchant number regExp authorized by y-note.
const yNoteMerchantNumber = "^(237)?(69\\d{7}$|65[5-9]\\d{6}$)"

func isOmNumber(fl validator.FieldLevel) bool {
	value := fl.Field().Interface().(string)

	haveMatch, _ := regexp.MatchString(omNumberRegex, value)

	return haveMatch
}

func isyNoteMerchantNumber(fl validator.FieldLevel) bool {
	value := fl.Field().Interface().(string)

	haveMatch, _ := regexp.MatchString(yNoteMerchantNumber, value)

	return haveMatch
}
