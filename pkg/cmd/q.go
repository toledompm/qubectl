package cmd

import (
	"context"
	"fmt"
	"regexp"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SubCommand uint8
type SubCommandHandler func(*QueryOptions) error

const (
	Custom SubCommand = iota
	PortForward
)

var (
	queryExample = `
	// Port forward to a pod matching the regex
	kubectl q port-forward -r my-pod-regex

	// Port forward, with address
	kubectl q port-forward -r my-pod-regex -- --address 0.0.0.0

	// Folow logs of a pod matching the regex
	kubectl q custom -r my-pod-regex -- logs -f

	// Get pod yaml (custom is the default subcommand)
	kubectl q -r my-pod-regex -- get pod -o yaml
	`
	subCommandStringMap = map[string]SubCommand{
		"custom":       Custom,
		"port-forward": PortForward,
	}
	subCommandHandlerMap = map[SubCommand]SubCommandHandler{
		Custom:      custom,
		PortForward: portForward,
	}
)

type namespaceOpts struct {
	name string
	flag string
}

type QueryOptions struct {
	configFlags *genericclioptions.ConfigFlags
	client      *kubernetes.Clientset

	selectedPod          string
	selectedPodNamespace string

	handler              SubCommandHandler
	forwardedKubectlArgs []string

	podRegex       string
	allNamespaces  bool
	queryNamespace string

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
		Use:          "q [command] [flags] -- [kubectl args]",
		Short:        "Queries for a pod and executes the given command",
		Long:         "q serves as a wrapper to other kubectl commands. It queries for a pod matching the given regex and executes the given command with the pod name appended. Some commands, such as port-forward, are handled by q itself before being forwarded to kubectl.",
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
	cmd.Flags().BoolVarP(&o.allNamespaces, "all-namespaces", "A", false, "Query all namespaces")

	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

func (o *QueryOptions) Complete(cmd *cobra.Command, args []string) error {
	var err error

	if o.queryNamespace, err = cmd.Flags().GetString("namespace"); err != nil {
		return err
	}

	if o.queryNamespace != "" && o.allNamespaces {
		return fmt.Errorf("cannot specify both namespace and all-namespaces")
	}

	if o.queryNamespace == "" && o.allNamespaces {
		o.queryNamespace = metav1.NamespaceAll
	}

	if o.queryNamespace == "" && !o.allNamespaces {
		if o.queryNamespace, _, err = o.configFlags.ToRawKubeConfigLoader().Namespace(); err != nil {
			return err
		}
	}

	if o.podRegex, err = cmd.Flags().GetString("pod-regex"); err != nil {
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

	argsLenAtDash := cmd.Flags().ArgsLenAtDash()

	if len(args) > 0 && argsLenAtDash != 0 {
		if subCommand, ok = subCommandStringMap[args[0]]; !ok {
			return fmt.Errorf("Invalid subcommand %s", args[0])
		}
	}

	if argsLenAtDash >= 0 {
		o.forwardedKubectlArgs = args[argsLenAtDash:]
	}

	o.handler = subCommandHandlerMap[subCommand]

	return nil
}

func (o *QueryOptions) Validate() error {
	return nil
}

func (o *QueryOptions) Run() error {
	pods, err := o.client.CoreV1().Pods(o.queryNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	matchingPods := []corev1.Pod{}

	for _, pod := range pods.Items {
		if match, err := regexp.MatchString(o.podRegex, pod.Name); err != nil {
			return err
		} else if match {
			matchingPods = append(matchingPods, pod)
		}
	}

	matchingPodNames := []string{}
	for _, pod := range matchingPods {
		matchingPodNames = append(matchingPodNames, fmt.Sprintf("%s - %s", pod.Name, pod.Namespace))
	}

	if len(matchingPodNames) == 0 {
		return fmt.Errorf("No pods found matching regex %s in namespace %s", o.podRegex, o.queryNamespace)
	} else if len(matchingPodNames) == 1 {
		o.selectedPod = matchingPods[0].Name
		o.selectedPodNamespace = matchingPods[0].Namespace
	} else {
		prompt := promptui.Select{
			Label: "Select Pod (name - namespace)",
			Items: matchingPodNames,
		}

		index, _, err := prompt.Run()

		o.selectedPod = matchingPods[index].Name
		o.selectedPodNamespace = matchingPods[index].Namespace

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
