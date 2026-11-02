// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and overlay-golang contributors
// SPDX-License-Identifier: Apache-2.0
// Adapted from overlay-golang/internal/common/xml2json for metadata compaction.

package xml2json

import "github.com/open-resource-discovery/metadata-compactor-golang/internal/common/utils"

type Attributes map[string]string

func NewAttributes(values ...string) Attributes {
	result := make(Attributes, len(values)/2)
	for index := 0; index+1 < len(values); index += 2 {
		result[values[index]] = values[index+1]
	}
	return result
}

func (self Attributes) Has(name string) bool {
	return utils.ContainsKey(self, name)
}

func (self Attributes) Get(name string) string {
	return self[name]
}
