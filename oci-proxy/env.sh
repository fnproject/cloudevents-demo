#!/usr/bin/env bash

export OCI_USER=`cat ~/.oci/config | grep user | awk '{split($0,array,"=")} END{print array[2]}'`
export OCI_TENANCY=`cat ~/.oci/config | grep tenancy | awk '{split($0,array,"=")} END{print array[2]}'`
export OCI_REGION=`cat ~/.oci/config | grep region | awk '{split($0,array,"=")} END{print array[2]}'`
export OCI_FINGERPRINT=`cat ~/.oci/config | grep fingerprint | awk '{split($0,array,"=")} END{print array[2]}'`
export OCI_KEY=`cat $(cat ~/.oci/config | grep key_file | awk '{split($0,array,"=")} END{print array[2]}')`
export OCI_KEY_PASS=${1}
export OCI_COMPARTMENT=${2}
export FN_INVOKE_ENDPOINT=${3}
