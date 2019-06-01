package k8s

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sync"
)

type webhookCaPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

type k8s struct {
	clientset kubernetes.Interface
}

var (
	clientm     sync.Mutex
	client      *k8s
	clientError error
)

func getk8s() *k8s {
	if client == nil && clientError == nil {
		clientm.Lock()
		if client == nil && clientError == nil {
			client, clientError = initClient()
		}
		clientm.Unlock()
	}

	if clientError != nil {
		log.WithField("err", clientError).Fatal("Failed getting client")
	}

	return client
}

func initClient() (*k8s, error) {
	// Create a config from the config file, or a InCluster config if empty.
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		return nil, errors.Wrap(err, "error building kubernetes config")
	}

	// Create the client.
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "error creating kubernetes client")
	}

	return &k8s{clientset: c}, err
}

// PatchWebhookConfigurations will patch validatingWebhook and mutatingWebhook clientConfig configurations with
// the provided ca data. If failurePolicy is provided, patch all webhooks with this value
func PatchWebhookConfigurations(
	configurationNames string, ca []byte,
	failurePolicy *admissionv1beta1.FailurePolicyType,
	patchMutating bool, patchValidating bool) {

	log.Debugf("Patching webhook configurations '%s' mutating=%t, validating=%t", configurationNames, patchMutating, patchValidating)
	k8s := getk8s()

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

// GetCaFromCertificate will check for the presence of a secret. If it exists, will return the content of the
// "ca" from the secret, otherwise will return nil
func GetCaFromCertificate(secretName string, namespace string) []byte {

	k8s := getk8s()
	log.Debugf("Getting secret '%s' in namespace '%s'", secretName, namespace)
	secret, err := k8s.clientset.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.WithField("err", err).Info("No secret found")
			return nil
		}
		log.WithField("err", err).Fatal("Error getting secret")
	}

	data := secret.Data["ca"]
	if data == nil {
		log.Fatal("Got secret, but it did not contain a 'ca' key")
	}
	log.Debug("Got secret")
	return data
}

// SaveCertsToSecret saves the provided ca, cert and key into a secret in the specified namespace.
func SaveCertsToSecret(secretName string, namespace string, ca []byte, cert []byte, key []byte) {

	log.Debugf("Saving to secret '%s' in namespace '%s'", secretName, namespace)
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{"ca": ca, "cert": cert, "key": key},
	}

	log.Debug("Saving secret")
	_, err := getk8s().clientset.CoreV1().Secrets(namespace).Create(secret)
	if err != nil {
		log.WithField("err", err).Fatal("Failed creating secret")
	}
	log.Debug("Saved secret")
}
