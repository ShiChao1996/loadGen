package lib

import (
	"time"
	"sync/atomic"
	"context"
	"log"
	"fmt"
)

var logger = log.Logger{}

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
	waitStatis chan struct{} // wait for statistic
}

func NewGen(caller Caller, timeout time.Duration, lps uint32, duration time.Duration) LoadGenerator {
	g := &generator{
		lps:        lps,
		caller:     caller,
		timeout:    timeout,
		duration:   duration,
		waitStatis: make(chan struct{}),
		//resultCh: make(chan *Result, ), //todo:use cached channel
	}
	g.concurrency = uint32(timeout) / (1e9 / lps)
	g.tickets = NewTickets(g.concurrency)
	g.resultCh = make(chan *Result, g.concurrency)
	return g
}

func (g *generator) Start() bool {
	if ok := atomic.CompareAndSwapUint32(&g.status, STATUS_ORIGIN, STATUS_STARTTING); !ok {
		return false
	}
	g.tickets.Init()
	g.callCount = 0
	g.ctx, g.cancelFunc = context.WithTimeout(context.Background(), g.duration)
	ticker := time.NewTicker(time.Duration(1e9 / g.lps))

	go func() {
		count := 0
		for {
			select {
			case <-g.ctx.Done():
				{
					fmt.Printf("use %d goroutines", count)
					return // todo: gracefully stop
				}
			default:
			}

			// g.tickets.Get() note: shouldn't be here, if it blocks, never get done from the loop
			if ok := g.tickets.Get(); ok { // note: when we want to stop it, we close the channel, then it can run along, but shouldn't setup a new goroutine
				go g.call(ticker)
				count++
			}
		}
	}()

	go g.statistic()
	g.wait()
	a := make(chan int)
	a <- 0
	return true
}

func (g *generator) Stop() {
	if atomic.SwapUint32(&g.status, STATUS_STOPPED) == STATUS_STOPPED {
		logger.Println("already closed !")
		return
	}

	g.tickets.Close()
	g.cancelFunc()
	close(g.resultCh)

}

func (g *generator) Status() uint32 {
	return atomic.LoadUint32(&g.status)
}

func (g *generator) CallCount() uint32 {
	return atomic.LoadUint32(&g.callCount)
}

func (g *generator) call(ticker *time.Ticker) {
	defer func() {
		g.tickets.Put()
	}()
	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			go g.callOne()
		}
	}
}

func (g *generator) callOne() {
	defer func() {
		if err := recover(); err != nil {

		}
	}()

	var result *Result
	caller := g.caller
	req := caller.BuildReq()
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

	g.resultCh <- result

}

func (g *generator) statistic() {
	var (
		count       uint64 = 0
		minElapse   time.Duration
		maxElapse   time.Duration
		totalElapse time.Duration = 0
	)
	firstRes := <-g.resultCh
	minElapse = firstRes.Elapse
	totalElapse = firstRes.Elapse
	count++
	for result := range g.resultCh {
		if minElapse > result.Elapse {
			minElapse = result.Elapse
		}
		if maxElapse < result.Elapse {
			maxElapse = result.Elapse
		}
		totalElapse += result.Elapse
		count ++
	}
	averageElapse := uint64(totalElapse) / count
	fmt.Printf("total calls: %d \n minElapse: %d \n maxElapse: %d \n averageElapse: %d \n", count, minElapse, maxElapse, averageElapse)
	g.waitStatis <- struct{}{}
}

func (g *generator) wait() {
	<-g.ctx.Done()
	close(g.resultCh)
	<-g.waitStatis

}
