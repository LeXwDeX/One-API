package openai

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	relaymodel "github.com/LeXwDeX/one-api/relay/model"
	"github.com/LeXwDeX/one-api/relay/relaymode"
)

func TestConvertRequestGPT5MapsMaxTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	adaptor := &Adaptor{}

	req := &relaymodel.GeneralOpenAIRequest{
		Model:     "gpt-5",
		MaxTokens: 128,
	}

	converted, err := adaptor.ConvertRequest(ctx, relaymode.ChatCompletions, req)
	require.NoError(t, err)

	convertedReq, ok := converted.(*relaymodel.GeneralOpenAIRequest)
	require.True(t, ok)
	require.NotNil(t, convertedReq.MaxCompletionTokens)
	assert.Equal(t, 128, *convertedReq.MaxCompletionTokens)
	assert.Zero(t, convertedReq.MaxTokens)

	body, err := json.Marshal(convertedReq)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "max_tokens")
	assert.Contains(t, string(body), "max_completion_tokens")
}

func TestConvertRequestGPT5PrefersMaxCompletionTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	adaptor := &Adaptor{}

	mct := 64
	req := &relaymodel.GeneralOpenAIRequest{
		Model:               "gpt-5-pro",
		MaxTokens:           256,
		MaxCompletionTokens: &mct,
	}

	converted, err := adaptor.ConvertRequest(ctx, relaymode.ChatCompletions, req)
	require.NoError(t, err)

	convertedReq, ok := converted.(*relaymodel.GeneralOpenAIRequest)
	require.True(t, ok)
	require.NotNil(t, convertedReq.MaxCompletionTokens)
	assert.Equal(t, 64, *convertedReq.MaxCompletionTokens)
	assert.Zero(t, convertedReq.MaxTokens)

	body, err := json.Marshal(convertedReq)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "max_tokens")
	assert.Contains(t, string(body), "max_completion_tokens")
}
