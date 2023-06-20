package cmd

import (
	"fmt"

	k "github.com/toledompm/qubectl/pkg/kubectl-wrapper"
)

var REPLACEMENT_MARKER = "%%"

func custom(o *QueryOptions) error {

	if o.forwardedKubectlArgs == nil {
		// Print pod name and namespace if no args are passed
		fmt.Printf("%s -n %s\n", o.selectedPod, o.selectedPodNamespace)
		return nil
	}

	args := []string{"-n", o.selectedPodNamespace}
	replacementMarkerFound := false

	for _, arg := range o.forwardedKubectlArgs {
		if arg == REPLACEMENT_MARKER {
			arg = o.selectedPod
			replacementMarkerFound = true
		}
		args = append(args, arg)
	}

	if !replacementMarkerFound {
		args = append(args, o.selectedPod)
	}

	return k.Exec(args)
}
