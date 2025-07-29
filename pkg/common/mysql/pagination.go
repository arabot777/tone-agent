package mysql

import (
	"tone/agent/pkg/common/model"

	"gorm.io/gorm"
)

func pretreatment(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 10
	}
	return page, pageSize
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func PageWrapper[T any](page, pageSize int, db *gorm.DB, orderBy string) (*model.Page[T], error) {
	page, pageSize = pretreatment(page, pageSize)
	var total int64
	var items []T
	if err := db.Count(&total).Scopes(Paginate(page, pageSize)).Order(orderBy).Find(&items).Error; err != nil {
		return nil, err
	}
	return &model.Page[T]{
		Page:     page,
		PageSize: pageSize,
		Total:    int(total),
		Items:    items,
	}, nil
}
