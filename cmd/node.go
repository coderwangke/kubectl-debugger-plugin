package cmd

import (
	"github.com/coderwangke/kubectl-debugger-plugin/src/k8s"
	"github.com/coderwangke/kubectl-debugger-plugin/src/plugin"
	"github.com/spf13/cobra"
	"log"
)

type NodeCmdOptions struct {
	NodeName string
	Image    string
	Rm       bool
}

func newNodeCmd() *cobra.Command {
	var opt = &NodeCmdOptions{}
	var nodeCmd = &cobra.Command{
		Use:   "node <node-name>",
		Short: "A kubectl plugin to create debug pods on node",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opt.NodeName = args[0]
			opt.run()
		},
	}

	nodeCmd.Flags().BoolVarP(&opt.Rm, "rm", "r", true, "Remove the debug pod after it exits")
	nodeCmd.Flags().StringVarP(&opt.Image, "image", "i", "docker.io/library/alpine", "Image of the debug pod")

	return nodeCmd
}

func (o *NodeCmdOptions) run() {
	var err error
	client, err := k8s.NewKubernetesClient(kubeConfig)
	if err != nil {
		log.Fatalf("New k8s Client failed, error: %v", err)
	}

	nodeType, err := client.GetNodeType(o.NodeName)
	if err != nil {
		log.Fatalf("GetNode failed, error: %v", err)
	}

	helper := &plugin.DebuggerPodHelper{
		NodeName: o.NodeName,
		Image:    o.Image,
		Rm:       o.Rm,
	}

	switch nodeType {
	case k8s.SuperNodeType:
		log.Println("kubectl debugger node 子命令不支持超级节点, 请使用 kubectl debugger pod 子命令.")
	case k8s.NormalNodeType:
		log.Fatal(plugin.SpawnDebuggerPodOnNormalNode(client, helper))
	}
}
