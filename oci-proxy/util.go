package main

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/oracle/oci-go-sdk/common"
)

func LookUp(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("env: %v is not set", key)
	}
	log.Printf("%v=%v\n", key, v)
	return v
}

func getBaseEndpoint(invokeEndpoint string) (*string, error) {
	apiEndpointURL, err := url.Parse(invokeEndpoint)
	if err != nil {
		return nil, err
	}
	apiEndpoint := strings.Replace(
		apiEndpointURL.String(),
		apiEndpointURL.EscapedPath(), "", 1,
	)
	return common.String(apiEndpoint), nil
}
