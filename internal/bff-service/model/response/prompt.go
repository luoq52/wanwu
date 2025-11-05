package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type CustomPromptIDResp struct {
	CustomPromptID string `json:"customPromptId"` // 自定义提示词ID
}

type CustomPrompt struct {
	CustomPromptIDResp                // 自定义提示词ID
	Avatar             request.Avatar `json:"avatar"`   // 图标
	Name               string         `json:"name"`     // 名称
	Desc               string         `json:"desc"`     // 描述
	Prompt             string         `json:"prompt"`   // 提示词
	UpdateAt           string         `json:"updateAt"` // 更新时间
}
