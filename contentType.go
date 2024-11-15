package burptomap

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type ContentType int

const (
	FormData ContentType = iota
	ApplicationJSON

	UnsupportedContentType ContentType = -1
)

func detectContentType(req *http.Request) (ContentType, error) {
	bodyBytes, err := readFromBody(req)
	if err != nil {
		return UnsupportedContentType, err
	}

	// Restore the body
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var jsonObject map[string]interface{}
	jsonErr := json.Unmarshal(bodyBytes, &jsonObject)
	if jsonErr == nil {
		return ApplicationJSON, nil
	}

	_, formDataErr := url.ParseQuery(string(bodyBytes))
	if formDataErr == nil {
		return FormData, nil
	}

	return UnsupportedContentType, nil
}
