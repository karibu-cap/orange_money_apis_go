package orange_money_apis

import (
    "github.com/go-playground/validator/v10"
 "fmt"
"net/http"
)

type GenerateAccessTokenParams struct {
	key    string `validate:"required"`
	secret string `validate:"required"`
    endPoint string `validate:"required,datauri"`
	logger   *DebugLogger
}


func generateAccesToken(config GenerateAccessTokenParams) (accessToken *string, requestError error) {
	validate := validator.New()
	validationError := validate.Struct(config)

	if validate.Struct(config) != nil {
		validationErrors, _ := validationError.(validator.ValidationErrors)
		return nil, &validationErrors
	}
   
    tokenPath := fmt.Sprint(config.endPoint, "/token")
    
    req, requestError := http.NewRequest("POST", tokenPath, nil)
    
    if requestError == nil {
        return nil, requestError
    }

    req.PostForm.Add("grant_type", "client_credentials")
    basicKey := hash(config.key, config.secret)
    req.Header.Add("Authorization", fmt.Sprintf("Basic %s", basicKey)) 
    
    httpClient := &http.Client{}
    response, postError := httpClient.Do(req)
    
    if postError != nil {
        return nil, postError
    }

    defer response.Body.Close()
    
}
 
