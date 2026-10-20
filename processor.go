package metadata_compactor_golang

import (
	"fmt"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/processor/csn"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/processor/oas"
	"github.com/open-resource-discovery/metadata-compactor-golang/model"
)

type Format int

const (
	CSN Format = iota
	OAS Format = iota
)

type Processor struct {
	csn *csn.CSNProcessor
	oas *oas.OASProcessor
}

func (self *Processor) Process(format Format, document string, baseline ...string) string {
	switch format {
	case CSN:
		return self.csn.Process(document, baseline...)
	case OAS:
		return self.oas.Process(document, baseline...)
	default:
		panic(fmt.Sprintf("Unsupported format: %+v", format))
	}
}

func Create(ruleset *model.Ruleset) *Processor {
	return &Processor{
		csn: csn.CreateProcessor(ruleset.CSN.Options, ruleset.CSN.Rules),
		oas: oas.CreateProcessor(ruleset.OAS.Rules),
	}
}
