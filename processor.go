package metadata_compactor_golang

import (
	"fmt"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/processor/csn"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/processor/edmx"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/processor/oas"
	"github.com/open-resource-discovery/metadata-compactor-golang/model"
)

type Format int

const (
	CSN Format = iota
	OAS
	EDMX
)

type Processor struct {
	csn  *csn.CSNProcessor
	oas  *oas.OASProcessor
	edmx *edmx.EDMXProcessor
}

// Process preserves the existing panic-on-error API. Use ProcessE to handle XML errors.
func (self *Processor) Process(format Format, document string, baselines ...string) string {
	result, err := self.ProcessE(format, document, baselines...)
	if err != nil {
		panic(err)
	}
	return result
}

// ProcessE compacts a document using configured and baseline-derived rules.
// Malformed CSN and OAS retain the existing JSON processor panic behavior.
func (self *Processor) ProcessE(format Format, document string, baselines ...string) (string, error) {
	switch format {
	case CSN:
		return self.csn.Process(document, baselines...), nil
	case OAS:
		return self.oas.Process(document, baselines...), nil
	case EDMX:
		return self.edmx.Process(document, baselines...)
	default:
		return "", fmt.Errorf("Unsupported format: %+v", format)
	}
}

func Create(ruleset *model.Ruleset) *Processor {
	return &Processor{
		csn:  csn.CreateProcessor(ruleset.CSN.Options, ruleset.CSN.Rules),
		oas:  oas.CreateProcessor(ruleset.OAS.Rules),
		edmx: edmx.CreateProcessor(ruleset.EDMX.Rules),
	}
}
