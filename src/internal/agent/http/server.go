// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package http provides a http server for the webhook and proxy.
package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/defenseunicorns/zarf/src/internal/agent/hooks"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewAdmissionServer creates a http.Server for the mutating webhook admission handler.
func NewAdmissionServer(port string) *http.Server {
	message.Debugf("http.NewServer(%s)", port)

	// Instances hooks
	podsMutation := hooks.NewPodMutationHook()
	fluxGitRepositoryMutation := hooks.NewGitRepositoryMutationHook()
	fluxHelmRepositoryMutation := hooks.NewHelmRepositoryMutationHook()
	fluxOCIRepositoryMutation := hooks.NewOCIRepositoryMutationHook()
	argocdApplicationMutation := hooks.NewApplicationMutationHook()
	argocdRepositoryMutation := hooks.NewRepositoryMutationHook()

	// Routers
	admissonHandler := newAdmissionHandler()
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthz())
	mux.Handle("/mutate/pod", admissonHandler.Serve(podsMutation))
	mux.Handle("/mutate/flux-gitrepository", admissonHandler.Serve(fluxGitRepositoryMutation))
	mux.Handle("/mutate/flux-helmrepository", admissonHandler.Serve(fluxHelmRepositoryMutation))
	mux.Handle("/mutate/flux-ocirepository", admissonHandler.Serve(fluxOCIRepositoryMutation))
	mux.Handle("/mutate/argocd-application", admissonHandler.Serve(argocdApplicationMutation))
	mux.Handle("/mutate/argocd-repository", admissonHandler.Serve(argocdRepositoryMutation))
	mux.Handle("/metrics", promhttp.Handler())

	return &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // Set ReadHeaderTimeout to avoid Slowloris attacks
	}
}

// NewProxyServer creates and returns an http proxy server.
func NewProxyServer(port string) *http.Server {
	message.Debugf("http.NewHTTPProxy(%s)", port)

	mux := http.NewServeMux()
	mux.Handle("/healthz", healthz())
	mux.Handle("/", ProxyHandler())
	mux.Handle("/metrics", promhttp.Handler())

	return &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // Set ReadHeaderTimeout to avoid Slowloris attacks
	}
}

func healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}
