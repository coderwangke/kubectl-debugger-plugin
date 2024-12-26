package plugin

import (
	"fmt"
	"github.com/coderwangke/kubectl-debugger-plugin/src/k8s"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"time"
)

const defaultRetryTimeout = 1

type DebuggerPodHelper struct {
	PodName   string
	NodeName  string
	Namespace string
	Command   []string
	Rm        bool
	Image     string
}

func SpawnDebuggerPodOnSuperNode(client *k8s.KubernetesClient, helper *DebuggerPodHelper) error {
	var err error
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return client.AnnotatePod(helper.PodName, helper.Namespace, helper.Image)
	})
	if err != nil {
		return err
	}

	err = wait.PollImmediate(5*time.Second, defaultRetryTimeout*time.Minute, func() (bool, error) {
		return client.IsDebuggerContainerRunning(helper.PodName, helper.Namespace)
	})
	if err != nil {
		return err
	}

	fmt.Printf("spawning debugger pod in pod %s success\n", helper.PodName)

	err = client.ExecCommand(helper.PodName, helper.Namespace, "[pod-debugger]debugger", helper.Command)
	// 仅输出报错，依然根据 --rm 参数删除pod
	if err != nil {
		klog.Errorf("Error executing command: %v\n", err)
	}

	if helper.Rm {
		err = client.RemovePodAnnotation(helper.PodName, helper.Namespace)
		if err != nil {
			return fmt.Errorf("Error remove debugger pod: %v\n", err)
		}

		klog.Infof("Debug pod %s deleted\n", helper.PodName)
	} else {
		klog.Infof("Debug pod %s exited\n", helper.PodName)
	}

	return nil
}

func SpawnDebuggerPodOnNormalNode(client *k8s.KubernetesClient, helper *DebuggerPodHelper) error {
	var err error
	var nsenterPodName string
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		nsenterPodName, err = client.CreateNsenterPod(helper.NodeName, helper.Image)

		return err
	})
	if err != nil {
		klog.Fatal(err)
	}

	err = wait.PollImmediate(5*time.Second, defaultRetryTimeout*time.Minute, func() (bool, error) {
		return client.IsPodRunning(nsenterPodName, "default")
	})
	if err != nil {
		return err
	}

	klog.Infof("spawning debugger pod  %s on node %s success\n", nsenterPodName, helper.NodeName)

	err = client.ExecCommand(nsenterPodName, "default", "debugger", helper.Command)
	// 仅输出报错，依然根据 --rm 参数删除pod
	if err != nil {
		klog.Errorf("Error executing command: %v\n", err)
	}

	if helper.Rm {
		err = client.DeletePod(nsenterPodName, "default", int64(10))
		if err != nil {
			return fmt.Errorf("Error remove debugger pod: %v\n", err)
		}

		klog.Infof("Debug pod %s deleted\n", nsenterPodName)
	} else {
		klog.Infof("Debug pod %s exited\n", nsenterPodName)
	}

	return nil
}
