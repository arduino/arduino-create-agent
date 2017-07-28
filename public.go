package main

import (
	"github.com/goadesign/goa"
)

// PublicController implements the public resource.
type PublicController struct {
	*goa.Controller
}

// NewPublicController creates a public controller.
func NewPublicController(service *goa.Service) *PublicController {
	return &PublicController{Controller: service.NewController("PublicController")}
}
