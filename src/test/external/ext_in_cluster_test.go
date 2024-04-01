// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package external provides a test for interacting with external resources
package external

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/Racer159/jackal/src/pkg/cluster"
	"github.com/Racer159/jackal/src/pkg/utils/exec"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var inClusterCredentialArgs = []string{
	"--git-push-username=git-user",
	"--git-push-password=superSecurePassword",
	"--git-url=http://gitea-http.git-server.svc.cluster.local:3000",
	"--registry-push-username=push-user",
	"--registry-push-password=superSecurePassword",
	"--registry-url=127.0.0.1:31999"}

type ExtInClusterTestSuite struct {
	suite.Suite
	*require.Assertions
}

func (suite *ExtInClusterTestSuite) SetupSuite() {
	suite.Assertions = require.New(suite.T())

	// Install a gitea chart to the k8s cluster to act as the 'remote' git server
	giteaChartURL := "https://dl.gitea.io/charts/gitea-8.3.0.tgz"
	helmInstallArgs := []string{"install", "gitea", giteaChartURL, "-f", "gitea-values.yaml", "-n=git-server", "--create-namespace"}
	err := exec.CmdWithPrint("helm", helmInstallArgs...)
	suite.NoError(err, "unable to install gitea chart")

	// Install docker-registry chart to the k8s cluster to act as the 'remote' container registry
	helmAddArgs := []string{"repo", "add", "twuni", "https://helm.twun.io"}
	err = exec.CmdWithPrint("helm", helmAddArgs...)
	suite.NoError(err, "unable to add the docker-registry chart repo")

	helmInstallArgs = []string{"install", "external-registry", "twuni/docker-registry", "-f=docker-registry-values.yaml", "-n=external-registry", "--create-namespace"}
	err = exec.CmdWithPrint("helm", helmInstallArgs...)
	suite.NoError(err, "unable to install the docker-registry chart")

	// Verify the registry and gitea helm charts installed successfully
	registryWaitCmd := []string{"wait", "deployment", "-n=external-registry", "external-registry-docker-registry", "--for", "condition=Available=True", "--timeout=5s"}
	registryErrStr := "unable to verify the docker-registry chart installed successfully"
	giteaWaitCmd := []string{"wait", "pod", "-n=git-server", "gitea-0", "--for", "condition=Ready=True", "--timeout=5s"}
	giteaErrStr := "unable to verify the gitea chart installed successfully"
	success := verifyKubectlWaitSuccess(suite.T(), 2, registryWaitCmd, registryErrStr)
	suite.True(success, registryErrStr)
	success = verifyKubectlWaitSuccess(suite.T(), 3, giteaWaitCmd, giteaErrStr)
	suite.True(success, giteaErrStr)
}

func (suite *ExtInClusterTestSuite) TearDownSuite() {
	// Uninstall gitea to clean things up
	helmUninstallArgs := []string{"uninstall", "gitea", "-n=git-server"}
	err := exec.CmdWithPrint("helm", helmUninstallArgs...)
	suite.NoError(err, "unable to uninstall gitea chart")

	// Uninstall registry to clean things up
	helmUninstallArgs = []string{"uninstall", "external-registry", "-n=external-registry"}
	err = exec.CmdWithPrint("helm", helmUninstallArgs...)
	suite.NoError(err, "unable to uninstall external-registry chart")
}

func (suite *ExtInClusterTestSuite) Test_0_Mirror() {
	// Use Jackal to mirror a package to the services (do this as test 0 so that the registry is unpolluted)
	mirrorArgs := []string{"package", "mirror-resources", "../../../build/jackal-package-argocd-amd64.tar.zst", "--confirm"}
	mirrorArgs = append(mirrorArgs, inClusterCredentialArgs...)
	err := exec.CmdWithPrint(jackalBinPath, mirrorArgs...)
	suite.NoError(err, "unable to mirror the package with jackal")

	c, err := cluster.NewCluster()
	suite.NoError(err)

	// Check that the registry contains the images we want
	tunnelReg, err := c.NewTunnel("external-registry", "svc", "external-registry-docker-registry", "", 0, 5000)
	suite.NoError(err)
	_, err = tunnelReg.Connect()
	suite.NoError(err)
	defer tunnelReg.Close()

	regCatalogURL := fmt.Sprintf("http://push-user:superSecurePassword@%s/v2/_catalog", tunnelReg.Endpoint())
	respReg, err := http.Get(regCatalogURL)
	suite.NoError(err)
	regBody, err := io.ReadAll(respReg.Body)
	suite.NoError(err)
	fmt.Println(string(regBody))
	suite.Equal(200, respReg.StatusCode)
	suite.Contains(string(regBody), "stefanprodan/podinfo", "registry did not contain the expected image")

	// Check that the git server contains the repos we want (TODO VERIFY NAME AND PORT)

	tunnelGit, err := c.NewTunnel("git-server", "svc", "gitea-http", "", 0, 3000)
	suite.NoError(err)
	_, err = tunnelGit.Connect()
	suite.NoError(err)
	defer tunnelGit.Close()

	gitRepoURL := fmt.Sprintf("http://git-user:superSecurePassword@%s/api/v1/repos/search", tunnelGit.Endpoint())
	respGit, err := http.Get(gitRepoURL)
	suite.NoError(err)
	gitBody, err := io.ReadAll(respGit.Body)
	fmt.Println(string(gitBody))
	suite.NoError(err)
	suite.Equal(200, respGit.StatusCode)
	suite.Contains(string(gitBody), "podinfo", "git server did not contain the expected repo")
}

func (suite *ExtInClusterTestSuite) Test_1_Deploy() {
	// Use Jackal to initialize the cluster
	initArgs := []string{"init", "--confirm"}
	initArgs = append(initArgs, inClusterCredentialArgs...)
	err := exec.CmdWithPrint(jackalBinPath, initArgs...)
	suite.NoError(err, "unable to initialize the k8s server with jackal")

	// Deploy the flux example package
	deployArgs := []string{"package", "deploy", "../../../build/jackal-package-podinfo-flux-amd64.tar.zst", "--confirm"}
	err = exec.CmdWithPrint(jackalBinPath, deployArgs...)
	suite.NoError(err, "unable to deploy flux example package")

	// Verify flux was able to pull from the 'external' repository
	podinfoWaitCmd := []string{"wait", "deployment", "-n=podinfo", "podinfo", "--for", "condition=Available=True", "--timeout=3s"}
	errorStr := "unable to verify flux deployed the podinfo example"
	success := verifyKubectlWaitSuccess(suite.T(), 2, podinfoWaitCmd, errorStr)
	suite.True(success, errorStr)

	_, _, err = exec.CmdWithContext(context.TODO(), exec.PrintCfg(), jackalBinPath, "destroy", "--confirm")
	suite.NoError(err, "unable to teardown jackal")
}

func TestExtInClusterTestSuite(t *testing.T) {
	suite.Run(t, new(ExtInClusterTestSuite))
}
