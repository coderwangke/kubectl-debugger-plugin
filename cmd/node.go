package cmd

import (
	"github.com/coderwangke/kubectl-debugger-plugin/src/k8s"
	"github.com/coderwangke/kubectl-debugger-plugin/src/plugin"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

type NodeCmdOptions struct {
	NodeName string
	Image    string
	Command  []string
	Rm       bool
}

func newNodeCmd() *cobra.Command {
	var opt = &NodeCmdOptions{
		Command: []string{"/bin/sh"},
	}
	var nodeCmd = &cobra.Command{
		Use:   "node <node-name> [flags] -- COMMAND",
		Short: "A kubectl plugin to create debug pods on node",
		Run: func(cmd *cobra.Command, args []string) {
			argsLenAtDash := cmd.ArgsLenAtDash()
			if argsLenAtDash > -1 {
				opt.Command = args[argsLenAtDash:]
			}

			opt.NodeName = args[0]
			opt.run()
		},
	}

	nodeCmd.Flags().BoolVarP(&opt.Rm, "rm", "r", false, "Remove the debug pod after it exits")
	nodeCmd.Flags().StringVarP(&opt.Image, "image", "i", "docker.io/library/alpine", "Image of the debug pod")

	return nodeCmd
}

func (o *NodeCmdOptions) run() {
	var err error
	client, err := k8s.NewKubernetesClient(kubeConfig)
	if err != nil {
		klog.Fatalf("New k8s Client failed, error: %v", err)
	}

	nodeType, err := client.GetNodeType(o.NodeName)
	if err != nil {
		klog.Fatalf("GetNode failed, error: %v", err)
	}

	helper := &plugin.DebuggerPodHelper{
		NodeName: o.NodeName,
		Image:    o.Image,
		Command:  o.Command,
		Rm:       o.Rm,
	}

	switch nodeType {
	case k8s.SuperNodeType:
		klog.Warning("kubectl debugger node 子命令不支持超级节点, 请使用 kubectl debugger pod 子命令.")
	case k8s.NormalNodeType:
		if err = plugin.SpawnDebuggerPodOnNormalNode(client, helper); err != nil {
			klog.Fatal(err)
		}
	}
}
