package curl

import (
	"fmt"
	"testing"
)

func TestVersionInfo(t *testing.T) {
	info := VersionInfo(VERSION_FIRST)

	fmt.Println(info)
	expectedProtocols := []string{"dict", "file", "ftp", "ftps", "gopher", "http", "https", "imap", "imaps", "ldap", "ldaps", "pop3", "pop3s", "rtmp", "rtsp", "smtp", "smtps", "telnet", "tftp", "scp", "sftp", "smb", "smbs"}
	protocols := info.Protocols
	for _, protocol := range protocols {
		found := false
		for _, expectedProtocol := range expectedProtocols {
			fmt.Println(expectedProtocol, protocol, expectedProtocol == protocol)
			if expectedProtocol == protocol {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("protocol should be in %v and is %s.", expectedProtocols, protocol)
		}
	}
}
