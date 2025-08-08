package ratio

import (
	"encoding/json"
	"sync"

	"github.com/LeXwDeX/one-api/common/logger"
)

const (
	USD2RMB   = 7
	USD       = 500 // $0.002 = 1 -> $1 = 500
	MILLI_USD = 1.0 / 1000 * USD
	RMB       = USD / USD2RMB
)

var modelRatioLock sync.RWMutex

// ModelRatio
// https://platform.openai.com/docs/models/model-endpoint-compatibility
// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9dlf
// https://openai.com/pricing
// 1 === $0.002 / 1K tokens
// 1 === ￥0.014 / 1k tokens
var ModelRatio = map[string]float64{
	// https://openai.com/pricing
	"o1":     1.0, // $15.00 / 1M input tokens
	"o3":     1.0,
	"gpt4.1": 1.0,
	// https://docs.anthropic.com/en/docs/about-claude/models
	"claude-3-opus": 15.0 / 1000 * USD,
	// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/hlrk4akp7
	"ERNIE-4.0": 0.120 * RMB,
	// https://ai.google.dev/pricing
	// https://cloud.google.com/vertex-ai/generative-ai/pricing
	// "gemma-2-2b-it":                       0,
	// "gemma-2-9b-it":                       0,
	// "gemma-2-27b-it":                      0,
	"gemini-pro": 0.25 * MILLI_USD, // $0.00025 / 1k characters -> $0.001 / 1k tokens
	// https://open.bigmodel.cn/pricing
	"glm-zero-preview": 0.01 * RMB,
	// https://help.aliyun.com/zh/dashscope/developer-reference/tongyi-thousand-questions-metering-and-billing
	"qwen-turbo": 0.0003 * RMB,
	// https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
	"hunyuan-turbo": 0.015 * RMB,
	// https://platform.moonshot.cn/pricing
	"moonshot-v1-8k": 0.012 * RMB,
	// https://platform.baichuan-ai.com/price
	"Baichuan2-Turbo": 0.008 * RMB,
	// https://api.minimax.chat/document/price
	"abab6.5-chat": 0.03 * RMB,
	// https://docs.mistral.ai/platform/pricing/
	"open-mistral-7b": 0.25 / 1000 * USD,
	// https://wow.groq.com/#:~:text=inquiries%C2%A0here.-,Model,-Current%20Speed
	"gemma-7b-it": 0.07 / 1000000 * USD,
	// https://platform.lingyiwanwu.com/docs#-计费单元
	"yi-34b-chat-0205": 2.5 / 1000 * RMB,
	// https://platform.stepfun.com/docs/pricing/details
	"step-1-8k": 0.005 / 1000 * RMB,
	// aws llama3 https://aws.amazon.com/cn/bedrock/pricing/
	"llama3-8b-8192(33)": 0.0003 / 0.002, // $0.0003 / 1K tokens
	// https://cohere.com/pricing
	"command": 0.5,
	// https://platform.deepseek.com/api-docs/pricing/
	"deepseek-chat":     0.14 * MILLI_USD,
	"deepseek-reasoner": 0.55 * MILLI_USD,
	// https://www.deepl.com/pro?cta=header-prices
	"deepl-zh": 25.0 / 1000 * USD,
	"deepl-en": 25.0 / 1000 * USD,
	"deepl-ja": 25.0 / 1000 * USD,
	// https://console.x.ai/
	"grok-beta": 5.0 / 1000 * USD,
	// replicate charges based on the number of generated images
	// https://replicate.com/pricing
	"black-forest-labs/flux-1.1-pro": 0.04 * USD,
	// replicate chat models
	"ibm-granite/granite-20b-code-instruct-8k": 0.100 * USD,
	//https://openrouter.ai/models
	"01-ai/yi-large": 1.5,
}

var CompletionRatio = map[string]float64{
	// aws llama3
	"llama3-8b-8192(33)": 0.0006 / 0.0003,
	// whisper
	"whisper-1": 0, // only count input tokens
	// deepseek
	"deepseek-chat":     0.28 / 0.14,
	"deepseek-reasoner": 2.19 / 0.55,
}

var (
	DefaultModelRatio      map[string]float64
	DefaultCompletionRatio map[string]float64
)

func init() {
	DefaultModelRatio = make(map[string]float64)
	for k, v := range ModelRatio {
		DefaultModelRatio[k] = v
	}
	DefaultCompletionRatio = make(map[string]float64)
	for k, v := range CompletionRatio {
		DefaultCompletionRatio[k] = v
	}
}

func AddNewMissingRatio(oldRatio string) string {
	newRatio := make(map[string]float64)
	err := json.Unmarshal([]byte(oldRatio), &newRatio)
	if err != nil {
		logger.SysError("error unmarshalling old ratio: " + err.Error())
		return oldRatio
	}
	for k, v := range DefaultModelRatio {
		if _, ok := newRatio[k]; !ok {
			newRatio[k] = v
		}
	}
	jsonBytes, err := json.Marshal(newRatio)
	if err != nil {
		logger.SysError("error marshalling new ratio: " + err.Error())
		return oldRatio
	}
	return string(jsonBytes)
}

func ModelRatio2JSONString() string {
	jsonBytes, err := json.Marshal(ModelRatio)
	if err != nil {
		logger.SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRatioByJSONString(jsonStr string) error {
	modelRatioLock.Lock()
	defer modelRatioLock.Unlock()
	ModelRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &ModelRatio)
}

func GetModelRatio(name string, channelType int) float64 {
	return 1
}

func CompletionRatio2JSONString() string {
	jsonBytes, err := json.Marshal(CompletionRatio)
	if err != nil {
		logger.SysError("error marshalling completion ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateCompletionRatioByJSONString(jsonStr string) error {
	CompletionRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &CompletionRatio)
}

func GetCompletionRatio(name string, channelType int) float64 {
	return 1
}
