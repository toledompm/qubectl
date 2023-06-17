package cmd

import (
	"context"
	"fmt"
	"net"
	"os/exec"

	"github.com/manifoldco/promptui"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func portForward(o *QueryOptions) error {
	pod, err := o.client.CoreV1().Pods(o.namespace).Get(context.TODO(), o.selectedPod, metav1.GetOptions{})
	if err != nil {
		return err
	}

	containers := pod.Spec.Containers

	if len(containers) == 0 {
		return fmt.Errorf("no containers found in pod %s, namespace: %s", o.selectedPod, o.namespace)
	}

	var selectedContainer corev1.Container

	if o.portForwardOptions.containter != "" {
		for _, container := range containers {
			if container.Name == o.portForwardOptions.containter {
				selectedContainer = container
				break
			}
		}
	} else if len(containers) == 1 {
		selectedContainer = containers[0]
	} else {
		prompt := promptui.Select{
			Label: "Select Container",
			Items: containers,
		}

		index, _, err := prompt.Run()
		if err != nil {
			fmt.Printf("Container selection prompt failed %v\n", err)
			return err
		}

		selectedContainer = containers[index]
	}

	ports := selectedContainer.Ports

	if len(ports) == 0 {
		return fmt.Errorf("no ports found in container %s, pod %s, namespace: %s", selectedContainer.Name, o.selectedPod, o.namespace)
	}

	var selectedPort corev1.ContainerPort

	if o.portForwardOptions.containerPort != 0 {
		for _, port := range ports {
			if port.ContainerPort == o.portForwardOptions.containerPort {
				selectedPort = port
				break
			}
		}
	} else if len(ports) == 1 {
		selectedPort = ports[0]
	} else {
		prompt := promptui.Select{
			Label: "Select Port",
			Items: ports,
		}

		index, _, err := prompt.Run()
		if err != nil {
			fmt.Printf("Port selection prompt failed %v\n", err)
			return err
		}

		selectedPort = ports[index]
	}

	var selectedHostPort int32

	if o.portForwardOptions.hostPort != 0 {
		selectedHostPort = o.portForwardOptions.hostPort
	} else {
		selectedHostPort = selectedPort.ContainerPort
		for checkPortInUse(selectedHostPort) {
			selectedHostPort++
		}
	}

	exec.Command("kubectl", "port-forward", "-n", o.namespace, o.selectedPod, fmt.Sprintf("%d:%d", selectedHostPort, selectedPort.ContainerPort)).Start()

	return nil
}

func checkPortInUse(port int32) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return true
	}

	_ = ln.Close()
	return false
}
