package cmd

import (
	. "github.com/golang/mock/gomock"
	"github.com/jet/kube-webhook-certgen/k8s"
	mock_k8s "github.com/jet/kube-webhook-certgen/mock"
	"testing"
)

func TestPostOperation(t *testing.T) {
	ctrl := NewController(t)
	m := mock_k8s.NewMockK8s(ctrl)
	defer ctrl.Finish()

	goodSecret := "goodSecret"
	notExists := "notExists"

	cfg.namespace = "namespace"
	cfg.caName = "ca"
	cfg.patchValidating = []string{"va", "vb"}
	cfg.patchMutating = []string{"ma", "mb"}
	ca := []byte("cabytes")

	// Good configs
	m.EXPECT().
		GetCaFromSecret(Eq(goodSecret), Eq(cfg.namespace), Eq(cfg.caName)).
		Return(ca, true).AnyTimes()
	for _, v := range cfg.patchValidating {
		m.EXPECT().
			UpdateWebhook(Eq(v), Eq(ca), Eq(cfg.patchFailurePolicy), Eq(k8s.ValidatingHook)).
			Return(true).AnyTimes()
	}
	for _, v := range cfg.patchMutating {
		m.EXPECT().
			UpdateWebhook(Eq(v), Eq(ca), Eq(cfg.patchFailurePolicy), Eq(k8s.MutatingHook)).
			Return(true).AnyTimes()
	}

	// Bad configs
	m.EXPECT().
		UpdateWebhook(Eq(notExists), Any(), Any(), Any()).
		Return(false).AnyTimes()
	m.EXPECT().
		GetCaFromSecret(Eq(notExists), Any(), Any()).
		Return(nil, false).AnyTimes()

	// Check normal behaviour
	cfg.secretName = goodSecret
	exitCode := post(m)
	if exitCode != 0 {
		t.Fatal("Expected 0 exit code")
	}

	// No secret should exit with 1
	cfg.secretName = notExists
	exitCode = post(m)
	if exitCode != 1 {
		t.Fatal("Expected 1 exit code")
	}

	// Single hook update should exit with 1
	cfg.secretName = goodSecret
	cfg.patchMutating = append(cfg.patchMutating, notExists)
	exitCode = post(m)
	if exitCode != 1 {
		t.Fatal("Expected 1 exit code")
	}
}
