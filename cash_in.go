package orange_money_apis

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type CashInParams struct {
	CustomerKey    string `validate:"required"`
	CustomerSecret string `validate:"required"`
	XAuthToken     string `validate:"required"`
	MerchantNumber string `validate:"required,ynoteMerchantNumber"`
	Pin            string `validate:"required"`
	IsProd         bool
	Logger         DebugLogger
}

type InitializeCashInParams struct {
	NotificationUrl      string `validate:"required,http_url"`
	Amount               uint32 `validate:"required"` // todo: check if 0 is valid.
	ReferenceId, comment string `validate:"required"`
	BuyerAccountPhone    string `validate:"required,omNumber"`
}

type CashIn struct {
	Config CashInParams
}

type _PayTokenRes struct {
	Message string `json:"message"`
	Data    struct {
		PayToken string `json:"payToken"`
	} `json:"data"`
}

type _CashInRes struct {
	json    interface{}
	Message string `json:"message"`
	Data    struct {
		Id                int    `json:"id"`
		Createtime        string `json:"createtime"`
		SubscriberMsisdn  string `json:"subscriberMsisdn"`
		Amount            string `json:"amount"`
		PayToken          string `json:"payToken"`
		Txnid             string `json:"txnid"`
		Txnmode           string `json:"txnmode"`
		Inittxnmessage    string `json:"inittxnmessage"`
		Inittxnstatus     string `json:"inittxnstatus"`
		Confirmtxnstatus  string `json:"confirmtxnstatus"`
		Confirmtxnmessage string `json:"confirmtxnmessage"`
		Status            string `json:"status"`
		NotifUrl          string `json:"notifUrl"`
		Description       string `json:"description"`
		ChannelUserMsisdn string `json:"channelUserMsisdn"`
	} `json:"data"`
}

type NewCashInRes struct {
	Raw      string
	Status   int8
	PayToken string
}

const (
	apiLocationProd string = "https://api-s1.orange.cm"
	apiLocationDev         = "https://mockapi.taurs.dev/karibu-cap/orange_money_apis"
)

func (this *CashIn) getApiLocation() string {
	if this.Config.IsProd {
		return apiLocationProd
	}
	return apiLocationDev
}

func (this *CashIn) requestNewPayToken() (payToken string, error error) {
	const loggingID = "requestNewPayToken"

	accessToken, tokenError := requestNewAccesToken(this.Config.CustomerKey, this.Config.CustomerSecret, this.getApiLocation())
	if tokenError != nil {
		return "", tokenError
	}

	header := map[string][]string{
		"X-AUTH-TOKEN":  {this.Config.XAuthToken},
		"Authorization": {utils.join("Bearer ", accessToken)},
	}

	this.Config.Logger.Debug(
		fmt.Sprintf("%s:start", loggingID),
		map[string]any{"message": "Initializing payment(generating pay token)"},
	)
	endPoint := utils.join(this.getApiLocation(), "/omcoreapis/1.0.2/mp/init")

	response, reqError := request.post(endPoint, nil, header)
	if reqError != nil {
		return "", reqError
	}

	if response.status != 200 && response.status != 201 {
		return "", utils.newError(map[string]any{
			"errorMessage": "Failed to request a new payToken",
			"response":     response.asText(),
			"endPoint":     endPoint,
			"body":         nil,
			"header":       header,
		})
	}

	var parsedResponse _PayTokenRes

	jsonError := response.asJson(&parsedResponse)
	if jsonError != nil {
		return "", jsonError
	}

	return parsedResponse.Data.PayToken, nil
}

