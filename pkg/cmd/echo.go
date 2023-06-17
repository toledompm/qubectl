package cmd

import "fmt"

func echo(o *QueryOptions) error {
	fmt.Printf("%s", o.selectedPod)

	return nil
}
