# CRD Testing with CTY

From version `v0.8.0` cty supports the command `test`.

`test` supports testing CRD schemas against snapshots of generated YAML files that satisfy the schema
or small snippets of yaml strings.

Adding this test to your CRDs makes sure that any modification on the CRD will not break a generated snapshot of the
CRD. This is from version to version, meaning the tests will make sure that API version is respected.

## Example

Let's look at an example.

Consider the following test suite definition that loosely follows the syntax of helm unittest definitions:

```yaml
suite: test crd bootstrap
template: crd-bootstrap/crds/bootstrap_crd.yaml # should point to a CRD that is regularly updated like in a helm chart.
tests:
  - it: matches bootstrap crds correctly
    asserts:
      - matchSnapshot:
          # this will generate one snapshot / CRD version and match all of them to the right version of the CRD
          path: sample-tests/__snapshots__
          ignoreErrors:
            - "ignore this match"
      - matchSnapshot:
          path: sample-tests/__snapshots__
          # generates a yaml file
          minimal: true
  - it: matches some custom stuff
    asserts:
      - matchString:
          apiVersion: v1alpha1 # this will match this exact version only from the list of versions in the CRD
          kind: Bootstrap
          spec:
            source:
              url:
                url: https://github.com/Skarlso/test
```

Put this into a file called `bootstrap_test.yaml`.

**IMPORTANT**: `test` will only consider yaml files that end with `_test.yaml`.

One test is per CRD file. A single CRD file, however, can contain multiple apiVersions. Therefore, it's important to
only target a specific version with a snapshot, otherwise, we might be testing something that is broken intentionally.

Now, we can run test like this:

```
./bin/cty test sample-tests
```

The locations in the suite are relative to the execution location.

Running test like this, will match the following snapshots with the given CRD assuming we have two version v1alpha1 and
v1beta1:
- bootstrap_crd-v1alpha1.yaml
- bootstrap_crd-v1alpha1.min.yaml
- bootstrap_crd-v1beta1.yaml
- bootstrap_crd-v1beta1.min.yaml

**_Note_**: At this release version the minimal version needs to be adjusted because it will generate an empty object without the closing `{}`.

If everything is okay, it will generate an output like this:

```
./bin/cty test sample-tests
+--------+----------------------------------+---------------+-------+--------------------------------------+
| STATUS | IT                               | MATCHER       | ERROR | TEMPLATE                             |
+--------+----------------------------------+---------------+-------+--------------------------------------+
| PASS   | matches bootstrap crds correctly | matchSnapshot |       | sample-tests/crds/bootstrap_crd.yaml |
| PASS   | matches bootstrap crds correctly | matchSnapshot |       | sample-tests/crds/bootstrap_crd.yaml |
| PASS   | matches some custom stuff        | matchString   |       | sample-tests/crds/bootstrap_crd.yaml |
+--------+----------------------------------+---------------+-------+--------------------------------------+

Tests total: 3, failed: 0, passed: 3
```

If there _was_ an error, it should look something like this:

```
./bin/cty test sample-tests
+--------+----------------------------------+---------------+----------------------------------------------------------------------------------+--------------------------------------+
| STATUS | IT                               | MATCHER       | ERROR                                                                            | TEMPLATE                             |
+--------+----------------------------------+---------------+----------------------------------------------------------------------------------+--------------------------------------+
| PASS   | matches bootstrap crds correctly | matchSnapshot |                                                                                  | sample-tests/crds/bootstrap_crd.yaml |
| PASS   | matches bootstrap crds correctly | matchSnapshot |                                                                                  | sample-tests/crds/bootstrap_crd.yaml |
| FAIL   | matches some custom stuff        | matchString   | matcher returned failure: failed to validate kind Bootstrap and version v1alpha1 | sample-tests/crds/bootstrap_crd.yaml |
|        |                                  |               | : spec.source.url in body must be of type object: "null"                         |                                      |
+--------+----------------------------------+---------------+----------------------------------------------------------------------------------+--------------------------------------+

Tests total: 3, failed: 1, passed: 2
```

In the above failing example, we forgot to define the URL field for source. Similarly, if the regex changes for a field
we should error and be alerted that it's a breaking change for any existing users.

```
./bin/cty test sample-tests
+--------+-----------------------------------+---------------+----------------------------------------------------------------------------------+--------------------------------------------------------------------+
| STATUS | IT                                | MATCHER       | ERROR                                                                            | TEMPLATE                                                           |
+--------+-----------------------------------+---------------+----------------------------------------------------------------------------------+--------------------------------------------------------------------+
| PASS   | matches AWSCluster crds correctly | matchSnapshot |                                                                                  | sample-tests/crds/infrastructure.cluster.x-k8s.io_awsclusters.yaml |
| PASS   | matches AWSCluster crds correctly | matchSnapshot |                                                                                  | sample-tests/crds/infrastructure.cluster.x-k8s.io_awsclusters.yaml |
| PASS   | matches AWSCluster crds correctly | matchString   |                                                                                  | sample-tests/crds/infrastructure.cluster.x-k8s.io_awsclusters.yaml |
| FAIL   | matches AWSCluster crds correctly | matchString   | matcher returned failure: failed to validate kind AWSCluster and version v1beta2 | sample-tests/crds/infrastructure.cluster.x-k8s.io_awsclusters.yaml |
|        |                                   |               | : spec.s3Bucket.name in body should match '^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$'   |                                                                    |
| PASS   | matches bootstrap crds correctly  | matchSnapshot |                                                                                  | sample-tests/crds/bootstrap_crd.yaml                               |
| PASS   | matches bootstrap crds correctly  | matchSnapshot |                                                                                  | sample-tests/crds/bootstrap_crd.yaml                               |
| PASS   | matches some custom stuff         | matchString   |                                                                                  | sample-tests/crds/bootstrap_crd.yaml                               |
+--------+-----------------------------------+---------------+----------------------------------------------------------------------------------+--------------------------------------------------------------------+

Tests total: 7, failed: 1, passed: 6
```

## Updating Snapshots

In order to generate snapshots for CRDs, simply add `--update` to the command:

```
./bin/cty test sample-tests --update
```

It should generate all snapshots and overwrite existing snapshots under the specified folder of the snapshot matcher.
Meaning consider the following yaml snippet from the above test:

```yaml
    asserts:
      - matchSnapshot:
          # this will generate one snapshot / CRD version and match all of them to the right version of the CRD
          path: sample-tests/__snapshots__
```

Provided this `path` the generated snapshots would end up under `sample-tests/__snapshots__` folder with a generated
name that will match the `template` field in the suite. If `template` field is changed, regenerate the tests and
delete any outdated snapshots.

## Examples

For further examples, please see under [sample-tests](./sample-tests).
