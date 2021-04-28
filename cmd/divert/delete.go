// Copyright 2020 The Okteto Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package divert

import (
	"context"
	"fmt"

	"github.com/okteto/okteto/cmd/utils"
	"github.com/okteto/okteto/pkg/analytics"
	k8Client "github.com/okteto/okteto/pkg/k8s/client"
	"github.com/okteto/okteto/pkg/k8s/deployments"
	"github.com/okteto/okteto/pkg/k8s/ingress"
	"github.com/okteto/okteto/pkg/k8s/services"
	"github.com/okteto/okteto/pkg/okteto"
	"github.com/spf13/cobra"
)

//Delete deletes a divert
func Delete(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "divert",
		Short: "Deletes a divert",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := executeDeleteDivert(ctx)
			analytics.TrackDeleteDivert(err == nil)
			return err
		},
	}
}

//TODO: spinners
func executeDeleteDivert(ctx context.Context) error {
	//TODO: check everything missing in up sequence

	dev, err := utils.LoadDev("okteto.yml", "", "")
	if err != nil {
		return err
	}

	if dev.Divert == nil {
		return fmt.Errorf("'divert' field in your okteto manifest is not configured")
	}

	c, _, err := k8Client.GetLocalWithContext(dev.Context)
	if err != nil {
		return err
	}

	//TODO: sanitize username to have azAZ09-
	username := okteto.GetUsername()

	//Delete Deployment
	dName := fmt.Sprintf("%s-%s", dev.Name, username)
	if err := deployments.Destroy(ctx, dName, dev.Namespace, c); err != nil {
		//TODO: better error?
		return fmt.Errorf("delete deployment '%s': %s", dName, err.Error())
	}

	//Delete Divert CRD

	//Delete Service
	sName := fmt.Sprintf("%s-%s", dev.Divert.Service, username)
	if err := services.Destroy(ctx, sName, dev.Namespace, c); err != nil {
		//TODO: better error?
		return fmt.Errorf("delete service '%s': %s", sName, err.Error())
	}

	//Delete Ingress
	iName := fmt.Sprintf("%s-%s", dev.Divert.Ingress, username)
	if err := ingress.Destroy(ctx, iName, dev.Namespace, c); err != nil {
		//TODO: better error?
		return fmt.Errorf("delete ingress '%s': %s", iName, err.Error())
	}

	return nil

}
