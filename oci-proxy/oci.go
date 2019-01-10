package main

import (
	"log"

	"github.com/fnproject/fn_go/provider"
	"github.com/fnproject/fn_go/provider/oracle"
	"github.com/oracle/oci-go-sdk/common"
)

func NewFromConfig(cfg common.ConfigurationProvider, compartmentID string, FnAPIURL string) (provider.Provider, error) {
	apiUrl, err := provider.CanonicalFnAPIUrl(FnAPIURL)
	if err != nil {
		return nil, err
	}
	log.Printf("canonical Fn URL: %v\n", apiUrl.String())

	keyID, err := cfg.KeyID()
	if err != nil {
		return nil, err
	}
	log.Printf("OCI user key ID: %v\n", keyID)

	key, err := cfg.PrivateRSAKey()
	if err != nil {
		return nil, err
	}
	log.Println("OCI user private key parsed")
	err = key.Validate()
	if err != nil {
		return nil, err
	}
	log.Println("OCI user private key is valid")

	return &oracle.Provider{
		FnApiUrl:      apiUrl,
		KeyId:         keyID,
		PrivateKey:    key,
		DisableCerts:  false,
		CompartmentID: compartmentID,
	}, nil
}

func setupOCI(compartmentID, tenancy, user, region, fingerprint, pKey, pKeyPass, invokeEndpoint string) (provider.Provider, error) {
	log.Println("setting up OCI provider")
	cfg := common.NewRawConfigurationProvider(
		tenancy, user, region, fingerprint, pKey, common.String(pKeyPass))
	log.Println("OCI config provisioned")
	apiEndpoint, err := getBaseEndpoint(invokeEndpoint)
	if err != nil {
		return nil, err
	}
	log.Printf("Fn endpoint: %v\n", apiEndpoint)
	return NewFromConfig(cfg, compartmentID, *apiEndpoint)
}
