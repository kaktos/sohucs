package scs

import (
	"errors"
	"os"
	"testing"
)

var (
	scs    *SCS
	bucket *Bucket
)

func setUp() {
	auth, err := envAuth()
	if err != nil {
		panic(err)
	}
	scs = New(auth, "http://bjcnc.scs.sohucs.com")
	bucket = scs.Bucket("pptest")
}

func TestPut(t *testing.T) {
	setUp()
	data := []byte("Hello, SCS!!")
	err := bucket.Put("sample111.txt", data, "text/plain")

	if err != nil {
		t.Errorf("error is %#v", err)
	}
}

func TestListBuckets(t *testing.T) {
	setUp()
	result, err := scs.Buckets()

	if err != nil {
		t.Errorf("error is %#v", err)
	} else {
		t.Logf("result is %#v\n", result)
	}
}

func TestGet(t *testing.T) {
	setUp()
	data, err := bucket.Get("sample111.txt")
	if err != nil {
		t.Errorf("error is %#v", err)
	} else {
		t.Logf("result is %#v\n", string(data))
	}
}

func TestDel(t *testing.T) {
	setUp()
	err := bucket.Del("hello01.txt")

	if err != nil {
		t.Errorf("error is %#v", err)
	}
}

func envAuth() (auth Auth, err error) {
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
