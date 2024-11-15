package burptomap

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type EliminatedResults map[string]struct {
	hash   string
	inputs []string
	Item
}

func Eliminate(root *Items) (EliminatedResults, error) {

	finalResults := EliminatedResults{}

	for _, item := range root.Items {

		httpReq, _, err := loadRequestFromString(item.Request.Value)
		if err != nil {
			return nil, err
		}

		allInputs, err := getAllInputKeys(httpReq)
		if err != nil {
			return nil, err
		}

		reqHash := makeHashOfReq(httpReq, allInputs)

		// If the hash already exists eliminate the request
		if _, ok := finalResults[reqHash]; ok {
			continue
		}

		finalResults[reqHash] = struct {
			hash   string
			inputs []string
			Item
		}{
			hash:   reqHash,
			inputs: allInputs,
			Item:   item,
		}
	}

	return finalResults, nil
}

func makeHashOfReq(req *http.Request, inputs []string) string {
	h := sha256.New()

	message := fmt.Sprintf("%s-%s-%s", req.Method, req.URL.Path, strings.Join(inputs, ","))
	h.Write([]byte(message))

	return string(h.Sum(nil))
}

func getAllInputKeys(req *http.Request) ([]string, error) {
	var allInputs []string

	allInputs = append(allInputs, getQueryKeys(req)...)

	switch req.Method {
	case "POST", "PUT", "PATCH":
		contentType, err := detectContentType(req)
		if err != nil {
			return nil, err
		}

		if contentType == ApplicationJSON {
			jsonKeys, err := getJsonKeys(req)
			if err != nil {
				return nil, err
			}
			allInputs = append(allInputs, jsonKeys...)
		} else if contentType == FormData {
			formKeys, err := getFormDataKeys(req)
			if err != nil {
				return nil, err
			}
			allInputs = append(allInputs, formKeys...)
		}
	}

	return allInputs, nil
}

func getQueryKeys(req *http.Request) []string {
	queryParams := req.URL.Query()

	var keys []string
	for key := range queryParams {
		keys = append(keys, key)
	}

	return keys
}

func getJsonKeys(req *http.Request) ([]string, error) {
	body, err := readFromBody(req)
	if err != nil {
		return nil, err
	}

	var jsonObject map[string]interface{}
	if err := json.Unmarshal(body, &jsonObject); err != nil {
		return nil, err
	}

	var keys []string
	for key := range jsonObject {
		keys = append(keys, key)
	}

	return keys, nil
}

func getFormDataKeys(req *http.Request) ([]string, error) {
	formData, err := readFromBody(req)
	if err != nil {
		return nil, err
	}

	values, err := url.ParseQuery(string(formData))
	if err != nil {
		return nil, err
	}

	var keys []string
	for key := range values {
		keys = append(keys, key)
	}

	return keys, nil
}
