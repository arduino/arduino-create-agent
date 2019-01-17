package pkgs

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Indexes struct {
	Log *logrus.Logger
}

func (c *Indexes) Add(context.Context) error {
	return nil
}

func (c *Indexes) List(context.Context) ([]string, error) {
	return nil, nil
}

func (c *Indexes) Remove(context.Context) error {
	return nil
}
