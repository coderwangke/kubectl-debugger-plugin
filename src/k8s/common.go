package k8s

const (
	EKS_DEBUGGER_POD_ANNOTATION_KEY = "eks.tke.cloud.tencent.com/debug-pod"
	EKS_INGEST_POD_NAME = `{"apiVersion":"v1","kind":"Pod","metadata":{"name":"pod-debugger"},"spec":{"containers":[{"image":"busybox","name":"debugger","command":["/bin/sh"],"args":["-c","tail -f /dev/null"],"resources":{"limits":{"memory":"1Gi","cpu":"500m"}},"securityContext":{"privileged":true},"volumeMounts":[{"mountPath":"/host","name":"host-root"}]}],"dnsPolicy":"ClusterFirst","hostIPC":true,"hostNetwork":true,"hostPID":true,"volumes":[{"hostPath":{"path":"/","type":""},"name":"host-root"}]}}`
)
