# Treedoc

This package takes pulumi schema (checked in as `./azure-native.json`, `./kube.json`) and converts it into a filter spec designed to be consumed by a tree-view filter control. It also computes stats about how many nodes will appear in the filter tree. 

This tool and format are designed for testing purposes. It does not for instance handle multiple nested submodules, or do things that the pulumi docs site does like trim `kubernetes.io` suffixes from module names.

## Running the tool

Run via a simple `go run main.go`:

```console
$ go run main.go
Building filter for: ./kube.json...

Filter stats for: ./kube.json
Module Count:        25
Sub-module Count:     51
Leaf-node Count:      214
Total rendered nodes: 290


Writing filter spec output to: ./kube-filter.json

Building filter for: ./azure-native.json...

Filter stats for: ./azure-native.json
Module Count:        169
Sub-module Count:     169
Leaf-node Count:      2727
Total rendered nodes: 3065


Writing filter spec output to: ./azure-native-filter.json
```

## FilterSpec Data Format

See `./kube-filter.json` or `./azure-native-filter.json` for resulting filter spec.

At a high level, there is a top level array `Nodes` (`name`, `parent`, `type`, `children`) that is sorted by Name. All children are also sorted by `name`. In practice, there are three levels in the hierarchy:

```
- Module 1
  - SubModule 1
    - LeafNode 1 (Resource | Function)
      ...
    - LeafNode N
    ...
  - SubModule N
...
- Module N
```

```json
{
    "nodes": [
        {
            "name": "admissionregistration.k8s.io",
            "type": "module",
            "parentName": "",
            "children": [
                {
                    "name": "v1",
                    "type": "module",
                    "parentName": "admissionregistration.k8s.io",
                    "children": [
                        {
                            "name": "MutatingWebhookConfiguration",
                            "type": "resource",
                            "parentName": "v1",
                            "children": null,
                            "token": "kubernetes:admissionregistration.k8s.io/v1:MutatingWebhookConfiguration"
                        },
                        {
                            "name": "MutatingWebhookConfigurationList",
                            "type": "resource",
                            "parentName": "v1",
                            "children": null,
                            "token": "kubernetes:admissionregistration.k8s.io/v1:MutatingWebhookConfigurationList"
                        },
                        {
                            "name": "ValidatingWebhookConfiguration",
                            "type": "resource",
                            "parentName": "v1",
                            "children": null,
                            "token": "kubernetes:admissionregistration.k8s.io/v1:ValidatingWebhookConfiguration"
                        },
                        {
                            "name": "ValidatingWebhookConfigurationList",
                            "type": "resource",
                            "parentName": "v1",
                            "children": null,
                            "token": "kubernetes:admissionregistration.k8s.io/v1:ValidatingWebhookConfigurationList"
                        }
                    ],
                    "token": ""
                },
            ]
        }
    ]
}
```
