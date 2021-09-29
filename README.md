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

At a high level, there is a top level entity `Modules` that has a sorted list of `SubModules` and each submodule has a sorted list of `Resources` and `Functions`. For example:

```json
{
    "Modules": [
        {
            "Name": "admissionregistration.k8s.io",
            "SubModules": [
                {
                    "Name": "v1",
                    "Resources": [
                        "MutatingWebhookConfiguration",
                        "MutatingWebhookConfigurationList",
                        "ValidatingWebhookConfiguration",
                        "ValidatingWebhookConfigurationList"
                    ],
                    "Functions": null
                },
                {
                    "Name": "v1beta1",
                    "Resources": [
                        "MutatingWebhookConfiguration",
                        "MutatingWebhookConfigurationList",
                        "ValidatingWebhookConfiguration",
                        "ValidatingWebhookConfigurationList"
                    ],
                    "Functions": null
                }
            ]
        }
    ]
}
```
