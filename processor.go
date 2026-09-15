package metadata_compactor_golang

import (
	"fmt"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/processor/csn"
	"github.com/open-resource-discovery/metadata-compactor-golang/model"
)

type Format int

const (
	CSN Format = iota
)

type Processor struct {
	csn *csn.CSNProcessor
}

func (self *Processor) Process(format Format, document string) string {
	switch format {
	case CSN:
		return self.csn.Process(document)
	default:
		panic(fmt.Sprintf("Unsupported format: %+v", format))
	}
}

func Create(ruleset *model.Ruleset) *Processor {
	return &Processor{
		csn: csn.CreateProcessor(ruleset.CSN.Options, ruleset.CSN.Rules),
	}
}
