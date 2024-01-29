// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package external provides a test for interacting with external resources
package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/pkg/utils/exec"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"helm.sh/helm/v3/pkg/repo"
)

// Docker/k3d networking constants
const (
	network        = "k3d-k3s-external-test"
	subnet         = "172.31.0.0/16"
	gateway        = "172.31.0.1"
	giteaIP        = "172.31.0.99"
	giteaHost      = "gitea.localhost"
	registryHost   = "registry.localhost"
	clusterName    = "zarf-external-test"
	giteaUser      = "git-user"
	registryUser   = "push-user"
	commonPassword = "superSecurePassword"
)

var outClusterCredentialArgs = []string{
	"--git-push-username=" + giteaUser,
	"--git-push-password=" + commonPassword,
	"--git-url=http://" + giteaHost + ":3000",
	"--registry-push-username=" + registryUser,
	"--registry-push-password=" + commonPassword,
	"--registry-url=k3d-" + registryHost + ":5000"}

type ExtOutClusterTestSuite struct {
	suite.Suite
	*require.Assertions
}

func (suite *ExtOutClusterTestSuite) SetupSuite() {

	suite.Assertions = require.New(suite.T())

	// Teardown any leftovers from previous tests
	_ = exec.CmdWithPrint("k3d", "cluster", "delete", clusterName)
	_ = exec.CmdWithPrint("k3d", "registry", "delete", registryHost)
	_ = exec.CmdWithPrint("docker", "network", "remove", network)

	// Setup a network for everything to live inside
	err := exec.CmdWithPrint("docker", "network", "create", "--driver=bridge", "--subnet="+subnet, "--gateway="+gateway, network)
	suite.NoError(err, "unable to create the k3d registry")

	// Install a k3d-managed registry server to act as the 'remote' container registry
	err = exec.CmdWithPrint("k3d", "registry", "create", registryHost, "--port", "5000")
	suite.NoError(err, "unable to create the k3d registry")

	// Create a k3d cluster with the proper networking and aliases
	err = exec.CmdWithPrint("k3d", "cluster", "create", clusterName, "--registry-use", "k3d-"+registryHost+":5000", "--host-alias", giteaIP+":"+giteaHost, "--network", network)
	suite.NoError(err, "unable to create the k3d cluster")

	// Install a gitea server via docker compose to act as the 'remote' git server
	err = exec.CmdWithPrint("docker", "compose", "up", "-d")
	suite.NoError(err, "unable to install the gitea-server")

	// Wait for gitea to deploy properly
	giteaArgs := []string{"inspect", "-f", "{{.State.Status}}", "gitea.init"}
	giteaErrStr := "unable to verify the gitea container installed successfully"
	success := verifyWaitSuccess(suite.T(), 2, "docker", giteaArgs, "exited", giteaErrStr)
	suite.True(success, giteaErrStr)

	// Connect gitea to the k3d network
	err = exec.CmdWithPrint("docker", "network", "connect", "--ip", giteaIP, network, giteaHost)
	suite.NoError(err, "unable to connect the gitea-server top k3d")
}

func (suite *ExtOutClusterTestSuite) TearDownSuite() {
	// Tear down all of that stuff we made for local runs
	err := exec.CmdWithPrint("k3d", "cluster", "delete", clusterName)
	suite.NoError(err, "unable to teardown cluster")

	err = exec.CmdWithPrint("docker", "compose", "down")
	suite.NoError(err, "unable to teardown the gitea-server")

	err = exec.CmdWithPrint("k3d", "registry", "delete", registryHost)
	suite.NoError(err, "unable to teardown the k3d registry")

	err = exec.CmdWithPrint("docker", "network", "remove", network)
	suite.NoError(err, "unable to teardown the docker test network")
}

func (suite *ExtOutClusterTestSuite) Test_0_Mirror() {
	// Use Zarf to mirror a package to the services (do this as test 0 so that the registry is unpolluted)
	mirrorArgs := []string{"package", "mirror-resources", "../../../build/zarf-package-argocd-amd64.tar.zst", "--confirm"}
	mirrorArgs = append(mirrorArgs, outClusterCredentialArgs...)
	err := exec.CmdWithPrint(zarfBinPath, mirrorArgs...)
	suite.NoError(err, "unable to mirror the package with zarf")

	// Check that the registry contains the images we want
	regCatalogURL := fmt.Sprintf("http://%s:%s@k3d-%s:5000/v2/_catalog", registryUser, commonPassword, registryHost)
	respReg, err := http.Get(regCatalogURL)
	suite.NoError(err)
	regBody, err := io.ReadAll(respReg.Body)
	suite.NoError(err)
	fmt.Println(string(regBody))
	suite.Equal(200, respReg.StatusCode)
	suite.Contains(string(regBody), "stefanprodan/podinfo", "registry did not contain the expected image")

	// Check that the git server contains the repos we want
	gitRepoURL := fmt.Sprintf("http://%s:%s@%s:3000/api/v1/repos/search", giteaUser, commonPassword, giteaHost)
	respGit, err := http.Get(gitRepoURL)
	suite.NoError(err)
	gitBody, err := io.ReadAll(respGit.Body)
	suite.NoError(err)
	fmt.Println(string(gitBody))
	suite.Equal(200, respGit.StatusCode)
	suite.Contains(string(gitBody), "podinfo", "git server did not contain the expected repo")
}

