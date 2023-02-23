package wfapi

import (
	"fmt"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"

	_ "github.com/warptools/warpforge/pkg/testutil"
)

type CatalogLineage struct {
	Name     string
	Metadata struct {
		Keys   []string
		Values map[string]string
	}
	Releases []CatalogRelease
}

type CatalogMirrorsByWare struct {
	Keys   []WareID
	Values map[WareID][]WarehouseAddr
}

type CatalogMirrorsByModule struct {
	Keys   []ModuleName
	Values map[ModuleName]CatalogMirrorsByPacktype
}

type CatalogMirrorsByPacktype struct {
	Keys   []Packtype
	Values map[Packtype][]WarehouseAddr
}

type CatalogMirrorsCapsule struct {
	CatalogMirrors *CatalogMirrors
}

type CatalogMirrors struct {
	ByWare   *CatalogMirrorsByWare
	ByModule *CatalogMirrorsByModule
}

type CatalogReleaseCID string

type Catalog struct {
	Keys   []ModuleName
	Values map[ModuleName]CatalogModule
}

type CatalogModuleCapsule struct {
	CatalogModule *CatalogModule
}

type CatalogModule struct {
	Name     ModuleName
	Releases struct {
		Keys   []ReleaseName
		Values map[ReleaseName]CatalogReleaseCID
	}
	Metadata struct {
		Keys   []string
		Values map[string]string
	}
}

type CatalogRelease struct {
	ReleaseName ReleaseName
	Items       struct {
		Keys   []ItemLabel
		Values map[ItemLabel]WareID
	}
	Metadata struct {
		Keys   []string
		Values map[string]string
	}
}

func (rel *CatalogRelease) Cid() CatalogReleaseCID {
	// convert parsed release to node
	nRelease := bindnode.Wrap(rel, TypeSystem.TypeByName("CatalogRelease"))

	// compute CID of parsed release data
	lsys := cidlink.DefaultLinkSystem()
	lnk, errRaw := lsys.ComputeLink(cidlink.LinkPrototype{Prefix: cid.Prefix{
		Version:  1,    // Usually '1'.
		Codec:    0x71, // 0x71 means "dag-cbor" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhType:   0x20, // 0x20 means "sha2-384" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhLength: 48,   // sha2-384 hash has a 48-byte sum.
	}}, nRelease.(schema.TypedNode).Representation())
	if errRaw != nil {
		// panic! this should never fail unless IPLD is broken
		panic(fmt.Sprintf("Fatal IPLD Error: lsys.ComputeLink failed for CatalogRelease: %s", errRaw))
	}
	cid, errRaw := lnk.(cidlink.Link).StringOfBase('z')
	if errRaw != nil {
		panic(fmt.Sprintf("Fatal IPLD Error: failed to encode CID for CatalogRelease: %s", errRaw))
	}
	return CatalogReleaseCID(cid)
}

type ReplayCapsule struct {
	Plot *Plot
}
