package cache

import "sigs.k8s.io/controller-runtime/pkg/client"

type Cache struct{}

func NewCache(cli *client.Client) *Cache {
	return &Cache{}
}
