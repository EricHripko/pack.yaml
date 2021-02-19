package cmd

import (
	frontend "github.com/EricHripko/pack.yaml/internal/app/packer-frontend"

	"github.com/moby/buildkit/frontend/gateway/grpcclient"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/sirupsen/logrus"
)

// Main entrypoint for the frontend.
func Main() {
	if err := grpcclient.RunFromEnvironment(appcontext.Context(), frontend.Build); err != nil {
		logrus.Errorf("fatal error: %+v", err)
		panic(err)
	}
}
