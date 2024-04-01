// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package http provides a http server for the webhook and proxy.
package http

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Racer159/jackal/src/config/lang"
	"github.com/Racer159/jackal/src/internal/agent/state"
	"github.com/Racer159/jackal/src/pkg/message"
	"github.com/Racer159/jackal/src/pkg/transform"
)

// ProxyHandler constructs a new httputil.ReverseProxy and returns an http handler.
func ProxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := proxyRequestTransform(r)
		if err != nil {
			message.Debugf("%#v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(lang.AgentErrUnableTransform))
			return
		}

		proxy := &httputil.ReverseProxy{Director: func(_ *http.Request) {}, ModifyResponse: proxyResponseTransform}
		proxy.ServeHTTP(w, r)
	}
}

func proxyRequestTransform(r *http.Request) error {
	message.Debugf("Before Req %#v", r)
	message.Debugf("Before Req URL %#v", r.URL)

	// We add this so that we can use it to rewrite urls in the response if needed
	r.Header.Add("X-Forwarded-Host", r.Host)

	// We remove this so that go will encode and decode on our behalf (see https://pkg.go.dev/net/http#Transport DisableCompression)
	r.Header.Del("Accept-Encoding")

	jackalState, err := state.GetJackalStateFromAgentPod()
	if err != nil {
		return err
	}

	var targetURL *url.URL

	// Setup authentication for each type of service based on User Agent
	switch {
	case isGitUserAgent(r.UserAgent()):
		r.SetBasicAuth(jackalState.GitServer.PushUsername, jackalState.GitServer.PushPassword)
	case isNpmUserAgent(r.UserAgent()):
		r.Header.Set("Authorization", "Bearer "+jackalState.ArtifactServer.PushToken)
	default:
		r.SetBasicAuth(jackalState.ArtifactServer.PushUsername, jackalState.ArtifactServer.PushToken)
	}

	// Transform the URL; if we see the NoTransform prefix, strip it; otherwise, transform the URL based on User Agent
	if strings.HasPrefix(r.URL.Path, transform.NoTransform) {
		switch {
		case isGitUserAgent(r.UserAgent()):
			targetURL, err = transform.NoTransformTarget(jackalState.GitServer.Address, r.URL.Path)
		default:
			targetURL, err = transform.NoTransformTarget(jackalState.ArtifactServer.Address, r.URL.Path)
		}
	} else {
		switch {
		case isGitUserAgent(r.UserAgent()):
			targetURL, err = transform.GitURL(jackalState.GitServer.Address, getTLSScheme(r.TLS)+r.Host+r.URL.String(), jackalState.GitServer.PushUsername)
		case isPipUserAgent(r.UserAgent()):
			targetURL, err = transform.PipTransformURL(jackalState.ArtifactServer.Address, getTLSScheme(r.TLS)+r.Host+r.URL.String())
		case isNpmUserAgent(r.UserAgent()):
			targetURL, err = transform.NpmTransformURL(jackalState.ArtifactServer.Address, getTLSScheme(r.TLS)+r.Host+r.URL.String())
		default:
			targetURL, err = transform.GenTransformURL(jackalState.ArtifactServer.Address, getTLSScheme(r.TLS)+r.Host+r.URL.String())
		}
	}

	if err != nil {
		return err
	}

	r.Host = targetURL.Host
	r.URL = targetURL
	r.RequestURI = getRequestURI(targetURL.Path, targetURL.RawQuery, targetURL.Fragment)

	message.Debugf("After Req %#v", r)
	message.Debugf("After Req URL%#v", r.URL)

	return nil
}

func proxyResponseTransform(resp *http.Response) error {
	message.Debugf("Before Resp %#v", resp)

	// Handle redirection codes (3xx) by adding a marker to let Jackal know this has been redirected
	if resp.StatusCode/100 == 3 {
		message.Debugf("Before Resp Location %#v", resp.Header.Get("Location"))

		locationURL, err := url.Parse(resp.Header.Get("Location"))
		message.Debugf("%#v", err)
		locationURL.Path = transform.NoTransform + locationURL.Path
		locationURL.Host = resp.Request.Header.Get("X-Forwarded-Host")

		resp.Header.Set("Location", locationURL.String())

		message.Debugf("After Resp Location %#v", resp.Header.Get("Location"))
	}

	contentType := resp.Header.Get("Content-Type")

	// Handle text content returns that may contain links
	if strings.HasPrefix(contentType, "text") || strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "application/xml") {
		err := replaceBodyLinks(resp)

		if err != nil {
			message.Debugf("%#v", err)
		}
	}

	message.Debugf("After Resp %#v", resp)

	return nil
}

func replaceBodyLinks(resp *http.Response) error {
	message.Debugf("Resp Request: %#v", resp.Request)

	// Create the forwarded (online) and target (offline) URL prefixes to replace
	forwardedPrefix := fmt.Sprintf("%s%s%s", getTLSScheme(resp.Request.TLS), resp.Request.Header.Get("X-Forwarded-Host"), transform.NoTransform)
	targetPrefix := fmt.Sprintf("%s%s", getTLSScheme(resp.TLS), resp.Request.Host)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = resp.Body.Close()
	if err != nil {
		return err
	}

	bodyString := string(body)
	message.Warnf("%s", bodyString)

	bodyString = strings.ReplaceAll(bodyString, targetPrefix, forwardedPrefix)

	message.Warnf("%s", bodyString)

	// Setup the new reader, and correct the content length
	resp.Body = io.NopCloser(strings.NewReader(bodyString))
	resp.ContentLength = int64(len(bodyString))
	resp.Header.Set("Content-Length", fmt.Sprint(int64(len(bodyString))))

	return nil
}

func getTLSScheme(tls *tls.ConnectionState) string {
	scheme := "https://"

	if tls == nil {
		scheme = "http://"
	}

	return scheme
}

func getRequestURI(path, query, fragment string) string {
	uri := path

	if query != "" {
		uri += "?" + query
	}

	if fragment != "" {
		uri += "#" + fragment
	}

	return uri
}

func isGitUserAgent(userAgent string) bool {
	return strings.HasPrefix(userAgent, "git")
}

func isPipUserAgent(userAgent string) bool {
	return strings.HasPrefix(userAgent, "pip") || strings.HasPrefix(userAgent, "twine")
}

func isNpmUserAgent(userAgent string) bool {
	return strings.HasPrefix(userAgent, "npm") || strings.HasPrefix(userAgent, "pnpm") || strings.HasPrefix(userAgent, "yarn") || strings.HasPrefix(userAgent, "bun")
}
