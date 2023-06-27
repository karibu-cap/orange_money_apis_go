package orange_money_apis

// Makes a request to generate a newaccessToken.
func requestNewAccesToken(key, secret, endPoint string) (accessToken string, requestError error) {
	tokenPath := utils.join(endPoint, "/token")

	basicKey := utils.hash(key, secret)

	header := map[string][]string{
		"Authorization": {utils.join("Basic ", basicKey)},
	}

	body := []byte("grant_type=client_credentials")

	res, requestError := request.post(tokenPath, body, header)

	if requestError != nil {
		return "", requestError
	}

	if res.status != 200 && res.status != 201 {
		return "", utils.newError("Backend failed to generate the access Token with message:", res.asText())
	}

	return res.asText(), nil
}
