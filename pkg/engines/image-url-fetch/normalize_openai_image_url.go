package image_url_fetch

import (
	"encoding/json"
	"fmt"

	"github.com/buger/jsonparser"
)

type openaiStringImageURLTarget struct {
	msgIdx  int
	field   string // "content" or "reasoning_content"
	partIdx int
	url     string
}

// normalizeRawOpenAIImageURLStrings rewrites multipart parts where "image_url" is a JSON string
// into an object {"url":"..."}, using jsonparser only (no full ChatCompletionRequest round-trip).
// It is a cheap path aligned with collectFromOpenAIMessageContent / llm_router.
func normalizeRawOpenAIImageURLStrings(body []byte) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}
	_, dt, _, err := jsonparser.Get(body, "messages")
	if err != nil || dt != jsonparser.Array {
		return body, false, nil
	}
	targets, err := collectOpenAIStringImageURLTargets(body)
	if err != nil {
		return body, false, err
	}
	if len(targets) == 0 {
		return body, false, nil
	}
	out := body
	for _, t := range targets {
		obj, mErr := json.Marshal(map[string]string{"url": t.url})
		if mErr != nil {
			return body, false, mErr
		}
		var setErr error
		out, setErr = jsonparser.Set(out, obj,
			"messages", fmt.Sprintf("[%d]", t.msgIdx), t.field, fmt.Sprintf("[%d]", t.partIdx), "image_url")
		if setErr != nil {
			return body, false, setErr
		}
	}
	return out, true, nil
}

func collectOpenAIStringImageURLTargets(body []byte) ([]openaiStringImageURLTarget, error) {
	var targets []openaiStringImageURLTarget
	var collectErr error
	msgIdx := 0
	_, err := jsonparser.ArrayEach(body, func(msg []byte, dt jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}
		if collectErr != nil {
			return
		}
		if dt != jsonparser.Object {
			msgIdx++
			return
		}
		objErr := jsonparser.ObjectEach(msg, func(key []byte, val []byte, typ jsonparser.ValueType, offset int) error {
			ks := string(key)
			if ks != "content" && ks != "reasoning_content" {
				return nil
			}
			if typ != jsonparser.Array {
				return nil
			}
			partIdx := 0
			_, aerr := jsonparser.ArrayEach(val, func(part []byte, ptyp jsonparser.ValueType, offset int, perr error) {
				if perr != nil {
					return
				}
				defer func() { partIdx++ }()
				if ptyp != jsonparser.Object {
					return
				}
				if typStr, e := jsonparser.GetString(part, "type"); e == nil {
					if typStr != "" && typStr != "image_url" {
						return
					}
				}
				urlStr, e := jsonparser.GetString(part, "image_url")
				if e != nil || urlStr == "" {
					return
				}
				targets = append(targets, openaiStringImageURLTarget{
					msgIdx: msgIdx, field: ks, partIdx: partIdx, url: urlStr,
				})
			})
			return aerr
		})
		if objErr != nil {
			collectErr = objErr
			msgIdx++
			return
		}
		msgIdx++
	}, "messages")
	if err != nil {
		return nil, err
	}
	if collectErr != nil {
		return nil, collectErr
	}
	return targets, nil
}
