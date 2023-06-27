package orange_money_apis

import (
	"encoding/base64"
	"fmt"
)

type DebugLogger interface {
	Debug(context string, data map[string]string)
}

func hash(key, secret string) string {
    format := fmt.Sprintf("%s:%s", key, secret)
    return base64.StdEncoding.EncodeToString([]byte(format))
}
