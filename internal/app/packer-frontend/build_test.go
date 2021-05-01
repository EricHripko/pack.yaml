package frontend

import (
	"context"
	"errors"
	"path"
	"testing"

	"github.com/EricHripko/pack.yaml/pkg/packer2llb"
	packer2llb_mock "github.com/EricHripko/pack.yaml/pkg/packer2llb/mock"

	cib_mock "github.com/EricHripko/buildkit-fdk/pkg/cib/mock"
	"github.com/golang/mock/gomock"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/moby/buildkit/frontend/gateway/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	fsutil "github.com/tonistiigi/fsutil/types"
)

type singleTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	ctx    context.Context
	build  *cib_mock.MockService
	client *cib_mock.MockClient
}

func (suite *singleTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.ctx = context.Background()
	suite.build = cib_mock.NewMockService(suite.ctrl)
	suite.client = cib_mock.NewMockClient(suite.ctrl)

	suite.build.EXPECT().
		GetOpts().
		Return(map[string]string{})
	platforms := []*specs.Platform{
		{OS: "linux", Architecture: "amd64"},
	}
	suite.build.EXPECT().
		GetTargetPlatforms().
		Return(platforms, nil)
}

func (suite *singleTestSuite) TearDownTest() {
	suite.ctrl.Finish()
	packer2llb.Clear()
}

func (suite *singleTestSuite) TestGetMetadataFails() {
	// Arrange
	expected := errors.New("something went wrong")
	suite.build.EXPECT().
		GetMetadata().
		Return(nil, expected)

	// Act
	_, actual := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.Same(suite.T(), expected, actual)
}

func (suite *singleTestSuite) TestReadMetadataFails() {
	// Arrange
	data := []byte("!;'not_val1d")
	suite.build.EXPECT().
		GetMetadata().
		Return(data, nil)

	// Act
	_, err := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.NotNil(suite.T(), err)
}

