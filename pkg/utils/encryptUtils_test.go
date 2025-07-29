package utils

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {

	as := ""
	fmt.Println("as:" + as)
	encryptAs := AesEncryptCBC(as, SecurityKey)
	fmt.Println("encryptAs:" + encryptAs)

	v := ""
	decryptAs := AesDecryptCBC(v, SecurityKey)
	fmt.Println("decryptAs:" + decryptAs)
}
