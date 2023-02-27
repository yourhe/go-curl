package curl

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"sync"
	"syscall"

	"golang.org/x/exp/slices"
)

type CC struct {
	req      *http.Request
	C        *CURL
	Callback chan error
}
type CurlTransport struct {
	tr        http.RoundTripper
	lock      sync.RWMutex
	m         *CURLM
	r         chan *CC
	c         chan *CURL
	ctx       context.Context
	closer    chan struct{}
	ctxcancel func()
}

func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

// RoundTrip executes a single HTTP transaction, returning
// a Response for the provided Request.
//
// RoundTrip should not attempt to interpret the response. In
// particular, RoundTrip must return err == nil if it obtained
// a response, regardless of the response's HTTP status code.
// A non-nil err should be reserved for failure to obtain a
// response. Similarly, RoundTrip should not attempt to
// handle higher-level protocol details such as redirects,
// authentication, or cookies.
//
// RoundTrip should not modify the request, except for
// consuming and closing the Request's Body. RoundTrip may
// read fields of the request in a separate goroutine. Callers
// should not mutate or reuse the request until the Response's
// Body has been closed.
//
// RoundTrip must always close the body, including on errors,
// but depending on the implementation may do so in a separate
// goroutine even after RoundTrip returns. This means that
// callers wanting to reuse the body for subsequent requests
// must arrange to wait for the Close call before doing so.
//
// The Request's URL and Header fields must be initialized.
func (c *CurlTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	easy := EasyInit()
	callback := make(chan error, 1)
	easy.Setopt(OPT_URL, req.URL.String())
	switch req.Method {
	case "GET":
	case "POST":
		easy.Setopt(OPT_POST, true)
	case "PUT":
		easy.Setopt(OPT_PUT, true)
	case "HEAD":
		easy.Setopt(OPT_NOBODY, true)
	default:
		easy.Setopt(OPT_CUSTOMREQUEST, req.Method)
	}
	headerbytes := bytes.NewBuffer(nil)
	req.Header.Write(headerbytes)
	var headers []string
	rl := bufio.NewReader(headerbytes)

	for {
		s, e := Readln(rl)
		if e != nil {
			break
		}
		headers = append(headers, s)
	}
	if len(headers) > 0 {
		easy.Setopt(OPT_HTTPHEADER, headers)
	}
	r, w := io.Pipe()
	revc := make(chan *http.Response, 1)
	headerBuf := bytes.NewBuffer(nil)
	easy.Setopt(OPT_READFUNCTION, func(ptr []byte, userdata interface{}) int {
		if req.Body == nil {
			return 0
		}
		n, _ := req.Body.Read(ptr)
		return n
	})
	f := false
	easy.Setopt(OPT_HTTP_CONTENT_DECODING, 0)
	easy.Setopt(OPT_WRITEFUNCTION, func(ptr []byte, userdata interface{}) bool {
		if !f {
			// fmt.Println(req, headerBuf.String())
			resp, err := http.ReadResponse(bufio.NewReader(headerBuf), req)
			if err != nil {
				fmt.Println("err", err)
				fmt.Println(headerBuf.String())
			}
			resp.Body = r
			revc <- resp
			f = true
		}
		_, err := io.Copy(w, bytes.NewReader(ptr))
		return err == nil
	})
	easy.Setopt(OPT_HEADERFUNCTION, func(ptr []byte, userdata interface{}) bool {
		var err error
		if len(ptr) > 6 && string(ptr[:7]) == "HTTP/2 " {
			_, err = io.Copy(headerBuf, io.MultiReader(bytes.NewBufferString("HTTP/2.0 "), bytes.NewBuffer(ptr[7:])))
		} else {
			_, err = io.Copy(headerBuf, bytes.NewReader(ptr))
		}
		return err == nil
	})
	easy.Setopt(OPT_SSL_VERIFYPEER, 0)
	easy.Setopt(OPT_CONNECTTIMEOUT, 5)
	easy.Setopt(OPT_VERBOSE, false)
	var err error
	go func() {
		select {
		case <-req.Context().Done():
			c.c <- easy
		case e := <-callback:
			err = e
			if !f {
				if e != nil {
					revc <- nil
				} else {
					resp, _ := http.ReadResponse(bufio.NewReader(headerBuf), req)
					if req.Method == "HEAD" {
						resp.Body = http.NoBody
					}
					revc <- resp
				}

			}
		}
		w.Close()

	}()
	c.r <- &CC{
		req:      req,
		C:        easy,
		Callback: callback,
	}
	return <-revc, err
}

