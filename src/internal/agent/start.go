// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package agent holds the mutating webhook server.
package agent

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Racer159/jackal/src/config/lang"
	agentHttp "github.com/Racer159/jackal/src/internal/agent/http"
	"github.com/Racer159/jackal/src/pkg/message"
)

// Heavily influenced by https://github.com/douglasmakey/admissioncontroller and
// https://github.com/slackhq/simple-kubernetes-webhook

// We can hard-code these because we control the entire thing anyway.
const (
	httpPort = "8443"
	tlsCert  = "/etc/certs/tls.crt"
	tlsKey   = "/etc/certs/tls.key"
)

// StartWebhook launches the Jackal agent mutating webhook in the cluster.
func StartWebhook() {
	message.Debug("agent.StartWebhook()")

	startServer(agentHttp.NewAdmissionServer(httpPort))
}

// StartHTTPProxy launches the jackal agent proxy in the cluster.
func StartHTTPProxy() {
	message.Debug("agent.StartHttpProxy()")

	startServer(agentHttp.NewProxyServer(httpPort))
}

func startServer(server *http.Server) {
	go func() {
		if err := server.ListenAndServeTLS(tlsCert, tlsKey); err != nil && err != http.ErrServerClosed {
			message.Fatal(err, lang.AgentErrStart)
		}
	}()

	message.Infof(lang.AgentInfoPort, httpPort)

	// listen shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	message.Infof(lang.AgentInfoShutdown)
	if err := server.Shutdown(context.Background()); err != nil {
		message.Fatal(err, lang.AgentErrShutdown)
	}
}
