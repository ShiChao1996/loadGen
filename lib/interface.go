package lib

import "time"

const (
	STATUS_ORIGIN    uint32 = iota
	STATUS_STARTTING
	STATUS_STARTTED
	STATUS_STOPPED
)

type LoadGenerator interface {
	// Start start to generator load
	Start() bool

	// Stop stops the generator
	Stop()

	Status() uint32

	// CallCount returns number of calls
	CallCount() uint32
}

type Caller interface {
	BuildReq() RawReq

	Call(req []byte) ([]byte, error)

	CheckResp(req RawReq, resp RawResp) bool
}

/*
type RawReq interface {
	ID() uint64

	// Bytes marshals RawReq to []byte to send
	Bytes() []byte
}

type RawResp interface {
	ID() uint64

	// Bytes marshals RawResp to []byte
	Bytes() []byte

	Err() error

	// Elapse returns time duration that the server used to handle request
	Elapse() time.Duration
}
*/
type Tickets interface {
	Init()
	Put()
	Get() bool
	//Active() bool
	Total() uint32
	Remainder() uint32
	Close()
}

type Result struct {
	ID     int32
	Req    RawReq
	Resp   RawResp
	Code   int32 // retcode
	Msg    string
	Elapse time.Duration
}

type RawReq struct {
	id  int32
	req []byte
}

func NewReq(id int32, req []byte) RawReq {
	return RawReq{id, req}
}

type RawResp struct {
	id  int32
	res []byte
	err error
}

func NewResp(id int32, res []byte, err error) RawResp {
	return RawResp{
		id,
		res,
		err,
	}
}
