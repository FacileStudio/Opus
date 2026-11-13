package config

import "context"

type controller struct {
	service *Service
}

func newController(service *Service) *controller {
	return &controller{service: service}
}

func (c *controller) get(ctx context.Context) ConfigResponse {
	return c.service.Get(ctx)
}
