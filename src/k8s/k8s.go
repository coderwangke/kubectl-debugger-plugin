package k8s

import (
	"context"
	"encoding/hex"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	"math/rand"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	kubectlexec "k8s.io/kubectl/pkg/cmd/exec"
)

// KubernetesClient 存储 Kubernetes 客户端
type KubernetesClient struct {
	clientset     *kubernetes.Clientset
	config        *rest.Config
	streamOptions *kubectlexec.StreamOptions
}

// NewKubernetesClient 创建一个新的 Kubernetes 客户端
func NewKubernetesClient(kubeconfig string) (*KubernetesClient, error) {
	// 创建 Kubernetes 配置
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// 创建 Kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	streamOptions := kubectlexec.StreamOptions{
		IOStreams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		Stdin: true,
		TTY:   true,
	}

	return &KubernetesClient{
		clientset:     clientset,
		config:        config,
		streamOptions: &streamOptions,
	}, nil
}

// GetNode 根据 pod 名称和命名空间获取 pod 所在的节点名称
func (c *KubernetesClient) GetNode(podName, namespace string) (string, error) {
	// 获取 pod 所在的节点
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return "", fmt.Errorf("Pod %s in namespace %s not found\n", podName, namespace)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		return "", fmt.Errorf("Error getting Pod %s in namespace %s: %v\n",
			podName, namespace, statusError.ErrStatus.Message)
	} else if err != nil {
		return "", err
	}

	node := pod.Spec.NodeName
	if node == "" {
		return "", fmt.Errorf("pod %s in namespace %s is not running on any node", podName, namespace)
	}

	return node, nil
}

// GetNodeType 检查节点类型，如果节点有标签：node.kubernetes.io/instance-type: eklet 就是超级节点，否则就是普通节点
func (c *KubernetesClient) GetNodeType(node string) (string, error) {
	// 获取节点标签
	nodeObj, err := c.clientset.CoreV1().Nodes().Get(context.Background(), node, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return "", fmt.Errorf("Node %s not found\n", node)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		return "", fmt.Errorf("Error getting Node %s: %v\n",
			node, statusError.ErrStatus.Message)
	} else if err != nil {
		return "", err
	}

	labels := nodeObj.GetLabels()
	if labels["node.kubernetes.io/instance-type"] == "eklet" {
		return SuperNodeType, nil
	}

	return NormalNodeType, nil
}

// AnnotatePod 给 pod 打上注解
func (c *KubernetesClient) AnnotatePod(podName, namespace, image string) error {
	// 获取 pod 对象
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod %s in namespace %s: %v", podName, namespace, err)
	}

	// 给 pod 打上注解
	annotations := pod.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	ingestPodStr, err := getEksIngestPodStr(image)
	if err != nil {
		return err
	}

	// 如果注解存在，则进行更新；如果注解不存在，则进行新增
	annotations[EksDebuggerPodAnnotationKey] = ingestPodStr
	pod.SetAnnotations(annotations)

	_, err = c.clientset.CoreV1().Pods(namespace).Update(context.Background(), pod, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed annotate pod %s in namespace %s: %v", podName, namespace, err)
	}

	return nil
}

func (c *KubernetesClient) RemovePodAnnotation(podName, namespace string) error {
	// 获取 pod 对象
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod %s in namespace %s: %v", podName, namespace, err)
	}

	// 移除指定的注解
	annotations := pod.GetAnnotations()
	if annotations != nil {
		if _, exists := annotations[EksDebuggerPodAnnotationKey]; !exists {
			return nil
		}
		delete(annotations, EksDebuggerPodAnnotationKey)
		pod.SetAnnotations(annotations)
	}

	_, err = c.clientset.CoreV1().Pods(namespace).Update(context.Background(), pod, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove debug annotation from pod %s in namespace %s: %v", podName, namespace, err)
	}

	return nil
}

