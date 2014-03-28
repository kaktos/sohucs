package scs

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"strings"
)

func sign(auth Auth, method string, canonicalPath string, headers map[string][]string) {
	var md5, ctype, date string
	method = strings.ToLower(method)
	for k, v := range headers {
		k = strings.ToLower(k)
		switch k {
		case "content-md5":
			md5 = v[0]
		case "content-type":
			ctype = v[0]
		case "date":
			date = v[0]
		}
	}

	payload := method + "\n" + md5 + "\n" + ctype + "\n" + date + "\n" + canonicalPath
	hmac := hmac.New(sha1.New, []byte(auth.SecretKey))
	hmac.Write([]byte(payload))
	signature := make([]byte, base64.StdEncoding.EncodedLen(hmac.Size()))
	base64.StdEncoding.Encode(signature, hmac.Sum(nil))

	headers["Authorization"] = []string{"AWS " + auth.AccessKey + ":" + string(signature)}
}
