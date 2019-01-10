package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/fnproject/fn_go/provider"
)

var (
	MaximumRequestBodySize int64 = 5 * 1024 * 1024
)

func Invoke(provider provider.Provider, invokeUrl string, content io.Reader, output io.Writer, headers http.Header) error {

	method := http.MethodPost

	// Read the request body (up to the maximum size), as this is used in the
	// authentication signature
	var req *http.Request
	if content != nil {
		b, err := ioutil.ReadAll(io.LimitReader(content, MaximumRequestBodySize))
		buffer := bytes.NewBuffer(b)
		req, err = http.NewRequest(method, invokeUrl, buffer)
		if err != nil {
			return fmt.Errorf("Error creating request to service: %s", err)
		}
	} else {
		var err error
		req, err = http.NewRequest(method, invokeUrl, nil)
		if err != nil {
			return fmt.Errorf("Error creating request to service: %s", err)
		}
	}

	req.Header = headers

	transport := provider.WrapCallTransport(http.DefaultTransport)
	httpClient := http.Client{Transport: transport}

	b, err := httputil.DumpRequestOut(req, content != nil)
	if err != nil {
		return err
	}
	fmt.Printf(string(b) + "\n")
	os.Stderr.Write(b)

	resp, err := httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("Error invoking fn: %s", err)
	}

	b, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	fmt.Printf(string(b) + "\n")
	os.Stderr.Write(b)

	io.Copy(output, resp.Body)

	return nil
}
