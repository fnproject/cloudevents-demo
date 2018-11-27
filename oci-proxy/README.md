# OCI proxy application

## Purpose

Provide auth-less proxy to auth-protected Fn.


## Application configuration

This application has k8s templates to deploy:

 - [oci auth config map](kube/oci_cfg.yaml)
 - [oci proxy app deployment](kube/oci_proxy.yaml)
 - [oci proxy app deployment service](kube/oci_proxy_service.yaml)

