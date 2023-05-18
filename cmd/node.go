package cmd

import (
	"fmt"
	"github.com/coderwangke/kubectl-debugger-plugin/src/k8s"
	"github.com/coderwangke/kubectl-debugger-plugin/src/plugin"
	"github.com/spf13/cobra"
	"log"
)

type NodeCmdOptions struct {
	NodeName string
	Rm bool
}

func NewNodeCmd() *cobra.Command {
	var opt = &NodeCmdOptions{}
	var nodeCmd = &cobra.Command{
		Use:   "node <node-name>",
		Short: "A kubectl plugin to create debug pods on node",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opt.NodeName = args[0]
			opt.run()
		},
	}

	nodeCmd.Flags().BoolVarP(&opt.Rm, "rm", "r", true, "Remove the debug pod after it exits")

	return nodeCmd
}

func (o *NodeCmdOptions) run() {
	var err error
	client, err := k8s.NewKubernetesClient(".kube/config")
	if err != nil {
		log.Fatalf("New k8s Client failed, error: %v", err)
	}


	nodeType, err := client.GetNodeType(o.NodeName)
	if err != nil {
		log.Fatalf("GetNode failed, error: %v", err)
	}

	if nodeType == "super" {
		fmt.Println("超级节点")
		log.Fatal("Don't support SuperNode!!!")
	} else if nodeType == "normal" {
		fmt.Println("普通节点")
		log.Fatal(plugin.SpawnDebuggerPodOnNormalNode(client, o.NodeName, o.Rm))
	}
}
