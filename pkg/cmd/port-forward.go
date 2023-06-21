package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manifoldco/promptui"
	k "github.com/toledompm/qubectl/pkg/kubectl-wrapper"
	"github.com/toledompm/qubectl/pkg/prompt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func portForward(o *QueryOptions) error {
	pod, err := o.client.CoreV1().Pods(o.selectedPodNamespace).Get(context.TODO(), o.selectedPod, metav1.GetOptions{})
	if err != nil {
		return err
	}

	containers := pod.Spec.Containers

	if len(containers) == 0 {
		return fmt.Errorf("no containers found in pod %s, namespace: %s", o.selectedPod, o.selectedPodNamespace)
	}

	var selectedContainer corev1.Container

	if len(containers) == 1 {
		selectedContainer = containers[0]
	} else {
		containerItems := []string{}
		for _, container := range containers {
			containerItems = append(containerItems, container.Name)
		}

		prompt := promptui.Select{
			Label: "More than one Container found, select one",
			Items: containerItems,
		}

		index, _, err := prompt.Run()
		if err != nil {
			fmt.Printf("Container selection prompt failed %v\n", err)
			return err
		}

		selectedContainer = containers[index]
	}

	ports := selectedContainer.Ports

	var selectedPorts []corev1.ContainerPort

	if len(ports) == 0 {
		return fmt.Errorf("no ports found in container %s, pod %s, namespace: %s", selectedContainer.Name, o.selectedPod, o.selectedPodNamespace)
	} else if len(ports) == 1 {
		selectedPorts = append(selectedPorts, ports[0])
	} else {
		portItems := []*prompt.Item{}
		for i, port := range ports {
			portItems = append(portItems, &prompt.Item{
				ID:         fmt.Sprintf("%d %s", port.ContainerPort, port.Name),
				Index:      i,
				IsSelected: false,
			})
		}
		if selectedPortItems, err := prompt.MultiSelect("Select ports to forward", portItems, 0); err != nil {
			return err
		} else {
			for _, port := range selectedPortItems {
				selectedPorts = append(selectedPorts, ports[port.Index])
			}
		}
	}

	var portArgs []string

	for _, port := range selectedPorts {
		var hostPort int32

		selectedHostPort, err := prompt.Text(fmt.Sprintf("Enter host port for %d (default: 0)", port.ContainerPort))
		if err != nil {
			return err
		}

		if selectedHostPort != "" {
			parsedSelectedHostPort, err := strconv.ParseInt(selectedHostPort, 10, 32)
			if err != nil {
				return err
			}
			hostPort = int32(parsedSelectedHostPort)
		} else {
			hostPort = 0
		}

		portArgs = append(portArgs, fmt.Sprintf("%d:%d", hostPort, port.ContainerPort))
	}

	args := append([]string{"port-forward", "-n", o.selectedPodNamespace, o.selectedPod}, portArgs...)
	args = append(args, o.forwardedKubectlArgs...)

	return k.Exec(args)
}
