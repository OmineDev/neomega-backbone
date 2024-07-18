package neomega_backbone

import (
	"context"

	"github.com/OmineDev/neomega-core/neomega"
)

type PreInitOmega interface {
	ConfigRead
	StorageAndPathAccess
	BackendIO
	FlexEnhance
}

type ExtendOmega interface {
	PreInitOmega
	GameMenuSetter
	neomega.MicroOmega
	CQHTTPAccess
	FlexEnhance
}

type OmegaFrame interface {
	Init(ctx context.Context, omega neomega.MicroOmega)
	PreInit() error
}
