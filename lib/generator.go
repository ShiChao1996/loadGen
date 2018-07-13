/*
 * Revision History:
 *     Initial: 2018/7/04        ShiChao
 */

package lib

import (
	"time"
	"sync/atomic"
	"context"
	"log"
	"fmt"
	"errors"
	"os"
)

var logger = log.Logger{}

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
	g.tickets.Init()
	g.callCount = 0
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
		result *Result
		req    RawReq
	)

	defer func() {
		g.tickets.Put()
		if err := recover(); err != nil {
			result = &Result{
				ID:  req.id,
				Req: req,
				Resp: RawResp{
					id:  req.id,
					res: nil,
					err: errors.New("call func crashed"),
				},
				Elapse: 0,
			}
			g.sendResult(result)
		}
	}()

	caller := g.caller
	req = caller.BuildReq()

	start := time.Now().Nanosecond()
	resp, err := caller.Call(req.req)
	end := time.Now().Nanosecond()

	result = &Result{
		ID:  req.id,
		Req: req,
		Resp: RawResp{
			id:  req.id,
			res: resp,
			err: err,
		},
		Elapse: time.Duration(end - start),
	}

	g.sendResult(result)
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
