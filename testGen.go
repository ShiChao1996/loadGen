/*
 * Revision History:
 *     Initial: 2018/7/05        ShiChao
 */

 package main

import (
	"time"
	"fmt"
	"github.com/ShiChao1996/loadGen/lib"
	"github.com/ShiChao1996/loadGen/testHelp"
)

func main() {
	c := &testHelp.Caller{}
	resultCh := make(chan *lib.Result, 20)
	gen := lib.NewGen(c, 1*time.Second, 50, 2*time.Second, resultCh)
	gen.Start()

	var (
		count         uint64 = 0
		minElapse            = 10 * time.Second
		maxElapse     time.Duration
		totalElapse   time.Duration = 0
		averageElapse uint64
	)

	gen.BeforeExit(func() {
		averageElapse = uint64(totalElapse) / count
		fmt.Printf("total calls: %d \n minElapse: %d ns\n maxElapse: %d ns\n averageElapse: %d ns\n", count, minElapse, maxElapse, averageElapse)
	})

	for result := range resultCh { // note: eg,there can do more thing to analise the res.
		if minElapse > result.Elapse {
			minElapse = result.Elapse
		}
		if maxElapse < result.Elapse {
			maxElapse = result.Elapse
		}
		totalElapse += result.Elapse
		count ++
	}
	averageElapse = uint64(totalElapse) / count
	fmt.Printf("total calls: %d \n minElapse: %d ns\n maxElapse: %d ns\n averageElapse: %d ns\n", count, minElapse, maxElapse, averageElapse)
}
