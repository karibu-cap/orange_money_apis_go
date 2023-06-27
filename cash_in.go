package orange_money_apis

import "github.com/go-playground/validator/v10"
import "regexp"

type CashInParams struct {
	CustomerKey    string `validate:"required"`
	CustomerSecret string `validate:"required"`
	XAuthToken     string `validate:"required"`
	MerchantNumber string `validate:"required"`
	Pin            string `validate:"required"`
	IsProd         bool
	Logger         *DebugLogger
}

type InitializeCashInParams struct {
    notificationUrl string `validate:"required,datauri"`
    amount uint32 `validate:"required"` // todo: check if 0 is valid.
    referenceId, comment string `validate:"required"`
    buyerAccountPhone string `validate:"required,omNumber"`
}
    
type CashIn struct {
	config CashInParams
}

const (
	apiLocationProd string = "https://api-s1.orange.cm"
	apiLocationDev         = "https://mockapi.taurs.dev/karibu-cap/orange_money_apis"
)

const omNumberRegex = "^(69\\d{7}|65[5-9]\\d{6})$"

func isOmNumber(fl validator.FieldLevel) bool {
		value := fl.Field().Interface().(string)

        haveMatch, _ := regexp.MatchString(omNumberRegex, value)
        
        return haveMatch
	}

func (this *CashIn) getApiLocation() string {
	if this.config.IsProd {
		return apiLocationProd
	}
	return apiLocationDev
}

func initializeCashIn(config InitializeCashInParams) (interface{}, error) {
    validate := validator.New()
    validate.RegisterValidation("omNumber", isOmNumber)
    err := validate.Struct(config)
    
    if err != nil {
        return nil, err
    }

    
}

func New(config CashInParams) (*CashIn, *validator.ValidationErrors) {
	validate := validator.New()
	err := validate.Struct(config)

	if validate.Struct(config) != nil {
		validationErrors, _ := err.(validator.ValidationErrors)
		return nil, &validationErrors
	}

	return &CashIn{config: config}, nil
}

