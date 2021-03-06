package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

func main() {
	// local schemas to process, add more entries here if you want to test more providers
	// but know that filter spec is written relative to the filepath...
	schemaPaths := []string{"./kube.json", "./azure-native.json"}

	for _, schemaPath := range schemaPaths {
		fmt.Printf("Building filter for: %v...\n\n", schemaPath)

		pkg, err := getSchemaPackage(schemaPath)
		if err != nil {
			fmt.Printf("failed to parse schema: %v", schemaPath)
			panic(err)
		}

		nodes, err := collectNodes(pkg)
		if err != nil {
			panic(err)
		}

		filterSpec, err := buildFilterSpec(nodes)
		if err != nil {
			panic(err)
		}

		// compute some stats on filters
		modules := len(filterSpec.Nodes)
		subModules := 0
		leafNodes := 0

		// manually traverse to depth 3...
		for _, m := range filterSpec.Nodes {
			subModules += len(m.Children)
			for _, s := range m.Children {
				leafNodes += len(s.Children)
			}
		}

		fmt.Printf("Filter stats for: %v\n", schemaPath)
		fmt.Printf("Module Count:        %v\n", modules)
		fmt.Printf("Sub-module Count:     %v\n", subModules)
		fmt.Printf("Leaf-node Count:      %v\n", leafNodes)
		fmt.Printf("Total rendered nodes: %v\n\n", modules+subModules+leafNodes)
		fmt.Println()

		vv, _ := json.MarshalIndent(filterSpec, "", "    ")
		fname := strings.TrimSuffix(schemaPath, ".json") + "-filter.json"
		fmt.Printf("Writing filter spec output to: %v\n\n", fname)
		ioutil.WriteFile(fname, vv, 0777)
	}
}

type FilterNode struct {
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	ParentName string       `json:"parentName"`
	Children   []FilterNode `json:"children"`
	Token      string       `json:"token"`
}

type FilterSpec struct {
	Nodes []FilterNode `json:"nodes"`
}

type Node struct {
	Module     string
	SubModule  string
	Name       string
	IsFunction bool
	Token      string
}

func buildFilterSpec(nodes []Node) (FilterSpec, error) {
	filterSpec := FilterSpec{}

	modSpec := FilterNode{
		Name:       nodes[0].Module,
		Type:       "module",
		ParentName: "",
	}
	subModule := FilterNode{
		Name:       nodes[0].SubModule,
		Type:       "module",
		ParentName: modSpec.Name,
	}
	// run through nodes in sorted order and build up the filter tree
	for _, n := range nodes {
		if n.Module != modSpec.Name {
			// done with this module, pop into the filter spec
			modSpec.Children = append(modSpec.Children, subModule)
			filterSpec.Nodes = append(filterSpec.Nodes, modSpec)
			modSpec = FilterNode{
				Name:       n.Module,
				Type:       "module",
				ParentName: "",
			}
			subModule = FilterNode{
				Name:       n.SubModule,
				Type:       "module",
				ParentName: modSpec.Name,
			}
		}

		if n.SubModule != subModule.Name {
			// done with this submod, pop it into the module
			modSpec.Children = append(modSpec.Children, subModule)
			subModule = FilterNode{
				Name:       n.SubModule,
				Type:       "module",
				ParentName: modSpec.Name,
			}
		}

		leaf := FilterNode{
			Name:       n.Name,
			ParentName: subModule.Name,
			Token:      n.Token,
		}

		// pop the resource/function into the submod
		if n.IsFunction {
			leaf.Type = "function"
		} else {
			leaf.Type = "resource"
		}

		subModule.Children = append(subModule.Children, leaf)
	}

	// push the final submod and module into the filter spec
	modSpec.Children = append(modSpec.Children, subModule)
	filterSpec.Nodes = append(filterSpec.Nodes, modSpec)

	return filterSpec, nil
}

// turns schema into a colleciton of nodes
func collectNodes(pkg *schema.Package) ([]Node, error) {
	nodes := []Node{}

	for _, r := range pkg.Resources {
		parts := strings.Split(r.Token, ":")
		if len(parts) != 3 {
			return nil, errors.New("couldn't parse package token")
		}
		modPart := parts[1]
		modParts := strings.Split(modPart, "/")
		numSubMods := len(modParts) - 1
		// default empty name for module level resources
		subModuleName := ""
		// current code hardcodes a module/submodule/resource hierarchy which is (mostly?) true in practice
		// but not a limitation of the schema. We could see community components/resources that fail this check
		if numSubMods == 1 {
			subModuleName = modParts[1]
		} else if numSubMods > 1 {
			return nil, errors.New("did not expect multiple sub modules: " + r.Token)
		}

		node := Node{
			Module:     modParts[0],
			SubModule:  subModuleName,
			Name:       parts[2],
			IsFunction: false,
			Token:      r.Token,
		}
		nodes = append(nodes, node)
	}

	for _, f := range pkg.Functions {
		parts := strings.Split(f.Token, ":")
		if len(parts) != 3 {
			return nil, errors.New("couldn't parse package token")
		}
		modPart := parts[1]
		modParts := strings.Split(modPart, "/")
		numSubMods := len(modParts) - 1
		// default empty name for module level resources
		subModuleName := ""
		// current code hardcodes a module/submodule/resource hierarchy which is (mostly?) true in practice
		// but not a limitation of the schema. We could see community components/resources that fail this check
		if numSubMods == 1 {
			subModuleName = modParts[1]
		} else if numSubMods > 1 {
			return nil, errors.New("did not expect multiple sub modules: " + f.Token)
		}

		node := Node{
			Module:     modParts[0],
			SubModule:  subModuleName,
			Name:       parts[2],
			IsFunction: true,
			Token:      f.Token,
		}
		nodes = append(nodes, node)
	}

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Module == nodes[j].Module {
			if nodes[i].SubModule == nodes[j].SubModule {
				return nodes[i].Name < nodes[j].Name
			}

			return nodes[i].SubModule < nodes[j].SubModule
		}
		return nodes[i].Module < nodes[j].Module
	})
	return nodes, nil
}

func getSchemaPackage(path string) (*schema.Package, error) {
	schemaBytes, err := os.ReadFile(filepath.FromSlash(path))
	if err != nil {
		return nil, err
	}

	var pSpec schema.PackageSpec
	err = json.Unmarshal(schemaBytes, &pSpec)
	if err != nil {
		return nil, err
	}

	p, err := schema.ImportSpec(pSpec, nil)
	if err != nil {
		return nil, err
	}

	return p, nil
}
