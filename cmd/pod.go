package cmd

import (
	"github.com/coderwangke/kubectl-debugger-plugin/src/k8s"
	"github.com/coderwangke/kubectl-debugger-plugin/src/plugin"
	"github.com/spf13/cobra"
	"log"
)

type PodCmdOptions struct {
	PodName   string
	Namespace string
	Image     string
	Rm        bool
}

func newPodCmd() *cobra.Command {
	var opt = &PodCmdOptions{}
	var podCmd = &cobra.Command{
		Use:   "pod <pod-name>",
		Short: "A kubectl plugin to create debug pods",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
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

	helper := &plugin.DebuggerPodHelper{
		PodName:   o.PodName,
		Namespace: o.Namespace,
		Image:     o.Image,
		NodeName:  nodeName,
		Rm:        o.Rm,
	}

	switch nodeType {
	case k8s.NormalNodeType:
		log.Fatal(plugin.SpawnDebuggerPodOnNormalNode(client, helper))

	case k8s.SuperNodeType:
		log.Println("超级节点")
		// check
		//running, err := client.IsPodRunning(podName, namespace)
		//if !running ||  err != nil {
		//	log.Fatal("Pod status is abnormal. Please check using kubectl.")
		//}
		log.Fatal(plugin.SpawnDebuggerPodOnSuperNode(client, helper))
	}
}
