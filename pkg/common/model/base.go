package model

import (
	"time"
)

type Base struct {
	ID         uint64    `json:"id" gorm:"primaryKey"`
	CreateTime time.Time `json:"createTime" gorm:"autoCreateTime"`
	UpdateTime time.Time `json:"updateTime" gorm:"autoUpdateTime"`
	Status     int8      `json:"status" gorm:"type:tinyint(4);default:1"`
}

type Page[T any] struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
	Items    []T `json:"items"`
}
