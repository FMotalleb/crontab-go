package event

import (
	"fmt"
	"maps"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/generator"
)

var eg = generator.New[*config.JobEvent, abstraction.EventGenerator]()

func Build(log *zap.Logger, cfg *config.JobEvent) (abstraction.EventGenerator, error) {
	if g, ok := eg.Get(log, cfg); ok {
		return g, nil
	}
	return nil, fmt.Errorf("no event generator matched %+v", *cfg)
}

type MetaData struct {
	Emitter string
	Extra   map[string]any
}

func NewMetaData(emitter string, extra map[string]any) *MetaData {
	var e map[string]any
	if extra != nil {
		e = maps.Clone(extra)
	} else {
		e = make(map[string]any)
	}
	e["emitter"] = emitter
	return &MetaData{
		Emitter: emitter,
		Extra:   e,
	}
}

func NewErrMetaData(emitter string, err error) *MetaData {
	return &MetaData{
		Emitter: emitter,
		Extra: map[string]any{
			"error": err.Error(),
		},
	}
}

func (m *MetaData) GetData() map[string]any {
	return m.Extra
}
