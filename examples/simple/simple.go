package main

import (
	"fmt"

	curl "github.com/yourhe/go-curl"
)

func main() {
	f := curl.VersionInfo(9)
	fmt.Println(f)
	easy := curl.EasyInit()
	defer easy.Cleanup()
	if easy != nil {
		easy.Setopt(curl.OPT_URL, "http://www.baidu.com/")
		easy.Perform()
	}
}
