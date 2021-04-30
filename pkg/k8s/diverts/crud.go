package diverts

import (
	"context"
	"fmt"

	"github.com/okteto/okteto/pkg/errors"
	"github.com/okteto/okteto/pkg/k8s/deployments"
	"github.com/okteto/okteto/pkg/k8s/ingress"
	okLabels "github.com/okteto/okteto/pkg/k8s/labels"
	"github.com/okteto/okteto/pkg/k8s/services"
	"github.com/okteto/okteto/pkg/model"
	"github.com/okteto/okteto/pkg/okteto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Create(ctx context.Context, dev *model.Dev, c kubernetes.Interface) error {

	if dev.Divert == nil {
		return nil
	}

	//TODO: error if not okteto namespace

	//TODO: sanitize username to have azAZ09-
	username := okteto.GetUsername()

	//Duplicate ingress
	i, err := ingress.Get(ctx, dev.Divert.Ingress, dev.Namespace, c)
	if err != nil {
		//TODO: better error?
		return fmt.Errorf("get ingress '%s': %s", dev.Divert.Ingress, err.Error())
	}

	//TODO: move to translate ingress function
	i.Name = fmt.Sprintf("%s-%s", username, i.Name)
	if i.Annotations == nil {
		i.Annotations = map[string]string{}
	}
	if host := i.Annotations[okLabels.OktetoIngressAutoGenerateHost]; host != "" {
		if host != "true" {
			i.Annotations[okLabels.OktetoIngressAutoGenerateHost] = fmt.Sprintf("%s-%s", username, host)
		}
	} else {
		i.Annotations[okLabels.OktetoIngressAutoGenerateHost] = "true"
	}
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
	s.Name = fmt.Sprintf("%s-%s", username, s.Name)
	//TODO: take into account if sourcce service is already diverted
	//TODO: make idempotent
	s.ResourceVersion = ""
	s.Spec.ClusterIP = ""
	delete(s.Annotations, "dev.okteto.com/auto-ingress")
	if err := services.Create(ctx, s, c); err != nil {
		//TODO: better error?
		return fmt.Errorf("create service '%s': %s", dev.Divert.Service, err.Error())
	}

	dClient, err := GetClient(dev.Context)
	if err != nil {
		//TODO: better error?
		return err
	}
	div, err := dClient.Diverts(dev.Namespace).Get(ctx, s.Name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			//TODO: better error?
			return fmt.Errorf("getting divert: %s", err.Error())
		}
	}
	//TODO: move to translate divert function
	div = &Divert{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Divert",
			APIVersion: "weaver.okteto.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: dev.Namespace,
		},
		Spec: DivertSpec{
			Ingress: IngressDivertSpec{
				Name:      i.Name,
				Namespace: dev.Namespace,
				Value:     username,
			},
			FromService: ServiceDivertSpec{
				Name:      dev.Divert.Service,
				Namespace: dev.Namespace,
				Port:      dev.Divert.Port,
			},
			ToService: ServiceDivertSpec{
				Name:      s.Name,
				Namespace: dev.Namespace,
				Port:      dev.Divert.Port,
			},
		},
	}

	if _, err := dClient.Diverts(dev.Namespace).Create(ctx, div); err != nil {
		//TODO: better error?
		return fmt.Errorf("creating divert: %s", err.Error())
	}

	//Duplicate deployment
	d, err := deployments.Get(ctx, dev, dev.Namespace, c)
	if err != nil {
		//TODO: better error?
		return fmt.Errorf("get deployment '%s': %s", dev.Name, err.Error())
	}
	//TODO: move to translate deployment function
	d.Name = fmt.Sprintf("%s-%s", username, d.Name)
	delete(d.Spec.Template.Annotations, "divert.okteto.com/inject-sidecar")
	d.ResourceVersion = ""
	//TODO: make idempotent
	if err := deployments.Deploy(ctx, d, true, c); err != nil {
		//TODO: better error?
		return fmt.Errorf("deploy deployment '%s': %s", d.Name, err.Error())
	}

	dev.Name = fmt.Sprintf("%s-%s", username, dev.Name)
	dev.Labels = nil
	return nil
}

func Delete(ctx context.Context, dev *model.Dev, c kubernetes.Interface) error {
	if dev.Divert == nil {
		return nil
	}

	//TODO: error if not okteto namespace

	//TODO: sanitize username to have azAZ09-
	username := okteto.GetUsername()

	//Delete Deployment
	dName := fmt.Sprintf("%s-%s", username, dev.Name)
	if err := deployments.Destroy(ctx, dName, dev.Namespace, c); err != nil {
		//TODO: better error?
		return fmt.Errorf("delete deployment '%s': %s", dName, err.Error())
	}

	//Delete Divert CRD
	dClient, err := GetClient(dev.Context)
	if err != nil {
		//TODO: better error?
		return fmt.Errorf("getting divert client: %s", err.Error())
	}
	sName := fmt.Sprintf("%s-%s", username, dev.Divert.Service)
	if err := dClient.Diverts(dev.Namespace).Delete(ctx, sName, metav1.DeleteOptions{}); err != nil {
		if !errors.IsNotFound(err) {
			//TODO: better error?
			return fmt.Errorf("getting divert: %s", err.Error())
		}
	}

	//Delete Service
	if err := services.Destroy(ctx, sName, dev.Namespace, c); err != nil {
		//TODO: better error?
		return fmt.Errorf("delete service '%s': %s", sName, err.Error())
	}

	//Delete Ingress
	iName := fmt.Sprintf("%s-%s", username, dev.Divert.Ingress)
	if err := ingress.Destroy(ctx, iName, dev.Namespace, c); err != nil {
		//TODO: better error?
		return fmt.Errorf("delete ingress '%s': %s", iName, err.Error())
	}

	return nil
}
