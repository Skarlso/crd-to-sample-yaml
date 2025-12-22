package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
)

// KubeHandler contains data for a kubernetes resource.
type KubeHandler struct {
	crd             string
	group           string
	resourceGroup   string
	resourceVersion string
	resource        string
}

// CRDs returns schemas found in a cluster that have been installed.
func (h *KubeHandler) CRDs() ([]*pkg.SchemaType, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building config from flags: %w", err)
	}

	cliSet, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic client: %w", err)
	}

	ctx := context.Background()

	result, err := cliSet.Resource(schema.GroupVersionResource{
		Group:    h.resourceGroup,
		Version:  h.resourceVersion,
		Resource: h.resource,
	}).Get(ctx, h.crd, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting CRD: %w", err)
	}

	schemaType, err := pkg.ExtractSchemaType(result)
	if err != nil {
		return nil, fmt.Errorf("failed to extract schema type: %w", err)
	}

	if schemaType == nil {
		return nil, nil
	}

	if h.group != "" {
		schemaType.Rendering = pkg.Rendering{Group: h.group}
	}

	return []*pkg.SchemaType{schemaType}, nil
}
