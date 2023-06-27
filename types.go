package orange_money_apis

type DebugLogger interface {
	Debug(context string, data map[string]string)
}