func (c *KubernetesClient) IsDebuggerContainerRunning(podName, namespace string) (bool, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, fmt.Errorf("Pod %s in namespace %s not found\n", podName, namespace)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		return false, fmt.Errorf("Error getting Pod %s in namespace %s: %v\n",
			podName, namespace, statusError.ErrStatus.Message)
	} else if err != nil {
		return false, err
	}

	// 检查指定容器的状态
	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == "[pod-debugger]debugger" {
			return container.State.Running != nil, nil
		}
	}

	return false, nil
}

func (c *KubernetesClient) IsPodRunning(podName, namespace string) (bool, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, fmt.Errorf("Pod %s in namespace %s not found\n", podName, namespace)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		return false, fmt.Errorf("Error getting Pod %s in namespace %s: %v\n",
			podName, namespace, statusError.ErrStatus.Message)
	} else if err != nil {
		return false, err
	}

	if pod.Status.Phase == corev1.PodRunning {
		return true, nil
	} else {
		return false, nil
	}
}

// CreateNsenterPod 在节点上创建一个命名为 nsenter-xxxxx 的 pod
func (c *KubernetesClient) CreateNsenterPod(node, image string) (string, error) {
	var podName string
	// 生成随机字符
	rand.Seed(time.Now().UnixNano())
	randBytes := make([]byte, 5)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string: %v", err)
	}
	randStr := hex.EncodeToString(randBytes)

	// 创建 PodSpec
	var podSpec = &corev1.PodSpec{
		NodeName: node,
		HostPID:  true,
		Containers: []corev1.Container{
			{
				SecurityContext: &corev1.SecurityContext{
					Privileged: pointer.Bool(true),
				},
				Image:     image,
				Name:      "debugger",
				Stdin:     true,
				StdinOnce: true,
				TTY:       true,
				Command: []string{
					"nsenter",
					"--target",
					"1",
					"--mount",
					"--uts",
					"--ipc",
					"--net",
					"--pid",
					"--",
					"bash",
					"-l",
				},
			},
		},
	}
	// 创建 Pod
	podName = fmt.Sprintf("nsenter-%s", randStr)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("nsenter-%s", randStr),
			Namespace: "default",
		},
		Spec: *podSpec,
	}

	_, err = c.clientset.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create nsenter pod on node %s in namespace %s: %v", node, "default", err)
	}

	return podName, nil
}

func (c *KubernetesClient) DeletePod(podName, namespace string, timeout int64) error {
	deleteOptions := &metav1.DeleteOptions{}
	if timeout > 0 {
		duration := int64(time.Duration(timeout).Seconds())
		deleteOptions.GracePeriodSeconds = &duration
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+10)*time.Second)
	defer cancel()
	err := c.clientset.CoreV1().Pods(namespace).Delete(ctx, podName, *deleteOptions)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		if statusError, isStatus := err.(*errors.StatusError); isStatus {
			return fmt.Errorf("Error deleting Pod %s in namespace %s: %v",
				podName, namespace, statusError.ErrStatus.Message)
		}
		return err
	}
	return nil
}

func (c *KubernetesClient) ExecCommand(podName, namespace, container string, command []string) error {
	// 创建REST请求
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t := c.streamOptions.SetupTTY()

	var sizeQueue remotecommand.TerminalSizeQueue

	if t.Raw {
		// this call spawns a goroutine to monitor/update the terminal size
		sizeQueue = t.MonitorSize(t.GetSize())

		// unset p.Err if it was previously set because both stdout and stderr go over p.Out when tty is
		// true
		c.streamOptions.ErrOut = nil
	}

	// 执行命令
	opts := remotecommand.StreamOptions{
		Stdin:             c.streamOptions.In,
		Stdout:            c.streamOptions.Out,
		Stderr:            c.streamOptions.ErrOut,
		Tty:               t.Raw,
		TerminalSizeQueue: sizeQueue,
	}

	fn := func() error {
		return executor.StreamWithContext(ctx, opts)
	}
	err = t.Safe(fn)
	if err != nil {
		return err
	}

	return nil
}
