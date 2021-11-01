package plotexec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/warpfork/warpforge/pkg/formulaexec"
	"github.com/warpfork/warpforge/pkg/workspace"
	"github.com/warpfork/warpforge/wfapi"
)

type pipeMap map[wfapi.StepName]map[wfapi.LocalLabel]wfapi.FormulaInput

// Returns a WareID for a given StepName and LocalLabel, if it exists
func (m pipeMap) lookup(stepName wfapi.StepName, label wfapi.LocalLabel) (*wfapi.FormulaInput, error) {
	if step, ok := m[stepName]; ok {
		if input, ok := step[label]; ok {
			// located a valid input
			return &input, nil
		} else {
			// located step, but no input by label
			if stepName == "" {
				return nil, fmt.Errorf("no label '%s' in plot inputs ('pipe::%s' not defined)", label, label)
			} else {
				return nil, fmt.Errorf("no label '%s' for step '%s' (pipe:%s:%s not defined)", label, stepName, stepName, label)
			}
		}
	} else {
		// did not locate step
		return nil, fmt.Errorf("no step '%s'", stepName)
	}
}

// Resolves a PlotInput to a WareID and optionally a WarehouseAddr.
// This will resolve various input types (Pipes, CatalogRefs, etc...)
// to allow them to be used in a Formula.
func plotInputToFormulaInput(wss []*workspace.Workspace, plotInput wfapi.PlotInput, pipeCtx pipeMap) (wfapi.FormulaInput, *wfapi.WarehouseAddr, error) {
	basis, addr, err := plotInputToFormulaInputSimple(wss, plotInput, pipeCtx)
	if err != nil {
		return wfapi.FormulaInput{}, nil, err
	}

	switch {
	case plotInput.PlotInputSimple != nil:
		return wfapi.FormulaInput{
			FormulaInputSimple: &basis,
		}, addr, nil
	case plotInput.PlotInputComplex != nil:
		return wfapi.FormulaInput{
			FormulaInputComplex: &wfapi.FormulaInputComplex{
				Basis:   basis,
				Filters: plotInput.PlotInputComplex.Filters,
			}}, addr, nil
	default:
		panic("unreachable")
	}
}

func plotInputToFormulaInputSimple(wss []*workspace.Workspace, plotInput wfapi.PlotInput, pipeCtx pipeMap) (wfapi.FormulaInputSimple, *wfapi.WarehouseAddr, error) {
	var basis wfapi.PlotInputSimple

	switch {
	case plotInput.PlotInputSimple != nil:
		basis = *plotInput.PlotInputSimple
	case plotInput.PlotInputComplex != nil:
		basis = plotInput.PlotInputComplex.Basis
	default:
		return wfapi.FormulaInputSimple{}, nil, fmt.Errorf("invalid plot input")
	}

	switch {
	case basis.WareID != nil:
		// convert WareID PlotInput to FormulaInput
		return wfapi.FormulaInputSimple{
			WareID: basis.WareID,
		}, nil, nil
	case basis.Mount != nil:
		// convert WareID PlotInput to FormulaInput
		return wfapi.FormulaInputSimple{
			Mount: basis.Mount,
		}, nil, nil
	case basis.CatalogRef != nil:
		// search the warehouse stack for this CatalogRef
		// this will return the WareID and WarehouseAddr to use
		var wareId *wfapi.WareID
		for _, ws := range wss {
			wareId, wareAddr, err := ws.GetCatalogWare(*basis.CatalogRef)
			if err != nil {
				return wfapi.FormulaInputSimple{}, nil, err
			}
			if wareId != nil {
				// found a matching ware in a catalog, stop searching
				return wfapi.FormulaInputSimple{
					WareID: wareId,
				}, wareAddr, nil
			}
		}

		if wareId == nil {
			// failed to find a match in the catalog
			return wfapi.FormulaInputSimple{},
				nil,
				fmt.Errorf("no definition found for %q", basis.CatalogRef.String())
		}
	case basis.Pipe != nil:
		// resolve the pipe to a WareID using the pipeCtx
		input, err := pipeCtx.lookup(basis.Pipe.StepName, basis.Pipe.Label)
		return *input.Basis(), nil, err

	case basis.Ingest != nil && basis.Ingest.GitIngest != nil:
		input := wfapi.FormulaInputSimple{
			WareID: &wfapi.WareID{
				Packtype: "git",
				Hash:     basis.Ingest.GitIngest.Ref,
			},
		}
		path, err := filepath.Abs(basis.Ingest.GitIngest.HostPath)
		if err != nil {
			return wfapi.FormulaInputSimple{}, nil, fmt.Errorf("failed to get absolute path for git ingest")
		}

		// NOTE: we should be using go-git and not git exec here.
		// however, this does work, because it will be checked out and owned by the same user that invokes runc,
		// resulting in all files being owned by uid 0 within the container. this doesn't work for tarballs (which
		// preserve persmissions) but does work for git.
		// this checks the repo out to
		ws, _ := workspace.OpenHomeWorkspace(os.DirFS("/"))

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		gitCmd := exec.Command(
			"git",
			"--git-dir",
			filepath.Join(path, ".git"),
			"rev-parse",
			basis.Ingest.GitIngest.Ref)
		gitCmd.Stdout = &stdout
		gitCmd.Stderr = &stderr
		err = gitCmd.Run()
		if err != nil {
			fmt.Println(stderr.String())
			return input, nil, fmt.Errorf("git rev-parse failed: %s", err)
		}
		hash := strings.TrimSpace(stdout.String())
		input.WareID.Hash = hash
		input.WareID.Packtype = "git"

		cachePath, err := ws.CachePath(*input.WareID)
		if err != nil {
			return input, nil, err
		}
		if _, err = os.Stat(cachePath); os.IsNotExist(err) {
			var stdout bytes.Buffer
			cmd := exec.Command("git", "clone", "file://"+path, cachePath)
			cmd.Stdout = &stdout
			cmd.Stderr = &stdout
			err = cmd.Run()
			if err != nil {
				fmt.Println(stdout.String())
				return input, nil, fmt.Errorf("git failed: %s", err)
			}
		}
		return input, nil, nil

	}
	return wfapi.FormulaInputSimple{}, nil, fmt.Errorf("invalid type in plot input")
}

