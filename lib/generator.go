package lib

import (
	"time"
	"sync/atomic"
	"context"
)

type generator struct {
	// loads per second
	lps         uint32
	callCount   uint32
	concurrency uint32
	status      uint32

	ctx        context.Context
	cancelFunc context.CancelFunc
	timeout    time.Duration
	duration   time.Duration
	tickets    Tickets
	caller     Caller
	resultCh   chan *Result
}

func NewGen(caller Caller, timeout time.Duration, lps uint32, duration time.Duration) LoadGenerator {
	g := &generator{
		lps:      lps,
		caller:   caller,
		timeout:  timeout,
		duration: duration,
		resultCh: make(chan *Result), //todo:use cached channel
	}
	g.concurrency = uint32(timeout) / (1e9 / lps)
	g.tickets = NewTickets(g.concurrency)
	return g
}

func (g *generator) Start() bool {
	if ok := atomic.CompareAndSwapUint32(&g.status, STATUS_ORIGIN, STATUS_STARTTING); !ok {
		return false
	}
	g.tickets.Init()
	g.callCount = 0
	g.ctx, g.cancelFunc = context.WithCancel(context.Background())
	ticker := time.NewTicker(time.Duration(1e9 / g.lps))

	go func() {
		for {
			select {
			case g.ctx.Done():
				{
					return // todo: gracefully stop
				}
			default:
			}

			g.tickets.Get()
			go g.call(ticker)
		}
	}()
	return true
}

func (g *generator) Stop() bool {
	return false
}

func (g *generator) Status() uint32 {
	return atomic.LoadUint32(&g.status)
}

func (g *generator) CallCount() uint32 {
	return atomic.LoadUint32(&g.callCount)
}

func (g *generator) call(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			go g.callOne()
		}
	}

}

func (g *generator) callOne() (result *Result) {
	defer func() {
		if err := recover(); err != nil {

		}
	}()

	caller := g.caller
	req := caller.BuildReq()
	resp, err := caller.Call(req.req)

	result = &Result{
		ID:  req.id,
		Req: req,
		Resp: RawResp{
			id:  req.id,
			res: resp,
			err: err,
		},
	}

	return result
}
