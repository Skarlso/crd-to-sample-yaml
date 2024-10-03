# crd-to-sample-yaml or cty ( city )

![logo](./imgs/cty_logo.png)

Generate a sample YAML file from a CRD definition.

## CRD Testing using CTY

For more information about how to use `cty` for helm-like unit testing your CRD schemas,
please follow the [How to test CRDs with CTY Readme](./crd-testing-README.md).

![crd-unittest-sample-output](./imgs/crd-unittest-outcome.png)

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

### HTML output

It's possible to generate a pre-rendered HTML based output for self-hosting what the website produces online.

To get an HTML output provide the format flag like this:

```
cty generate -c delivery.krok.app_krokcommands --comments --format html
```

![parsed1_cli](./imgs/parsed1_cli.png)
![parsed2_cli](./imgs/parsed2_cli.png)

In case of multiple CRD files being parsed using a `folder` target, the CRDs will be listed
in collapsed drop-down menus where their KIND is the title.

![parsed3_cli](./imgs/parsed3_cli.png)

### Minimal required CRD sample

It's possible to generate a sample YAML for a CRD that will make the CRD validation pass. Meaning, it will only contain
samples for fields that are actually required. All other fields will be ignored.

For example, a CRD having a single required field with an example and the rest being optional would generate something
like this:

```yaml
apiVersion: delivery.krok.app/v1alpha1
kind: KrokCommand
spec:
  image: "krok-hook/slack-notification:v0.0.1"
```

To run cty with minimal required fields, pass in `--minimal` to the command like this:

```
cty generate -c delivery.krok.app_krokcommands --comments --minimal --format html
```

### Folder source

To parse multiple CRDs in a single folder, just pass in the whole folder like this:

```
cty generate -r folder
```

Any other flag will work as before.

## WASM frontend

There is a WASM based frontend that can be started by navigating into the `wasm` folder and running the following make
target:

```shell
make run
```

This will start a front-end that can be used to paste in and parse CRDs.

## Shareable Link

It's possible to provide a link that can be shared using a url parameter like this:

```
https://crdtoyaml.com/share?url=https://raw.githubusercontent.com/Skarlso/crd-to-sample-yaml/main/sample-crd/infrastructure.cluster.x-k8s.io_awsclusters.yaml
```

Will load the content, or display an appropriate error message.

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

## Templated CRDs

It's possible to provide a templated CRD like this one for flux: [Helm Controller](https://raw.githubusercontent.com/fluxcd-community/helm-charts/main/charts/flux2/templates/helm-controller.crds.yaml).

It contains template definition like:

```yaml
{{- if and .Values.installCRDs .Values.helmController.create }}
```

These are trimmed so that the CRD parses correctly. Any values that might be in-lined are replaced with `replaced`.
This is done to avoid trying to parse a breaking yaml.

Things like this:
```yaml
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.15.0
  labels:
    app.kubernetes.io/component: helm-controller
    app.kubernetes.io/instance: {{ .Release.Namespace }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    app.kubernetes.io/part-of: flux
    app.kubernetes.io/version: {{ .Chart.AppVersion }}
    helm.sh/chart: '{{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}'
  name: helmreleases.helm.toolkit.fluxcd.io
```

Where some templated value isn't escaped with `'` will create an invalid YAML that fails to parse.

## Showcase

![frontpage](./imgs/frontend.png)

Parsed Yaml output on the website:

![parsed1](./imgs/parsed1_website.png)
![parsed2](./imgs/parsed2_website.png)
![parsed3](./imgs/parsed1_sample_website.png)