func (c *CurlTransport) Stop() {
	c.ctxcancel()
}
func (c *CurlTransport) Start() {
	c.m = MultiInit()
	c.ctx, c.ctxcancel = context.WithCancel(context.Background())
	c.r = make(chan *CC, 100)
	c.c = make(chan *CURL, 100)

	handlers := []*CC{}
	runingHandlers := []*CC{}
	rlock := &sync.RWMutex{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			req := <-c.r
			rlock.Lock()
			handlers = append(handlers, req)
			rlock.Unlock()
		}
	}()
	go func() {
		var (
			rset, wset, eset            syscall.FdSet
			still_running, curl_timeout int = 0, 0
			err                         error
		)

		for {
			mh := c.m
			FD_ZERO(&rset)
			FD_ZERO(&wset)
			FD_ZERO(&eset)

			timeout := syscall.Timeval{Sec: 1, Usec: 0}
			curl_timeout, err = mh.Timeout()
			if err != nil {
				fmt.Printf("Error multi_timeout: %s\n", err)
			}
			if curl_timeout >= 0 {
				timeout.Sec = int64(curl_timeout / 1000)
				if timeout.Sec > 1 {
					timeout.Sec = 1
				} else {
					timeout.Usec = int64((curl_timeout % 1000)) * 1000
				}
			}

			max_fd, err := mh.Fdset(&rset, &wset, &eset)
			if err != nil {
				fmt.Printf("Error FDSET: %s\n", err)
			}
			// fmt.Println("maxfd", max_fd)
			_, err = syscall.Select(int(max_fd+1), &rset, &wset, &eset, &timeout)

			if err != nil {
				fmt.Printf("Error select: %s\n", err)
				continue
			}

			still_running, err = mh.Perform()
			msg, _ := mh.Info_read()
			for msg != nil {
				// if msg != nil {
				switch msg.Msg {
				case CURLMSG_DONE:
					code := binary.LittleEndian.Uint32(msg.Data[:])
					err := newCurlError(_Ctype_CURLcode(code))
					idx := slices.IndexFunc(runingHandlers, func(c *CC) bool {
						return c.C.handle == msg.Easy_handle.handle
					})

					if idx > -1 {
						mh.RemoveHandle(msg.Easy_handle)
						msg.Easy_handle.Cleanup()
						runingHandlers[idx].Callback <- err
						runingHandlers = slices.Delete(runingHandlers, idx, idx+1)
					}
				}
				// }
				// fmt.Println(msg, ri)
				msg, _ = mh.Info_read()
			}

			if still_running > 0 {
				fmt.Printf("Still running: %d\n", still_running)
				fmt.Printf("queue: %d\n", len(handlers))
				// } else {
				// 	break
			}
			rlock.Lock()
			if len(handlers) > 0 {
				for _, req := range handlers {
					runingHandlers = append(runingHandlers, req)
					mh.AddHandle(req.C)
				}
				handlers = []*CC{}
			}
			rlock.Unlock()
			select {
			case cc := <-c.c:
				idx := slices.IndexFunc(runingHandlers, func(c *CC) bool {
					return c.C.handle == cc.handle
				})
				if idx > -1 {
					mh.RemoveHandle(cc)
					cc.Cleanup()
					runingHandlers = slices.Delete(runingHandlers, idx, idx+1)
				}
			// case req := <-c.r:
			// 	handlers = append(handlers, req)
			// 	mh.AddHandle(req.C)
			case <-c.ctx.Done():
				for _, v := range runingHandlers {
					mh.RemoveHandle(v.C)
				}
				runingHandlers = nil
				mh.Cleanup()
				wg.Done()
				// close(c.r)
				// close(c.c)
				return
			default:
				// if still_running == 0 && len(handlers) == 0 {
				// 	select {
				// 	case req := <-c.r:
				// 		handlers = append(handlers, req)
				// 		mh.AddHandle(req.C)
				// 	case <-c.ctx.Done():
				// 		for _, v := range handlers {
				// 			mh.RemoveHandle(v.C)
				// 		}
				// 		handlers = nil
				// 		mh.Cleanup()
				// 		wg.Done()
				// 		return
				// 	}
				// }
			}
		}

	}()
	wg.Wait()
}
