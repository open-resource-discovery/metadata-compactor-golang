package jputils

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/utils"
	"golang.org/x/exp/constraints"
)

func Frag(value string) jp.Frag {
	switch value {
	case "@":
		return jp.At('@')
	case "$":
		return jp.Root('$')
	case "*":
		return jp.Wildcard('*')
	default:
		if utils.First(regexp.MatchString(`^\[\d+]$`, value)) {
			return jp.Nth(utils.First(strconv.Atoi(value)))
		}

		return jp.Child(value)
	}
}

func Expr(values ...any) jp.Expr {
	result := make([]jp.Frag, 0, len(values))

	for _, node := range values {
		switch node.(type) {
		case string:
			result = append(result, Frag(utils.SafeCast[string](node)))
		case jp.Expr, []jp.Frag:
			result = append(result, utils.SafeCast[jp.Expr](node)...)
		case *jp.Equation:
			result = append(result, utils.SafeCast[*jp.Equation](node).Filter())
		default:
			result = append(result, node.(jp.Frag))
		}
	}

	return result
}

func Const(value any) *jp.Equation {
	switch value.(type) {
	case bool:
		return jp.ConstBool(value.(bool))
	case string:
		return jp.ConstString(value.(string))
	case float32, float64:
		return jp.ConstFloat(utils.SafeCast[float64](value))
	case int, int8, int16, int32, int64:
		return jp.ConstInt(utils.SafeCast[int64](value))
	default:
		panic(fmt.Sprintf("unsupported constant type: %v", reflect.TypeOf(value)))
	}
}

func Eq[T bool | string | constraints.Float | constraints.Integer](expression string, value T) *jp.Equation {
	return jp.Eq(
		jp.Get(
			Expr(
				utils.Map(
					strings.Split(expression, "."),
					func(_ int, node string) any { return node },
				)...,
			),
		),
		Const(value),
	)
}
