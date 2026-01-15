package service

import (
	"xianyu_aner/pkg/feishu"
)

// Deduplicator 去重服务（通用）
type Deduplicator struct{}

// NewDeduplicator 创建去重服务
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{}
}

// DeduplicateProducts 去重商品（通过飞书表格查询）
func (d *Deduplicator) DeduplicateProducts(bitableService *feishu.BitableService, tableID string, products []feishu.Product) ([]feishu.Product, error) {
	return bitableService.DeduplicateProducts(tableID, products)
}
