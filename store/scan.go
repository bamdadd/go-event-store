package store

type QueryConstraint interface {
	isQueryConstraint()
}

type SequenceNumberAfterConstraint struct {
	SequenceNumber int
}

func (SequenceNumberAfterConstraint) isQueryConstraint() {}

type ScanConfig struct {
	Constraints []QueryConstraint
}

type ScanOption func(*ScanConfig)

func WithSequenceNumberAfter(seq int) ScanOption {
	return func(cfg *ScanConfig) {
		cfg.Constraints = append(cfg.Constraints, SequenceNumberAfterConstraint{SequenceNumber: seq})
	}
}

func BuildScanConfig(opts ...ScanOption) ScanConfig {
	cfg := ScanConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}
