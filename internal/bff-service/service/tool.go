// @Author wangxm 10/24/星期五 14:46:00
package service

import (
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/gin-gonic/gin"
)

func GetToolSelect(ctx *gin.Context, userID, orgID string, name string) (*response.ListResult, error) {
	resp, err := mcp.GetToolSelect(ctx.Request.Context(), &mcp_service.GetToolSelectReq{
		Name: name,
		Identity: &mcp_service.Identity{
			UserId: userID,
			OrgId:  orgID,
		},
	})
	if err != nil {
		return nil, err
	}

	var list []response.ToolSelect
	for _, item := range resp.List {
		list = append(list, response.ToolSelect{
			UniqueId: "tool-" + item.ToolId,
			ToolInfo: response.ToolInfo{
				ToolId:          item.ToolId,
				ToolName:        item.ToolName,
				ToolType:        item.ToolType,
				Desc:            item.Desc,
				APIKey:          item.ApiKey,
				NeedApiKeyInput: item.NeedApiKeyInput,
			},
		})
	}
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}
