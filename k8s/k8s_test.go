package k8s_test

import (
	"bytes"
	"context"
	"github.com/jet/kube-webhook-certgen/certs"
	"github.com/jet/kube-webhook-certgen/k8s"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	kdyn "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	kfake "k8s.io/client-go/kubernetes/fake"

	"testing"
)

const (
	testWebhookName = "c7c95710-d8c3-4cc3-a2a8-8d2b46909c76"
	testSecretName  = "15906410-af2a-4f9b-8a2d-c08ffdd5e129"
	testNamespace   = "7cad5f92-c0d5-4bc9-87a3-6f44d5a5619d"
)

var (
	fail   = admissionv1.Fail
	ignore = admissionv1.Ignore
)

func genSecretData() (ca, cert, key []byte) {
	ca, cert, key, _ = certs.GenerateCerts([]string{"test"})
	return
}

type t8s struct {
	client kubernetes.Interface
	dyn    dynamic.Interface
}

func clients() (kubernetes.Interface, dynamic.Interface) {
	return kfake.NewSimpleClientset(), kdyn.NewSimpleDynamicClient(runtime.NewScheme())
}

func TestGetCaFromCertificate(t *testing.T) {
	client, dyn := clients()
	k := k8s.NewFake(client, dyn)
	ca, cert, key := genSecretData()

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: testSecretName,
		},
		Data: map[string][]byte{"ca": ca, "cert": cert, "key": key},
	}

	client.CoreV1().Secrets(testNamespace).Create(context.Background(), secret, metav1.CreateOptions{})

	retrievedCa, _ := k.GetCaFromSecret(testSecretName, testNamespace, "ca")
	if !bytes.Equal(retrievedCa, ca) {
		t.Error("Was not able to retrieve CA information that was saved")
	}
}

func TestGetCaFromCertificateShouldFailWhenMissing(t *testing.T) {
	client, dyn := clients()
	k := k8s.NewFake(client, dyn)
	ca, cert, key := genSecretData()

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: testSecretName,
		},
		Data: map[string][]byte{"ca": ca, "cert": cert, "key": key},
	}
	client.CoreV1().Secrets(testNamespace).Create(context.Background(), secret, metav1.CreateOptions{})

	// Should not retrieve data when wrong secret name
	retrievedCa, ok := k.GetCaFromSecret("junk", testNamespace, "ca")
	if ok || retrievedCa != nil {
		t.Fatal("Expected error due to ca data missing")
	}

	// Should not retrieve data when wrong key
	retrievedCa, ok = k.GetCaFromSecret(testSecretName, testNamespace, "junk")
	if ok || retrievedCa != nil {
		t.Fatal("Expected error due to ca data missing")
	}
}

func TestSaveCertsToSecret(t *testing.T) {
	client, dyn := clients()
	k := k8s.NewFake(client, dyn)

	ca, cert, key := genSecretData()

	k.SaveCertsToSecret(testSecretName, testNamespace, "cert", "key", "ca", ca, cert, key)

	secret, _ := client.CoreV1().Secrets(testNamespace).Get(context.Background(), testSecretName, metav1.GetOptions{})

	if !bytes.Equal(secret.Data["cert"], cert) {
		t.Error("'cert' saved data does not match retrieved")
	}

	if !bytes.Equal(secret.Data["key"], key) {
		t.Error("'key' saved data does not match retrieved")
	}
}

func TestSaveThenLoadSecret(t *testing.T) {
	client, dyn := clients()
	k := k8s.NewFake(client, dyn)
	ca, cert, key := genSecretData()
	k.SaveCertsToSecret(testSecretName, testNamespace, "cert", "key", "ca", ca, cert, key)
	retrievedCert, _ := k.GetCaFromSecret(testSecretName, testNamespace, "ca")
	if !bytes.Equal(retrievedCert, ca) {
		t.Error("Was not able to retrieve CA information that was saved")
	}
}

