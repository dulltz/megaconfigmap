package megaconfigmap

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
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
}

// NewMegaConfigMapOptions provides an instance of MegaConfigMapOptions with default values
func NewMegaConfigMapOptions(streams genericclioptions.IOStreams) (*Options, error) {
	return &Options{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
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
	o.configFlags.AddFlags(cmd.Flags())
	return cmd, nil
}
