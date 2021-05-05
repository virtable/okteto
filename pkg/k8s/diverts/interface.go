package diverts

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

type DivertInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*DivertList, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*Divert, error)
	Create(ctx context.Context, divert *Divert) (*Divert, error)
	Delete(ctx context.Context, name string, options metav1.DeleteOptions) error
}

type divertClient struct {
	restClient rest.Interface
	scheme     *runtime.Scheme
	ns         string
}

func (c *divertClient) List(ctx context.Context, opts metav1.ListOptions) (*DivertList, error) {
	result := DivertList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("diverts").
		VersionedParams(&opts, runtime.NewParameterCodec(c.scheme)).
		Do(ctx).
		Into(&result)
	return &result, err
}

func (c *divertClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*Divert, error) {
	result := Divert{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("diverts").
		Name(name).
		VersionedParams(&opts, runtime.NewParameterCodec(c.scheme)).
		Do(ctx).
		Into(&result)
	return &result, err
}

func (c *divertClient) Create(ctx context.Context, divert *Divert) (*Divert, error) {
	result := Divert{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("diverts").
		Body(divert).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *divertClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.restClient.
		Delete().
		Namespace(c.ns).
		Resource("diverts").
		Name(name).
		// VersionedParams(&opts, runtime.NewParameterCodec(c.scheme)).
		Do(ctx).Error()
}
