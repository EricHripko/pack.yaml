package golang

import (
	"context"
	"errors"
	"testing"

	"github.com/EricHripko/pack.yaml/pkg/cib"
	cib_mock "github.com/EricHripko/pack.yaml/pkg/cib/mock"
	"github.com/EricHripko/pack.yaml/pkg/packer2llb"

	"github.com/golang/mock/gomock"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/moby/buildkit/frontend/gateway/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	fsutil "github.com/tonistiigi/fsutil/types"
)

type golangTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	ctx    context.Context
	build  *cib_mock.MockService
	src    *cib_mock.MockReference
	plugin *Plugin
}

func (suite *golangTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.ctx = context.Background()
	suite.build = cib_mock.NewMockService(suite.ctrl)
	suite.src = cib_mock.NewMockReference(suite.ctrl)
	suite.plugin = &Plugin{}
}

func (suite *golangTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *golangTestSuite) TestDetectNotFound() {
	// Arrange
	req := client.ReadDirRequest{Path: "."}
	files := []*fsutil.Stat{
		{Path: "README.md"},
	}
	suite.src.EXPECT().
		ReadDir(suite.ctx, req).
		Return(files, nil)

	// Act
	err := suite.plugin.Detect(suite.ctx, suite.src, nil)

	// Assert
	require.Nil(suite.T(), err)
}

func (suite *golangTestSuite) TestDetectFoundGoSource() {
	// Arrange
	req := client.ReadDirRequest{Path: "."}
	files := []*fsutil.Stat{
		{Path: "hello.go"},
	}
	suite.src.EXPECT().
		ReadDir(suite.ctx, req).
		Return(files, nil)
	suite.src.EXPECT().
		ReadFile(suite.ctx, gomock.Any()).
		Return(nil, errors.New("not found"))

	// Act
	err := suite.plugin.Detect(suite.ctx, suite.src, nil)

	// Assert
	require.Same(suite.T(), ErrUnknownDep, err)
}

func (suite *golangTestSuite) TestDetectGoModFails() {
	// Arrange
	req := client.ReadDirRequest{Path: "."}
	files := []*fsutil.Stat{
		{Path: "hello.go"},
	}
	suite.src.EXPECT().
		ReadDir(suite.ctx, req).
		Return(files, nil)
	goMod := []byte("\"")
	suite.src.EXPECT().
		ReadFile(suite.ctx, gomock.Any()).
		Return(goMod, nil).
		Times(2)

	// Act
	err := suite.plugin.Detect(suite.ctx, suite.src, nil)

	// Assert
	require.NotNil(suite.T(), err)
}

func (suite *golangTestSuite) TestDetectGoModIncomplete() {
	// Arrange
	req := client.ReadDirRequest{Path: "."}
	files := []*fsutil.Stat{
		{Path: "hello.go"},
	}
	suite.src.EXPECT().
		ReadDir(suite.ctx, req).
		Return(files, nil)
	goMod := []byte("")
	suite.src.EXPECT().
		ReadFile(suite.ctx, gomock.Any()).
		Return(goMod, nil).
		Times(2)

	// Act
	err := suite.plugin.Detect(suite.ctx, suite.src, nil)

	// Assert
	require.Same(suite.T(), ErrModIncomplete, err)
}

func (suite *golangTestSuite) TestDetectGoModSucceeds() {
	// Arrange
	req := client.ReadDirRequest{Path: "."}
	files := []*fsutil.Stat{
		{Path: "hello.go"},
	}
	suite.src.EXPECT().
		ReadDir(suite.ctx, req).
		Return(files, nil)
	goMod := []byte(`
module github.com/notareal/project

go 1.15
`)
	suite.src.EXPECT().
		ReadFile(suite.ctx, gomock.Any()).
		Return(goMod, nil).
		Times(2)

	// Act
	err := suite.plugin.Detect(suite.ctx, suite.src, nil)

	// Assert
	require.Same(suite.T(), packer2llb.ErrActivate, err)
}

func (suite *golangTestSuite) TestBuildFailsFrom1() {
	// Arrange
	suite.plugin.version = "1.14"

	platform := &specs.Platform{OS: "linux", Architecture: "amd64"}
	expected := errors.New("something went wrong")
	suite.build.EXPECT().
		From("golang:1.14", platform, gomock.Any()).
		Return(llb.Scratch(), nil, expected)

	// Act
	_, _, actual := suite.plugin.Build(suite.ctx, platform, suite.build)

	// Assert
	require.Same(suite.T(), expected, actual)
}

func (suite *golangTestSuite) TestBuildFailsSrc() {
	// Arrange
	suite.plugin.version = "1.14"

	platform := &specs.Platform{OS: "linux", Architecture: "amd64"}
	suite.build.EXPECT().
		From("golang:1.14", platform, gomock.Any()).
		Return(llb.Scratch(), nil, nil)

	expected := errors.New("something went wrong")
	suite.build.EXPECT().
		SrcState().
		Return(llb.Scratch(), expected)

	// Act
	_, _, actual := suite.plugin.Build(suite.ctx, platform, suite.build)

	// Assert
	require.Same(suite.T(), expected, actual)
}

func (suite *golangTestSuite) TestBuildFailsFrom2() {
	// Arrange
	suite.plugin.version = "1.14"
	suite.plugin.config = cib.NewConfig()

	platform := &specs.Platform{OS: "linux", Architecture: "amd64"}
	suite.build.EXPECT().
		From("golang:1.14", platform, gomock.Any()).
		Return(llb.Scratch(), nil, nil)
	suite.build.EXPECT().
		SrcState().
		Return(llb.Scratch(), nil)

	expected := errors.New("something went wrong")
	suite.build.EXPECT().
		From("gcr.io/distroless/base:debug", platform, gomock.Any()).
		Return(llb.Scratch(), nil, expected)

	// Act
	_, _, actual := suite.plugin.Build(suite.ctx, platform, suite.build)

	// Assert
	require.Same(suite.T(), expected, actual)
}

func (suite *golangTestSuite) TestBuildSucceeds() {
	// Arrange
	suite.plugin.version = "1.14"
	suite.plugin.config = cib.NewConfig()
	suite.plugin.config.Debug = false

	platform := &specs.Platform{OS: "linux", Architecture: "amd64"}
	suite.build.EXPECT().
		From("golang:1.14", platform, gomock.Any()).
		Return(llb.Scratch(), nil, nil)
	suite.build.EXPECT().
		SrcState().
		Return(llb.Scratch(), nil)
	expected := &dockerfile2llb.Image{}
	suite.build.EXPECT().
		From("gcr.io/distroless/base", platform, gomock.Any()).
		Return(llb.Scratch(), expected, nil)

	// Act
	state, actual, err := suite.plugin.Build(suite.ctx, platform, suite.build)

	// Assert
	require.Nil(suite.T(), err)
	require.NotNil(suite.T(), state)
	require.Same(suite.T(), expected, actual)
}

func TestGolangPlugin(t *testing.T) {
	suite.Run(t, new(golangTestSuite))
}
