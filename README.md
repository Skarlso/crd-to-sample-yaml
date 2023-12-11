# crd-to-sample-yaml or cty ( city )

Generate a sample YAML file from a CRD definition.

## Getting started
- Prerequisites: Go installed on your machine. (Check out this link for details: https://go.dev/doc/install)
- Clone the repository
- Execute `make build` to build the binary

Now you can simply run:

```
cty generate -c delivery.krok.app_krokcommands
```

Optionally, define a URL at which a CRD is located:

```
cty generate -u https://raw.githubusercontent.com/kubernetes-sigs/cluster-api-provider-aws/main/config/crd/bases/infrastructure.cluster.x-k8s.io_awsclusters.yaml
```

`cty` does not support authentication modes, therefore the CRD needs to be publicly accessible.

This will result in a file similar to this:

```yaml
apiVersion: delivery.krok.app/v1alpha1
kind: KrokCommand
metadata: {}
spec:
  commandHasOutputToWrite: true
  dependencies: ["string"]
  enabled: true
  image: string
  platforms: ["string"]
  readInputFromSecret:
    name: string
    namespace: string
  schedule: string
status: {}
```

A single file will be created containing all versions in the CRD delimited by `---`.

Optionally, you can provide the flag `-s` which will output the generated content to `stdout`.

Future plans include generating proper, schema validated values for all fields.

## WASM frontend

There is a WASM based frontend that can be started by navigating into the `wasm` folder and running the following make
target:

```shell
make run
```

This will start a front-end that can be used to paste in and parse CRDs.

## Comments

Comments can be added to each line of the generated YAML content where descriptions are available. This looks something
like this:

```yaml
# APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
# Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
kind: AWSCluster
metadata: {}
# AWSClusterSpec defines the desired state of an EC2-based Kubernetes cluster.
spec:
  # AdditionalTags is an optional set of tags to add to AWS resources managed by the AWS provider, in addition to the ones added by default.
  additionalTags: {}
  # Bastion contains options to configure the bastion host.
  bastion:
  ...
```

To add comments simply run cty with:
```console
cty generate -c sample-crd/infrastructure.cluster.x-k8s.io_awsclusters.yaml --comments
```

The frontend also has a checkbox to add comments to the generated yaml output.

## Showcase

![frontpage](./imgs/frontend.png)

![parsed1](./imgs/parsed1.png)
![parsed2](./imgs/parsed2.png)
