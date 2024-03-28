package cmd

import (
	"github.com/coderwangke/kubectl-debugger-plugin/src/k8s"
	"github.com/coderwangke/kubectl-debugger-plugin/src/plugin"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

type PodCmdOptions struct {
	PodName   string
	Namespace string
	Image     string
	Command   []string
	Rm        bool
}

func newPodCmd() *cobra.Command {
	var opt = &PodCmdOptions{
		Command: []string{"/bin/sh"},
	}
	var podCmd = &cobra.Command{
		Use:   "pod <pod-name> [flags] -- COMMAND",
		Short: "A kubectl plugin to create debug pods",
		Run: func(cmd *cobra.Command, args []string) {
			argsLenAtDash := cmd.ArgsLenAtDash()
			if argsLenAtDash > -1 {
				opt.Command = args[argsLenAtDash:]
			}

			opt.PodName = args[0]
			opt.run()
		},
	}

	podCmd.Flags().StringVarP(&opt.Namespace, "namespace", "n", "default", "Namespace of the debug pod")
	podCmd.Flags().StringVarP(&opt.Image, "image", "i", "busybox", "Image of the debug pod")
	podCmd.Flags().BoolVarP(&opt.Rm, "rm", "r", true, "Remove the debug pod after it exits")

	return podCmd
}

func (o *PodCmdOptions) run() {
	var err error
	client, err := k8s.NewKubernetesClient(kubeConfig)
	if err != nil {
		klog.Fatalf("New k8s Client failed, error: %v", err)
	}

	nodeName, err := client.GetNode(o.PodName, o.Namespace)
	if err != nil {
		klog.Fatalf("GetNode failed, error: %v", err)
	}

	nodeType, err := client.GetNodeType(nodeName)
	if err != nil {
		klog.Fatalf("GetNode failed, error: %v", err)
	}

	helper := &plugin.DebuggerPodHelper{
		PodName:   o.PodName,
		Namespace: o.Namespace,
		Image:     o.Image,
		Command:   o.Command,
		NodeName:  nodeName,
		Rm:        o.Rm,
	}

	switch nodeType {
	case k8s.NormalNodeType:
		if err = plugin.SpawnDebuggerPodOnNormalNode(client, helper); err != nil {
			klog.Fatal(err)
		}

	case k8s.SuperNodeType:
		klog.Info("超级节点")
		if err = plugin.SpawnDebuggerPodOnSuperNode(client, helper); err != nil {
			klog.Fatal(err)
		}
	}
}
