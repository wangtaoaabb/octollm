package engines

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/infinigence/octollm/pkg/exprenv"
	"github.com/infinigence/octollm/pkg/octollm"
)

// AddHeaderByExprEngine adds HTTP headers based on expr-lang expressions
// The expressions are evaluated at request time using the exprenv context
type AddHeaderByExprEngine struct {
	// HeaderExprs maps header names to compiled expressions
	// The expressions can access request data via exprenv (req.Header(), req.RawReq(), etc.)
	HeaderExprs map[string]*vm.Program
	Next        octollm.Engine
}

var _ octollm.Engine = (*AddHeaderByExprEngine)(nil)

// NewAddHeaderByExprEngine creates a new AddHeaderByExprEngine with compiled expressions
// exprMap: map of header name to expression string
// Returns error if any expression fails to compile
func NewAddHeaderByExprEngine(exprMap map[string]string, next octollm.Engine) (*AddHeaderByExprEngine, error) {
	if len(exprMap) == 0 {
		return &AddHeaderByExprEngine{
			HeaderExprs: make(map[string]*vm.Program),
			Next:        next,
		}, nil
	}

	compiledExprs := make(map[string]*vm.Program, len(exprMap))
	for headerName, exprStr := range exprMap {
		program, err := expr.Compile(exprStr, expr.Env(exprenv.Sentinel))
		if err != nil {
			return nil, fmt.Errorf("failed to compile expression for header %s: %w", headerName, err)
		}
		compiledExprs[headerName] = program
	}

	return &AddHeaderByExprEngine{
		HeaderExprs: compiledExprs,
		Next:        next,
	}, nil
}

func (e *AddHeaderByExprEngine) Process(req *octollm.Request) (*octollm.Response, error) {
	if len(e.HeaderExprs) == 0 {
		return e.Next.Process(req)
	}

	env := exprenv.Get(req)
	for headerName, program := range e.HeaderExprs {
		output, err := expr.Run(program, env)
		if err != nil {
			continue
		}

		headerValue := fmt.Sprintf("%v", output)
		req.Header.Set(headerName, headerValue)
	}

	return e.Next.Process(req)
}
