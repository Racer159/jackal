# Test Initializing Jackal w/ An External Git Repository and A External Container Registry

> Note: For this test case, we deploy an 'external' Git server and container registry as pods running within the k8s cluster. These are still considered 'external' servers since they already existed inside the k8s cluster before `jackal init` command is executed

This directory holds the tests that verify Jackal can initialize a cluster to use an already existing Git server and container registry that is external to the resources Jackal manages. The tests in this directory are currently only run when manually executed.

## Running Tests Locally

### Dependencies

Running the tests locally have the same prerequisites as running and building Jackal:

1. GoLang >= `1.19.x`
2. Make
3. Access to a cluster to test against

### Actually Running The Test

Here are a few different ways to run the tests, based on your specific situation:

``` bash
# The default way, from the root directory of the repo. This will automatically build any Jackal related resources if they don't already exist (i.e. binary, init-package, example packages):
make test-external
```

``` bash
# If you are in the root folder of the repository and already have everything built (i.e., the binary, the init-package and the flux-test example package):
go test ./src/test/external/... -v
```
