package kubectl_wrapper

import (
	"fmt"
	"os/exec"
	"strings"
)

func Exec(args []string) error {
	fmt.Printf("kubectl %s\n", strings.Join(args, " "))

	cmd := exec.Command(
		"kubectl",
		args...,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	cmd.Stderr = cmd.Stdout

	if err = cmd.Start(); err != nil {
		return err
	}

	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}

	return cmd.Wait()
}
