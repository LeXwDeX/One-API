package openai

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/LeXwDeX/one-api/common/logger"
	"github.com/LeXwDeX/one-api/relay/adaptor"
	"github.com/LeXwDeX/one-api/relay/adaptor/alibailian"
	"github.com/LeXwDeX/one-api/relay/adaptor/baiduv2"
	"github.com/LeXwDeX/one-api/relay/adaptor/doubao"
	"github.com/LeXwDeX/one-api/relay/adaptor/geminiv2"
	"github.com/LeXwDeX/one-api/relay/adaptor/minimax"
	"github.com/LeXwDeX/one-api/relay/adaptor/novita"
	"github.com/LeXwDeX/one-api/relay/channeltype"
	"github.com/LeXwDeX/one-api/relay/meta"
	"github.com/LeXwDeX/one-api/relay/model"
	"github.com/LeXwDeX/one-api/relay/relaymode"
)

type Adaptor struct {
	ChannelType int
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.ChannelType = meta.ChannelType
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.ChannelType {
	case channeltype.Azure:
		if meta.Mode == relaymode.ImagesGenerations {
			// https://learn.microsoft.com/en-us/azure/ai-services/openai/dall-e-quickstart?tabs=dalle3%2Ccommand-line&pivots=rest-api
			// https://{resource_name}.openai.azure.com/openai/deployments/dall-e-3/images/generations?api-version=2024-03-01-preview
			fullRequestURL := fmt.Sprintf("%s/openai/deployments/%s/images/generations?api-version=%s", meta.BaseURL, meta.ActualModelName, meta.Config.APIVersion)
			return fullRequestURL, nil
		}

		// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/chatgpt-quickstart?pivots=rest-api&tabs=command-line#rest-api
		requestURL := strings.Split(meta.RequestURLPath, "?")[0]
		requestURL = fmt.Sprintf("%s?api-version=%s", requestURL, meta.Config.APIVersion)
		task := strings.TrimPrefix(requestURL, "/v1/")
		model_ := meta.ActualModelName
		// 保留原始模型名，兼容 Azure deployment 名称如 gpt-4.1
		//https://github.com/LeXwDeX/one-api/issues/1191
		// {your endpoint}/openai/deployments/{your azure_model}/chat/completions?api-version={api_version}
		requestURL = fmt.Sprintf("/openai/deployments/%s/%s", model_, task)
		return GetFullRequestURL(meta.BaseURL, requestURL, meta.ChannelType), nil
	case channeltype.Minimax:
		return minimax.GetRequestURL(meta)
	case channeltype.Doubao:
		return doubao.GetRequestURL(meta)
	case channeltype.Novita:
		return novita.GetRequestURL(meta)
	case channeltype.BaiduV2:
		return baiduv2.GetRequestURL(meta)
	case channeltype.AliBailian:
		return alibailian.GetRequestURL(meta)
	case channeltype.GeminiOpenAICompatible:
		return geminiv2.GetRequestURL(meta)
	default:
		return GetFullRequestURL(meta.BaseURL, meta.RequestURLPath, meta.ChannelType), nil
	}
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	if meta.ChannelType == channeltype.Azure {
		req.Header.Set("api-key", meta.APIKey)
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	if meta.ChannelType == channeltype.OpenRouter {
		req.Header.Set("HTTP-Referer", "https://github.com/LeXwDeX/one-api")
		req.Header.Set("X-Title", "One API")
	}
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	if request.Stream {
		// always return usage in stream mode
		if request.StreamOptions == nil {
			request.StreamOptions = &model.StreamOptions{}
		}
		request.StreamOptions.IncludeUsage = true
	}
	// 兼容 temperature，仅对 o3、o3-mini 和 o4-mini 模型强制设为 1
	if request.Model == "o3" || request.Model == "o3-mini" || request.Model == "o4-mini" {
		v := 1.0
		request.Temperature = &v
	}
	// 兼容 max_completion_tokens，仅对 o3、o3-mini 和 o4-mini 模型做参数适配
	if request.Model == "o3" || request.Model == "o3-mini" || request.Model == "o4-mini" {
		if request.MaxCompletionTokens != nil {
			// 转为 map 以便动态调整字段
			m, err := adaptor.StructToMap(request)
			if err != nil {
				return nil, err
			}
			delete(m, "max_tokens")
			m["max_completion_tokens"] = *request.MaxCompletionTokens
			// add reasoning_effort for o4-mini model
			if request.ReasoningEffort == nil {
				m["reasoning_effort"] = "high"
			}
			return m, nil
		}
		// 自动适配 max_tokens -> max_completion_tokens
		if request.MaxCompletionTokens == nil && request.MaxTokens != 0 {
			m, err := adaptor.StructToMap(request)
			if err != nil {
				return nil, err
			}
			m["max_completion_tokens"] = request.MaxTokens
			delete(m, "max_tokens")
			// add reasoning_effort for o4-mini model
			if request.ReasoningEffort == nil {
				m["reasoning_effort"] = "high"
			}
			return m, nil
		}
	}
	// add reasoning_effort for o4-mini model
	if request.Model == "o4-mini" {
		if request.ReasoningEffort == nil {
			s := "high"
			request.ReasoningEffort = &s
		}
	}
	return request, nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	// 记录 meta 信息
	logger.SysLog(fmt.Sprintf("[DoRequest] meta: %+v", meta))
	// 记录请求体前 512 字符，防止 base64 刷屏
	var bodyPreview string
	if requestBody != nil {
		buf := make([]byte, 512)
		n, _ := requestBody.Read(buf)
		bodyPreview = string(buf[:n])
		// 由于 Read 会消耗 requestBody，需要重置
		requestBody = io.MultiReader(strings.NewReader(bodyPreview), requestBody)
	}
	logger.SysLog(fmt.Sprintf("[DoRequest] request body preview: %s", bodyPreview))
	resp, err := adaptor.DoRequestHelper(a, c, meta, requestBody)
	if resp != nil {
		logger.SysLog(fmt.Sprintf("[DoRequest] downstream status: %d", resp.StatusCode))
	}
	if err != nil {
		logger.SysError(fmt.Sprintf("[DoRequest] error: %v", err))
	}
	return resp, err
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		var responseText string
		err, responseText, usage = StreamHandler(c, resp, meta.Mode)
		if usage == nil || usage.TotalTokens == 0 {
			usage = ResponseText2Usage(responseText, meta.ActualModelName, meta.PromptTokens)
		}
		if usage.TotalTokens != 0 && usage.PromptTokens == 0 { // some channels don't return prompt tokens & completion tokens
			usage.PromptTokens = meta.PromptTokens
			usage.CompletionTokens = usage.TotalTokens - meta.PromptTokens
		}
	} else {
		switch meta.Mode {
		case relaymode.ImagesGenerations:
			err, _ = ImageHandler(c, resp)
		default:
			err, usage = Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	_, modelList := GetCompatibleChannelMeta(a.ChannelType)
	return modelList
}

func (a *Adaptor) GetChannelName() string {
	channelName, _ := GetCompatibleChannelMeta(a.ChannelType)
	return channelName
}
