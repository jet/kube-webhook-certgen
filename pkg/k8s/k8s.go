package k8s

import (
	log "github.com/sirupsen/logrus"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
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

// PatchWebhookConfigurations will patch validatingWebhook and mutatingWebhook clientConfig configurations with
// the provided ca data. If failurePolicy is provided, patch all webhooks with this value
func (k8s *k8s) PatchWebhookConfigurations(
	configurationNames string, ca []byte,
	failurePolicy *admissionv1beta1.FailurePolicyType,
	patchMutating bool, patchValidating bool) {

	log.Debugf("Patching webhook configurations '%s' mutating=%t, validating=%t", configurationNames, patchMutating, patchValidating)

	if patchValidating {
		valHook, err := k8s.clientset.
			AdmissionregistrationV1beta1().
			ValidatingWebhookConfigurations().
			Get(configurationNames, metav1.GetOptions{})

		if err != nil {
			log.WithField("err", err).Fatal("Failed getting validating webhook")
		}

		for i := range valHook.Webhooks {
			h := &valHook.Webhooks[i]
			h.ClientConfig.CABundle = ca
			if failurePolicy != nil {
				h.FailurePolicy = failurePolicy
			}
		}

		if _, err = k8s.clientset.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Update(valHook); err != nil {
			log.WithField("err", err).Fatal("Failed patching validating webhook")
		}
		log.Debug("Patched validating hook")
	} else {
		log.Debug("Validating hook patching not required")
	}

	if patchMutating {
		mutHook, err := k8s.clientset.
			AdmissionregistrationV1beta1().
			MutatingWebhookConfigurations().
			Get(configurationNames, metav1.GetOptions{})
		if err != nil {
			log.WithField("err", err).Fatal("Failed getting validating webhook")
		}

		for i := range mutHook.Webhooks {
			h := &mutHook.Webhooks[i]
			h.ClientConfig.CABundle = ca
			if failurePolicy != nil {
				h.FailurePolicy = failurePolicy
			}
		}

		if _, err = k8s.clientset.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Update(mutHook); err != nil {
			log.WithField("err", err).Fatal("Failed patching validating webhook")
		}
		log.Debug("Patched mutating hook")
	} else {
		log.Debug("Mutating hook patching not required")
	}
}

// GetCertFromSecret will check for the presence of a secret. If it exists, will return the content of the
// "cert" from the secret, otherwise will return nil
func (k8s *k8s) GetCertFromSecret(secretName string, namespace string) []byte {
	log.Debugf("Getting secret '%s' in namespace '%s'", secretName, namespace)
	secret, err := k8s.clientset.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.WithField("err", err).Info("No secret found")
			return nil
		}
		log.WithField("err", err).Fatal("Error getting secret")
	}

	data := secret.Data["cert"]
	if data == nil {
		log.Fatal("Got secret, but it did not contain a 'cert' key")
	}
	log.Debug("Got secret")
	return data
}

// SaveCertsToSecret saves the provided ca, cert and key into a secret in the specified namespace.
func (k8s *k8s) SaveCertsToSecret(secretName string, namespace string, cert []byte, key []byte) {

	log.Debugf("Saving to secret '%s' in namespace '%s'", secretName, namespace)
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{"cert": cert, "key": key},
	}

	log.Debug("Saving secret")
	_, err := k8s.clientset.CoreV1().Secrets(namespace).Create(secret)
	if err != nil {
		log.WithField("err", err).Fatal("Failed creating secret")
	}
	log.Debug("Saved secret")
}
