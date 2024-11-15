package burptomap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func MarkAllInjectionPoints(requestString string) (int, string, error) {
	httpReq, isHTTP2Detected, err := loadRequestFromString(requestString)
	if err != nil {
		return 0, "", nil
	}

	finalInjectionCount := 0

	// Do all conversion stuff here
	switch httpReq.Method {
	case "GET", "DELETE":
		// just modify the query parameters
		finalInjectionCount += handleQueryParameters(httpReq)
	case "POST", "PUT", "PATCH":
		// first modify the query parameters
		iC := handleQueryParameters(httpReq)
		finalInjectionCount += iC

		contentType, err := detectContentType(httpReq)
		if err != nil {
			return 0, "", err
		}

		if contentType == ApplicationJSON {
			iC, err := handleJSONBody(httpReq)
			if err != nil {
				return 0, "", err
			}
			finalInjectionCount += iC

		} else if contentType == FormData {
			iC, err := handleFormBody(httpReq)
			if err != nil {
				return 0, "", err
			}
			finalInjectionCount += iC
		}
	}

	// Convert the modified request back to bytes.
	finalReqString, err := dumpRequestToString(httpReq, isHTTP2Detected)
	if err != nil {
		return 0, "", err
	}

	return finalInjectionCount, finalReqString, nil
}

// handleQueryParameters adds asteriks to all query parameters
func handleQueryParameters(req *http.Request) int {
	injectionCount := 0

	queryParams := req.URL.Query()
	var modifiedParams []string
	for key, values := range queryParams {
		for _, value := range values {
			// Modify each query parameter value by adding an asterisk at the end without URL encoding
			modifiedParam := fmt.Sprintf("%s=%s*", key, value)
			modifiedParams = append(modifiedParams, modifiedParam)
			injectionCount += 1
		}
	}
	// Re-encode the modified query parameters back into the URL
	req.URL.RawQuery = strings.Join(modifiedParams, "&")

	return injectionCount
}

// hadnleJSONBody adds asteriks to all string values if content json parseable
func handleJSONBody(req *http.Request) (int, error) {
	body, err := readFromBody(req)
	if err != nil {
		return 0, err
	}

	injectionCount := 0
	var jsonObject map[string]interface{}
	if err := json.Unmarshal(body, &jsonObject); err != nil {
		return 0, err
	}

	for key, value := range jsonObject {
		if str, ok := value.(string); ok {
			jsonObject[key] = str + "*"
			injectionCount += 1
		}
	}

	modifiedBody, err := json.Marshal(jsonObject)
	if err != nil {
		return 0, err
	}

	writeToBody(req, modifiedBody)
	return injectionCount, nil
}

// handleFormBody adds asteriks to all body attributes if content parseable as formdata
func handleFormBody(req *http.Request) (int, error) {
	formData, err := readFromBody(req)
	if err != nil {
		return 0, err
	}

	values, err := url.ParseQuery(string(formData))
	if err != nil {
		return 0, err
	}

	injectionCount := 0
	var modifiedValues []string
	for key := range values {
		modifiedValue := fmt.Sprintf("%s=%s*", key, values.Get(key))
		modifiedValues = append(modifiedValues, modifiedValue)
		injectionCount += 1
	}

	modifiedBody := []byte(strings.Join(modifiedValues, "&"))
	writeToBody(req, modifiedBody)
	return injectionCount, nil
}
