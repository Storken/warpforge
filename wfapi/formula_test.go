package wfapi

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/ipld/go-ipld-prime/codec/json"
	"github.com/ipld/go-ipld-prime/node/bindnode"
)

func TestParseFormulaAndContext(t *testing.T) {
	serial := `{
	"formula": {
		"inputs": {
			"/mount/path": "ware:tar:qwerasdf",
			"$ENV_VAR": "literal:hello",
			"/more/mounts": {"basis": "ware:tar:fghjkl", "filters":{"uid":"10"}}
		},
		"action": {
			"exec": {
				"command": ["/bin/bash", "-c", "echo hey there"]
			}
		},
		"outputs": {
			"theoutputlabel": {
				"from": "/collect/here",
				"packtype": "tar"
			},
			"another": {
				"from": "$VAR",
			}
		}
	},
	"context": {
		"warehouses": {
			"tar:qwerasdf": "ca+file:///somewhere/",
			"tar:fghjkl": "ca+file:///elsewhere/"
		}
	}
}`

	np := bindnode.Prototype((*FormulaAndContext)(nil), TypeSystem.TypeByName("FormulaAndContext"))
	nb := np.Representation().NewBuilder()
	err := json.Decode(nb, strings.NewReader(serial))
	qt.Assert(t, err, qt.IsNil)
	n := bindnode.Unwrap(nb.Build()).(*FormulaAndContext)
	_ = n
}
