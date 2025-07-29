package rpccode

import (
	_errors "tone/agent/pkg/common/pkgerror"
)

var errors Errors

type Errors interface {
	Render(code int) ErrInfo
	Register(code int, fields ...string)
	ParseErrInfo(err error) (int, ErrInfo)
}

type set struct {
	list map[int]ErrInfo
}

func (s *set) Render(code int) ErrInfo {
	return s.list[code]
}

func newErrorSet() Errors {
	s := &set{
		list: make(map[int]ErrInfo),
	}
	return s
}

// Register(300001, "msg", "title", "reference")
// 注册 错误码
func (s *set) Register(code int, fields ...string) {
	fieldsSz := len(fields)
	ei := ErrInfo{}
	switch {
	case fieldsSz > 2:
		ei.Reference = fields[2]
		fallthrough
	case fieldsSz > 1:
		ei.Title = fields[1]
		fallthrough
	case fieldsSz > 0:
		ei.Msg = fields[0]
	}
	s.list[code] = ei
	register(code, ei.Msg, ei.Reference)
}

func init() {
	errors = newErrorSet()
}

func (s *set) ParseErrInfo(err error) (int, ErrInfo) {
	coder := _errors.ParseCoder(err)

	// 获取预定义的错误模板
	preDefinedErrInfo := Render(coder.Code())

	// 优先使用WithCode包装的具体错误信息
	actualErrMsg := _errors.ParseError(err)

	// 如果WithCode包装了具体的错误信息，优先使用它
	if actualErrMsg != "" {
		return coder.Code(), ErrInfo{
			Title:     preDefinedErrInfo.Title,     // 使用预定义的标题
			Msg:       actualErrMsg,                // 使用WithCode包装的具体错误信息
			Reference: preDefinedErrInfo.Reference, // 使用预定义的参考信息
		}
	}

	// 如果没有具体错误信息，使用预定义的错误模板
	if preDefinedErrInfo.Msg != "" {
		return coder.Code(), preDefinedErrInfo
	}

	// 兜底处理：既没有具体错误信息，也没有预定义的错误模板
	return coder.Code(), ErrInfo{
		Title: "Uncaught exception",
		Msg:   err.Error(), // 使用原始错误信息作为兜底
	}
}
