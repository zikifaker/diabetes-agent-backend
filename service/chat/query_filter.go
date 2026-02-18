package chat

import (
	"diabetes-agent-server/config"
	"encoding/json"
	"errors"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	green "github.com/alibabacloud-go/green-20220302/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	clientConnectTimeout  = 3000
	clientReadTimeout     = 6000
	serviceConnectTimeout = 10000
	serviceReadTimeout    = 10000
)

var ErrQueryContainsHighRiskContent = errors.New("query contains high risk content")

type FilterQueryRequest struct {
	Content string `json:"content"`
}

// 调用阿里云大语言模型输入文字检测服务，检测用户输入是否包含高风险内容
func filterQuery(query string) error {
	config := &openapi.Config{
		AccessKeyId:     tea.String(config.Cfg.OSS.AccessKeyID),
		AccessKeySecret: tea.String(config.Cfg.OSS.AccessKeySecret),
		RegionId:        tea.String("cn-shanghai"),
		Endpoint:        tea.String("green-cip.cn-shanghai.aliyuncs.com"),
		ConnectTimeout:  tea.Int(serviceConnectTimeout),
		ReadTimeout:     tea.Int(serviceReadTimeout),
	}
	client, err := green.NewClient(config)
	if err != nil {
		return err
	}

	parameters, _ := json.Marshal(FilterQueryRequest{Content: query})
	request := green.TextModerationPlusRequest{
		Service:           tea.String("llm_query_moderation"),
		ServiceParameters: tea.String(string(parameters)),
	}

	runtime := &util.RuntimeOptions{
		ConnectTimeout: tea.Int(clientConnectTimeout),
		ReadTimeout:    tea.Int(clientReadTimeout),
	}
	result, err := client.TextModerationPlusWithOptions(&request, runtime)
	if err != nil {
		return err
	}

	body := result.Body
	riskLevel := body.Data.RiskLevel
	if riskLevel != nil && *riskLevel == "high" {
		return ErrQueryContainsHighRiskContent
	}

	return nil
}
