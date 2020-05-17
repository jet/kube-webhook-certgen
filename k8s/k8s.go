package k8s

import (
	"context"
	"encoding/base64"
	"errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type WebhookType string

const (
	group                      = "admissionregistration.k8s.io"
	ValidatingHook WebhookType = "validatingwebhookconfigurations"
	MutatingHook   WebhookType = "mutatingwebhookconfigurations"
)

var (
	versions = []string{"v1", "v1beta1"}
)

type K8s interface {
	UpdateWebhook(name string, ca []byte, policyType string, hookType WebhookType) bool
	GetCaFromSecret(secretName, namespace, key string) ([]byte, bool)
	SaveCertsToSecret(secretName, namespace, certName, keyName, caName string, ca, cert, key []byte) bool
}

type k8s struct {
	client kubernetes.Interface
	dyn    dynamic.Interface
}

func New(kubeconfig string) K8s {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.WithError(err).Fatal("error building kubernetes config")
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithError(err).Fatal("error creating kubernetes client")
	}
	d, err := dynamic.NewForConfig(config)
	if err != nil {
		log.WithError(err).Fatal("error creating dynamic client")
	}

	return &k8s{client: c, dyn: d}
}

func NewFake(client kubernetes.Interface, dyn dynamic.Interface) K8s {
	return &k8s{
		client: client,
		dyn:    dyn,
	}
}

func (k8s *k8s) UpdateWebhook(name string, ca []byte, policyType string, hookType WebhookType) bool {
	l := log.WithField("name", name).WithField("type", hookType)
	l.Debug("Patching hook")
	resource, gvk, err := k8s.getWebhookDynamic(name, hookType)
	if err != nil {
		l.WithError(err).Error("Resource not found")
	}
	w, ok := resource.Object["webhooks"]
	if !ok {
		l.Error("Unable to read 'spec.webhooks'")
		return false
	}

	wh, ok := w.([]interface{})
	if !ok {
		l.Error("Unable to interpret 'spec.webhooks' as '[]interface{}'")
		return false
	}

	for _, h := range wh {
		hook, ok := h.(map[string]interface{})
		if !ok {
			l.Error("Unable to interpret single element in 'webhooks' as 'map[string]interface{}'")
			return false
		}
		if policyType != "" {
			hook["failurePolicy"] = policyType
		}

		cc, ok := hook["clientConfig"]
		if !ok {
			cc = make(map[string]interface{}, 1)
		}
		clientConfig, ok := cc.(map[string]interface{})
		if !ok {
			l.Error("Unable to interpret 'clientConfig' as 'map[string]interface{}'")
			return false
		}

		clientConfig["caBundle"] = base64.StdEncoding.EncodeToString(ca)
		hook["clientConfig"] = clientConfig
	}

	if _, err := k8s.dyn.Resource(*gvk).Update(context.Background(), resource, metav1.UpdateOptions{}); err != nil {
		l.WithError(err).Error("Unable to update configuration")
		return false
	}
	l.Debug("Patched successfully")
	return true
}

func (k8s *k8s) getWebhookDynamic(name string, typ WebhookType) (*unstructured.Unstructured, *schema.GroupVersionResource, error) {
	for _, v := range versions {
		gvk := &schema.GroupVersionResource{Group: group, Version: v, Resource: string(typ)}
		res, err := k8s.dyn.Resource(*gvk).Get(context.Background(), name, metav1.GetOptions{})
		if err == nil {
			return res, gvk, err
		}
		if !k8serrors.IsNotFound(err) {
			return nil, nil, err
		}
	}

	return nil, nil, errors.New("no webhook found")
}

// GetCaFromSecret will check for the presence of a secret. If it exists, will return the content of the
// key from the secret, otherwise will return nil
func (k8s *k8s) GetCaFromSecret(secretName, namespace, key string) ([]byte, bool) {
	log.Debugf("getting secret '%s' in namespace '%s'", secretName, namespace)
	secret, err := k8s.client.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.WithField("err", err).Info("no secret found")
			return nil, false
		}
		log.WithField("err", err).Error("error getting secret")
		return nil, false
	}

	data, ok := secret.Data[key]
	if !ok {
		log.Errorf("got secret, but it did not contain a '%s' key", key)
		return nil, false
	}
	log.Debug("got secret")
	return data, true
}

// SaveCertsToSecret saves the provided ca, cert and key into a secret in the specified namespace.
func (k8s *k8s) SaveCertsToSecret(secretName, namespace, certName, keyName, caName string, ca, cert, key []byte) bool {
	log.Debugf("saving to secret '%s' in namespace '%s'", secretName, namespace)
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{caName: ca, certName: cert, keyName: key},
	}

	log.Debug("saving secret")
	_, err := k8s.client.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		log.WithField("err", err).Error("failed creating secret")
		return false
	}
	log.Debug("saved secret")
	return true
}
