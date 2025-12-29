package middleware

import (
	"errors"
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
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, err.Error())
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

// func AuthOpenAPI(appType string) func(*gin.Context) {
// 	return func(ctx *gin.Context) {
// 		token, err := getAppKey(ctx)
// 		if err != nil {
// 			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, err.Error())
// 			ctx.Abort()
// 			return
// 		}
// 		appKey, err := service.GetAppKeyByKey(ctx, token)
// 		if err != nil {
// 			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, err.Error())
// 			ctx.Abort()
// 			return
// 		}
// 		if appKey.AppType != appType {
// 			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, "invalid appType")
// 			ctx.Abort()
// 			return
// 		}
// 		ctx.Set(gin_util.USER_ID, appKey.UserId)
// 		ctx.Set(gin_util.X_ORG_ID, appKey.OrgId)
// 		ctx.Set(gin_util.APP_ID, appKey.AppId)
// 	}
//}

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
		//1.获取value值
		value := getFieldValue(ctx, fieldName)
		if len(value) == 0 {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("knowledgeId is required"))
			ctx.Abort()
			return
		}
		//2.校验用户授权权限
		err := openAPIKnowledgeGrantUser(ctx, value, permissionType)
		//3.返回结果
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
	}
}

func openAPIKnowledgeGrantUser(ctx *gin.Context, knowledgeId string, permissionType int32) error {
	// userID orgID
	userID, orgID := getOpenAPIUserID(ctx), getOpenAPIOrgID(ctx)
	if len(userID) == 0 || len(orgID) == 0 {
		return errors.New("USER-ID or X-Org-Id is empty")
	}

	// check user knowledge permission
	if err := service.CheckKnowledgeUserPermission(ctx, userID, orgID, knowledgeId, permissionType); err != nil {
		return err
	}
	return nil
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

// 获取当前用户ID
func getOpenAPIUserID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.USER_ID)
}

// 获取当前组织ID
func getOpenAPIOrgID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.X_ORG_ID)
}
