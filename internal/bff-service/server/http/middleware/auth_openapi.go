package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
)

func AuthOpenAPIKey(openApiType string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		token, err := getApiKey(ctx)
		if err != nil {
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, err.Error())
			ctx.Abort()
			return
		}
		apiKey, err := service.GetApiKeyByKey(ctx, token)
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusUnauthorized, err)
			ctx.Abort()
			return
		}
		if !apiKey.Status {
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, "api key disabled")
			ctx.Abort()
			return
		}
		if apiKey.ExpiredAt != 0 && apiKey.ExpiredAt < time.Now().UnixMilli() {
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, "api key expired")
			ctx.Abort()
			return
		}
		ctx.Set(gin_util.USER_ID, apiKey.UserId)
		ctx.Set(gin_util.X_ORG_ID, apiKey.OrgId)
	}
}

func AuthAppKeyByQuery(appType string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		token, err := getAppKeyByQuery(ctx)
		if err != nil {
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, err.Error())
			ctx.Abort()
			return
		}
		appKey, err := service.GetAppKeyByKey(ctx, token)
		if err != nil {
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, err.Error())
			ctx.Abort()
			return
		}
		if appKey.AppType != appType {
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, "invalid appType")
			ctx.Abort()
			return
		}
		ctx.Set(gin_util.USER_ID, appKey.UserId)
		ctx.Set(gin_util.X_ORG_ID, appKey.OrgId)
		ctx.Set(gin_util.APP_ID, appKey.AppId)
	}

}

// AuthOpenAPIKnowledge 校验知识库权限
func AuthOpenAPIKnowledge(fieldName string, permissionType int32) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()

		// 1. 获取知识库ID
		knowledgeId := getFieldValue(ctx, fieldName)
		if len(knowledgeId) == 0 {
			gin_util.ResponseDetail(ctx, http.StatusBadRequest, codes.Code(err_code.Code_BFFAuth), nil, "knowledgeId is required")
			ctx.Abort()
			return
		}

		// 2. 获取用户和机构信息
		userID, orgID := ctx.GetString(gin_util.USER_ID), ctx.GetString(gin_util.X_ORG_ID)
		if len(userID) == 0 || len(orgID) == 0 {
			gin_util.ResponseDetail(ctx, http.StatusBadRequest, codes.Code(err_code.Code_BFFAuth), nil, "USER-ID or X-Org-Id is required")
			ctx.Abort()
			return
		}

		// 3. 校验用户对知识库的权限
		if err := service.CheckKnowledgeUserPermission(ctx, userID, orgID, knowledgeId, permissionType); err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}

		// 4. 权限验证通过，继续后续处理
		ctx.Next()
	}
}

// --- internal ---
func getApiKey(ctx *gin.Context) (string, error) {
	authorization := ctx.Request.Header.Get("Authorization")
	if authorization != "" {
		tks := strings.Split(authorization, " ")
		if len(tks) > 1 && tks[0] == "Bearer" {
			return tks[1], nil
		} else {
			return "", fmt.Errorf("not Bearer token format")
		}
	} else {
		return "", fmt.Errorf("token is nil")
	}
}

func getAppKeyByQuery(ctx *gin.Context) (string, error) {
	key := ctx.Query("key")
	if key != "" {
		return key, nil
	} else {
		return "", fmt.Errorf("token is nil")
	}
}
