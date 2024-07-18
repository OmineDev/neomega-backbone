package neomega_backbone

import (
	"context"

	"github.com/OmineDev/neomega-core/neomega"
)

type StorageAndLogAccess interface {
	StorageAndPathAccess
	BackendIO
}

type StorageAndLogProvider interface {
	StorageAndPathProvider
	BackendIO
}

type PreInitOmega interface {
	ConfigRead
	StorageAndLogProvider
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
