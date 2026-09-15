//go:build unit || integration

package testutils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ohler55/ojg/oj"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/utils"
)

func LoadFixture(path string) string {
	bytes, err := os.ReadFile(filepath.Join(utils.First(os.Getwd()), path))
	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %s", path, err.Error()))
	}

	return string(bytes)
}

func UnmarshalFixture[T any](path string) T {
	return oj.MustParseString(LoadFixture(path)).(T)
}
