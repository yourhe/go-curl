package main

import (
	"fmt"

	curl "github.com/yourhe/go-curl"
)

func main() {
	share := curl.ShareInit()
	defer share.Cleanup()
	share.Setopt(curl.SHOPT_SHARE, curl.LOCK_DATA_SSL_SESSION)
	share.Setopt(curl.SHOPT_SHARE, curl.LOCK_DATA_CONNECT)
	easy := curl.EasyInit()
	easy.Setopt(curl.OPT_SHARE, share)
	// defer easy.Cleanup()
	if easy != nil {
		easy.Setopt(curl.OPT_URL, "https://onlinelibrary.wiley.com?cookieSet=1")
		// skip_peer_verification
		easy.Setopt(curl.OPT_SSL_VERIFYPEER, false) // 0 is ok
		easy.Setopt(curl.OPT_HEADERFUNCTION, func(ptr []byte, userdata interface{}) bool {
			fmt.Println(string(ptr))
			return true
		}) // 0 is ok
		easy.Setopt(curl.OPT_VERBOSE, true)
		easy.Perform()
		easy.Cleanup()
	}
}
