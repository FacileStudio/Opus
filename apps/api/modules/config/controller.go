package config

import "context"

type controller struct {
	service *Service
}

func newController(service *Service) *controller {
	return &controller{service: service}
}

func (c *controller) get(_ context.Context) ConfigResponse {
	return c.service.Get()
}
