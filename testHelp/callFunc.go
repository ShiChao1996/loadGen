/*
 * Revision History:
 *     Initial: 2018/7/05        ShiChao
 */

package testHelp

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
	"log"
)

func call(req []byte) []byte {
	conn, err := net.Dial("tcp", "127.0.0.1:6666")
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Connecting to 127.0.0.1:6666")

	var wg sync.WaitGroup
	resch := make(chan []byte)
	wg.Add(1)
	go handleWrite(conn, req, &wg)
	go handleRead(conn, resch)
	wg.Wait()

	res := <-resch
	return res
}
func handleWrite(conn net.Conn, req []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := conn.Write(req)
	if err != nil {
		log.Panic(err)
	}
}
func handleRead(conn net.Conn, ch chan []byte) {
	reader := bufio.NewReader(conn)

	line, err := reader.ReadBytes(byte('\n'))
	if err != nil {
		fmt.Print("Error to read message because of ", err)
	}
	fmt.Println("line: " + string(line))
	line = append(line)
	ch <- line
}
