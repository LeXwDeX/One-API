package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/LeXwDeX/one-api/common/render"

	"github.com/gin-gonic/gin"
	"github.com/LeXwDeX/one-api/common"
	"github.com/LeXwDeX/one-api/common/conv"
	"github.com/LeXwDeX/one-api/common/logger"
	"github.com/LeXwDeX/one-api/relay/model"
	"github.com/LeXwDeX/one-api/relay/relaymode"
)

const (
	dataPrefix       = "data: "
	done             = "[DONE]"
	dataPrefixLength = len(dataPrefix)
)

func StreamHandler(c *gin.Context, resp *http.Response, relayMode int) (*model.ErrorWithStatusCode, string, *model.Usage) {
	responseText := ""
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)
	var usage *model.Usage

	common.SetEventStreamHeaders(c)

	doneRendered := false
	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < dataPrefixLength { // ignore blank line or wrong format
			continue
		}
		if data[:dataPrefixLength] != dataPrefix && data[:dataPrefixLength] != done {
			continue
		}
		if strings.HasPrefix(data[dataPrefixLength:], done) {
			render.StringData(c, data)
			doneRendered = true
			continue
		}
		switch relayMode {
		case relaymode.ChatCompletions:
			var streamResponse ChatCompletionsStreamResponse
			err := json.Unmarshal([]byte(data[dataPrefixLength:]), &streamResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				render.StringData(c, data) // if error happened, pass the data to client
				continue                   // just ignore the error
			}
			if len(streamResponse.Choices) == 0 && streamResponse.Usage == nil {
				// but for empty choice and no usage, we should not pass it to client, this is for azure
				continue // just ignore empty choice
			}
			render.StringData(c, data)
			for _, choice := range streamResponse.Choices {
				responseText += conv.AsString(choice.Delta.Content)
			}
			if streamResponse.Usage != nil {
				usage = streamResponse.Usage
			}
		case relaymode.Completions:
			render.StringData(c, data)
			var streamResponse CompletionsStreamResponse
			err := json.Unmarshal([]byte(data[dataPrefixLength:]), &streamResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				continue
			}
			for _, choice := range streamResponse.Choices {
				responseText += choice.Text
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	if !doneRendered {
		render.Done(c)
	}

	err := resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), "", nil
	}

	return nil, responseText, usage
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	var textResponse SlimTextResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &textResponse)
	if err != nil {
		return ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if textResponse.Error.Type != "" {
		return &model.ErrorWithStatusCode{
			Error:      textResponse.Error,
			StatusCode: resp.StatusCode,
		}, nil
	}
	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the HTTPClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return ErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	if textResponse.Usage.TotalTokens == 0 || (textResponse.Usage.PromptTokens == 0 && textResponse.Usage.CompletionTokens == 0) {
		completionTokens := 0
		for _, choice := range textResponse.Choices {
			completionTokens += CountTokenText(choice.Message.StringContent(), modelName)
		}
		textResponse.Usage = model.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		}
	}
	return nil, &textResponse.Usage
}

// Azure Responses API 流式与非流式专用处理

// AzureResponsesStreamPassthrough 将 Azure Responses 的 SSE 事件原样透传，并提取 response.output_text.delta 累积文本
func AzureResponsesStreamPassthrough(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, string, *model.Usage) {
	responseText := ""
	var usage *model.Usage

	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	common.SetEventStreamHeaders(c)

	currentEvent := ""
	doneRendered := false

	type outputDelta struct {
		Delta string `json:"delta"`
	}

	for scanner.Scan() {
		line := scanner.Text()
		// 直接透传每一行，保证客户端按 Azure Responses 事件结构解析
		render.StringData(c, line)

		if line == "" {
			// 事件块分隔
			continue
		}

		if strings.HasPrefix(line, "event:") {
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			// 累积 delta
			if currentEvent == "response.output_text.delta" {
				var d outputDelta
				if err := json.Unmarshal([]byte(data), &d); err == nil && d.Delta != "" {
					responseText += d.Delta
				} else {
					// 如果 data 不为 JSON，尝试去掉首尾引号后累积
					trimmed := strings.Trim(data, "\"")
					responseText += trimmed
				}
			}
			// 标记完成事件
			if currentEvent == "response.completed" || currentEvent == "response.error" {
				doneRendered = true
			}
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	if !doneRendered {
		render.Done(c)
	}

	if err := resp.Body.Close(); err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), "", nil
	}

	// Azure Responses 流式通常不返回 usage，留给上层用 ResponseText2Usage 推断
	return nil, responseText, usage
}

// AzureResponsesNonStreamHandler 读取非流式响应，转发给客户端，并基于 output_text 计算 usage
func AzureResponsesNonStreamHandler(c *gin.Context, resp *http.Response, modelName string, promptTokens int) (*model.ErrorWithStatusCode, *model.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	if err = resp.Body.Close(); err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	// 尝试解析 output_text
	var outputText string
	// 既支持根级 output_text，也兼容简单字符串体
	// 优先尝试 JSON 对象
	var obj map[string]any
	if json.Unmarshal(responseBody, &obj) == nil {
		if v, ok := obj["output_text"]; ok {
			switch vv := v.(type) {
			case string:
				outputText = vv
			default:
				// 若不是字符串，尝试转为字符串
				b, _ := json.Marshal(v)
				outputText = string(b)
			}
		}
	} else {
		// 非 JSON，按纯文本处理
		outputText = string(responseBody)
	}

	// 重置响应体并原样转发给客户端
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	for k, v := range resp.Header {
		if len(v) > 0 {
			c.Writer.Header().Set(k, v[0])
		}
	}
	c.Writer.WriteHeader(resp.StatusCode)
	if _, err = io.Copy(c.Writer, resp.Body); err != nil {
		return ErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	if err = resp.Body.Close(); err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	// 计算 usage（若服务端未提供）
	usage := ResponseText2Usage(outputText, modelName, promptTokens)
	return nil, usage
}