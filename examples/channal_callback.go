package main

import (
	"time"

	curl "github.com/yourhe/go-curl"
)

func write_data(ptr []byte, userdata interface{}) bool {
	ch, ok := userdata.(chan string)
	if ok {
		ch <- string(ptr)
		return true // ok
	} else {
		println("ERROR!")
		return false
	}
	return false
}

func main() {
	curl.GlobalInit(curl.GLOBAL_ALL)

	// init the curl session
	easy := curl.EasyInit()
	defer easy.Cleanup()

	easy.Setopt(curl.OPT_URL, "https://onlinelibrary.wiley.com/?cookieSet=1")
	// easy.Setopt(curl.OPT_COOKIEJAR, "./cookie.jar")
	easy.Setopt(curl.OPT_COOKIEFILE, "./cookie.jar")
	easy.Setopt(curl.OPT_HTTPHEADER, []string{"Accept:text/html"})
	easy.Setopt(curl.OPT_HTTPHEADER, []string{"Referer:http://www.google.com"})
	easy.Setopt(curl.OPT_HTTPHEADER, []string{"Cookie:bcd=asdf"})
	// easy.Setopt(curl.OPT_WRITEFUNCTION, write_data)
	easy.Setopt(curl.OPT_VERBOSE, true)
	// make a chan
	ch := make(chan string, 100)
	go func(ch chan string) {
		for {
			data := <-ch
			println("Got data size=", len(data))
			println("Got data=", data)
		}
	}(ch)

	// easy.Setopt(curl.OPT_WRITEDATA, ch)

	if err := easy.Perform(); err != nil {
		println("ERROR: ", err.Error())
	}

	time.Sleep(10000) // wait gorotine
}
