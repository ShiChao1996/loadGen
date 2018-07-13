/*
 * Revision History:
 *     Initial: 2018/7/05        ShiChao
 */

 package testHelp

import (
	"github.com/ShiChao1996/loadGen/lib"
	"sync/atomic"
)

type Caller struct {
	id   int32
	req  lib.RawReq
	resp lib.RawResp
}

var id = uint32(0)

func genID() uint32 {
	return atomic.AddUint32(&id, 1)
}

func (c *Caller) BuildReq() *lib.RawReq {
	return lib.NewReq(genID(), []byte("aaaaa \n"))
}

func (c *Caller) Call(req []byte) (resp []byte, err error) {
	resp = call(req)
	return resp, nil
}

func (c *Caller) CheckResp(req lib.RawReq, resp lib.RawResp) bool {
	return true
}
