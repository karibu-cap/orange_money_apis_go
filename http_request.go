package orange_money_apis

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Response struct {
	bytes  []byte
	status int
}

func (this *Response) asText() string {
	return string(this.bytes)
}

// Parses the JSON-encoded data and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Unmarshal returns an InvalidUnmarshalError.
func (this *Response) asJson(v any) error {
	return json.Unmarshal(this.bytes, v)
}

type Request struct{}

// Post request to server and return the response of the server.
func (*Request) post(endPoint string, body []byte, header http.Header) (*Response, error) {
	req, requestError := http.NewRequest("POST", endPoint, bytes.NewBuffer(body))

	if requestError != nil {
		return nil, requestError
	}

	req.Header = header
	httpClient := &http.Client{}
	response, postError := httpClient.Do(req)

	if postError != nil {
		return nil, postError
	}

	defer response.Body.Close()

	reqBody, ioError := io.ReadAll(response.Body)

	if ioError != nil {
		return nil, ioError
	}

	return &Response{bytes: reqBody, status: response.StatusCode}, nil
}

// Get request to server and return the response of the server.
func (*Request) get(endPoint string, header http.Header) (*Response, error) {
	req, requestError := http.NewRequest("GET", endPoint, nil)

	if requestError != nil {
		return nil, requestError
	}

	req.Header = header
	httpClient := &http.Client{}
	response, postError := httpClient.Do(req)

	if postError != nil {
		return nil, postError
	}

	defer response.Body.Close()

	reqBody, ioError := io.ReadAll(response.Body)

	if ioError != nil {
		return nil, ioError
	}

	return &Response{bytes: reqBody, status: response.StatusCode}, nil
}

var request Request
