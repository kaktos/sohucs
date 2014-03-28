package scs

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	debug = false
)

type Auth struct {
	AccessKey string
	SecretKey string
}

type SCS struct {
	Auth
	region string
}

type Bucket struct {
	*SCS
	Name string
}

type request struct {
	method  string
	bucket  string
	path    string
	baseurl string
	headers http.Header
	payload io.Reader
}

//for list all exist buckts
type BucketInfo struct {
	Name, CreationDate, Region string
}
type Owner struct {
	ID, DisplayName string
}
type ListAllBucketsResult struct {
	Owner   Owner
	Buckets []BucketInfo `xml:"Buckets>Bucket"`
}

//get new SCS instance
func New(auth Auth, region string) *SCS {
	return &SCS{auth, region}
}

func (s *SCS) Bucket(name string) *Bucket {
	return &Bucket{s, name}
}

func (s *SCS) Buckets() (*ListAllBucketsResult, error) {
	req := &request{
		method:  "GET",
		path:    "/",
		headers: make(map[string][]string),
	}

	result := &ListAllBucketsResult{}
	err := s.query(req, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Put Object into SCS Bucket
func (b *Bucket) Put(path string, data []byte, contentType string) error {
	headers := map[string][]string{
		"Content-Length": {strconv.FormatInt(int64(len(data)), 10)},
		"Content-Type":   {contentType},
	}
	req := &request{
		method:  "PUT",
		bucket:  b.Name,
		path:    path,
		headers: headers,
		payload: bytes.NewBuffer(data),
	}

	return b.SCS.query(req, nil)
}

// Get retrieves an object from an SCS bucket.
func (b *Bucket) Get(path string) (data []byte, err error) {
	body, err := b.GetReader(path)
	if err != nil {
		return nil, err
	}
	data, err = ioutil.ReadAll(body)
	body.Close()
	return data, err
}

// It is the caller's responsibility to call Close
func (b *Bucket) GetReader(path string) (rc io.ReadCloser, err error) {
	resp, err := b.GetResponse(path)
	if resp != nil {
		return resp.Body, err
	}
	return nil, err
}

// GetResponse retrieves an object from an SCS bucket
func (b *Bucket) GetResponse(path string) (*http.Response, error) {
	req := &request{
		method:  "GET",
		bucket:  b.Name,
		path:    path,
		headers: make(map[string][]string),
	}
	err := b.SCS.prepare(req)
	if err != nil {
		return nil, err
	}

	resp, err := b.SCS.run(req, nil)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (req *request) url() (*url.URL, error) {
	u, err := url.Parse(req.baseurl)
	if err != nil {
		return nil, fmt.Errorf("bad SCS endpoint URL %q: %v", req.baseurl, err)
	}

	u.Path = req.path
	return u, nil
}

//delete object in bucket
func (b *Bucket) Del(path string) error {
	req := &request{
		method:  "DELETE",
		bucket:  b.Name,
		path:    path,
		headers: make(map[string][]string),
	}
	return b.SCS.query(req, nil)
}

//make the request and get result back
func (s *SCS) query(req *request, resp interface{}) error {
	err := s.prepare(req)
	if err == nil {
		var httpResponse *http.Response
		httpResponse, err = s.run(req, resp)
		if resp == nil && httpResponse != nil {
			httpResponse.Body.Close()
		}
	}
	return err

}

func (s *SCS) prepare(req *request) error {
	if !strings.HasPrefix(req.path, "/") {
		req.path = "/" + req.path
	}
	if req.baseurl == "" {
		req.baseurl = s.region
		if req.bucket != "" {
			req.path = "/" + req.bucket + req.path
		}
	}
	u, err := req.url()
	if err != nil {
		return fmt.Errorf("bad SCS endpoint URL %q: %v", req.baseurl, err)
	}
	req.headers["Host"] = []string{u.Host}
	req.headers["Date"] = []string{time.Now().Format(time.ANSIC)}
	sign(s.Auth, req.method, req.path, req.headers)
	return nil
}

func (s *SCS) run(req *request, resp interface{}) (*http.Response, error) {
	u, err := req.url()
	if err != nil {
		return nil, err
	}
	hreq := http.Request{
		URL:    u,
		Method: req.method,
		Close:  true,
		Header: req.headers,
	}
	if v, ok := req.headers["Content-Length"]; ok {
		hreq.ContentLength, _ = strconv.ParseInt(v[0], 10, 64)
		delete(req.headers, "Content-Length")
	}
	if req.payload != nil {
		hreq.Body = ioutil.NopCloser(req.payload)
	}

	hresp, err := http.DefaultClient.Do(&hreq)
	if err != nil {
		return nil, err
	}

	if debug {
		dump, _ := httputil.DumpResponse(hresp, true)
		log.Printf("} -> %s\n", dump)
	}

	if hresp.StatusCode != 200 && hresp.StatusCode != 204 {
		return nil, buildError(hresp)
	}

	if resp != nil {
		err = xml.NewDecoder(hresp.Body).Decode(resp)
		hresp.Body.Close()
	}
	return hresp, err
}

// Error represents an error in an operation with SCS.
type Error struct {
	StatusCode int
	Code       string
	Message    string
	Resource   string
	RequestId  string
}

func (e *Error) Error() string {
	return e.Message
}

func buildError(r *http.Response) error {
	err := Error{}
	xml.NewDecoder(r.Body).Decode(&err)
	r.Body.Close()
	err.StatusCode = r.StatusCode
	if err.Message == "" {
		err.Message = r.Status
	}
	//log.Printf("error built: %#v\n", err)
	return &err
}
