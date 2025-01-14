# update-ksops-secrets

<!--mdtogo:Short-->

Update and generate the encrypted secrets preparing for Kustomize SOPS (KSOPS)

<!--mdtogo-->

### Overview

<!--mdtogo:Long-->

The `update-ksops-secrets` provides a declarative configuration for managing the secrets encryption and also generates the integrated manifests that could be rendered the `Secret` resource by kustomize with Kustomize SOPS plugin.

[Kustomize SOPS](https://raw.githubusercontent.com/viaduct-ai/kustomize-sops) manages the Kubernetes secret resources via the GitOps paradigm by integrating with [kustomize](https://github.com/kubernetes-sigs/kustomize/) and [SOPS](https://github.com/mozilla/sops).

### FunctionConfig

We use `UpdateKSopsSecrets` custom resource to configure the `update-ksops-secrets` function. The desired values are provided as a schema.

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: UpdateKSopsSecrets
metadata:
  name: string
  annotations:
    string: string
  labels:
    string: string
secret:
  type: string
  references:
    - string
  items:
    - string
recipients:
  - type: string
    recipient: string
    publicKeySecretReference:
      name: string
      key: string
```

#### apiVersion

The API version of the configuration resource

- `fn.kpt.dev/v1alpha1`

#### kind

The resource kind refers to this configuration

- `UpdateKSopsSecrets`

#### metadata

The metadata for describing the generated `Secret` resource

|         Field | Description                          | Example                                      |
| ------------: | ------------------------------------ | -------------------------------------------- |
|        `name` | The generated `Secret` resource name | `runtime-secrets`                            |
| `annotations` | The secret annotations               | `fn.kpt.dev/generator: update-ksops-secrets` |
|      `labels` | The secret labels                    | `fn.kpt.dev/generated-by: kpt`               |

#### secret

|                       Field | Description                                                                                                         | Example                                                         |
| --------------------------: | ------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
|                      `type` | Type of the generated `Secret` resource <br/>-`Opaque` (default)<br/>-`kubernetes.io/dockerconfigjson`<br/>-`...`   | `kubernetes.io/dockerconfigjson`                                |
|                `references` | The list of unencrypted secret resources that the `update-ksops-secrets` will look up and generates encrypted files | - `unencrypted-secrets`<br/> - `unencrypted-secrets-config-txt` |
|                     `items` | The list of secret keys for data look up in the referenced secret resources                                         |
| [`recipients`](#recipients) | The list of recipients who could decrypt the generated encrypted files                                              |

#### recipients

|                                                   Field | Description                                                                                     | Example                                                                                                         |
| ------------------------------------------------------: | ----------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
|                                                  `type` | The type of the recipient that supported by SOPS, eg. `age`, `pgp`                              | `age`                                                                                                           |
|                                             `recipient` | The recipient id<br/>-`age`: public key<br/>-`pgp`: fingerprint id                              | `age1x7pzjx4r05ar95pulf20knx0mkscaxa0zhtqr948wza3863fvees8tzaaa`<br/>`F532DA10E563EE84440977A19D0470BDA6CDC457` |
| [`publicKeySecretReference`](#publickeysecretreference) | Pass the PGP/GPG public key data with a secret reference, ignored for all other types but `pgp` |                                                                                                                 |

#### publicKeySecretReference

|  Field | Description                                                | Example                                        |
| -----: | ---------------------------------------------------------- | ---------------------------------------------- |
| `name` | The secret name contains PGP/GPG public keys data          | `gpg-publickeys`                               |
|  `key` | The secret key contains a specific PGP/GPG public key data | `380024A2AC1D3EBC9402BEE66E38309B4DA30118.gpg` |

`update-ksops-secrets` function performs the following steps when invoked:

1. Pass unencrypted secrets manifests referred by the configuration to the mutators pipeline.
2. Generate the encrypted secrets from the listed items in the configuration using SOPS,
   - Unavailable secrets would be skipped.
   - Existing encrypted files of unavailable secrets will be processed with untouch.
3. Generate or update the kustomization and KSOPS secrets resources.

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

#### Generate kustomization manifests with encrypted files

Let's start with the input resource in a package, see the [Note](#create-unencrypted-secrets) for how to prepare an `unencrypted-secrets` resource file

```yaml
# unencrypted-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: unencrypted-secrets
stringData:
  test2: test2
  UPPER_CASE: upper_case
data:
  test: dGVzdA==
```

```yaml
# unencrypted-secrets-config-txt.yaml
apiVersion: v1
kind: Secret
metadata:
  name: unencrypted-secrets-config-txt
data:
  config.txt: Y29uZmlnLnR4dAo=
```

Declare the new desired values for setters in the functionConfig file.

```yaml
# update-ksops-secrets.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: UpdateKSopsSecrets
metadata:
  name: test-update-ksops-secrets
secret:
  type: Opaque
  references:
    - unencrypted-secrets
    - unencrypted-secrets-config-txt
  items:
    - test
    - test2
    - UPPER_CASE
    - config.txt
recipients:
  - type: age
    recipient: age1x7pzjx4r05ar95pulf20knx0mkscaxa0zhtqr948wza3863fvees8tzaaa
```

```yaml
# Kptfile
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: update-ksops-secrets
pipeline:
  mutators:
    - image: ghcr.io/neutronth/kpt-update-ksops-secrets:0.11
      configPath: update-ksops-secrets.yaml
```

Invoke the function:

```shell
$ kpt fn render
```

Alternatively, invoke function directly without the `Kptfile`

```shell
$ kpt fn eval \
    --image=ghcr.io/neutronth/kpt-update-ksops-secrets:0.11 \
    --fn-config=update-ksops-secrets.yaml
```

If you encountered the error with the PGP/GPG recipients encryption, see the [Note](#gpg-receive-keys-requires-network-to-work-properly) to understand the limitation and working solution.

The above command will add files to your directory, which you can view

```shell
$ kpt pkg tree
Package "example"
├── [Kptfile]  Kptfile update-ksops-secrets
├── [kustomization.yaml]  Kustomization
├── [secrets.yaml]  Secret test-update-ksops-secrets
├── [update-ksops-secrets.yaml]  UpdateKSopsSecrets test-update-ksops-secrets
└── generated
    ├── [ksops-generator.yaml]  ksops ksops-generator-test-update-ksops-secrets-config.txt
    ├── [ksops-generator.yaml]  ksops ksops-generator-test-update-ksops-secrets-test
    ├── [ksops-generator.yaml]  ksops ksops-generator-test-update-ksops-secrets-test2
    ├── [ksops-generator.yaml]  ksops ksops-generator-test-update-ksops-secrets-upper_case
    ├── [secrets.config.txt.enc.yaml]  Secret test-update-ksops-secrets
    ├── [secrets.test.enc.yaml]  Secret test-update-ksops-secrets
    ├── [secrets.test2.enc.yaml]  Secret test-update-ksops-secrets
    └── [secrets.upper_case.enc.yaml]  Secret test-update-ksops-secrets
```

The generated files in the package could be rendered by the kustomization

```shell
$ kustomize build --enable-alpha-plugins .
```

```yaml
apiVersion: v1
data:
  UPPER_CASE: dXBwZXJfY2FzZQ==
  config.txt: Y29uZmlnLnR4dAo=
  test: dGVzdA==
  test2: dGVzdDI=
kind: Secret
metadata:
  annotations:
    kustomize.config.k8s.io/behavior: merge
  name: test-update-ksops-secrets
type: Opaque
```

<!--mdtogo-->

### Note

#### Create unencrypted secrets

We could manually create an `unencrypted secrets` file with your favorite text editor.

Or alternatively, create with `kubectl` as following command

```shell
$ kubectl create secret generic unencrypted-secrets --output=yaml --dry-run=client \
  --from-literal="test=test" \
  --from-literal="test2=test2" \
  --from-file=config.txt \
  | tee unencrypted-secrets.yaml
```

This created file must not stored in the repository, it was used only for data preparation that
kept in local storage. Make sure to add the `.gitignore` for those files exclusion.

```sh
# .gitignore
unencrypted-*
```

#### GPG receive-keys requires network to work properly

The limitation of Kpt for security reason, the running container network is not allowed by default. The user should explicitly instruct the Kpt to enable network.

There is no options to enable network within `render` function as described above. Only the alternative command works with additional parameter `--network`

Hence, if the encrypted files recipients include the PGP/GPG fingerprints, the `kpt-update-ksops-secrets` requires network to work properly as following command,

```shell
$ kpt fn eval \
    --image=ghcr.io/neutronth/kpt-update-ksops-secrets:0.11 \
    --fn-config=update-ksops-secrets.yaml \
    --network
```
