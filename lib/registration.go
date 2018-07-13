/*
 * Revision History:
 *     Initial: 2018/7/03        ShiChao
 */

package lib

import "time"

const (
	STATUS_ORIGIN    uint32 = iota
	STATUS_STARTTING
	STATUS_STARTTED
	STATUS_STOPPING
	STATUS_STOPPED
)

const (
	RET_CODE_SUCCESS              uint32 = 0    // 成功。
	RET_CODE_WARNING_CALL_TIMEOUT        = 1001 // 调用超时警告。
	RET_CODE_ERROR_CALL                  = 2001 // 调用错误。
	RET_CODE_ERROR_RESPONSE              = 2002 // 响应内容错误。
	RET_CODE_ERROR_CALEE                 = 2003 // 被调用方（被测软件）的内部错误。
	RET_CODE_FATAL_CALL                  = 3001 // 调用过程中发生了致命错误！
)

const (
	STATUS_UNCALL  uint32 = iota
	STATUS_CALLED
)

type LoadGenerator interface {
	// Start start to generator load
	Start() bool

	// Stop this will be invoked when actively stop the generator or unexpectedly. see signal. you can collection some resource before exit.
	Stop()

	Status() uint32

	// CallCount returns number of calls
	CallCount() uint32

	// BeforeStop this will be called while program exit invalidly
	BeforeExit(fn func())
}

type Caller interface {
	BuildReq() *RawReq

	Call(req []byte) ([]byte, error)

	CheckResp(req RawReq, resp RawResp) bool
}

type Tickets interface {
	Put()
	Get() bool
	Total() uint32
	Remainder() uint32
}

type Result struct {
	ID     uint32
	Req    *RawReq
	Resp   *RawResp
	Code   uint32 // retcode
	Msg    string
	Elapse time.Duration
}

type RawReq struct {
	ID  uint32
	Req []byte
}

func NewReq(id uint32, req []byte) *RawReq {
	return &RawReq{id, req}
}

type RawResp struct {
	Resp []byte
	Err error
	Elapse time.Duration
}
