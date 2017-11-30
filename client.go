package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
)

// MsgHandler is the callback executed when a message is received.
type MsgHandler func(subject, reply string, payload []byte)

// Msg represents message received by the server.
type Msg struct {
	Subject string
	Reply   string
	Data    []byte
	next    *Msg
}

// Subscription represents a subscription by the client.
type Subscription struct {
	pHead *Msg
	pTail *Msg
	pCond *sync.Cond
	mu    sync.Mutex
}

// Client is a basic client to the NATS server.
type Client struct {
	conn net.Conn
	w    *bufio.Writer

	// sid increments monotonically per subscription and it is
	// used to identify a subscription from the client when
	// receiving a message.
	sid int

	// subs maps a subscription identifier to a callback.
	subs map[int]*Subscription

	sync.Mutex

	fch chan struct{}

	// respSub is the wildcard subject on which the responses
	// are received.
	respSub string

	// respMap is a map of request inboxes to the channel that is
	// signaled when the response is received.
	respMap map[string]chan []byte

	// respMux is the subscription on which the requests
	// are received.
	respMux *Subscription

	// respSetup is used to set up the wildcard subscription
	// for requests.
	respSetup sync.Once
}

// NewClient returns a NATS client.
func NewClient() *Client {
	return &Client{
		subs: make(map[int]*Subscription),
	}
}

// Connect establishes a connection to a NATS server.
func (c *Client) Connect(netloc string) error {
	conn, err := net.Dial("tcp", netloc)
	if err != nil {
		return err
	}
	c.conn = conn
	c.w = bufio.NewWriter(conn)
	c.fch = make(chan struct{}, 1024)

	// Deactivate verbose mode
	connect := struct {
		Verbose bool `json:"verbose"`
	}{}
	connectOp, err := json.Marshal(connect)
	if err != nil {
		return err
	}
	connectCmd := fmt.Sprintf("CONNECT %s\r\n", connectOp)
	_, err = c.w.WriteString(connectCmd)
	if err != nil {
		return err
	}

	err = c.w.Flush()
	if err != nil {
		return err
	}

	// Spawn goroutine for the parser reading loop.
	go parserLoop(c, conn)

	// Spawn a flusher goroutine which waits to be signaled
	// that there is pending data to be flushed.
	go func() {
		for {
			if _, ok := <-c.fch; !ok {
				return
			}
			c.Lock()
			if c.w.Buffered() > 0 {
				c.w.Flush()
			}
			c.Unlock()
		}
	}()

	return nil
}

// Close terminates a connection to NATS.
func (c *Client) Close() {
	c.Lock()
	defer c.Unlock()
	c.conn.Close()
}

const (
	// Used for hand rolled itoa
	digits = "0123456789"
	_CRLF_ = "\r\n"
)

// Publish takes a subject and payload then sends
// the message to the server.
func (c *Client) Publish(subject string, data []byte) error {
	c.Lock()
	defer c.Unlock()

	msgh := []byte("PUB ")
	msgh = append(msgh, subject...)
	msgh = append(msgh, ' ')

	// msgh = strconv.AppendInt(msgh, int64(len(data)), 10)

	// faster alternative
	var b [12]byte
	var i = len(b)
	if len(data) > 0 {
		for l := len(data); l > 0; l /= 10 {
			i -= 1
			b[i] = digits[l%10]
		}
	} else {
		i -= 1
		b[i] = digits[0]
	}
	msgh = append(msgh, b[i:]...)
	msgh = append(msgh, _CRLF_...)
	_, err := c.w.Write(msgh)
	if err != nil {
		return err
	}
	_, err = c.w.Write(data)
	if err != nil {
		return err
	}
	_, err = c.w.WriteString("\r\n")
	if err != nil {
		return err
	}

	if len(c.fch) == 0 {
		c.fch <- struct{}{}
	}

	return nil
}

// Subscribe registers interest into a subject.
func (c *Client) Subscribe(subject string, cb MsgHandler) (*Subscription, error) {
	c.Lock()
	defer c.Unlock()
	c.sid += 1
	sid := c.sid

	sub := fmt.Sprintf("SUB %s %d\r\n", subject, sid)
	_, err := c.w.WriteString(sub)
	if err != nil {
		return nil, err
	}

	err = c.w.Flush()
	if err != nil {
		return nil, err
	}

	s := &Subscription{}
	s.pCond = sync.NewCond(&s.mu)
	c.subs[sid] = s

	// Wait for messages under a goroutine
	go func() {
		for {
			s.mu.Lock()
			if s.pHead == nil {
				s.pCond.Wait()
			}
			msg := s.pHead
			if msg != nil {
				s.pHead = msg.next
				if s.pHead == nil {
					s.pTail = nil
				}
			}
			s.mu.Unlock()

			if msg != nil {
				cb(msg.Subject, msg.Reply, msg.Data)
			}
		}
	}()

	return s, nil
}

