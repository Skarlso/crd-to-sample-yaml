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

There is also an option to run this as a server. Run:

```
cty serve
```

This will start a front-end that can be used to paste in and parse CRDs.

## Showcase

![frontpage](./imgs/frontend.png)

![parsed1](./imgs/parsed1.png)
![parsed2](./imgs/parsed2.png)
