# crd-to-sample-yaml

Generate a sample YAML file from a CRD definition.

Simply run:

```
crd-to-yaml delivery.krok.app_krokcommands
```

This will result in a file similar to this:

```yaml
apiVersion: delivery.krok.app/v1alpha1
kind: KrokCommand
metadata: {}
spec:
  commandHasOutputToWrite: true
  dependencies: []
  enabled: true
  image: string
  platforms: []
  readInputFromSecret:
    name: string
    namespace: string
  schedule: string
status: {}
```

Each version will contain its own file output.

Future plans include generating proper, schema validated values for all fields.
