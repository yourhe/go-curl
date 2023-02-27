package curl

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestCurlTransport_Start(t *testing.T) {
	c := &CurlTransport{}

	go func() {
		defer c.Stop()
		ctx, _ := context.WithCancel(context.Background())
		// req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.google.com.hk/pagead/1p-user-list/973395369/?random=1672189631155&cv=11&fst=1672189200000&bg=ffffff&guid=ON&async=1&gtm=2oabu0&u_w=1920&u_h=1080&frm=0&url=https%3A%2F%2Fchemistry-europe.onlinelibrary.wiley.com%2Fdoi%2Fepdf%2F10.1002%2Fchem.201903795&ref=https%3A%2F%2Fchemistry-europe.onlinelibrary.wiley.com%2Fdoi%2F10.1002%2Fchem.201903795&tiba=Isolation%20of%20Carbene%E2%80%90Stabilized%20Arsenic%20Monophosphide%20%5BAsP%5D%20and%20its%20Radical%20Cation%20%5BAsP%5D%2B.%20and%20Dication%20%5BAsP%5D2%2B%20-%20Doddi%20-%202019%20-%20Chemistry%20%26%238211%3B%20A%20European%20Journal%20-%20Wiley%20&data=event%3Dgtag.config&fmt=3&is_vtc=1&random=4145956300&rmt_tld=1&ipr=y", nil)
		req, _ := http.NewRequestWithContext(ctx, "HEAD", "https://test.dr2am.cn", nil)

		resp, err := c.RoundTrip(req)
		if err != nil {
			fmt.Println(err, resp)
			return
		}
		fs, err := ioutil.ReadAll(resp.Body)

		fmt.Println(resp.Body, len(fs), err)
		resp.Body.Close()
		// time.Sleep(1 * time.Second)
		// {
		// 	req, _ := http.NewRequest("GET", "https://www.baidu.com", nil)
		// 	c.RoundTrip(req)
		// }
		time.Sleep(10 * time.Second)
	}()
	c.Start()

}

// CURL_IMPERSONATE=chrome101 go test -exec "env DYLD_LIBRARY_PATH=/Users/yorhe/Desktop/work/dev/project/2022/dr2am/go-curl/go-curl/lib/curl-impersonate-chrome" -timeout 60s -run ^TestCurlTransport_Start$ github.com/yourhe/go-curl -v

func TestHead(t *testing.T) {
	// req, _ := http.NewRequest( "HEAD", , nil)
	resp, err := http.Head("https://www.baidu.com")
	fmt.Println(resp, err)
	fmt.Println(ioutil.ReadAll(resp.Body))
	resp.Body.Close()
}

// client hellow info &{[27242 4865 4866 4867 49195 49199 49196 49200 52393 52392 49171 49172 156 157 47 53] test.dr2am.cn [CurveID(64250) X25519 CurveP256 CurveP384] [0] [ECDSAWithP256AndSHA256 PSSWithSHA256 PKCS1WithSHA256 ECDSAWithP384AndSHA384 PSSWithSHA384 PKCS1WithSHA384 PSSWithSHA512 PKCS1WithSHA512] [h2 http/1.1] [39578 772 771] 0xc0006a2d68 0xc029300300 0xc026f85c40}

// client hellow info &{[2570 4865 4866 4867 49195 49199 49196 49200 52393 52392 49171 49172 156 157 47 53] test.dr2am.cn [CurveID(39578) X25519 CurveP256 CurveP384] [0] [ECDSAWithP256AndSHA256 PSSWithSHA256 PKCS1WithSHA256 ECDSAWithP384AndSHA384 PSSWithSHA384 PKCS1WithSHA384 PSSWithSHA512 PKCS1WithSHA512] [h2 http/1.1] [39578 772 771] 0xc026db20e0 0xc029300300 0xc0010cad00}
