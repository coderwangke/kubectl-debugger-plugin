package k8s

import "testing"

func TestGetEksIngestPodStr(t *testing.T) {
	podStr, err := getEksIngestPodStr("nginx:latest")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(podStr)
}