func TestPatchWebhookConfigurations(t *testing.T) {
	client, dyn := clients()
	k := k8s.NewFake(client, dyn)
	// Versions in tests are in reverse preference. If the v1 configurations exist they are patched in preference
	// to the v1beta1 configurations, which are ignored if a higher preference configuration is found.
	for _, version := range []string{"v1beta1", "v1"} {
		t.Run(version, func(t *testing.T) {
			gvk := schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: version}
			ca, _, _ := genSecretData()
			err := createDynamic(dyn, &admissionv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{Name: testWebhookName},
				Webhooks:   []admissionv1.MutatingWebhook{{Name: "m1"}, {Name: "m2"}}},
				gvkWithType(gvk, k8s.MutatingHook))
			if err != nil {
				t.Error(err)
			}

			err = createDynamic(dyn, &admissionv1.ValidatingWebhookConfiguration{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: testWebhookName,
				},
				Webhooks: []admissionv1.ValidatingWebhook{{Name: "v1"}, {Name: "v2"}}},
				gvkWithType(gvk, k8s.ValidatingHook))
			if err != nil {
				t.Error(err)
			}

			k.UpdateWebhook(testWebhookName, ca, "Fail", k8s.MutatingHook)
			k.UpdateWebhook(testWebhookName, ca, "Fail", k8s.ValidatingHook)

			whmut := &admissionv1.MutatingWebhookConfiguration{}
			err = getDynamic(dyn, testWebhookName, whmut,
				gvkWithType(gvk, k8s.MutatingHook))
			if err != nil {
				t.Error(err)
			}

			whval := &admissionv1.ValidatingWebhookConfiguration{}
			err = getDynamic(dyn, testWebhookName, whval,
				gvkWithType(gvk, k8s.ValidatingHook))
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(whmut.Webhooks[0].ClientConfig.CABundle, ca) {
				t.Error("Ca retrieved from first mutating webhook configuration does not match")
			}
			if !bytes.Equal(whmut.Webhooks[1].ClientConfig.CABundle, ca) {
				t.Error("Ca retrieved from second mutating webhook configuration does not match")
			}
			if !bytes.Equal(whval.Webhooks[0].ClientConfig.CABundle, ca) {
				t.Error("Ca retrieved from first validating webhook configuration does not match")
			}
			if !bytes.Equal(whval.Webhooks[1].ClientConfig.CABundle, ca) {
				t.Error("Ca retrieved from second validating webhook configuration does not match")
			}
			if whmut.Webhooks[0].FailurePolicy == nil {
				t.Errorf("Expected first mutating webhook failure policy to be set to %s", fail)
			}
			if whmut.Webhooks[1].FailurePolicy == nil {
				t.Errorf("Expected second mutating webhook failure policy to be set to %s", fail)
			}
			if whval.Webhooks[0].FailurePolicy == nil {
				t.Errorf("Expected first validating webhook failure policy to be set to %s", fail)
			}
			if whval.Webhooks[1].FailurePolicy == nil {
				t.Errorf("Expected second validating webhook failure policy to be set to %s", fail)
			}
		})
	}
}

func TestAccessingNonExistantWebhook(t *testing.T) {
	k := k8s.NewFake(clients())
	ok := k.UpdateWebhook(testWebhookName, []byte{}, "Fail", k8s.MutatingHook)
	if ok {
		t.Error("Should not succeed when webhook does not exist")
	}
}

var otherConfigs = []*admissionv1.MutatingWebhookConfiguration{
	{
		ObjectMeta: metav1.ObjectMeta{Name: testWebhookName},
		// No webhook configurations
	},
	{
		// Full configuration
		ObjectMeta: metav1.ObjectMeta{Name: testWebhookName},
		Webhooks: []admissionv1.MutatingWebhook{
			{
				Name: "",
				ClientConfig: admissionv1.WebhookClientConfig{
					CABundle: []byte("junk"),
				},
				FailurePolicy: &ignore,
			},
		},
	},
}

func TestResourceStrangeConfigurationShouldSucceed(t *testing.T) {
	ca, _, _ := genSecretData()
	for _, config := range otherConfigs {
		_, dyn := clients()
		k := k8s.NewFake(nil, dyn)
		gvk := schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: "v1"}

		err := createDynamic(dyn, config, gvkWithType(gvk, k8s.MutatingHook))
		if err != nil {
			t.Error(err)
		}

		ok := k.UpdateWebhook(testWebhookName, ca, "", k8s.MutatingHook)
		if !ok {
			t.Error("Expected ok")
		}
	}
}

func createDynamic(dyn dynamic.Interface, data interface{}, gvk schema.GroupVersionResource) error {
	var err error
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(data)
	if err != nil {
		return err
	}
	uns := &unstructured.Unstructured{Object: obj}
	_, err = dyn.Resource(gvk).Create(context.Background(), uns, metav1.CreateOptions{})
	return err
}

func getDynamic(dyn dynamic.Interface, name string, data interface{}, gvk schema.GroupVersionResource) error {
	resp, err := dyn.Resource(gvk).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), data); err != nil {
		return err
	}
	return nil
}

func gvkWithType(gvk schema.GroupVersionResource, kind k8s.WebhookType) schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: gvk.Group, Version: gvk.Version, Resource: string(kind)}

}
