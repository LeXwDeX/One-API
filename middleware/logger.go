package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/LeXwDeX/one-api/common/config"
	"github.com/LeXwDeX/one-api/common/helper"
	"github.com/LeXwDeX/one-api/common/logger"
	"github.com/gin-gonic/gin"
)

// responseBodyWriter captures the response body for logging
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// SetUpLogger registers both a concise summary logger and a detailed debug logger
func SetUpLogger(server *gin.Engine) {
	// concise summary logger
	server.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		var requestID string
		if param.Keys != nil {
			requestID = param.Keys[helper.RequestIdKey].(string)
		}
		return fmt.Sprintf("[GIN] %s | %s | %3d | %13v | %15s | %7s %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			requestID,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		)
	}))

	// detailed logging when LogConsumeEnabled is true
	server.Use(func(c *gin.Context) {
		if !config.LogConsumeEnabled {
			c.Next()
			return
		}
		start := time.Now()

		// read request body
		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = ioutil.ReadAll(c.Request.Body)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
		}
		// marshal request header
		reqHeader, _ := json.Marshal(c.Request.Header)

		// wrap response writer to capture response body
		rbw := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = rbw

		c.Next()

		// marshal response header and body
		resHeader, _ := json.Marshal(c.Writer.Header())
		resBody := rbw.body.String()
		latency := time.Since(start)

		// structured debug log
		logger.Debugf(c.Request.Context(),
			"HTTP %s %s\nRequest Header: %s\nRequest Body: %s\nResponse Status: %d\nResponse Header: %s\nResponse Body: %s\nLatency: %v",
			c.Request.Method,
			c.Request.URL.Path,
			string(reqHeader),
			string(reqBody),
			c.Writer.Status(),
			string(resHeader),
			resBody,
			latency,
		)
	})
}
