package utils

import (
	"time"

	"github.com/valyala/fasthttp"
)

var _req *fasthttp.Request
var _resp *fasthttp.Response
var _client *fasthttp.Client

func InitHttpClient() {
	_req = fasthttp.AcquireRequest()
	_resp = fasthttp.AcquireResponse()
	_client = &fasthttp.Client{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
}

func CleanHttpClient() {
	fasthttp.ReleaseRequest(_req)
	fasthttp.ReleaseResponse(_resp)
}

func SetHttpContentType(contentType string) {
	_req.Header.SetContentType(contentType)
}

func SetHttpHeader(key string, value string) {
	_req.Header.Set(key, value)
}

func DelHttpHeader(key string) {
	_req.Header.Del(key)
}

func ResetHttpHeader() {
	_req.Header.Reset()
}

func HttpGet(url string) (int, string, error) {
	body := ""
	_req.SetRequestURI(url)
	_req.Header.SetMethod("GET")

	var err error
	if err = _client.Do(_req, _resp); err == nil {
		body = string(_resp.Body())
	}

	return _resp.StatusCode(), body, err
}

func HttpPost(url string, reqBody []byte) (int, string, error) {
	body := ""
	_req.SetRequestURI(url)
	_req.Header.SetMethod("POST")
	_req.SetBody(reqBody)

	var err error
	if err = _client.Do(_req, _resp); err == nil {
		body = string(_resp.Body())
	}

	return _resp.StatusCode(), body, err
}
