package orange_money_apis

import (
	"encoding/base64"
	"errors"
	"fmt"
)

type u struct{}

var utils = u{}

const (
	PendingStatus   int8 = 1
	SucceededStatus      = 2
	FailedStatus         = 0
)

func (*u) hash(key, secret string) string {
	format := fmt.Sprintf("%s:%s", key, secret)
	return base64.StdEncoding.EncodeToString([]byte(format))
}

func (*u) join(a ...any) string {
	return fmt.Sprint(a)
}

func (*u) newError(a ...any) error {
	return errors.New(fmt.Sprint(a))
}

func getStatusFromProviderRawStatus(rawStatus string) int8 {
	switch rawStatus {
	case "PENDING":
	case "INITIATED":
		return PendingStatus
	case "SUCCESSFULL":
	case "SUCCESS":
		return SucceededStatus
	case "CANCELLED":
	case "EXPIRED":
	case "FAILED":
		return FailedStatus
	}
	return -1
}
