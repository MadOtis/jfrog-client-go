package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/madotis/jfrog-client-go/artifactory"
	"github.com/madotis/jfrog-client-go/auth"
	"github.com/madotis/jfrog-client-go/config"
	"github.com/stretchr/testify/assert"
)

func TestCtx(t *testing.T) {
	initArtifactoryTest(t)
	t.Run("ctx", testCtx)
	t.Run("ctxWithTimeout", testCtxTimeout)
}

func ctxMgr(t *testing.T, artDetails auth.ServiceDetails, ctx context.Context) (artifactory.ArtifactoryServicesManager, error) {
	cfg, err := config.NewConfigBuilder().SetServiceDetails(artDetails).SetContext(ctx).Build()
	assert.NoError(t, err)
	return artifactory.New(cfg)
}

func testCtx(t *testing.T) {
	artDetails := GetRtDetails()
	sm, err := ctxMgr(t, artDetails, context.Background())
	assert.NoError(t, err)
	_, err = sm.GetVersion()
	assert.NoError(t, err)
}

func testCtxTimeout(t *testing.T) {
	artDetails := GetRtDetails()
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond*250)
	defer cancel()
	sm, err := ctxMgr(t, artDetails, timeoutCtx)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 300)
	_, err = sm.GetVersion()
	assert.Error(t, err)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fail()
	}
}
