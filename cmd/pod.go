package cmd

import (
	"fmt"
	"github.com/coderwangke/kubectl-debugger-plugin/src/k8s"
	"github.com/coderwangke/kubectl-debugger-plugin/src/plugin"
	"github.com/spf13/cobra"
	"log"
)

type PodCmdOptions struct {
	PodName string
	Namespace string
	Rm bool
}

func NewPodCmd() *cobra.Command {
	var opt = &PodCmdOptions{}
	var podCmd = &cobra.Command{
		Use:   "pod <pod-name>",
		Short: "A kubectl plugin to create debug pods",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opt.PodName = args[0]
			opt.run()
		},
	}

	podCmd.Flags().StringVarP(&opt.Namespace, "namespace", "n", "default", "Namespace of the debug pod")
	podCmd.Flags().BoolVarP(&opt.Rm, "rm", "r", true, "Remove the debug pod after it exits")

	return podCmd
}

func (o *PodCmdOptions) run() {
	var err error
	client, err := k8s.NewKubernetesClient(".kube/config")
	if err != nil {
		log.Fatalf("New k8s Client failed, error: %v", err)
	}


	nodeName, err := client.GetNode(o.PodName, o.Namespace)
	if err != nil {
		log.Fatalf("GetNode failed, error: %v", err)
	}

	nodeType, err := client.GetNodeType(nodeName)
	if err != nil {
		log.Fatalf("GetNode failed, error: %v", err)
	}

	if nodeType == "super" {
		fmt.Println("超级节点")
		// check
		//running, err := client.IsPodRunning(podName, namespace)
		//if !running ||  err != nil {
		//	log.Fatal("Pod status is abnormal. Please check using kubectl.")
		//}
		log.Fatal(plugin.SpawnDebuggerPodOnSuperNode(client, o.PodName, o.Namespace, o.Rm))
	} else if nodeType == "normal" {
		fmt.Println("普通节点")
		log.Fatal(plugin.SpawnDebuggerPodOnNormalNode(client, nodeName, o.Rm))
	}
}