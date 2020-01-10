package combiner

import (
	"crypto/sha1"
	"encoding/json"
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
	idLabel        = labelNamespace + "/id"
	orderLabel     = labelNamespace + "/order"
	fileNameLabel  = labelNamespace + "/filename"
	partialItemKey = "partial-item"
)

// Combiner
type Combiner struct {
	megaConfigMapPath string
	shareDir          string
	k8s               *kubernetes.Clientset
}

// Run
func (c *Combiner) Run() error {
	megaConfig, err := decodeConfigMap(c.megaConfigMapPath)
	if err != nil {
		return err
	}
	mapID, ok := megaConfig.GetLabels()[idLabel]
	if !ok {
		return errors.New(orderLabel + " is not found in megaconfigmap")
	}
	configmaps, err := c.k8s.CoreV1().ConfigMaps("").List(
		metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", idLabel, mapID)})
	if err != nil {
		return err
	}
	tempFileName, err := c.WriteTemp(configmaps)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(tempFileName)
	if err != nil {
		return err
	}
	if mapID != MapID(data, megaConfig.Namespace, megaConfig.Name) {
		return errors.New("checksum is not matched")
	}
	fileName, ok := megaConfig.Labels[fileNameLabel]
	if !ok {
		return errors.New(fileNameLabel + " is not found in megaconfigmap")
	}
	return os.Rename(tempFileName, filepath.Join(c.shareDir, fileName))
}

// Write writes data from ConfigMap list
func (c *Combiner) WriteTemp(configmaps *corev1.ConfigMapList) (string, error) {
	tmp, err := ioutil.TempFile("", "megaconfigmap")
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	contents := make([]string, len(configmaps.Items))
	for _, cm := range configmaps.Items {
		orderingStr, ok := cm.Labels[orderLabel]
		if !ok {
			return "", fmt.Errorf("%s is not found in configmap %s/%s", orderLabel, cm.GetNamespace(), cm.GetName())
		}
		ordering, err := strconv.Atoi(orderingStr)
		if err != nil {
			return "", err
		}
		partial, ok := cm.Data[partialItemKey]
		if !ok {
			return "", fmt.Errorf("partial-item is not found in configmap %s/%s", cm.GetNamespace(), cm.GetName())
		}
		contents[ordering] = partial
	}

	for _, content := range contents {
		_, err = tmp.WriteString(content)
		if err != nil {
			return "", err
		}
	}
	return tmp.Name(), nil
}

// NewCombiner creates a Combiner instance
func NewCombiner(megaConfigMapPath, shareDir string) (*Combiner, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Combiner{
		megaConfigMapPath: megaConfigMapPath,
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

func decodeConfigMap(filename string) (*corev1.ConfigMap, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	configmap := &corev1.ConfigMap{}
	err = json.NewDecoder(f).Decode(configmap)
	if err != nil {
		return nil, err
	}
	return configmap, nil
}
