package k8s

import (
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

const (
	EksDebuggerPodAnnotationKey = "eks.tke.cloud.tencent.com/debug-pod"
	SuperNodeType               = "super"
	NormalNodeType              = "normal"
)

var eksIngestPod = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "pod-debugger",
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Image:   "busybox",
				Name:    "debugger",
				Command: []string{"/bin/sh"},
				Args:    []string{"-c", "tail -f /dev/null"},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("1Gi"),
						corev1.ResourceCPU:    resource.MustParse("500m"),
					},
				},
				SecurityContext: &corev1.SecurityContext{
					Privileged: pointer.Bool(true),
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						MountPath: "/host",
						Name:      "host-root",
					},
				},
			},
		},
		DNSPolicy:   corev1.DNSClusterFirst,
		HostIPC:     true,
		HostNetwork: true,
		HostPID:     true,
		Volumes: []corev1.Volume{
			{
				Name: "host-root",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/",
					},
				},
			},
		},
	},
}

func getEksIngestPodStr(image string) (string, error) {
	eksIngestPod.Spec.Containers[0].Image = image

	data, err := json.Marshal(eksIngestPod)
	if err != nil {
		return "", fmt.Errorf("failed get ingrest pod string, err: %v", err.Error())
	}

	return string(data), nil
}
