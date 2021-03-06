package main

import (
	"os"

	"github.com/dulltz/megaconfigmap/pkg/megaconfigmap"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-megaconfigmap", pflag.ExitOnError)
	pflag.CommandLine = flags

	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	root, err := megaconfigmap.NewCmdMegaConfigMap(streams)
	if err != nil {
		panic(err)
	}
	root.AddCommand(megaconfigmap.NewCmdCreate(streams))
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
