package k8s

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

func createPatch(numWebhooks int, caBundle string) []byte {
	patches := make([]webhookCaPatch, numWebhooks)
	for i := 0; i < numWebhooks; i++ {
		patches[i] = webhookCaPatch{
			Op:    "add",
			Path:  fmt.Sprintf("/webhooks/%d/clientConfig/caBundle", i),
			Value: caBundle}
	}

	dat, _ := json.Marshal(patches)
	return dat
}

// PatchWebhookConfigurations will patch validatingWebhook and mutatingWebhook clientConfig configurations with
// the provided ca data.
func PatchWebhookConfigurations(configurationNames string, ca []byte, patchMutating bool, patchValidating bool) {
	log.Debugf("Patching webhook configurations '%s' mutating=%t, validating=%t", configurationNames, patchMutating, patchValidating)
	caBase64 := base64.StdEncoding.EncodeToString(ca)
	k8s := getk8s()
	if patchValidating {
		valHook, err := k8s.clientset.
			AdmissionregistrationV1beta1().
			ValidatingWebhookConfigurations().
			Get(configurationNames, metav1.GetOptions{})
		if err != nil {
			log.WithField("err", err).Fatal("Failed getting validating webhook")
		}
		patch := createPatch(len(valHook.Webhooks), caBase64)
		_, err = k8s.clientset.
			AdmissionregistrationV1beta1().
			ValidatingWebhookConfigurations().
			Patch(configurationNames, types.JSONPatchType, patch)
		if err != nil {
			log.WithField("err", err).Fatal("Failed patching validating webhook")
		}
		log.Debug("Patched validating hook")
	} else {
		log.Debug("Validating hook patching not required")
	}

	if patchMutating {
		mut, err := k8s.clientset.
			AdmissionregistrationV1beta1().
			MutatingWebhookConfigurations().
			Get(configurationNames, metav1.GetOptions{})
		if err != nil {
			log.WithField("err", err).Fatal("Failed getting validating webhook")
		}
		patch := createPatch(len(mut.Webhooks), caBase64)
		_, err = k8s.clientset.
			AdmissionregistrationV1beta1().
			MutatingWebhookConfigurations().
			Patch(configurationNames, types.JSONPatchType, patch)
		if err != nil {
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
