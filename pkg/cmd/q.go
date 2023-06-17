package cmd

import (
	"context"
	"fmt"
	"regexp"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SubCommand uint8
type SubCommandHandler func(*QueryOptions) error

const (
	Echo SubCommand = iota
	PortForward
)

var (
	queryExample = `
	TODO
	`
	subCommandStringMap = map[string]SubCommand{
		"echo": Echo,
		"pf":   PortForward,
	}
	subCommandHandlerMap = map[SubCommand]SubCommandHandler{
		Echo:        echo,
		PortForward: portForward,
	}
)

type portForwardOptions struct {
	containter    string
	containerPort int32
	hostPort      int32
}

type QueryOptions struct {
	configFlags *genericclioptions.ConfigFlags
	client      *kubernetes.Clientset

	selectedPod string

	handler SubCommandHandler

	podRegex  string
	namespace string

	portForwardOptions portForwardOptions

	genericclioptions.IOStreams
}

func NewQueryOptions(streams genericclioptions.IOStreams) *QueryOptions {
	return &QueryOptions{
		configFlags: genericclioptions.NewConfigFlags(true),

		IOStreams: streams,
	}
}

// NewCmdQuery provides a cobra command wrapping QueryOptions
func NewCmdQuery(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewQueryOptions(streams)

	cmd := &cobra.Command{
		Use:          "q [pf] [flags]",
		Short:        "Queries for a pod and executes the given command",
		Example:      fmt.Sprintf(queryExample, "kubectl"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&o.podRegex, "pod-regex", "r", "", "The regex to match the pod name")

	cmd.Flags().StringVarP(&o.portForwardOptions.containter, "container", "c", "", "Container name")
	cmd.Flags().Int32VarP(&o.portForwardOptions.containerPort, "container-port", "p", 0, "Container port to forward to the host")
	cmd.Flags().Int32VarP(&o.portForwardOptions.hostPort, "host-port", "P", 0, "Host port to forward to the pod")
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

func (o *QueryOptions) Complete(cmd *cobra.Command, args []string) error {
	var err error
	if o.namespace, _, err = o.configFlags.ToRawKubeConfigLoader().Namespace(); err != nil {
		return err
	}

	if specifiedNamespace, err := cmd.Flags().GetString("namespace"); err != nil {
		return err
	} else if specifiedNamespace != "" {
		o.namespace = specifiedNamespace
	}

	if o.podRegex, err = cmd.Flags().GetString("pod-regex"); err != nil {
		return err
	}

	if o.portForwardOptions.containerPort, err = cmd.Flags().GetInt32("container-port"); err != nil {
		return err
	}

	if o.portForwardOptions.hostPort, err = cmd.Flags().GetInt32("host-port"); err != nil {
		return err
	}

	if restConfig, err := o.configFlags.ToRESTConfig(); err != nil {
		return err
	} else {
		o.client, err = kubernetes.NewForConfig(restConfig)
		if err != nil {
			return err
		}
	}

	var (
		subCommand = SubCommand(0)
		ok         bool
	)

	if len(args) > 0 {
		if subCommand, ok = subCommandStringMap[args[0]]; !ok {
			return fmt.Errorf("Invalid subcommand %s", args[0])
		}
	}

	o.handler = subCommandHandlerMap[subCommand]

	return nil
}

func (o *QueryOptions) Validate() error {
	return nil
}

func (o *QueryOptions) Run() error {
	pods, err := o.client.CoreV1().Pods(o.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	matchingPods := []string{}

	for _, pod := range pods.Items {
		if match, err := regexp.MatchString(o.podRegex, pod.Name); err != nil {
			return err
		} else if match {
			matchingPods = append(matchingPods, pod.Name)
		}
	}

	if len(matchingPods) == 0 {
		return fmt.Errorf("No pods found matching regex %s in namespace %s", o.podRegex, o.namespace)
	} else if len(matchingPods) == 1 {
		o.selectedPod = matchingPods[0]
	} else {
		prompt := promptui.Select{
			Label: "Select Pod",
			Items: matchingPods,
		}

		_, o.selectedPod, err = prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}
	}

	return o.handler(o)
}

func parseSubCommandArg(s string) (SubCommand, bool) {
	subCommand, ok := subCommandStringMap[s]
	return subCommand, ok
}
