package utils

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/smitendu1997/auto-message-dispatcher/logger"
)

func HttpCall(functionName string, ctx context.Context, method, url string, client http.Client, reqBody []byte, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Error(functionName, "http make request Error: ", err)
		return nil, 0, err
	}
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(functionName, "make http call Error: ", err)
		return nil, 0, err
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(functionName, "Error during read response: ", err)
		return res, resp.StatusCode, err
	}
	return res, resp.StatusCode, nil
}
