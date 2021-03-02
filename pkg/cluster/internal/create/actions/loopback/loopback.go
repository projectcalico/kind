/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package kubeadminit implements the kubeadm init action
package loopback

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kind/pkg/cluster/internal/create/actions"
	"sigs.k8s.io/kind/pkg/errors"
)

// kubeadmInitAction implements action for executing the kubadm init
// and a set of default post init operations like e.g. install the
// CNI network plugin.
type action struct{}

// NewAction returns a new action for kubeadm init
func NewAction() actions.Action {
	return &action{}
}

// Execute runs the action
func (a *action) Execute(ctx *actions.ActionContext) error {
	allNodes, err := ctx.Nodes()
	if err != nil {
		return err
	}

	for _, node := range allNodes {
		loopAddress, err := node.Loopback()
		if err != nil {
			fmt.Printf("Loopback action error: %v\n", err)
			continue
		}
		if loopAddress != "" {
			fmt.Printf("Add loopback %v for node %v\n", loopAddress, node)
			cmd := node.Command(
				"ip",
				"a",
				"a",
				loopAddress+"/32",
				"dev",
				"lo",
			)
			err := cmd.Run()
			if err != nil {
				return errors.Wrap(err, "failed to set loopback address")
			}
		}

		routes, err := node.Routes()
		if err != nil {
			fmt.Printf("Loopback action error: %v\n", err)
			continue
		}
		for _, route := range strings.Split(routes, ",") {
			fmt.Printf("Install route: %v\n", route)
			args := []string{"r", "a"}
			args = append(args, strings.Split(route, " ")...)
			cmd := node.Command("ip", args...)
			err := cmd.Run()
			if err != nil {
				return errors.Wrap(err, "failed to install route")
			}
		}

	}
	return nil
}