func (this *CashIn) RequestNewCashIn(config *InitializeCashInParams) (*NewCashInRes, error) {
	const loggingID = "CashIn.RequestNewCashIn"
	this.Config.Logger.Debug(fmt.Sprintf("%s:start", loggingID), nil)

	validate := validator.New()
	validate.RegisterValidation("omNumber", isOmNumber)
	validate.RegisterValidation("ynoteMerchantNumber", isYnoteMerchantNumber)

	err := validate.Struct(config)
	if err != nil {
		return nil, err
	}

	payToken, payTokenResError := this.requestNewPayToken()
	if payTokenResError != nil {
		return nil, payTokenResError
	}

	accessToken, accessTokenError := requestNewAccesToken(this.Config.CustomerKey, this.Config.CustomerSecret, this.getApiLocation())
	if accessTokenError != nil {
		return nil, accessTokenError
	}

	header := map[string][]string{
		"X-AUTH-TOKEN":  {this.Config.XAuthToken},
		"Authorization": {utils.join("Bearer ", accessToken)},
		"Content-Type":  {"application/json"},
	}

	body := map[string]string{
		"subscriberMsisdn":  config.BuyerAccountPhone,
		"notifUrl":          config.NotificationUrl,
		"orderId":           config.ReferenceId,
		"description":       config.comment,
		"amount":            utils.join(config.Amount),
		"channelUserMsisdn": this.Config.MerchantNumber,
		"payToken":          payToken,
		"pin":               this.Config.Pin,
	}

	serializedBody, serializationError := json.Marshal(body)

	if serializationError != nil {
		return nil, serializationError
	}

	this.Config.Logger.Debug(
		fmt.Sprintf("%s:requesting payment", loggingID),
		map[string]any{"header": header, "body": body},
	)
	endPoint := utils.join(this.getApiLocation(), "/omcoreapis/1.0.2/mp/pay")

	response, requestError := request.post(endPoint, serializedBody, header)

	if requestError != nil {
		return nil, requestError
	}

	if response.status != 200 && response.status != 201 {
		return nil, utils.newError(map[string]any{
			"message":   "Cashin request failed",
			"response":  response.asText(),
			"enPoint":   endPoint,
			"reqHeader": header,
			"reqBody":   body,
		})
	}
	this.Config.Logger.Debug(fmt.Sprintf("%s:payment_request_end_with_data", loggingID), response.asText()),

	var parsedResponse _CashInRes
	resUnwrapError := response.asJson(&parsedResponse)
	if resUnwrapError != nil {
		return nil, resUnwrapError
	}

	return &NewCashInRes{
		Status:   getStatusFromProviderRawStatus(parsedResponse.Data.Status),
		PayToken: payToken,
		Raw:      response.asText(),
	}, nil
}

func (this *CashIn) FetchCashInStatus(payToken string) (*NewCashInRes, error) {
	accessToken, accessTokenError := requestNewAccesToken(this.Config.CustomerKey, this.Config.CustomerSecret, this.getApiLocation())
	if accessTokenError != nil {
		return nil, accessTokenError
	}

	header := map[string][]string{
		"X-AUTH-TOKEN":  {this.Config.XAuthToken},
		"Authorization": {utils.join("Bearer ", accessToken)},
	}

	var body []byte

	endPoint := utils.join(this.getApiLocation(), "/omcoreapis/1.0.2/mp/paymentstatus/", payToken)

	response, err := request.post(endPoint, body, header)
	if err != nil {
		return nil, err
	}

	if response.status != 200 && response.status != 201 {
		return nil, utils.newError(map[string]any{
			"message":   "Failed to retreive the status of the requested cash in.",
			"response":  response.asText(),
			"endPoint":  endPoint,
			"reqBody":   body,
			"reqHeader": header,
		})
	}

	var parsedResponse _CashInRes
	resUnwrapError := response.asJson(parsedResponse)
	if resUnwrapError != nil {
		return nil, resUnwrapError
	}

	return &NewCashInRes{
		Status:   getStatusFromProviderRawStatus(parsedResponse.Data.Status),
		PayToken: payToken,
		Raw:      response.asText(),
	}, nil
}

func New(config CashInParams) (*CashIn, *validator.ValidationErrors) {
	validate := validator.New()
	err := validate.Struct(config)

	if validate.Struct(config) != nil {
		validationErrors, _ := err.(validator.ValidationErrors)
		return nil, &validationErrors
	}

	return &CashIn{Config: config}, nil
}
