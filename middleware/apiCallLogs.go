package middleware

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/smitendu1997/auto-message-dispatcher/logger"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type logWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w logWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func APICallLogsMiddleware(c *gin.Context) {
	const functionName = "middleware.APICallLogsMiddleware"
	if !viper.GetBool("APICallLogs") {
		c.Next()
		return
	}
	blw := &logWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	body, err := io.ReadAll(c.Request.Body)
	if err != nil && err != io.EOF {
		logger.Error(functionName, "Error reading payload:", err)
	}
	requestBodyString := string(body)
	//add to the request body so the caller can read it again
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	c.Next()
	query := c.Request.URL.Query()
	path := cast.ToString(c.Request.URL)
	responseBody := cast.ToString(blw.body)
	headers := cast.ToString(c.Request.Header)
	payload := requestBodyString
	fmt.Println("APICallLogs", "http_method: ", c.Request.Method, "http_url: ", path, "headers: ", string(headers), "request_body: ", string(payload), "query_parameters: ", query, "response_body: ", string(responseBody))
}
