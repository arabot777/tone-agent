package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tone/agent/pkg/common/logger"

	"gorm.io/gorm"
)

// PrintSQL 将GORM查询的SQL语句和参数拼接并打印
// 返回完整的可执行SQL语句（参数已替换）
func PrintSQL(ctx context.Context, db *gorm.DB, operation string) string {
	// 创建一个DryRun会话来获取SQL而不执行
	stmt := db.Statement
	sqlStr := stmt.SQL.String()
	vars := stmt.Vars

	// 将参数拼接到SQL语句中
	completeSql := sqlStr
	for _, v := range vars {
		var strVal string
		switch val := v.(type) {
		case string:
			// 处理字符串，需要特别处理保留字
			// 如果是枚举值，可能是数据库保留字，需要用反引号包裹
			if isReservedWord(val) {
				strVal = "`" + val + "`"
			} else {
				strVal = "'" + strings.ReplaceAll(val, "'", "''") + "'" // 转义单引号
			}
		case []uint64:
			// 处理uint64切片
			ids := make([]string, len(val))
			for i, id := range val {
				ids[i] = fmt.Sprintf("%d", id)
			}
			strVal = "(" + strings.Join(ids, ",") + ")"
		case []int64:
			// 处理int64切片
			ids := make([]string, len(val))
			for i, id := range val {
				ids[i] = fmt.Sprintf("%d", id)
			}
			strVal = "(" + strings.Join(ids, ",") + ")"
		case []int:
			// 处理int切片
			ids := make([]string, len(val))
			for i, id := range val {
				ids[i] = fmt.Sprintf("%d", id)
			}
			strVal = "(" + strings.Join(ids, ",") + ")"
		case time.Time:
			strVal = fmt.Sprintf("'%s'", val.Format("2006-01-02 15:04:05"))
		case int, int64, uint, uint64:
			strVal = fmt.Sprintf("%d", val)
		case float32, float64:
			strVal = fmt.Sprintf("%f", val)
		case bool:
			if val {
				strVal = "1"
			} else {
				strVal = "0"
			}
		case nil:
			strVal = "NULL"
		default:
			// 处理枚举类型
			s := fmt.Sprintf("%v", val)
			if isReservedWord(s) {
				strVal = "`" + s + "`"
			} else {
				strVal = s
			}
		}
		// 替换第一个?
		completeSql = strings.Replace(completeSql, "?", strVal, 1)
	}

	// 打印SQL语句
	logger.Infof(ctx, "\n\n=== %s SQL ===\n%s\n=== 查询结束 ===\n\n", operation, completeSql)

	return completeSql
}

// PrintQuerySQL 打印查询SQL语句
func PrintQuerySQL(ctx context.Context, query *gorm.DB) string {
	return PrintSQL(ctx, query, "查询")
}

// PrintCountSQL 打印计数SQL语句
func PrintCountSQL(ctx context.Context, query *gorm.DB) string {
	return PrintSQL(ctx, query, "计数")
}

// PrintUpdateSQL 打印更新SQL语句
func PrintUpdateSQL(ctx context.Context, query *gorm.DB) string {
	return PrintSQL(ctx, query, "更新")
}

// PrintDeleteSQL 打印删除SQL语句
func PrintDeleteSQL(ctx context.Context, query *gorm.DB) string {
	return PrintSQL(ctx, query, "删除")
}

// PrintInsertSQL 打印插入SQL语句
func PrintInsertSQL(ctx context.Context, query *gorm.DB) string {
	return PrintSQL(ctx, query, "插入")
}

// isReservedWord 检查是否是MySQL保留字
func isReservedWord(word string) bool {
	// MySQL保留字列表，可以根据需要扩充
	reservedWords := map[string]bool{
		"group":        true,
		"order":        true,
		"key":          true,
		"level":        true,
		"user":         true,
		"organization": true,
		"tree_node":    true,
		"knowledge":    true,
		"agent":        true,
		"status":       true,
		"limit":        true,
		"offset":       true,
		"select":       true,
		"update":       true,
		"delete":       true,
		"insert":       true,
		"where":        true,
		"from":         true,
		"join":         true,
		"left":         true,
		"right":        true,
		"inner":        true,
		"outer":        true,
		"on":           true,
		"as":           true,
		"by":           true,
	}

	return reservedWords[strings.ToLower(word)]
}
