package orange_money_apis

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
)

type YNoteRefundApiConfig struct {
	ClientId          string `validate:"required"`
	ClientSecret      string `validate:"required"`
	CustomerKey       string `validate:"required"`
	CustomerSecret    string `validate:"required"`
	ChannelUserMsisdn string `validate:"required,ynoteMerchantNumber"`
	Pin               string `validate:"required"`
	IsProd            bool
	Logger            *DebugLogger
}

type RequestNewRefundParams struct {
	NotificationUrl                  string `validate:"required,datauri"`
	Amount                           uint32 `validate:"required"` // todo: check if 0 is valid.
	ReferenceId, CustomerAccountName string `validate:"required"`
	CustomerAccountPhone             string `validate:"required,omNumber"`
}

type YNoteRefundApi struct {
	config YNoteRefundApiConfig
}

type _RequestNewRefundApiRes struct {
	MD5OfMessageBody       string `json:"MD5OfMessageBody"`
	MD5OfMessageAttributes int    `json:"MD5OfMessageAttributes"`
	MessageId              string `json:"MessageId"`
	ResponseMetadata       struct {
		RequestId      string `json:"RequestId"`
		HTTPStatusCode int    `json:"HTTPStatusCode"`
		RetryAttempts  int    `json:"txnid"`
		HTTPHeaders    struct {
			XAmznRequestId string `json:"x-amzn-requestid"`
			XAmznTraceId   string `json:"x-amzn-trace-id"`
			Date           string `json:"date"`
			ContentType    string `json:"content-type"`
			ContentLength  string `json:"content-length"`
		} `json:"HTTPHeaders"`
	} `json:"ResponseMetadata"`
}

type _FetchRefundStatus struct {
	Result     _CashInRes `json:"result"`
	CreateAt   string     `json:"CreateAt"`
	MessageId  string     `json:"MessageId"`
	RefundStep string     `json:"RefundStep"`
	Parameters struct {
		Amount               string `json:"amount"`
		Xauth                string `json:"xauth"`
		Channel_user_msisdn  string `json:"channel_user_msisdn"`
		Customer_key         string `json:"customer_key"`
		Customer_secret      string `json:"customer_secret"`
		Final_customer_name  string `json:"final_customer_name"`
		Final_customer_phone string `json:"final_customer_phone"`
	} `json:"parameters"`
}

type RequestNewRefundRes struct {
	Raw       _RequestNewRefundApiRes
	MessageId string
}

type FetchRefundStatusRes struct {
	Raw        _FetchRefundStatus
	Status     int8
	RefundStep string
}

const (
	ynoteApiHost         string = "https://omapi.ynote.africa"
	yNoteApiTokenHost           = "https://omapi-token.ynote.africa/oauth2"
	TransferSent                = '2'
	InitializingTransfer        = '1'
)

func (this *YNoteRefundApi) getApiEnv() string {
	if this.config.IsProd {
		return "prod"
	}
	return "dev"
}

func (this *YNoteRefundApi) RequestNewRefund(config RequestNewRefundParams) (*RequestNewRefundRes, error) {
	validate := validator.New()
	validate.RegisterValidation("omNumber", isOmNumber)

	err := validate.Struct(config)
	if err != nil {
		return nil, err
	}

	accessToken, accessTokenError := requestNewAccesToken(this.config.ClientId, this.config.ClientSecret, yNoteApiTokenHost)
	if accessTokenError != nil {
		return nil, accessTokenError
	}

	header := map[string][]string{
		"Authorization": {utils.join("Bearer ", accessToken)},
		"Content-Type":  {"application/json"},
	}

	body := map[string]string{
		"pin":                  this.config.Pin,
		"customerkey":          this.config.CustomerKey,
		"customersecret":       this.config.CustomerSecret,
		"channelUserMsisdn":    this.config.ChannelUserMsisdn,
		"amount":               utils.join(config.Amount),
		"webhook":              config.NotificationUrl,
		"final_customer_phone": config.CustomerAccountPhone,
		"final_customer_name":  config.CustomerAccountName,
		"refund_method":        "OrangeMoney",
	}

	serializedBody, serializationError := json.Marshal(body)

	if serializationError != nil {
		return nil, serializationError
	}

	endPoint := utils.join(ynoteApiHost, "/", this.getApiEnv(), "/refund")

	response, requestError := request.post(endPoint, serializedBody, header)

	if requestError != nil {
		return nil, requestError
	}

	if response.status != 200 && response.status != 201 {
		return nil, utils.newError(map[string]any{
			"message":   "Cashout request failed",
			"response":  response.asText(),
			"enPoint":   endPoint,
			"reqHeader": header,
			"reqBody":   body,
		})
	}

	var parsedResponse _RequestNewRefundApiRes
	resUnwrapError := response.asJson(parsedResponse)
	if resUnwrapError != nil {
		return nil, resUnwrapError
	}

	return &RequestNewRefundRes{
		MessageId: parsedResponse.MessageId,
		Raw:       parsedResponse,
	}, nil
}

func (this *YNoteRefundApi) FetchRefundStatus(messageId string) (*FetchRefundStatusRes, error) {
	accessToken, accessTokenError := requestNewAccesToken(this.config.ClientId, this.config.ClientSecret, yNoteApiTokenHost)
	if accessTokenError != nil {
		return nil, accessTokenError
	}

	header := map[string][]string{
		"Authorization": {utils.join("Bearer ", accessToken)},
	}

	var body []byte

	endPoint := utils.join(ynoteApiHost, "/", this.getApiEnv(), "/refund/status/", messageId)

	response, err := request.get(endPoint, header)
	if err != nil {
		return nil, err
	}

	if response.status != 200 && response.status != 201 {
		return nil, utils.newError(map[string]any{
			"message":   "Failed to retreive the status of the requested cash out.",
			"response":  response.asText(),
			"endPoint":  endPoint,
			"reqBody":   body,
			"reqHeader": header,
		})
	}

	var parsedResponse _FetchRefundStatus
	resUnwrapError := response.asJson(parsedResponse)
	if resUnwrapError != nil {
		return nil, resUnwrapError
	}

	return &FetchRefundStatusRes{
		Status:     getStatusFromProviderRawStatus(parsedResponse.Result.Data.Status),
		RefundStep: parsedResponse.RefundStep,
		Raw:        parsedResponse,
	}, nil
}

func NewYNoteRefund(config YNoteRefundApiConfig) (*YNoteRefundApi, *validator.ValidationErrors) {
	validate := validator.New()
	validate.RegisterValidation("ynoteMerchantNumber", isyNoteMerchantNumber)

	err := validate.Struct(config)
	if err != nil {
		validationErrors, _ := err.(validator.ValidationErrors)
		return nil, &validationErrors
	}

	return &YNoteRefundApi{config: config}, nil
}