func (suite *singleTestSuite) TestDetectFails() {
	// Arrange
	expected := errors.New("something went wrong")
	plugin := packer2llb_mock.NewMockPlugin(suite.ctrl)
	plugin.EXPECT().
		Detect(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(expected)
	packer2llb.Register(plugin)

	suite.build.EXPECT().
		GetMetadata().
		Return([]byte(""), nil)
	ref := cib_mock.NewMockReference(suite.ctrl)
	suite.build.EXPECT().
		Src().
		Return(ref, nil)

	// Act
	_, actual := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.Same(suite.T(), expected, actual)
}

func (suite *singleTestSuite) TestBuildFails() {
	// Arrange
	plugin := packer2llb_mock.NewMockPlugin(suite.ctrl)
	plugin.EXPECT().
		Detect(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(packer2llb.ErrActivate)
	plugin.EXPECT().
		Build(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil, errors.New("something went wrong"))
	packer2llb.Register(plugin)

	suite.build.EXPECT().
		GetMetadata().
		Return([]byte(""), nil)
	ref := cib_mock.NewMockReference(suite.ctrl)
	suite.build.EXPECT().
		Src().
		Return(ref, nil)

	// Act
	_, err := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.NotNil(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "something went wrong")
}

func (suite *singleTestSuite) TestSolveFails() {
	// Arrange
	plugin := packer2llb_mock.NewMockPlugin(suite.ctrl)
	plugin.EXPECT().
		Detect(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(packer2llb.ErrActivate)
	state := llb.Scratch()
	plugin.EXPECT().
		Build(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&state, nil, nil)
	packer2llb.Register(plugin)

	suite.build.EXPECT().
		GetMetadata().
		Return([]byte(""), nil)
	ref := cib_mock.NewMockReference(suite.ctrl)
	suite.build.EXPECT().
		Src().
		Return(ref, nil)

	expected := errors.New("something went wrong")
	suite.client.EXPECT().
		Solve(gomock.Any(), gomock.Any()).
		Return(nil, expected)

	// Act
	_, actual := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.Same(suite.T(), expected, actual)
}

func (suite *singleTestSuite) TestRefFails() {
	// Arrange
	plugin := packer2llb_mock.NewMockPlugin(suite.ctrl)
	plugin.EXPECT().
		Detect(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(packer2llb.ErrActivate)
	state := llb.Scratch()
	plugin.EXPECT().
		Build(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&state, nil, nil)
	packer2llb.Register(plugin)

	suite.build.EXPECT().
		GetMetadata().
		Return([]byte(""), nil)
	ref := cib_mock.NewMockReference(suite.ctrl)
	suite.build.EXPECT().
		Src().
		Return(ref, nil)

	res := client.NewResult()
	res.AddRef("test", cib_mock.NewMockReference(suite.ctrl))
	suite.client.EXPECT().
		Solve(gomock.Any(), gomock.Any()).
		Return(res, nil)

	// Act
	_, err := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.NotNil(suite.T(), err)
}

func (suite *singleTestSuite) TestFindCommandFails() {
	// Arrange
	plugin := packer2llb_mock.NewMockPlugin(suite.ctrl)
	plugin.EXPECT().
		Detect(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(packer2llb.ErrActivate)
	state := llb.Scratch()
	img := &dockerfile2llb.Image{}
	plugin.EXPECT().
		Build(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&state, img, nil)
	packer2llb.Register(plugin)

	suite.build.EXPECT().
		GetMetadata().
		Return([]byte(""), nil)
	src := cib_mock.NewMockReference(suite.ctrl)
	suite.build.EXPECT().
		Src().
		Return(src, nil)

	ref := cib_mock.NewMockReference(suite.ctrl)
	ref.EXPECT().
		ReadDir(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("something went wrong"))

	res := client.NewResult()
	res.SetRef(ref)
	suite.client.EXPECT().
		Solve(gomock.Any(), gomock.Any()).
		Return(res, nil)

	// Act
	_, err := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.Same(suite.T(), errNoCommand, err)
}

func (suite *singleTestSuite) TestSucceedsImplicitCommand() {
	// Arrange
	plugin := packer2llb_mock.NewMockPlugin(suite.ctrl)
	plugin.EXPECT().
		Detect(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(packer2llb.ErrActivate)
	state := llb.Scratch()
	img := &dockerfile2llb.Image{}
	plugin.EXPECT().
		Build(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&state, img, nil)
	packer2llb.Register(plugin)

	suite.build.EXPECT().
		GetMetadata().
		Return([]byte(""), nil)
	src := cib_mock.NewMockReference(suite.ctrl)
	suite.build.EXPECT().
		Src().
		Return(src, nil)

	ref := cib_mock.NewMockReference(suite.ctrl)
	files := []*fsutil.Stat{
		{Path: path.Join(packer2llb.DirInstall, "command"), Mode: 0755},
	}
	ref.EXPECT().
		ReadDir(gomock.Any(), gomock.Any()).
		Return(files, nil)

	res := client.NewResult()
	res.SetRef(ref)
	suite.client.EXPECT().
		Solve(gomock.Any(), gomock.Any()).
		Return(res, nil)

	// Act
	res, err := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.Nil(suite.T(), err)
	require.NotNil(suite.T(), res.Ref)
}

func (suite *singleTestSuite) TestSucceedsExplicitCommand() {
	// Arrange
	plugin := packer2llb_mock.NewMockPlugin(suite.ctrl)
	plugin.EXPECT().
		Detect(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(packer2llb.ErrActivate)
	state := llb.Scratch()
	img := &dockerfile2llb.Image{}
	plugin.EXPECT().
		Build(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&state, img, nil)
	packer2llb.Register(plugin)

	metadata := []byte(`
entrypoint: ["entrypoint"]
`)
	suite.build.EXPECT().
		GetMetadata().
		Return(metadata, nil)
	src := cib_mock.NewMockReference(suite.ctrl)
	suite.build.EXPECT().
		Src().
		Return(src, nil)

	res := client.NewResult()
	res.SetRef(cib_mock.NewMockReference(suite.ctrl))
	suite.client.EXPECT().
		Solve(gomock.Any(), gomock.Any()).
		Return(res, nil)

	// Act
	res, err := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.Nil(suite.T(), err)
	require.NotNil(suite.T(), res.Ref)
}

func TestSinglePlatform(t *testing.T) {
	suite.Run(t, new(singleTestSuite))
}

type multiTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	ctx    context.Context
	build  *cib_mock.MockService
	client *cib_mock.MockClient
}

func (suite *multiTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.ctx = context.Background()
	suite.build = cib_mock.NewMockService(suite.ctrl)
	suite.client = cib_mock.NewMockClient(suite.ctrl)
}

func (suite *multiTestSuite) TearDownTest() {
	suite.ctrl.Finish()
	packer2llb.Clear()
}

func (suite *multiTestSuite) TestTargetPlatformFails() {
	// Arrange
	suite.build.EXPECT().
		GetOpts().
		Return(map[string]string{})
	expected := errors.New("something went wrong")
	suite.build.EXPECT().
		GetTargetPlatforms().
		Return(nil, expected)

	// Act
	_, actual := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.Same(suite.T(), expected, actual)
}

func (suite *multiTestSuite) TestParseFails() {
	// Arrange
	opts := map[string]string{
		keyMultiPlatform: "invalid",
	}
	suite.build.EXPECT().
		GetOpts().
		Return(opts)
	platforms := []*specs.Platform{
		{OS: "linux", Architecture: "amd64"},
		{OS: "linux", Architecture: "arm64"},
	}
	suite.build.EXPECT().
		GetTargetPlatforms().
		Return(platforms, nil)

	// Act
	_, err := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.NotNil(suite.T(), err)
}

func (suite *multiTestSuite) TestNotAllowed() {
	// Arrange
	opts := map[string]string{
		keyMultiPlatform: "false",
	}
	suite.build.EXPECT().
		GetOpts().
		Return(opts)
	platforms := []*specs.Platform{
		{OS: "linux", Architecture: "amd64"},
		{OS: "linux", Architecture: "arm64"},
	}
	suite.build.EXPECT().
		GetTargetPlatforms().
		Return(platforms, nil)

	// Act
	_, err := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.NotNil(suite.T(), err)
}

func (suite *multiTestSuite) TestSucceeds() {
	// Arrange
	suite.build.EXPECT().
		GetOpts().
		Return(map[string]string{})
	platforms := []*specs.Platform{
		{OS: "linux", Architecture: "amd64"},
		{OS: "linux", Architecture: "arm64"},
	}
	suite.build.EXPECT().
		GetTargetPlatforms().
		Return(platforms, nil)

	plugin := packer2llb_mock.NewMockPlugin(suite.ctrl)
	plugin.EXPECT().
		Detect(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(packer2llb.ErrActivate).
		Times(2)
	state := llb.Scratch()
	img := &dockerfile2llb.Image{}
	plugin.EXPECT().
		Build(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&state, img, nil).
		Times(2)
	packer2llb.Register(plugin)

	metadata := []byte(`
entrypoint: ["entrypoint"]
`)
	suite.build.EXPECT().
		GetMetadata().
		Return(metadata, nil).
		Times(2)
	src := cib_mock.NewMockReference(suite.ctrl)
	suite.build.EXPECT().
		Src().
		Return(src, nil).
		Times(2)

	res := client.NewResult()
	res.SetRef(cib_mock.NewMockReference(suite.ctrl))
	suite.client.EXPECT().
		Solve(gomock.Any(), gomock.Any()).
		Return(res, nil).
		Times(2)

	// Act
	res, err := BuildWithService(suite.ctx, suite.client, suite.build)

	// Assert
	require.Nil(suite.T(), err)
	require.NotNil(suite.T(), res.Refs)
	require.Len(suite.T(), res.Refs, 2)
}

func TestMultiPlatform(t *testing.T) {
	suite.Run(t, new(multiTestSuite))
}
