package common

import "tone/agent/pkg/common/gin/code/rpccode"

type Message struct {
	Code  int              `json:"code"`
	Error *rpccode.ErrInfo `json:"error,omitempty"`
	Data  any              `json:"data,omitempty"`
}
