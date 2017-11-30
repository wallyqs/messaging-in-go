package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

type client interface {
	processInfo(line string)
	processMsg(subj string, reply string, sid int, payload []byte)
	processPing()
	processPong()
	processErr(msg string)
	processOpErr(err error)
}

// parserLoop is ran as a goroutine and processes commands
// sent by the server.
func parserLoop(c client, conn net.Conn) {
	r := bufio.NewReader(conn)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			c.processOpErr(err)
			return
		}
		args := strings.SplitN(line, " ", 2)
		if len(args) < 1 {
			c.processOpErr(errors.New("Error: malformed control line"))
			return
		}

		op := strings.TrimSpace(args[0])
		switch op {
		case "MSG":
			var subject, reply string
			var sid, size int

			n := strings.Count(args[1], " ")
			switch n {
			case 2:
				// No reply inbox case.
				// MSG foo 1 3\r\n
				// bar\r\n
				_, err := fmt.Sscanf(args[1], "%s %d %d",
					&subject, &sid, &size)
				if err != nil {
					c.processOpErr(err)
					return
				}
			case 3:
				// With reply inbox case.
				// MSG foo 1 bar 4\r\n
				// quux\r\n
				_, err := fmt.Sscanf(args[1], "%s %d %s %d",
					&subject, &sid, &reply, &size)
				if err != nil {
					c.processOpErr(err)
					return
				}
			default:
				c.processOpErr(errors.New("nats: bad control line"))
				return
			}

			// Prepare buffer for the payload
			payload := make([]byte, size)
			_, err = io.ReadFull(r, payload)
			if err != nil {
				c.processOpErr(err)
				return
			}
			c.processMsg(subject, reply, sid, payload)
		case "INFO":
			c.processInfo(args[1])
		case "PING":
			c.processPing()
		case "PONG":
			c.processPong()
		case "+OK":
			// Do nothing.
		case "-ERR":
			c.processErr(args[1])
		}
	}
}
