package megaconfigmap

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"

	"github.com/dulltz/megaconfigmap/pkg/combiner"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	blockSize = int64(400 * 1024)
)

var (
	createExample = `
	# create megaconfigmap from file
	%[1]s megaconfigmap create my-config --from-file=<file-name>
`
)

// CreateOptions provides information required to create megaconfigmap
type CreateOptions struct {
	Options

	megaConfigMapName string
	namespace         string
	outputFile        string
	sourceFile        string
}

// Create MegaConfigMap
func (o *CreateOptions) Create() error {
	if len(o.sourceFile) > 0 {
		return o.createFromFile()
	}
	return errors.New("currently, --from-file is required")
}

func (o *CreateOptions) createFromFile() error {
	f, err := os.Open(o.sourceFile)
	if err != nil {
		return err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return errors.New("--from-file not support directory")
	}
	o.outputFile = stat.Name()
	checkSum, err := o.getCheckSum()
	if err != nil {
		return err
	}

	fmt.Printf("creating megaconfigmap %s...\n", o.megaConfigMapName)
	master, err := o.createMasterConfigMap(checkSum)
	if err != nil {
		return err
	}

	numPartials := int64(math.Ceil(float64(stat.Size()) / float64(blockSize)))
	fmt.Printf("creating %d partial configmaps from %s...\n", numPartials, o.megaConfigMapName)
	for i := int64(0); i < numPartials; i++ {
		err := func() error {
			buf := make([]byte, blockSize)
			n, err := f.ReadAt(buf, i*blockSize)
			if err != io.EOF && err != nil {
				defer o.k8s.CoreV1().ConfigMaps(o.namespace).Delete(master.Name, &metav1.DeleteOptions{})
				return err
			}
			err = o.createPartialConfigMap(buf[:n], i, checkSum, master)
			if err != nil {
				defer o.k8s.CoreV1().ConfigMaps(o.namespace).Delete(master.Name, &metav1.DeleteOptions{})
				return err
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *CreateOptions) getCheckSum() (string, error) {
	data, err := ioutil.ReadFile(o.sourceFile)
	if err != nil {
		return "", err
	}
	return combiner.MapID(data, o.namespace, o.megaConfigMapName), nil
}

func (o *CreateOptions) createPartialConfigMap(data []byte, order int64, sum string, master *v1.ConfigMap) error {
	_, err := o.k8s.CoreV1().ConfigMaps(o.namespace).Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: o.namespace,
			Name:      fmt.Sprintf("%s-%d", o.megaConfigMapName, order),
			Labels: map[string]string{
				combiner.IDLabel:       sum,
				combiner.OrderLabel:    fmt.Sprintf("%d", order),
				combiner.FileNameLabel: o.outputFile,
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "v1",
				Kind:       "ConfigMap",
				Name:       master.Name,
				UID:        master.UID,
			}},
		},
		Data: map[string]string{combiner.PartialItemKey: string(data)},
	})
	return err
}

func (o *CreateOptions) createMasterConfigMap(sum string) (*v1.ConfigMap, error) {
	_, err := o.k8s.CoreV1().ConfigMaps(o.namespace).Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: o.namespace,
			Name:      o.megaConfigMapName,
			Labels: map[string]string{
				combiner.IDLabel:       sum,
				combiner.FileNameLabel: o.outputFile,
				combiner.MasterLabel:   "true",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return o.k8s.CoreV1().ConfigMaps(o.namespace).Get(o.megaConfigMapName, metav1.GetOptions{})
}

// NewMegaConfigMapOptions provides an instance of MegaConfigMapOptions with default values
func NewCreateOptions(o *Options) (*CreateOptions, error) {
	ns := "default"
	if len(*o.configFlags.Namespace) > 0 {
		ns = *o.configFlags.Namespace
	}
	return &CreateOptions{
		Options:   *o,
		namespace: ns,
	}, nil
}

func newCmdCreate(rootOptions *Options) *cobra.Command {
	o, err := NewCreateOptions(rootOptions)
	if err != nil {
		return nil
	}
	cmd := &cobra.Command{
		Use:          "create my-config --from-file [flags]",
		Short:        "create megaconfigmap",
		Example:      fmt.Sprintf(createExample, "kubectl"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("exactly one NAME is required, got %d", len(args))
			}
			o.megaConfigMapName = args[0]
			return o.Create()
		},
	}
	cmd.Flags().StringVar(&o.sourceFile, "from-file", o.sourceFile, "Specify filename to be stored in megaconfigmap.")
	return cmd
}
