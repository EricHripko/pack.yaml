// Package packer2llb is the main integration point for plugins that implement
// support for various ecosystems.
package packer2llb

import (
	"context"
	"errors"

	"github.com/EricHripko/pack.yaml/pkg/cib"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/moby/buildkit/frontend/gateway/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// Detect if any of active integrations can process this project.
func Detect(ctx context.Context, build cib.Service, config *cib.Config) (plugin Plugin, err error) {
	src, err := build.Src()
	if err != nil {
		return
	}

	for _, candidate := range plugins {
		err = candidate.Detect(ctx, src, config)
		if err == nil {
			continue
		}
		if err == ErrActivate {
			err = nil
			plugin = candidate
			continue
		}
		return
	}
	return
}

//go:generate mockgen -package packer2llb_mock -destination mock/packer2llb.go . Plugin

// Plugin represents an ecosystem integration.
type Plugin interface {
	// Detect if this plugin is compatible with the project (in which case
	// ErrActivate is returned).
	Detect(ctx context.Context, src client.Reference, config *cib.Config) error
	// Build a container image for the project with this plugin.
	Build(ctx context.Context, platform *specs.Platform, build cib.Service) (*llb.State, *dockerfile2llb.Image, error)
}

// ErrActivate is returned by plugin's Detect function when plugin detected
// a compatible project.
var ErrActivate = errors.New("packer2llb: activate plugin")

// Register the plugin for the integration.
func Register(plugin Plugin) {
	plugins = append(plugins, plugin)
}

// Clear all plugin registrations.
func Clear() {
	plugins = []Plugin{}
}

var plugins []Plugin
