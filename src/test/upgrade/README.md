# Test Upgrading a Jackal package and Jackal itself

> Note: For this test case, we first deploy the podinfo package (from this directory) with version 6.3.3, then build a 6.3.4 package.  This package is then deployed to an upgraded cluster with the new jackal version and the new jackal version builds and deploys a 6.3.5 version.

This directory holds the tests that verify Jackal can perform these upgrade actions and that any deploy deprecations work as expected.

## Running Tests Locally

### Dependencies

Running the tests locally have the same prerequisites as running and building Jackal:

1. GoLang >= `1.19.x`
1. Make
1. Access to a cluster to test against
1. The jackal.yaml created and deployed with PODINFO_VERSION 6.3.3
1. The jackal.yaml created with PODINFO_VERSION 6.3.4

### Actually Running The Test

Here are a few different ways to run the tests, based on your specific situation:

``` bash
# The default way, from the root directory of the repo. This will automatically build any Jackal related resources if they don't already exist (i.e. binary, init-package, example packages):
make test-upgrade
```

``` bash
# If you are in the root folder of the repository and already have everything built (i.e., the binary, the init-package and the flux-test example package):
go test ./src/test/upgrade/...
```
