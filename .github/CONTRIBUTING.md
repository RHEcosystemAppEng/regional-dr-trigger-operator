# Contributing guidelines

## Contributions

Contributors to this repository must submit their contributions under the terms of the [Apache Public License 2.0](https://www.apache.org/licenses/LICENSE-2.0).

## Certificate of Origin

By contributing to this project, you agree to the Developer Certificate of
Origin (DCO). The Linux Kernel community created this document, and it is a simple statement that you, as a contributor, have the legal right to contribute the content you are contributing. See the [DCO](../DCO) file for details.

## Contributing A Patch

1. Submit an issue describing your proposed change to the repo.
2. The repo owners will respond to your issue promptly.
3. Fork the repo, develop and test your code changes.
4. Submit a pull request.
5. Ensure the PR title adheres to the [Conventional Commits Specifications](https://www.conventionalcommits.org/).
6. Keep each PR addressing only one concern.
7. Make sure to regenerate any manifests required.

## Manifests Generation

Regenerate the RBAC manifests with the following:

```shell
make generate manifests
```

## Pre-check before submitting a PR

After your PR is ready to commit, please run the following commands to check your code:

```shell
make build
make test
make lint
make chart USER=ecosystem-appeng
make lint-chart
```
