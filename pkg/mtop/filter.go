package mtop

import (
	"fmt"
	"time"
)

// FilterItems 根据 GuessYouLikeOptions 过滤商品列表
func FilterItems(items []FeedItem, options GuessYouLikeOptions) []FeedItem {
	if options.MinWantCount == 0 && options.DaysWithin == 0 {
		// 无过滤条件，直接返回全部
		return items
	}

	filtered := make([]FeedItem, 0, len(items))

	for _, item := range items {
		if options.MatchFilter(item) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// MatchFilter 检查商品是否匹配过滤条件
func (o GuessYouLikeOptions) MatchFilter(item FeedItem) bool {
	// 检查最低想要人数
	if o.MinWantCount > 0 && item.WantCount < o.MinWantCount {
		return false
	}

	// 检查发布时间范围
	if o.DaysWithin > 0 && item.PublishTimeTS > 0 {
		cutoffTime := time.Now().AddDate(0, 0, -o.DaysWithin).UnixMilli()
		if item.PublishTimeTS < cutoffTime {
			return false
		}
	}

	return true
}

// CheckResponseStatus 检查API响应状态
func CheckResponseStatus(resp *Response) error {
	success := false
	for _, r := range resp.Ret {
		if r == "SUCCESS::调用成功" || r == "SUCCESS" {
			success = true
			break
		}
	}
	if !success {
		return fmt.Errorf("API返回错误: ret=%v", resp.Ret)
	}
	return nil
}
