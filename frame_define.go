package neomega_backbone

import (
	"context"

	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/utils/pressure_metric"
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
}

type ExtendOmegaCmdBox struct {
	neomega.MicroOmegaCmdBox
	PreInitOmega
	GameMenuSetter
	CQHTTPAccess
}

func NewExtendOmegaCmdBox(o ExtendOmega, m *pressure_metric.FreqMetric) ExtendOmega {
	return &ExtendOmegaCmdBox{
		neomega.NewMicroOmegaCmdBox(o, m),
		o, o, o,
	}
}

type OmegaFrame interface {
	Init(ctx context.Context, omega neomega.MicroOmega)
	PreInit() error
}
