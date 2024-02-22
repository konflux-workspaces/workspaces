package kube

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	cli client.Client
}
