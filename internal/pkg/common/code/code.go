package code

import (
	"strings"
	"tone/agent/internal/pkg/common"
	"tone/agent/pkg/common/gin/code/rpccode"
)

//go:generate stringer -type ErrCode -linecomment -output code_string.go
type ErrCode int

const (
	codeSplit = " & "
)

/*
示例

	Msg & Title & Reference
*/

// 基础错误码
const (
	Success ErrCode = iota // 成功
	Fail                   // 失败
	NoCode                 // 错误码格式不正确
)

// handler 错误
const (
	ReqParseErr ErrCode = 1000 + iota //  请求参数错误
)

// control 错误
const (
	ReqUserIDEmptyErr ErrCode = 2000 + iota // controller 层错误
)

// dao mysql 错误
const (
	SQLInserErr  ErrCode = 3000 + iota // 数据库错误
	SQLQueryErr                        // 查询 出错
	SQLUpdateErr                       // 更新 出错
	SQLNotFound                        // 查询不存在
	SQLExist                           // 数据已经存在
)

// rpc  调用出错
const (
	RPCErr            ErrCode = 4000 + iota // rpc 错误
	RPCTooManyRequest                       // rpc 请求过多
)

// service 错误
const (
	ServiceErr           ErrCode = 5000 + iota //  业务层错误
	ServiceErrRetryLater                       //  服务内部错误，稍后重试

	ServiceUserExist         ErrCode = 5100 + iota // 用户存在
	ServiceUserTokenInvalid                        // 用户 token 无效
	ServiceUserPasswordErr                         // 用户密码错误
	ServiceUserPermissionErr                       // 用户权限错误

	ServiceWorkspaceInitErr               ErrCode = 5150 + iota // 工作空间初始化失败
	ServiceWorkspaceNotEnough                                   // 工作空间不足, 创建同步任务失败
	ServiceWorkspaceNotEnoughPairOpFailed                       // 工作空间不足, 部分同步任务创建失败

	ServiceModelSearchErr   ErrCode = 5200 + iota // 模型请求出错
	ServiceModelDirExist                          // 模型目录、文件已经存在
	ServiceModelDirNotFound                       // 模型目录、文件不存在
	ServiceModelIsSymlink                         // 模型目录、文件是符号链接
	ServiceModelCreateErr                         // 创建模型目录、文件失败
	ServiceModelDeleteErr                         // 删除模型目录、文件失败
)

const (
	VolcRpcVolumeErr ErrCode = 5300 + iota // 火山弹性块存储 rpc 错误
	VolcRpcTosErr                          // 火山对象存储 rpc 错误
)

const (
	PathIsNotDir           ErrCode = 5500 + iota // 路径不是目录
	PathIsNotFile                                // 路径不是文件
	PathIsNotExists                              // 路径不存在
	SrcPathIsNotExists                           // 源路径不存在
	DstPathIsNotExists                           // 目标路径不存在
	DstPathIsExists                              // 目标路径已经存在
	CreateFileErr                                // 创建文件失败
	DeleteFileErr                                // 删除文件失败
	RenameFileErr                                // 重命名文件失败
	MoveFileErr                                  // 移动文件失败
	MoveSrcDstNotParentErr                       // 移动源路径和目标路径不在同级目录
	GetFileMetaErr                               // 获取文件元信息失败
	IteratorFileErr                              // 迭代文件失败
)

// work
const (
	TaskNotExit ErrCode = 6000 + iota // task 错误
)

func (e ErrCode) Int() int {
	return int(e)
}

func (e ErrCode) Info() *rpccode.ErrInfo {
	errStr := e.String()
	info := &rpccode.ErrInfo{}
	if strings.HasPrefix(errStr, "ErrCode(") {
		nocodes := strings.Split(NoCode.String(), codeSplit)
		info.Title = nocodes[0]
		info.Msg = nocodes[1]
		info.Reference = nocodes[2]
	}

	errs := strings.Split(e.String(), codeSplit)
	switch len(errs) {
	case 0:
		nocodes := strings.Split(NoCode.String(), codeSplit)
		info.Title = nocodes[0]
		info.Msg = nocodes[1]
		info.Reference = nocodes[2]
	case 1:
		info.Msg = errs[0]
	case 2:
		info.Msg = errs[0]
		info.Title = errs[1]
	case 3:
		info.Msg = errs[0]
		info.Title = errs[1]
		info.Reference = errs[2]
	default:
		info.Msg = errs[0]
		info.Title = errs[1]
		info.Reference = strings.Join(errs[2:], " ")
	}
	return info
}

// TODO: 封装 utils
func (e ErrCode) Msg(data ...any) *common.Message {
	if e != Success {
		return &common.Message{Code: e.Int(), Error: e.Info()}
	}

	if e == Success && len(data) != 0 {
		if len(data) == 1 {
			return &common.Message{Code: e.Int(), Data: data[0]}
		} else {
			return &common.Message{Code: e.Int(), Data: data}
		}
	}

	return &common.Message{Code: e.Int()}
}
