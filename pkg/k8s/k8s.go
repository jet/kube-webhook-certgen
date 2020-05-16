package k8s

import (
	"context"
	log "github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type k8s struct {
	clientset kubernetes.Interface
}

func New(kubeconfig string) *k8s {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.WithError(err).Fatal("error building kubernetes config")
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithError(err).Fatal("error creating kubernetes client")
	}

	return &k8s{clientset: c}
}

// UpdateMutating will amend a validating webhook configuration with the CA and policyType specified
func (k8s *k8s) UpdateValidating(name string, ca []byte, policyType *admissionv1.FailurePolicyType) {
	valHook, err := k8s.clientset.
		AdmissionregistrationV1().
		ValidatingWebhookConfigurations().
		Get(context.Background(), name, metav1.GetOptions{})

	if err != nil {
		log.WithField("err", err).Fatal("failed getting validating webhook")
	}

	for i := range valHook.Webhooks {
		h := &valHook.Webhooks[i]
		h.ClientConfig.CABundle = ca
		if *policyType != "" {
			h.FailurePolicy = policyType
		}
	}

	if _, err = k8s.clientset.AdmissionregistrationV1().
		ValidatingWebhookConfigurations().
		Update(context.Background(), valHook, metav1.UpdateOptions{}); err != nil {
		log.WithField("err", err).Fatal("failed patching validating webhook")
	}
	log.Debug("patched validating hook")
}

// UpdateMutating will amend a mutating webhook configuration with the CA and policyType specified
func (k8s *k8s) UpdateMutating(names string, ca []byte, policyType *admissionv1.FailurePolicyType) {
	mutHook, err := k8s.clientset.
		AdmissionregistrationV1().
		MutatingWebhookConfigurations().
		Get(context.Background(), names, metav1.GetOptions{})
	if err != nil {
		log.WithField("err", err).Fatal("failed getting validating webhook")
	}

	for i := range mutHook.Webhooks {
		h := &mutHook.Webhooks[i]
		h.ClientConfig.CABundle = ca
		if *policyType != "" {
			h.FailurePolicy = policyType
		}
	}

	if _, err = k8s.clientset.AdmissionregistrationV1().
		MutatingWebhookConfigurations().
		Update(context.Background(), mutHook, metav1.UpdateOptions{}); err != nil {
		log.WithField("err", err).Fatal("failed patching validating webhook")
	}
	log.Debug("patched mutating hook")
}

// GetCaFromSecret will check for the presence of a secret. If it exists, will return the content of the
// "ca" from the secret, otherwise will return nil
func (k8s *k8s) GetCaFromSecret(secretName string, namespace string) []byte {
	log.Debugf("getting secret '%s' in namespace '%s'", secretName, namespace)
	secret, err := k8s.clientset.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.WithField("err", err).Info("no secret found")
			return nil
		}
		log.WithField("err", err).Fatal("error getting secret")
	}

	data := secret.Data["ca"]
	if data == nil {
		log.Fatal("got secret, but it did not contain a 'cert' key")
	}
	log.Debug("got secret")
	return data
}

// SaveCertsToSecret saves the provided ca, cert and key into a secret in the specified namespace.
func (k8s *k8s) SaveCertsToSecret(secretName, namespace, certName, keyName string, ca, cert, key []byte) {

	log.Debugf("saving to secret '%s' in namespace '%s'", secretName, namespace)
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{"ca": ca, certName: cert, keyName: key},
	}

	log.Debug("saving secret")
	_, err := k8s.clientset.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		log.WithField("err", err).Fatal("failed creating secret")
	}
	log.Debug("saved secret")
}