// Request takes a context, a subject, and a payload
// and returns a response as a byte array or an error.
func (c *Client) Request(ctx context.Context, subj string, payload []byte) ([]byte, error) {
	c.Lock()
	// Set up request subscription if we haven't done so already.
	if c.respMap == nil {
		u := make([]byte, 11)
		io.ReadFull(rand.Reader, u)
		c.respSub = fmt.Sprintf("%s.%s.*", "_INBOX", hex.EncodeToString(u))
		c.respMap = make(map[string]chan []byte)
	}

	// Buffered channel  awaits a single response.
	dataCh := make(chan []byte, 1)
	u := make([]byte, 11)
	io.ReadFull(rand.Reader, u)
	token := hex.EncodeToString(u)
	c.respMap[token] = dataCh

	ginbox := c.respSub
	prefix := c.respSub[:29] // _INBOX. + unique prefix
	respInbox := fmt.Sprintf("%s.%s", prefix, token)
	createSub := c.respMux == nil
	c.Unlock()

	if createSub {
		var err error
		c.respSetup.Do(func() {
			fn := func(subj, reply string, data []byte) {
				// _INBOX. + unique prefix + . + token
				respToken := subj[30:]

				// Dequeue the first response only.
				c.Lock()
				mch := c.respMap[respToken]
				delete(c.respMap, respToken)
				c.Unlock()
				select {
				case mch <- data:
				default:
					return
				}
			}

			var sub *Subscription
			sub, err = c.Subscribe(ginbox, fn)
			c.Lock()
			c.respMux = sub
			c.Unlock()
		})
		if err != nil {
			return nil, err
		}
	}
	// ...
	// Publish the request along with the payload, then wait
	// for the reply to be sent to the client or for the
	// context to timeout:
	//
	// PUB subject reply-inbox #number-of-bytes\r\n
	// <payload>\r\n
	//
	msgh := []byte("PUB ")
	msgh = append(msgh, subj...)
	msgh = append(msgh, ' ')
	msgh = append(msgh, respInbox...)
	msgh = append(msgh, ' ')

	var b [12]byte
	var i = len(b)
	if len(payload) > 0 {
		for l := len(payload); l > 0; l /= 10 {
			i -= 1
			b[i] = digits[l%10]
		}
	} else {
		i -= 1
		b[i] = digits[0]
	}
	msgh = append(msgh, b[i:]...)
	msgh = append(msgh, _CRLF_...)

	c.Lock()
	_, err := c.w.Write(msgh)
	if err == nil {
		_, err = c.w.Write(payload)
	}
	if err == nil {
		_, err = c.w.WriteString(_CRLF_)
	}

	if err != nil {
		c.Unlock()
		return nil, err
	}

	// Signal a flush for the request if there are none pending (length is zero).
	if len(c.fch) == 0 {
		select {
		case c.fch <- struct{}{}:
		default:
		}
	}
	c.Unlock()

	// Wait for the response via the data channel or give up if context is done
	select {
	case data := <-dataCh:
		return data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return nil, nil
}

func (c *Client) processInfo(line string) {}

func (c *Client) processMsg(subj string, reply string, sid int, payload []byte) {
	c.Lock()
	sub, ok := c.subs[sid]
	c.Unlock()

	if ok {
		msg := &Msg{subj, reply, payload, nil}
		sub.mu.Lock()
		if sub.pHead == nil {
			sub.pHead = msg
			sub.pTail = msg
		} else {
			sub.pTail.next = msg
			sub.pTail = msg
		}
		sub.pCond.Signal()
		sub.mu.Unlock()
	}
}

func (c *Client) processPing() {
	c.Lock()
	defer c.Unlock()

	// Reply back to prevent stale connection error.
	c.w.WriteString("PONG\r\n")
	c.w.Flush()
}

func (c *Client) processPong()           {}
func (c *Client) processErr(msg string)  {}
func (c *Client) processOpErr(err error) {}
