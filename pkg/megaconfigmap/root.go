package megaconfigmap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	example = `
	# create MegaConfigMap from file
	%[1]s megaconfigmap create --from-file=<file-name>
`
)

// Options provides information required for megaconfigmap
type Options struct {
	configFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams
	k8s *kubernetes.Clientset
}

// NewMegaConfigMapOptions provides an instance of MegaConfigMapOptions with default values
func NewMegaConfigMapOptions(streams genericclioptions.IOStreams) (*Options, error) {
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), "/.kube/config"))
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Options{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
		k8s:         clientset,
	}, nil
}

// NewCmdCreateMegaConfigMap provides a cobra command wrapping MegaConfigMapOptions
func NewCmdMegaConfigMap(streams genericclioptions.IOStreams) (*cobra.Command, error) {
	o, err := NewMegaConfigMapOptions(streams)
	if err != nil {
		return nil, err
	}
	cmd := &cobra.Command{
		Use:     "megaconfigmap [create,get] [flags]",
		Short:   "control megaconfigmap",
		Example: fmt.Sprintf(example, "kubectl"),
		RunE: func(c *cobra.Command, args []string) error {
			return c.Usage()
		},
	}

	cmd.AddCommand(newCmdCreate(o))
	o.configFlags.AddFlags(cmd.Flags())

	return cmd, nil
}
