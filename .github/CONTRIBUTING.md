# Contributing guidelines

## Contributions

All contributions to the repository must be submitted under the terms of the
[Apache Public License 2.0](https://www.apache.org/licenses/LICENSE-2.0).

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO](../DCO) file for details.

## Contributing A Patch

1. Submit an issue describing your proposed change to the repo in question.
2. The repo owners will respond to your issue promptly.
3. Fork the repo, develop and test your code changes.
4. Submit a pull request.
5. Make sure the PR title adhere to the [Conventional Commits Specifications](https://www.conventionalcommits.org/)

## Generating Code and Manifests

If your patch includes change to the API or to KubeBuilder's instructions, make sure to generate the code and/or
manifests with one/all of:

```shell
make generate/code
make generate/manifests
```

## Pre-check before submitting a PR

After your PR is ready to commit, please run following commands to check your code.

Based on your patch type, lint with one/all of:

```shell
make lint/code
make lint/containerFile
make lint/ci
```

Based on your development stage, test with one/all of:

```shell
make test
make test/cov
make test/mut
```

Based on your patch type, build with one/all of:

```shell
make build
make build/image
```
