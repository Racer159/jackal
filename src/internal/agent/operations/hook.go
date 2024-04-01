// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package operations provides functions for the mutating webhook.
package operations

import (
	"fmt"

	"github.com/racer159/jackal/src/config/lang"
	"github.com/racer159/jackal/src/pkg/message"
	admission "k8s.io/api/admission/v1"
)

// Result contains the result of an admission request.
type Result struct {
	Allowed  bool
	Msg      string
	PatchOps []PatchOperation
}

// AdmitFunc defines how to process an admission request.
type AdmitFunc func(request *admission.AdmissionRequest) (*Result, error)

// Hook represents the set of functions for each operation in an admission webhook.
type Hook struct {
	Create  AdmitFunc
	Delete  AdmitFunc
	Update  AdmitFunc
	Connect AdmitFunc
}

// Execute evaluates the request and try to execute the function for operation specified in the request.
func (h *Hook) Execute(r *admission.AdmissionRequest) (*Result, error) {
	message.Debugf("operations.Execute(*admission.AdmissionRequest) - %#v , %s/%s: %#v", r.Kind, r.Namespace, r.Name, r.Operation)

	switch r.Operation {
	case admission.Create:
		return wrapperExecution(h.Create, r)
	case admission.Update:
		return wrapperExecution(h.Update, r)
	case admission.Delete:
		return wrapperExecution(h.Delete, r)
	case admission.Connect:
		return wrapperExecution(h.Connect, r)
	}

	return &Result{Msg: fmt.Sprintf(lang.AgentErrInvalidOp, r.Operation)}, nil
}

// If the mutatingwebhook calls for an operation with no bound function--go tell on them.
func wrapperExecution(fn AdmitFunc, r *admission.AdmissionRequest) (*Result, error) {
	if fn == nil {
		return nil, fmt.Errorf(lang.AgentErrInvalidOp, r.Operation)
	}
	return fn(r)
}
