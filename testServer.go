/*
 * Revision History:
 *     Initial: 2018/7/05        ShiChao
 */

package main

import (
	"net"
	"bufio"
	"fmt"
)

func main() {
	listener, _ := net.Listen("tcp", ":6666")
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go handle(conn)
	}
}

func handle(conn net.Conn) {
	reader := bufio.NewReader(conn)
	str, _ := reader.ReadString(byte('\n'))
	conn.Write([]byte(fmt.Sprintf("hello, you said: %s \n", str)))
	conn.Close()
}
