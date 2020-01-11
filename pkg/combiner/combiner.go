package combiner

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	labelNamespace = "megaconfigmap.io"
	// IDLabel
	IDLabel = labelNamespace + "/id"
	// OrderLabel
	OrderLabel = labelNamespace + "/order"
	// FileNameLabel
	FileNameLabel = labelNamespace + "/filename"
	// MasterLabel
	MasterLabel = labelNamespace + "/master"
	// PartialItemKet is the configmap key to store partial data
	PartialItemKey = "partial-item"
)

// Combiner
type Combiner struct {
	megaConfigMapName string
	shareDir          string
	k8s               *kubernetes.Clientset
}

// Run
func (c *Combiner) Run() error {
	namespaceBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	namespace := string(namespaceBytes)
	if err != nil {
		return fmt.Errorf("failed to get current context namespace; %w", err)
	}

	megaConfig, err := c.k8s.CoreV1().ConfigMaps(namespace).Get(c.megaConfigMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get megaconfigmap %s; %w", c.megaConfigMapName, err)
	}
	mapID, ok := megaConfig.GetLabels()[IDLabel]
	if !ok {
		return errors.New(OrderLabel + " is not found in megaconfigmap " + c.megaConfigMapName)
	}
	configmaps, err := c.k8s.CoreV1().ConfigMaps(namespace).List(
		metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s,%s!=true", IDLabel, mapID, MasterLabel)})
	if err != nil {
		return fmt.Errorf("failed to list configmaps; %w", err)
	}
	tempFileName, err := c.WriteTemp(configmaps)
	if err != nil {
		return fmt.Errorf("failed to write to tempfile; %w", err)
	}
	//data, err := ioutil.ReadFile(tempFileName)
	//if err != nil {
	//	return err
	//}
	//if mapID != MapID(data, megaConfig.Namespace, megaConfig.Name) {
	//	return errors.New("checksum is not matched")
	//}
	fileName, ok := megaConfig.Labels[FileNameLabel]
	if !ok {
		return errors.New(FileNameLabel + " is not found in megaconfigmap")
	}
	return os.Rename(tempFileName, filepath.Join(c.shareDir, fileName))
}

// Write writes data from ConfigMap list
func (c *Combiner) WriteTemp(configmaps *corev1.ConfigMapList) (string, error) {
	tmp, err := ioutil.TempFile(c.shareDir, "megaconfigmap")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	contents, err := c.sortContents(configmaps)
	if err != nil {
		return "", err
	}

	for _, partialContent := range contents {
		_, err = tmp.WriteString(partialContent)
		if err != nil {
			return "", err
		}
	}
	return tmp.Name(), nil
}

func (c *Combiner) sortContents(configmaps *corev1.ConfigMapList) ([]string, error) {
	contents := make([]string, len(configmaps.Items))
	for _, cm := range configmaps.Items {
		orderingStr, ok := cm.Labels[OrderLabel]
		if !ok {
			return nil, fmt.Errorf("%s is not found in configmap %s/%s", OrderLabel, cm.GetNamespace(), cm.GetName())
		}
		ordering, err := strconv.Atoi(orderingStr)
		if err != nil {
			return nil, err
		}
		partial, ok := cm.Data[PartialItemKey]
		if !ok {
			return nil, fmt.Errorf("partial-item is not found in configmap %s/%s", cm.GetNamespace(), cm.GetName())
		}
		contents[ordering] = partial
	}
	return contents, nil
}

// NewCombiner creates a Combiner instance
func NewCombiner(megaConfigMapName, shareDir string) (*Combiner, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Combiner{
		megaConfigMapName: megaConfigMapName,
		shareDir:          shareDir,
		k8s:               clientset,
	}, nil
}

// MapID returns a hash string
func MapID(data []byte, namespace, name string) string {
	h := sha1.New()
	h.Write(data)
	h.Write([]byte(namespace))
	h.Write([]byte(name))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
