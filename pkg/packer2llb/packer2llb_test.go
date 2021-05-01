package packer2llb

import (
	"context"
	"errors"
	"testing"

	"github.com/EricHripko/pack.yaml/pkg/packer2llb/config"
	packer2llb_mock "github.com/EricHripko/pack.yaml/pkg/packer2llb/mock"

	cib_mock "github.com/EricHripko/buildkit-fdk/pkg/cib/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type pluginTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	ctx    context.Context
	build  *cib_mock.MockService
	plugin *packer2llb_mock.MockPlugin
}

func (suite *pluginTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.ctx = context.Background()
	suite.build = cib_mock.NewMockService(suite.ctrl)
	suite.plugin = packer2llb_mock.NewMockPlugin(suite.ctrl)
}

func (suite *pluginTestSuite) TearDownTest() {
	suite.ctrl.Finish()

	Clear()
}

func (suite *pluginTestSuite) TestRegister() {
	// Act
	Register(suite.plugin)

	// Assert
	require.Len(suite.T(), plugins, 1)
	require.Same(suite.T(), suite.plugin, plugins[0])
}

func (suite *pluginTestSuite) TestDetectSrcFails() {
	// Arrange
	cfg := &config.Config{}
	expected := errors.New("something went wrong")
	suite.build.EXPECT().
		Src().
		Return(nil, expected)

	// Act
	_, actual := Detect(suite.ctx, suite.build, cfg)

	// Assert
	require.Same(suite.T(), expected, actual)
}

func (suite *pluginTestSuite) TestDetectFails() {
	// Arrange
	cfg := &config.Config{}
	src := cib_mock.NewMockReference(suite.ctrl)
	suite.build.EXPECT().
		Src().
		Return(src, nil)
	Register(suite.plugin)
	expected := errors.New("something went wrong")
	suite.plugin.EXPECT().
		Detect(suite.ctx, src, cfg).
		Return(expected)

	// Act
	_, actual := Detect(suite.ctx, suite.build, cfg)

	// Assert
	require.Same(suite.T(), expected, actual)
}

func (suite *pluginTestSuite) TestDetectNotFound() {
	// Arrange
	cfg := &config.Config{}
	src := cib_mock.NewMockReference(suite.ctrl)
	suite.build.EXPECT().
		Src().
		Return(src, nil)
	Register(suite.plugin)
	suite.plugin.EXPECT().
		Detect(suite.ctx, src, cfg).
		Return(nil)

	// Act
	plugin, err := Detect(suite.ctx, suite.build, cfg)

	// Assert
	require.Nil(suite.T(), err)
	require.Nil(suite.T(), plugin)
}

func (suite *pluginTestSuite) TestDetectSucceeds() {
	// Arrange
	cfg := &config.Config{}
	src := cib_mock.NewMockReference(suite.ctrl)
	suite.build.EXPECT().
		Src().
		Return(src, nil)
	Register(suite.plugin)
	suite.plugin.EXPECT().
		Detect(suite.ctx, src, cfg).
		Return(ErrActivate)

	// Act
	plugin, err := Detect(suite.ctx, suite.build, cfg)

	// Assert
	require.Nil(suite.T(), err)
	require.Same(suite.T(), suite.plugin, plugin)
}

func TestPlugin(t *testing.T) {
	suite.Run(t, new(pluginTestSuite))
}