func execProtoformula(wss []*workspace.Workspace, pf wfapi.Protoformula, ctx wfapi.FormulaContext, pipeCtx pipeMap) (wfapi.RunRecord, error) {
	// create an empty Formula and FormulaContext
	formula := wfapi.Formula{
		Action: pf.Action,
	}
	formula.Inputs.Values = make(map[wfapi.SandboxPort]wfapi.FormulaInput)
	formula.Outputs.Values = make(map[wfapi.OutputName]wfapi.GatherDirective)

	// get the home workspace from the workspace stack
	var homeWs *workspace.Workspace
	for _, ws := range wss {
		if ws.IsHomeWorkspace() {
			homeWs = ws
			break
		}
	}

	// convert Protoformula inputs (of type PlotInput) to FormulaInputs
	for sbPort, plotInput := range pf.Inputs.Values {
		formula.Inputs.Keys = append(formula.Inputs.Keys, sbPort)
		input, wareAddr, err := plotInputToFormulaInput(wss, plotInput, pipeCtx)
		if err != nil {
			return wfapi.RunRecord{}, err
		}
		formula.Inputs.Values[sbPort] = input
		if wareAddr != nil {
			// input specifies a WarehouseAddr, add it to the formula's context
			ctx.Warehouses.Keys = append(ctx.Warehouses.Keys, *input.Basis().WareID)
			ctx.Warehouses.Values[*input.Basis().WareID] = *wareAddr
		}
	}

	// convert Protoformula outputs to Formula outputs
	for label, gatherDirective := range pf.Outputs.Values {
		label := wfapi.OutputName(label)
		formula.Outputs.Keys = append(formula.Outputs.Keys, label)
		formula.Outputs.Values[label] = gatherDirective
	}

	// execute the derived formula
	rr, err := formulaexec.Exec(homeWs, wfapi.FormulaAndContext{
		Formula: formula,
		Context: &ctx,
	})
	return rr, err
}

func Exec(wss []*workspace.Workspace, plot wfapi.Plot) (wfapi.PlotResults, error) {
	pipeCtx := make(pipeMap)
	results := wfapi.PlotResults{}

	// collect the plot inputs
	// these have an empty string for the step name (e.g., `pipe::foo`)
	pipeCtx[""] = make(map[wfapi.LocalLabel]wfapi.FormulaInput)
	inputContext := wfapi.FormulaContext{}
	inputContext.Warehouses.Values = make(map[wfapi.WareID]wfapi.WarehouseAddr)
	for name, input := range plot.Inputs.Values {
		input, wareAddr, err := plotInputToFormulaInput(wss, input, pipeCtx)
		if err != nil {
			return results, err
		}
		pipeCtx[""][name] = input
		if wareAddr != nil {
			// input specifies an address, add it to the context
			inputContext.Warehouses.Keys = append(inputContext.Warehouses.Keys, *input.Basis().WareID)
			inputContext.Warehouses.Values[*input.Basis().WareID] = *wareAddr
		}
	}

	// determine step execution order
	stepsOrdered, err := OrderSteps(plot)
	if err != nil {
		return results, err
	}

	// execute the plot steps
	for _, name := range stepsOrdered {
		step := plot.Steps.Values[name]
		switch {
		case step.Protoformula != nil:
			// execute Protoformula step
			rr, err := execProtoformula(wss, *step.Protoformula, inputContext, pipeCtx)
			if err != nil {
				return results, fmt.Errorf("failed to execute protoformula for step %s: %s", name, err)
			}
			// accumulate the results of the Protoformula our map of Pipes
			pipeCtx[name] = make(map[wfapi.LocalLabel]wfapi.FormulaInput)
			for result, input := range rr.Results.Values {
				pipeCtx[name][wfapi.LocalLabel(result)] = wfapi.FormulaInput{
					FormulaInputSimple: &input,
				}
			}
		case step.Plot != nil:
			// execute plot step
			stepResults, err := Exec(wss, *step.Plot)
			if err != nil {
				return results, fmt.Errorf("failed to execute plot for step %s: %s", name, err)
			}
			// accumulate the results of the Plot into our map of Pipes
			pipeCtx[name] = make(map[wfapi.LocalLabel]wfapi.FormulaInput)
			for result, wareId := range stepResults.Values {
				pipeCtx[name][wfapi.LocalLabel(result)] = wfapi.FormulaInput{
					FormulaInputSimple: &wfapi.FormulaInputSimple{
						WareID: &wareId,
					},
				}
			}
		default:
			return results, fmt.Errorf("invalid step %s", name)
		}
	}

	// collect the outputs of this plot
	results.Values = make(map[wfapi.LocalLabel]wfapi.WareID)
	for name, output := range plot.Outputs.Values {
		result, err := pipeCtx.lookup(output.Pipe.StepName, output.Pipe.Label)
		if err != nil {
			return results, err
		}
		results.Keys = append(results.Keys, name)
		results.Values[name] = *result.Basis().WareID
	}
	return results, nil
}
