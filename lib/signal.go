/*
 * Revision History:
 *     Initial: 2018/7/13        ShiChao
 */

package lib

import (
	"os/signal"
	"syscall"
)

func (g *generator) configureSignals() {
	signal.Notify(g.signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
}

func (g *generator) listenSignals() {
	for {
		sig := <-g.signals

		switch sig {
		case syscall.SIGUSR1:
		default:
			g.Stop()
			return
		}
	}
}
