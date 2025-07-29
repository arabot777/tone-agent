package code

import (
	"tone/agent/pkg/common/gin/code/rpccode"
)

const OK = 0

const (
	ErrBadParams           int = iota + 100001 // 参数错误接口
	ErrInternal                = 100002        // 服务内部错误
	ErrPermissionDenied        = 100003        // 禁止访问
	ErrNotFound                = 100004        // 未找到资源
	ErrServiceUnavailable      = 100005        // 服务不可用，(如：被熔断了)
	ErrBadID                   = 100006        // ID错误
	ErrBadPageSize             = 100007        // page size错误
	ErrBadPageNum              = 100008        // page number错误
	ErrCodeInvalidPassword     = 100009        // 密码错误
)

/**
通用错误码:     10xxxx
*/

// Common: basic errors.
// Code must start with 1xxxxx.
const (
	// ErrUnknown - : Internal server error.
	ErrUnknown int = iota + 100101

	// ErrBind - : Error occurred while binding the request body to the struct.
	ErrBind

	// ErrValidation - : Validation failed.
	ErrValidation

	// ErrTokenInvalid - : Token invalid.
	ErrTokenInvalid
)

// common: database errors.
const (
	// ErrDatabase - : Database error.
	ErrDatabase int = iota + 100201

	// ErrDatabaseTx - : Create database transaction error.
	ErrDatabaseTx

	// ErrDataExisted - : Data has already been created.
	ErrDataExisted

	// ErrDataReference - : Data has a reference and cannot be deleted.
	ErrDataReference

	// ErrRowsAffectedNone - : Data has a none rows affected
	ErrRowsAffectedNone
)

// common: encode/decode errors.
const (
	// ErrEncodingFailed - : Encoding failed due to an error with the data.
	ErrEncodingFailed int = iota + 100301

	// ErrDecodingFailed - : Decoding failed due to an error with the data.
	ErrDecodingFailed

	// ErrInvalidJSON - : Data is not valid JSON.
	ErrInvalidJSON

	// ErrEncodingJSON - : JSON data could not be encoded.
	ErrEncodingJSON

	// ErrDecodingJSON - : JSON data could not be decoded.
	ErrDecodingJSON

	// ErrInvalidYaml - : Data is not valid Yaml.
	ErrInvalidYaml

	// ErrEncodingYaml - : Yaml data could not be encoded.
	ErrEncodingYaml

	// ErrDecodingYaml - : Yaml data could not be decoded.
	ErrDecodingYaml
)

const (
	// ErrSignDecode - : Sign key decode error, Verify failed.
	ErrSignDecode int = iota + 100401
)

const (
	ErrHttpConnection int = iota + 100601
	ErrHttpRespCode
)

func init() {
	rpccode.Register(ErrBadParams, "请求的参数错误", "请求错误", "")
	rpccode.Register(ErrInternal, "服务器内部错误", "服务器错误", "")
	rpccode.Register(ErrServiceUnavailable, "服务器暂时不可用，请稍后重试", "服务器错误", "")
	rpccode.Register(ErrPermissionDenied, "没有权限访问该资源", "缺少权限", "")
	rpccode.Register(ErrNotFound, "未找到请求的资源", "not found", "")
	rpccode.Register(ErrBadID, "参数错误: 无效的ID", "请求错误", "")
	rpccode.Register(ErrBadPageSize, "分页错误: pgSz错误", "请求错误", "")
	rpccode.Register(ErrBadPageNum, "分页错误: pgNum错误", "请求错误", "")

	rpccode.Register(ErrUnknown, "Internal server error", "Basic errors", "")
	rpccode.Register(ErrBind, "Error occurred while binding the request body to the struct", "Basic errors", "")
	rpccode.Register(ErrValidation, "Validation failed", "Basic errors", "")
	rpccode.Register(ErrTokenInvalid, "Token invalid", "Basic errors", "")
	rpccode.Register(ErrDatabase, "Database error", "Database errors", "")
	rpccode.Register(ErrEncodingFailed, "Encoding failed due to an error with the data", "Encode/Decode errors", "")
	rpccode.Register(ErrDecodingFailed, "Decoding failed due to an error with the data", "Encode/Decode errors", "")
	rpccode.Register(ErrInvalidJSON, "Data is not valid JSON", "Encode/Decode errors", "")
	rpccode.Register(ErrEncodingJSON, "JSON data could not be encoded", "Encode/Decode errors", "")
	rpccode.Register(ErrDecodingJSON, "JSON data could not be decoded", "Encode/Decode errors", "")
	rpccode.Register(ErrInvalidYaml, "Data is not valid Yaml", "Encode/Decode errors", "")
	rpccode.Register(ErrEncodingYaml, "Yaml data could not be encoded", "Encode/Decode errors", "")
	rpccode.Register(ErrDecodingYaml, "Yaml data could not be decoded", "Encode/Decode errors", "")
	rpccode.Register(ErrDatabaseTx, "Create database transaction error", "Database errors", "")
	rpccode.Register(ErrDataExisted, "Data has already been created", "Database errors")
	rpccode.Register(ErrDataReference, "Data has a reference and cannot be deleted", "Database errors")

	rpccode.Register(ErrSignDecode, "Sign key decode error, Verify failed.", "Sign errors")

	rpccode.Register(ErrHttpConnection, "get http connection error.", "Http errors")
	rpccode.Register(ErrHttpRespCode, "get http connection error.", "Http errors")
	rpccode.Register(ErrRowsAffectedNone, "update rows affected is zero", "Database errors", "")

	rpccode.Register(ErrCodeInvalidPassword, "密码错误", "请求错误", "")
}
