package engines

import (
	"net/http"

	"github.com/infinigence/octollm/pkg/octollm"
)

type AddHeaderEngine struct {
	PassThroughHeaders []string
	SetHeaders         map[string]string
	Next               octollm.Engine
}

var _ octollm.Engine = (*AddHeaderEngine)(nil)

func (e *AddHeaderEngine) Process(req *octollm.Request) (*octollm.Response, error) {
	if len(e.PassThroughHeaders) > 0 {
		recvHeaders, ok := octollm.GetCtxValue[http.Header](req, octollm.ContextKeyReceivedHeader)
		if ok {
			for _, k := range e.PassThroughHeaders {
				if v := recvHeaders.Get(k); v != "" {
					req.Header.Set(k, v)
				}
			}
		}
	}
	for k, v := range e.SetHeaders {
		req.Header.Set(k, v)
	}
	return e.Next.Process(req)
}
