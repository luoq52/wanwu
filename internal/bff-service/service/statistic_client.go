package service

import (
	"context"
	"fmt"
	"strconv"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func GetCumulativeClientStatistic(ctx *gin.Context, endAt string) (*response.ClientCumulative, error) {
	resp, err := operate.GetCumulativeClient(ctx, &operate_service.GetCumulativeClientReq{
		EndAt: util.MustI64(endAt),
	})
	if err != nil {
		return nil, err
	}
	return &response.ClientCumulative{Total: resp.Total}, nil
}

func GetClientStatistic(ctx *gin.Context, startDate, endDate string) (*response.ClientStatistic, error) {
	// 客户端
	overview, err := getClientStatisticOverview(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	trend, err := getClientStatisticTrend(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 浏览量
	browseOverview, browseTrend, err := getGlobalBrowseStatistic(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	overview.Browse = *browseOverview
	trend.Browse = *browseTrend

	return &response.ClientStatistic{
		Overview: *overview,
		Trend:    *trend,
	}, nil
}

func getClientStatisticOverview(ctx *gin.Context, startDate, endDate string) (*response.ClientOverView, error) {
	resp, err := operate.GetClientOverview(ctx, &operate_service.GetClientOverviewReq{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, err
	}

	return &response.ClientOverView{
		NewClient:    clientOverviewPb2resp(resp.NewClient),
		ActiveClient: clientOverviewPb2resp(resp.ActiveClient),
	}, nil
}

func getClientStatisticTrend(ctx *gin.Context, startDate, endDate string) (*response.ClientTrend, error) {
	// resp, err := operate.GetClientTrend(ctx, &operate_service.GetClientTrendReq{
	// 	StartDate: startDate,
	// 	EndDate:   endDate,
	// })
	// if err != nil {
	// 	return nil, err
	// }
	// return &response.ClientTrends{
	// 	Client: convertStatisticChart(resp.Client),
	// }, nil

	//i18n替换表名
	// trend.Client.TableName = gin_util.I18nKey(ctx, trend.Client.TableName)
	// for i, line := range trend.Client.Lines {
	// 	trend.Client.Lines[i].LineName = gin_util.I18nKey(ctx, line.LineName)
	// }

	return &response.ClientTrend{}, nil
}

func clientOverviewPb2resp(item *operate_service.ClientOverviewItem) response.StatisticOverviewItem {
	valueStr := fmt.Sprintf("%.2f", item.Value)
	value, _ := strconv.ParseFloat(valueStr, 64)
	return response.StatisticOverviewItem{
		Value:            float32(value),
		PeriodOverPeriod: item.PeriodOverperiod,
	}
}

// --- global browse statistic ---

func getGlobalBrowseStatistic(ctx *gin.Context, startDate, endDate string) (*response.StatisticOverviewItem, *response.StatisticChart, error) {
	// 获取当前周期和上一个周期的日期列表
	prevDates, currentDates, err := util.PreviousDateRange(startDate, endDate)
	if err != nil {
		return nil, nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_template_stats", fmt.Sprintf("get date range error: %v", err))
	}

	// 获取浏览数据
	currentBrowseData, err := getBrowseDataFromRedis(ctx.Request.Context(), currentDates)
	if err != nil {
		return nil, nil, err
	}
	prevBrowseData, err := getBrowseDataFromRedis(ctx.Request.Context(), prevDates)
	if err != nil {
		return nil, nil, err
	}

	// 计算总览数据
	overview := calculateGlobalBrowseOverview(currentBrowseData, prevBrowseData)

	// 计算趋势数据
	trend := calculateGlobalBrowseTrend(ctx, currentBrowseData, currentDates)

	return &overview, &trend, nil
}

// 从Redis获取多个日期的浏览数据
func getBrowseDataFromRedis(ctx context.Context, dates []string) (map[string]int64, error) {
	items, err := redis.OP().HGetAll(ctx, redisGlobalBrowseKey)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_template_stats", fmt.Sprintf("redis HGetAll key %v fields %v err: %v", redisGlobalBrowseKey, dates, err))
	}

	data := make(map[string]int64)
	if len(items) == 0 {
		return data, nil
	}
	for _, date := range dates {
		for _, item := range items {
			if item.K == date {
				data[date] = util.MustI64(item.V)
				break
			}
		}
		// 如果某个日期没有数据，默认值为0
		if _, exist := data[date]; !exist {
			data[date] = 0
		}
	}

	return data, nil
}

// 计算总览数据
func calculateGlobalBrowseOverview(currentData, prevData map[string]int64) response.StatisticOverviewItem {
	// 计算当前周期总浏览量
	var currentTotal int64
	for _, count := range currentData {
		currentTotal += count
	}

	// 计算上一个周期总浏览量
	var prevTotal int64
	for _, count := range prevData {
		prevTotal += count
	}

	// 计算环比
	var pop float32
	if prevTotal > 0 {
		pop = (float32(currentTotal) - float32(prevTotal)) / float32(prevTotal) * 100
	} else if currentTotal > 0 {
		// 如果上期为0，本期有数据，增长率为100%
		pop = 100
	}

	return response.StatisticOverviewItem{
		Value:            float32(currentTotal),
		PeriodOverPeriod: pop,
	}
}

// 计算趋势数据
func calculateGlobalBrowseTrend(ctx *gin.Context, browseData map[string]int64, dates []string) response.StatisticChart {
	var items []response.StatisticChartLineItem
	for _, date := range dates {
		count := browseData[date]
		items = append(items, response.StatisticChartLineItem{
			Key:   date,
			Value: float32(count),
		})
	}
	return response.StatisticChart{
		TableName: gin_util.I18nKey(ctx, "ope_workflow_template_browse_table"),
		Lines: []response.StatisticChartLine{
			{
				LineName: gin_util.I18nKey(ctx, "ope_workflow_template_browse_line"),
				Items:    items,
			},
		},
	}
}
