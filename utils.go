package burptomap

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
)

// readFromBody reads the all bytes from body and closes it.
func readFromBody(req *http.Request) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body.Close()
	return body, nil
}

// writeToBody writes the given body to the request body and fix all the other requirements
func writeToBody(req *http.Request, body []byte) {
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
}

func loadRequestFromString(rawRequest string) (*http.Request, bool, error) {
	isHTTP2Detected := false
	reqBytes := bytes.NewBufferString(rawRequest).Bytes()

	// http.ReadRequest does not support HTTP/2 change to HTTP/1.1 temporarily
	if bytes.Contains(bytes.Split(reqBytes, []byte("\n"))[0], []byte("HTTP/2")) {
		isHTTP2Detected = true
		reqBytes = bytes.Replace(reqBytes, []byte("HTTP/2"), []byte("HTTP/1.1"), 1)
	}

	// Read Request with standart lib.
	reqBytesReader := bytes.NewReader(reqBytes)
	reqBufioReader := bufio.NewReader(reqBytesReader)
	httpReq, err := http.ReadRequest(reqBufioReader)
	if err != nil {
		return nil, isHTTP2Detected, err
	}

	return httpReq, isHTTP2Detected, nil
}

func dumpRequestToString(httpReq *http.Request, isHTTP2Detected bool) (string, error) {
	modifiedRequest := bytes.NewBuffer([]byte{})
	httpReq.Write(modifiedRequest)

	// Convert back to HTTP/2 if it is replaced before.
	if isHTTP2Detected {
		if bytes.Contains(bytes.Split(modifiedRequest.Bytes(), []byte("\n"))[0], []byte("HTTP/1.1")) {
			modifiedRequest = bytes.NewBuffer(bytes.Replace(modifiedRequest.Bytes(), []byte("HTTP/1.1"), []byte("HTTP/2"), 1))
		}
	}

	return modifiedRequest.String(), nil
}
