package wfapi

import (
	"fmt"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/schema"
)

func init() {
	TypeSystem.Accumulate(schema.SpawnStruct("Plot",
		[]schema.StructField{
			schema.SpawnStructField("inputs", "Map__LocalLabel__PlotInput", false, false),
			schema.SpawnStructField("steps", "Map__StepName__Step", false, false),
			schema.SpawnStructField("outputs", "Map__LocalLabel__PlotOutput", false, false),
		},
		schema.SpawnStructRepresentationMap(nil)))
	TypeSystem.Accumulate(schema.SpawnMap("Map__LocalLabel__PlotInput",
		"LocalLabel", "PlotInput", false))
	TypeSystem.Accumulate(schema.SpawnMap("Map__StepName__Step",
		"StepName", "Step", false))
	TypeSystem.Accumulate(schema.SpawnMap("Map__LocalLabel__PlotOutput",
		"LocalLabel", "PlotOutput", false))
	TypeSystem.Accumulate(schema.SpawnString("StepName"))
	TypeSystem.Accumulate(schema.SpawnString("LocalLabel"))

}

type Plot struct {
	Inputs struct {
		Keys   []LocalLabel
		Values map[LocalLabel]PlotInput
	}
	Steps struct {
		Keys   []StepName
		Values map[StepName]Step
	}
	Outputs struct {
		Keys   []LocalLabel
		Values map[LocalLabel]PlotOutput
	}
}

// StepName is for assigning string names to Steps in a Plot.
// StepNames will be part of wiring things together using Pipes.
//
// Must not contain ':' charaters or unprintables or whitespace.
type StepName string

// LocalLabel is for referencing data within a Plot.
// Input data gets assigned a LocalLabel;
// Pipes pull info from a LocalLabel (possibly together with a StepName to scope it);
// when a Step is evaluated (e.g. turned into a Formula, executed, and produces results),
// the results will become identifiable by a LocalLabel (scoped by the StepName).
//
// (LocalLabel and OutputName are essentially the same thing: an OutputName
// gets casted to being considered a LocalLabel when a Formula's results are hoisted
// into the Plot.)
//
// Must not contain ':' charaters or unprintables or whitespace.
type LocalLabel string

func init() {
	TypeSystem.Accumulate(schema.SpawnUnion("PlotInput",
		[]schema.TypeName{
			"PlotInputSimple",
			"PlotInputComplex",
		},
		schema.SpawnUnionRepresentationKinded(map[ipld.Kind]schema.TypeName{
			ipld.Kind_String: "PlotInputSimple",
			ipld.Kind_Map:    "PlotInputComplex",
		})))
	TypeSystem.Accumulate(schema.SpawnUnion("PlotInputSimple",
		[]schema.TypeName{
			"WareID",
			"Mount",
			"String",
			"Pipe",
			"CatalogRef",
			"Ingest",
		},
		schema.SpawnUnionRepresentationStringprefix("", map[string]schema.TypeName{
			"ware:":    "WareID",
			"mount:":   "Mount",
			"literal:": "String",
			"pipe:":    "Pipe",
			"catalog:": "CatalogRef",
			"ingest:":  "Ingest",
		})))
	TypeSystem.Accumulate(schema.SpawnStruct("PlotInputComplex",
		[]schema.StructField{
			schema.SpawnStructField("basis", "PlotInputSimple", false, false),
			schema.SpawnStructField("filters", "FilterMap", false, false),
		},
		schema.SpawnStructRepresentationMap(nil)))
}

type PlotInput struct {
	PlotInputSimple  *PlotInputSimple
	PlotInputComplex *PlotInputComplex
}

func (pi *PlotInput) Basis() *PlotInputSimple {
	switch {
	case pi.PlotInputSimple != nil:
		return pi.PlotInputSimple
	case pi.PlotInputComplex != nil:
		return &pi.PlotInputComplex.Basis
	default:
		panic("unreachable")
	}
}

type PlotInputSimple struct {
	WareID     *WareID
	Mount      *Mount
	Literal    *string
	Pipe       *Pipe
	CatalogRef *CatalogRef
	Ingest     *Ingest
}

type PlotInputComplex struct {
	Basis   PlotInputSimple
	Filters FilterMap
}

func init() {
	TypeSystem.Accumulate(schema.SpawnUnion("PlotOutput",
		[]schema.TypeName{
			"Pipe",
		},
		schema.SpawnUnionRepresentationStringprefix("", map[string]schema.TypeName{
			"pipe:": "Pipe",
		})))
}

type PlotOutput struct {
	Pipe *Pipe
}

func init() {
	TypeSystem.Accumulate(schema.SpawnStruct("Pipe",
		[]schema.StructField{
			schema.SpawnStructField("stepName", "StepName", false, false),
			schema.SpawnStructField("label", "LocalLabel", false, false),
		},
		schema.SpawnStructRepresentationStringjoin(":")))
}

type Pipe struct {
	StepName StepName
	Label    LocalLabel
}

func init() {
	TypeSystem.Accumulate(schema.SpawnUnion("Step",
		[]schema.TypeName{
			"Plot",
			"Protoformula",
		},
		schema.SpawnUnionRepresentationKeyed(map[string]schema.TypeName{
			"plot":         "Plot",
			"protoformula": "Protoformula",
		})))

	TypeSystem.Accumulate(schema.SpawnStruct("Protoformula",
		[]schema.StructField{
			schema.SpawnStructField("inputs", "Map__SandboxPort__PlotInput", false, false),
			schema.SpawnStructField("action", "Action", false, false),
			schema.SpawnStructField("outputs", "Map__LocalLabel__GatherDirective", false, false),
		},
		schema.SpawnStructRepresentationMap(nil)))
	TypeSystem.Accumulate(schema.SpawnMap("Map__SandboxPort__PlotInput",
		"SandboxPort", "PlotInput", false))
	TypeSystem.Accumulate(schema.SpawnMap("Map__LocalLabel__GatherDirective",
		"LocalLabel", "GatherDirective", false))
}

type Step struct {
	Plot         *Plot
	Protoformula *Protoformula
}

type Protoformula struct {
	Inputs struct {
		Keys   []SandboxPort
		Values map[SandboxPort]PlotInput
	}
	Action  Action
	Outputs struct {
		Keys   []LocalLabel
		Values map[LocalLabel]GatherDirective
	}
}

func init() {
	TypeSystem.Accumulate(schema.SpawnStruct("CatalogRef",
		[]schema.StructField{
			schema.SpawnStructField("moduleName", "ModuleName", false, false),
			schema.SpawnStructField("releaseName", "String", false, false),
			schema.SpawnStructField("itemName", "String", false, false),
		},
		schema.SpawnStructRepresentationStringjoin(":")))
	TypeSystem.Accumulate(schema.SpawnString("ModuleName"))
	TypeSystem.Accumulate(schema.SpawnString("ReleaseName"))
	TypeSystem.Accumulate(schema.SpawnString("ItemLabel"))
}

type ModuleName string
type ReleaseName string
type ItemLabel string

type CatalogRef struct {
	ModuleName  ModuleName
	ReleaseName ReleaseName
	ItemName    ItemLabel
}

func (c *CatalogRef) String() string {
	return fmt.Sprintf("catalog:%s:%s:%s", c.ModuleName, c.ReleaseName, c.ItemName)
}

func init() {
	TypeSystem.Accumulate(schema.SpawnMap("PlotResults",
		"LocalLabel", "WareID", false))
}

type PlotResults struct {
	Keys   []LocalLabel
	Values map[LocalLabel]WareID
}
