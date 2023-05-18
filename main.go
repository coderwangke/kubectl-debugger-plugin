package main

import (
	"fmt"
	"github.com/coderwangke/kubectl-debugger-plugin/src/k8s"
	"github.com/coderwangke/kubectl-debugger-plugin/src/plugin"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	podName   string
	namespace string
	rm        bool
)

var rootCmd = &cobra.Command{
	Use: "kubectl-debugger",
	Short: "A kubectl plugin to create debugger pods",
}

func main() {
	var podCmd = &cobra.Command{
		Use:   "pod <pod-name>",
		Short: "A kubectl plugin to create debug pods",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			podName = args[0]
			run()
		},
	}

	podCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace of the debug pod")
	podCmd.Flags().BoolVarP(&rm, "rm", "r", true, "Remove the debug pod after it exits")
	rootCmd.AddCommand(podCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() {
	var err error
	client, err := k8s.NewKubernetesClient(".kube/config")
	if err != nil {
		log.Fatalf("New k8s Client failed, error: %v", err)
	}


	nodeName, err := client.GetNode(podName, namespace)
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
		log.Fatal(plugin.SpawnDebuggerPodOnSuperNode(client, podName, namespace, rm))
	} else if nodeType == "normal" {
		fmt.Println("普通节点")
		log.Fatal(plugin.SpawnDebuggerPodOnNormalNode(client, nodeName, rm))
	}
}
