package neomega_backbone

import (
	"context"

	"github.com/OmineDev/neomega-core/neomega"
)

type PreInitOmega interface {
	ConfigRead
	StorageAndLogAccess
	BackendIO
	FrameLevelShare
}

type ExtendOmega interface {
	PreInitOmega
	GameMenuSetter
	neomega.MicroOmega
	CQHTTPAccess
	FlexEnhance
	Logger() LineDst
}

type OmegaFrame interface {
	Init(ctx context.Context, omega neomega.MicroOmega)
	PreInit() error
}
