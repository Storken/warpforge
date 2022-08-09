package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/json"
	"github.com/warpfork/warpforge/pkg/workspace"
	"github.com/warpfork/warpforge/wfapi"
)

// special file names for plot and module files
// these are json files with special formatting for detection
const PLOT_FILE_NAME = "plot.wf"
const MODULE_FILE_NAME = "module.wf"

// Returns the file type, which is the file name without extension
// e.g., formula.wf -> formula, module.wf -> module, etc...
func getFileType(name string) (string, error) {
	split := strings.Split(filepath.Base(name), ".")
	if len(split) != 2 {
		// ignore files without extensions
		return "", fmt.Errorf("unsupported file: %q", name)
	}
	return split[0], nil
}

// helper function for finding the path to internally used binaries (e.g, rio, runc)
func binPath(bin string) (string, error) {
	path, override := os.LookupEnv("WARPFORGE_PATH")
	if override {
		return filepath.Join(path, bin), nil
	}

	path, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(path), bin), nil
}

// Opens the default WorkspaceSet.
// This consists of:
// stack: a workspace stack starting at the current working directory,
// root workspace: the first marked root workspace in the stack, or the home workspace if none are marked,
// home workspace: the workspace at the user's homedir
func openWorkspaceSet() (workspace.WorkspaceSet, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return workspace.WorkspaceSet{}, fmt.Errorf("failed to get working directory: %s", err)
	}

	wss, err := workspace.OpenWorkspaceSet(os.DirFS("/"), "", pwd[1:])
	if err != nil {
		return workspace.WorkspaceSet{}, fmt.Errorf("failed to open workspace: %s", err)
	}
	return wss, nil
}

// takes a path to a plot file, returns a plot
func plotFromFile(filename string) (wfapi.Plot, error) {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return wfapi.Plot{}, err
	}

	plotCapsule := wfapi.PlotCapsule{}
	_, err = ipld.Unmarshal(f, json.Decode, &plotCapsule, wfapi.TypeSystem.TypeByName("PlotCapsule"))
	if err != nil {
		return wfapi.Plot{}, err
	}
	if plotCapsule.Plot == nil {
		return wfapi.Plot{}, fmt.Errorf("no v1 Plot in PlotCapsule")
	}

	return *plotCapsule.Plot, nil
}

// takes a path to a module file, returns a module
func moduleFromFile(filename string) (wfapi.Module, error) {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return wfapi.Module{}, err
	}

	moduleCapsule := wfapi.ModuleCapsule{}
	_, err = ipld.Unmarshal(f, json.Decode, &moduleCapsule, wfapi.TypeSystem.TypeByName("ModuleCapsule"))
	if err != nil {
		return wfapi.Module{}, err
	}
	if moduleCapsule.Module == nil {
		return wfapi.Module{}, fmt.Errorf("no v1 Module in ModuleCapsule")
	}

	return *moduleCapsule.Module, nil
}
