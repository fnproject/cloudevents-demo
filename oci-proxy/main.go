package main

import (
	"io"
	"log"
	"net/http"

	"github.com/fnproject/fn_go/provider"
)

func setupHandler(p provider.Provider, invokeURL string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		err := Invoke(p, invokeURL, r.Body, w, r.Header)
		if err != nil {
			w.WriteHeader(500)
			io.WriteString(w, err.Error())
		}
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	user := LookUp("OCI_USER")
	fingerprint := LookUp("OCI_FINGERPRINT")
	pKey := LookUp("OCI_KEY")
	pKeyPass := LookUp("OCI_KEY_PASS")
	tenancy := LookUp("OCI_TENANCY")
	region := LookUp("OCI_REGION")
	compartmentID := LookUp("OCI_COMPARTMENT")
	invokeEndpoint := LookUp("FN_INVOKE_ENDPOINT")

	log.Println("env is set properly")
	p, err := setupOCI(
		compartmentID,
		tenancy,
		user,
		region,
		fingerprint,
		pKey,
		pKeyPass,
		invokeEndpoint,
	)
	if err != nil {
		log.Fatalf(err.Error())
	}

	http.HandleFunc("/word-generator", setupHandler(p, invokeEndpoint))
	http.HandleFunc("/ping", ping)
	log.Fatal(http.ListenAndServe(":9999", nil))
}
