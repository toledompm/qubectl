package cmd

import (
	"fmt"

	k "github.com/toledompm/qubectl/pkg/kubectl-wrapper"
)

func custom(o *QueryOptions) error {

	if o.forwardedKubectlArgs == nil {
		// Print pod name if no args are passed
		fmt.Printf("%s\n", o.selectedPod)
		return nil
	}

	args := append(o.forwardedKubectlArgs, o.selectedPod, "-n", o.namespace)

	return k.Exec(args)
}
