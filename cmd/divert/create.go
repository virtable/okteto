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

	"github.com/okteto/okteto/cmd/up"
	"github.com/okteto/okteto/cmd/utils"
	"github.com/okteto/okteto/pkg/analytics"
	k8Client "github.com/okteto/okteto/pkg/k8s/client"
	"github.com/okteto/okteto/pkg/k8s/deployments"
	"github.com/okteto/okteto/pkg/k8s/ingress"
	okLabels "github.com/okteto/okteto/pkg/k8s/labels"
	"github.com/okteto/okteto/pkg/k8s/services"
	"github.com/okteto/okteto/pkg/okteto"
	"github.com/spf13/cobra"
)

//Create creates a divert
func Create(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "divert",
		Short: "Creates a divert",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := executeCreateDivert(ctx)
			analytics.TrackCreateDivert(err == nil)
			return err
		},
	}

	return cmd
}

//TODO: spinners
//TODO: show divert URL!
func executeCreateDivert(ctx context.Context) error {

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

	//Duplicate ingress
	i, err := ingress.Get(ctx, dev.Divert.Ingress, dev.Namespace, c)
	if err != nil {
		//TODO: better error?
		return fmt.Errorf("get ingress '%s': %s", dev.Divert.Ingress, err.Error())
	}

	//TODO: move to translate ingress function
	i.Name = fmt.Sprintf("%s-%s", i.Name, username)
	if i.Annotations == nil {
		i.Annotations = map[string]string{}
	}
	i.Annotations[okLabels.OktetoIngressAutoGenerateHost] = "true"
	i.ResourceVersion = ""
	//TODO: make idempotent
	if err := ingress.Create(ctx, i, c); err != nil {
		//TODO: better error?
		return fmt.Errorf("create ingress '%s': %s", dev.Divert.Ingress, err.Error())
	}

	//Duplicate service

	s, err := services.Get(ctx, dev.Divert.Service, dev.Namespace, c)
	if err != nil {
		//TODO: better error?
		return fmt.Errorf("get service '%s': %s", dev.Divert.Service, err.Error())
	}
	//TODO: move to translate service function
	s.Name = fmt.Sprintf("%s-%s", s.Name, username)
	//TODO: take into account if sourcce service is already diverted
	//TODO: make idempotent
	s.ResourceVersion = ""
	s.Spec.ClusterIP = ""
	if err := services.Create(ctx, s, c); err != nil {
		//TODO: better error?
		return fmt.Errorf("create service '%s': %s", dev.Divert.Service, err.Error())
	}

	//Create Divert

	//Duplicate deployment
	d, err := deployments.Get(ctx, dev, dev.Namespace, c)
	if err != nil {
		//TODO: better error?
		return fmt.Errorf("get deployment '%s': %s", dev.Name, err.Error())
	}
	//TODO: move to translate deployment function
	d.Name = fmt.Sprintf("%s-%s", d.Name, username)
	d.ResourceVersion = ""
	//TODO: make idempotent
	if err := deployments.Deploy(ctx, d, true, c); err != nil {
		//TODO: better error?
		return fmt.Errorf("deploy deployment '%s': %s", d.Name, err.Error())
	}

	dev.Name = fmt.Sprintf("%s-%s", d.Name, username)
	dev.Labels = nil
	return up.ExecuteUp("okteto.yml", "", "", 0, false, false, false, false)
}
