package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/toledompm/qubectl/pkg/cmd"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-q", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := cmd.NewCmdQuery(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
