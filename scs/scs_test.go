package scs

import (
	"errors"
	"os"
	"testing"
)

func TestPut(t *testing.T) {
	auth, err := EnvAuth()
	if err != nil {
		t.Fatal(err)
	}

	scs := New(auth, "http://bjcnc.scs.sohucs.com")
	bucket := scs.Bucket("pptest")

	data := []byte("Hello, SCS!!")
	err = bucket.Put("sample111.txt", data, "text/plain")
	if err != nil {
		t.Errorf("error is %#v", err)
	}
}

func EnvAuth() (auth Auth, err error) {
	auth.AccessKey = os.Getenv("SOHUCS_ACCESS_KEY")

	auth.SecretKey = os.Getenv("SOHUCS_SECRET_KEY")

	if auth.AccessKey == "" {
		err = errors.New("SOHUCS_ACCESS_KEY not found in environment")
	}
	if auth.SecretKey == "" {
		err = errors.New("SOHUCS_SECRET_KEY not found in environment")
	}
	return
}