func (suite *ExtOutClusterTestSuite) Test_1_Deploy() {
	// Use Zarf to initialize the cluster
	initArgs := []string{"init", "--confirm"}
	initArgs = append(initArgs, outClusterCredentialArgs...)
	err := exec.CmdWithPrint(zarfBinPath, initArgs...)
	suite.NoError(err, "unable to initialize the k8s server with zarf")

	// Deploy the flux example package
	deployArgs := []string{"package", "deploy", "../../../build/zarf-package-podinfo-flux-amd64.tar.zst", "--confirm"}
	err = exec.CmdWithPrint(zarfBinPath, deployArgs...)
	suite.NoError(err, "unable to deploy flux example package")

	// Verify flux was able to pull from the 'external' repository
	podinfoArgs := []string{"wait", "deployment", "-n=podinfo", "podinfo", "--for", "condition=Available=True", "--timeout=3s"}
	errorStr := "unable to verify flux deployed the podinfo example"
	success := verifyKubectlWaitSuccess(suite.T(), 2, podinfoArgs, errorStr)
	suite.True(success, errorStr)
}

func (suite *ExtOutClusterTestSuite) Test_2_AuthToPrivateHelmChart() {
	baseURL := fmt.Sprintf("http://%s:3000", giteaHost)

	suite.createHelmChartInGitea(baseURL, giteaUser, commonPassword)
	suite.makeGiteaUserPrivate(baseURL, giteaUser, commonPassword)

	tempDir := suite.T().TempDir()
	repoPath := filepath.Join(tempDir, "repositories.yaml")
	os.Setenv("HELM_REPOSITORY_CONFIG", repoPath)
	defer os.Unsetenv("HELM_REPOSITORY_CONFIG")

	packagePath := filepath.Join("..", "packages", "external-helm-auth")
	findImageArgs := []string{"dev", "find-images", packagePath}
	err := exec.CmdWithPrint(zarfBinPath, findImageArgs...)
	suite.Error(err, "Since auth has not been setup, this should fail")

	repoFile := repo.NewFile()

	chartURL := fmt.Sprintf("%s/api/packages/%s/helm", baseURL, giteaUser)
	entry := &repo.Entry{
		Name:     "temp_entry",
		Username: giteaUser,
		Password: commonPassword,
		URL:      chartURL,
	}
	repoFile.Add(entry)
	utils.WriteYaml(repoPath, repoFile, 0600)

	err = exec.CmdWithPrint(zarfBinPath, findImageArgs...)
	suite.NoError(err, "Unable to find images, helm auth likely failed")

	packageCreateArgs := []string{"package", "create", packagePath, fmt.Sprintf("--output=%s", tempDir), "--confirm"}
	err = exec.CmdWithPrint(zarfBinPath, packageCreateArgs...)
	suite.NoError(err, "Unable to create package, helm auth likely failed")
}

func (suite *ExtOutClusterTestSuite) createHelmChartInGitea(baseURL string, username string, password string) {
	tempDir := suite.T().TempDir()
	podInfoVersion := "6.4.0"
	podinfoChartPath := filepath.Join("..", "..", "..", "examples", "helm-charts", "chart")
	err := exec.CmdWithPrint("helm", "package", podinfoChartPath, "--destination", tempDir)
	podinfoTarballPath := filepath.Join(tempDir, fmt.Sprintf("podinfo-%s.tgz", podInfoVersion))
	suite.NoError(err, "Unable to package chart")

	utils.DownloadToFile(fmt.Sprintf("https://stefanprodan.github.io/podinfo/podinfo-%s.tgz", podInfoVersion), podinfoTarballPath, "")
	url := fmt.Sprintf("%s/api/packages/%s/helm/api/charts", baseURL, username)

	file, err := os.Open(podinfoTarballPath)
	suite.NoError(err)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", podinfoTarballPath)
	suite.NoError(err)
	_, err = io.Copy(part, file)
	suite.NoError(err)
	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	suite.NoError(err)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth(username, password)

	client := &http.Client{}

	resp, err := client.Do(req)
	suite.NoError(err)
	resp.Body.Close()
}

func (suite *ExtOutClusterTestSuite) makeGiteaUserPrivate(baseURL string, username string, password string) {
	url := fmt.Sprintf("%s/api/v1/admin/users/%s", baseURL, username)

	userOption := map[string]interface{}{
		"visibility": "private",
		"login_name": username,
	}

	jsonData, err := json.Marshal(userOption)
	suite.NoError(err)

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(jsonData))
	suite.NoError(err)

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	client := &http.Client{}
	resp, err := client.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	suite.NoError(err)
}

func TestExtOurClusterTestSuite(t *testing.T) {
	suite.Run(t, new(ExtOutClusterTestSuite))
}
