package openapi

import (
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/openapi"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	"github.com/UnicomAI/wanwu/pkg/constant"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func Register(openAPI *gin.RouterGroup) {
	// agent
	mid.Sub("openapi").Reg(openAPI, "/agent/conversation", http.MethodPost, openapi.CreateAgentConversation, "智能体创建对话OpenAPI", middleware.AuthOpenAPIKey(constant.OpenAPITypeAgent))
	mid.Sub("openapi").Reg(openAPI, "/agent/chat", http.MethodPost, openapi.ChatAgent, "智能体问答OpenAPI", middleware.AuthOpenAPIKey(constant.OpenAPITypeAgent))
	// rag
	mid.Sub("openapi").Reg(openAPI, "/rag/chat", http.MethodPost, openapi.ChatRag, "文本问答OpenAPI", middleware.AuthOpenAPIKey(constant.OpenAPITypeRag))
	// workflow
	mid.Sub("openapi").Reg(openAPI, "/workflow/run", http.MethodPost, openapi.WorkflowRun, "工作流OpenAPI", middleware.AuthOpenAPIKey(constant.OpenAPITypeWorkflow))
	mid.Sub("openapi").Reg(openAPI, "/workflow/file/upload", http.MethodPost, openapi.WorkflowFileUpload, "工作流OpenAPI文件上传", middleware.AuthOpenAPIKey(constant.OpenAPITypeChatflow))
	// chatflow
	mid.Sub("openapi").Reg(openAPI, "/chatflow/conversation", http.MethodPost, openapi.CreateChatflowConversation, "对话流创建对话OpenAPI", middleware.AuthOpenAPIKey(constant.OpenAPITypeChatflow))
	mid.Sub("openapi").Reg(openAPI, "/chatflow/conversation/message/list", http.MethodPost, openapi.GetConversationMessageList, "对话流根据conversationId获取历史对话", middleware.AuthOpenAPIKey(constant.OpenAPITypeChatflow))
	mid.Sub("openapi").Reg(openAPI, "/chatflow/chat", http.MethodPost, openapi.ChatflowChat, "对话流OpenAPI", middleware.AuthOpenAPIKey(constant.OpenAPITypeChatflow))

	// mcp server
	mid.Sub("openapi").Reg(openAPI, "/mcp/server/sse", http.MethodGet, openapi.GetMCPServerSSE, "新建MCP服务sse连接", middleware.AuthAppKeyByQuery(constant.AppTypeMCPServer))
	mid.Sub("openapi").Reg(openAPI, "/mcp/server/message", http.MethodPost, openapi.GetMCPServerMessage, "获取MCP服务sse消息", middleware.AuthAppKeyByQuery(constant.AppTypeMCPServer))
	mid.Sub("openapi").Reg(openAPI, "/mcp/server/streamable", http.MethodGet, openapi.GetMCPServerStreamable, "获取MCP服务streamable消息(GET)", middleware.AuthAppKeyByQuery(constant.AppTypeMCPServer))
	mid.Sub("openapi").Reg(openAPI, "/mcp/server/streamable", http.MethodPost, openapi.GetMCPServerStreamable, "获取MCP服务streamable消息(POST)", middleware.AuthAppKeyByQuery(constant.AppTypeMCPServer))

	// oauth
	mid.Sub("openapi").Reg(openAPI, "/oauth/jwks", http.MethodGet, openapi.OAuthJWKS, "JWT公钥")
	mid.Sub("openapi").Reg(openAPI, "/oauth/login", http.MethodGet, openapi.OAuthLogin, "OAuth登录授权")
	mid.Sub("openapi").Reg(openAPI, "/oauth/code/authorize", http.MethodGet, openapi.OAuthAuthorize, "获取授权码")
	mid.Sub("openapi").Reg(openAPI, "/oauth/code/token", http.MethodPost, openapi.OAuthToken, "授权码获取token")
	mid.Sub("openapi").Reg(openAPI, "/oauth/code/token/refresh", http.MethodPost, openapi.OAuthRefresh, "刷新Access Token")
	mid.Sub("openapi").Reg(openAPI, "/.well-known/openid-configuration", http.MethodGet, openapi.OAuthConfig, "返回Endpoint配置")
	// oauth user
	mid.Sub("openapi").Reg(openAPI, "/oauth/userinfo", http.MethodGet, openapi.OAuthGetUserInfo, "OAuth获取用户信息", middleware.JWTOAuthAccess)
}
