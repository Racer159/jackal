package test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindImages(t *testing.T) {
	t.Log("E2E: Find Images")

	t.Run("jackal prepare find-images", func(t *testing.T) {
		t.Parallel()
		// Test `jackal prepare find-images` for a remote asset
		stdOut, stdErr, err := e2e.Jackal("prepare", "find-images", "examples/helm-charts")
		require.NoError(t, err, stdOut, stdErr)
		require.Contains(t, stdOut, "ghcr.io/stefanprodan/podinfo:6.4.0", "The chart image should be found by Jackal")
		// Test `jackal prepare find-images` with a chart that uses helm annotations
		stdOut, stdErr, err = e2e.Jackal("prepare", "find-images", "src/test/packages/00-helm-annotations")
		require.NoError(t, err, stdOut, stdErr)
		require.Contains(t, stdOut, "registry1.dso.mil/ironbank/opensource/istio/pilot:1.17.2", "The pilot image should be found by Jackal")
	})

	t.Run("jackal prepare find-images --kube-version", func(t *testing.T) {
		t.Parallel()
		controllerImageWithTag := "quay.io/jetstack/cert-manager-controller:v1.11.1"
		controlImageWithSignature := "quay.io/jetstack/cert-manager-controller:sha256-4f1782c8316f34aae6b9ab823c3e6b7e6e4d92ec5dac21de6a17c3da44c364f1.sig"

		// Test `jackal prepare find-images` on a chart that has a `kubeVersion` declaration greater than the Helm default (v1.20.0)
		// This should pass because we build Jackal specifying the kubeVersion value from the kubernetes client-go library instead
		stdOut, stdErr, err := e2e.Jackal("prepare", "find-images", "src/test/packages/00-kube-version-override")
		require.NoError(t, err, stdOut, stdErr)
		require.Contains(t, stdOut, controllerImageWithTag, "The chart image should be found by Jackal")
		require.Contains(t, stdOut, controlImageWithSignature, "The image signature should be found by Jackal")

		// Test `jackal prepare find-images` with `--kube-version` specified and less than than the declared minimum (v1.21.0)
		stdOut, stdErr, err = e2e.Jackal("prepare", "find-images", "--kube-version=v1.20.0", "src/test/packages/00-kube-version-override")
		require.Error(t, err, stdOut, stdErr)
		require.Contains(t, stdErr, "Problem rendering the helm template for cert-manager", "The kubeVersion declaration should prevent this from templating")
		require.Contains(t, stdErr, "following charts had errors: [cert-manager]", "Jackal should print an ending error message")
	})

	t.Run("jackal dev find-images with helm or manifest vars", func(t *testing.T) {
		t.Parallel()
		registry := "coolregistry.gov"
		agentTag := "test"

		stdOut, _, err := e2e.Jackal("prepare", "find-images", ".", "--registry-url", registry, "--create-set", fmt.Sprintf("agent_image_tag=%s", agentTag))
		require.NoError(t, err)
		internalRegistryImage := fmt.Sprintf("%s/%s:%s", registry, "defenseunicorns/jackal/agent", agentTag)
		require.Contains(t, stdOut, internalRegistryImage, "registry image should be found with registry url")
		require.Contains(t, stdOut, "busybox:latest", "Busybox image should be found as long as helm chart doesn't error")

		path := filepath.Join("src", "test", "packages", "00-find-images-with-vars")
		stdOut, _, err = e2e.Jackal("prepare", "find-images", path, "--deploy-set", "BUSYBOX_IMAGE=busybox:earliest")
		require.NoError(t, err)
		require.Contains(t, stdOut, "nginx:latest", "Manifests aren't interpreting vars")
		require.Contains(t, stdOut, "busybox:earliest", "Values files aren't interpreting vars")
	})

	t.Run("jackal test find images --why w/ helm chart success", func(t *testing.T) {
		t.Parallel()
		testPackagePath := filepath.Join("examples", "wordpress")
		sets := []string{"WORDPRESS_USERNAME=jackal", "WORDPRESS_PASSWORD=fake", "WORDPRESS_EMAIL=hello@defenseunicorns.com", "WORDPRESS_FIRST_NAME=jackal", "WORDPRESS_LAST_NAME=jackal", "WORDPRESS_BLOG_NAME=blog"}
		deploysSet := strings.Join(sets, ",")
		stdOut, _, err := e2e.Jackal("dev", "find-images", testPackagePath, "--why", "docker.io/bitnami/apache-exporter:0.13.3-debian-11-r2", "--deploy-set", deploysSet)
		require.NoError(t, err)
		require.Contains(t, stdOut, "component: wordpress")
		require.Contains(t, stdOut, "chart: wordpress")
		require.Contains(t, stdOut, "image: docker.io/bitnami/wordpress:6.2.0-debian-11-r18")
	})

	t.Run("jackal test find images --why w/ manifests success", func(t *testing.T) {
		t.Parallel()
		testPackagePath := filepath.Join("examples", "manifests")

		stdOut, _, err := e2e.Jackal("dev", "find-images", testPackagePath, "--why", "httpd:alpine3.18")
		require.NoError(t, err)
		require.Contains(t, stdOut, "component: httpd-local")
		require.Contains(t, stdOut, "manifest: simple-httpd-deployment")
		require.Contains(t, stdOut, "image: httpd:alpine3.18")

	})

}
