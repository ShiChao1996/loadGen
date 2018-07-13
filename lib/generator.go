/*
 * Revision History:
 *     Initial: 2018/7/04        ShiChao
 */

package lib

import (
	"time"
	"sync/atomic"
	"context"
	"fmt"
	"os"
)

type generator struct {
	// loads per second
	lps       uint32
	callCount uint32
	// concurrency: single call response time div interval to send request
	concurrency     uint32
	status          uint32
	ctx             context.Context
	cancelFunc      context.CancelFunc
	timeout         time.Duration // single request max time, this is used to estimate the concurrency
	duration        time.Duration
	tickets         Tickets
	caller          Caller
	resultCh        chan *Result
	signals         chan os.Signal
	beforeStopFuncs []func()
}

func NewGen(caller Caller, timeout time.Duration, lps uint32, duration time.Duration, resultCh chan *Result) LoadGenerator {
	g := &generator{
		lps:      lps,
		caller:   caller,
		timeout:  timeout,
		duration: duration,
		resultCh: resultCh, // note: this should be passed in by user because the genLoad should not handle the result,just keep it simple.
		signals:  make(chan os.Signal),
	}
	g.concurrency = uint32(timeout) / (1e9 / lps) // note: = lps * second
	g.tickets = NewTickets(g.concurrency)
	return g
}

func (g *generator) Start() bool {
	if ok := atomic.CompareAndSwapUint32(&g.status, STATUS_ORIGIN, STATUS_STARTTING); !ok {
		return false
	}
	g.ctx, g.cancelFunc = context.WithTimeout(context.Background(), g.duration)

	go g.genLoad()
	g.configureSignals()
	go g.listenSignals()

	return true
}

func (g *generator) Stop() {
	g.cancelFunc()
	g.stopGraceful()
}

func (g *generator) Status() uint32 {
	return atomic.LoadUint32(&g.status)
}

func (g *generator) CallCount() uint32 {
	return atomic.LoadUint32(&g.callCount)
}

func (g *generator) BeforeExit(fn func()) {
	if fn == nil {
		fmt.Println("fn should not be nil")
		return
	}
	g.beforeStopFuncs = append(g.beforeStopFuncs, fn)
}

func (g *generator) genLoad() {
	ticker := time.NewTicker(time.Duration(1e9 / g.lps))
	go g.callOne() // note: immediately invoke one
	for {
		select {
		case <-g.ctx.Done():
			g.stopGraceful()
			return
		case <-ticker.C:
			g.tickets.Get()
			if atomic.LoadUint32(&g.status) == STATUS_STOPPED {
				return // note: in case that g.ctx timeout while waiting for tickets.And remember close tickets channel after timeout, or not it may block here for time.
			}
			go g.callOne()
		}
	}
}

func (g *generator) stopGraceful() {
	atomic.CompareAndSwapUint32(&g.status, STATUS_STARTTED, STATUS_STOPPING)
	fmt.Println("closing the result channel...")
	close(g.resultCh)
	atomic.StoreUint32(&g.status, STATUS_STOPPED)
}

func (g *generator) callOne() {
	var (
		callStatus = STATUS_UNCALL
		result     *Result
		req        = g.caller.BuildReq()
	)

	defer func() {
		g.tickets.Put()
		if err := recover(); err != nil {
			fmt.Println("errrrr: ", err)
			result = &Result{
				ID:     0,
				Req:    req,
				Resp:   nil,
				Code:   RET_CODE_FATAL_CALL,
				Msg:    "call fetal error",
				Elapse: 0,
			}
			g.sendResult(result)
		}
	}()

	timer := time.AfterFunc(g.timeout, func() {
		if !atomic.CompareAndSwapUint32(&callStatus, STATUS_UNCALL, STATUS_CALLED) {
			return
		}
		result = &Result{
			ID:     req.ID,
			Req:    req,
			Resp:   nil,
			Code:   RET_CODE_WARNING_CALL_TIMEOUT,
			Msg:    "call did not receive response until timeout",
			Elapse: g.timeout,
		}
		g.sendResult(result)
	})

	resp := g.doCallOne(req.Req, &callStatus)
	if !atomic.CompareAndSwapUint32(&callStatus, STATUS_UNCALL, STATUS_CALLED) {
		return
	}
	timer.Stop()

	if resp.Err != nil {
		result = &Result{
			ID:     req.ID,
			Req:    req,
			Resp:   resp,
			Code:   RET_CODE_ERROR_CALL,
			Msg:    "call error",
			Elapse: resp.Elapse,
		}
	} else {
		result = &Result{
			ID:     req.ID,
			Req:    req,
			Resp:   resp,
			Code:   RET_CODE_SUCCESS,
			Msg:    "call success",
			Elapse: resp.Elapse,
		}
	}

	g.sendResult(result)
}

func (g *generator) doCallOne(req []byte, callStatus *uint32) (res *RawResp) {
	start := time.Now().Nanosecond()
	resp, err := g.caller.Call(req)
	end := time.Now().Nanosecond()

	atomic.AddUint32(&g.callCount, 1)
	return &RawResp{
		Resp:   resp,
		Err:    err,
		Elapse: time.Duration(end - start),
	}
}

func (g *generator) sendResult(res *Result) bool {
	if atomic.LoadUint32(&g.status) == STATUS_STOPPED {
		return false
	}
	select {
	case g.resultCh <- res:
		return true
	default:
		fmt.Printf("Ingnore one result : %v\n", res)
		return false
	}
}
