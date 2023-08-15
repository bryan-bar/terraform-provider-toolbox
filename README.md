# Terraform Provider: Toolbox

The `Toolbox` provider is a provider that provides an interface between Terraform and external programs.
Using this provider, it is possible to write separate programs that can participate in the Terraform lifecycle by implementing a specific protocol.
This provider is for last resort and existing providers should be preferred as external programs escape the lifecycle and can depend on the environment.
[Issue #610 (opened - 2014/11/27) (closed - 2022/10/29) - capture output of provisioners into variables](https://github.com/hashicorp/terraform/issues/610)
[Issue #5 (opened - 2017/07/19) - External resource provider](https://github.com/hashicorp/terraform-provider-external/issues/5)

Uses:
* Allow execution of external programs during Terraform's lifecycle.
  * Direct CLI/API calls instead of relying on the provider to handle all cases.
  * Ansible playbooks to bootstrap a system after all resources are created/attached.
  * Execution of existing scripts/executables instead of creating a provider.
* Allow external program data to be returned back into the Terraform lifecycle.
  * Additional volume's UUID after formatting for tracking

Alternative Solutions:
* [External provider](https://registry.terraform.io/providers/hashicorp/external/latest/docs) - Data source which allows for external program execution and exposes output for use in a terraform (expects no side-effects as it will execute with every terraform run)
* [remote-exec provisioner](https://developer.hashicorp.com/terraform/language/resources/provisioners/remote-exec) - Remote executable defined within a resource and executed after resource is created (requires temp files to read ouput back into terraform).
* [local-exec provisioner](https://developer.hashicorp.com/terraform/language/resources/provisioners/local-exec) - Local executable defined within a resource and executed after a resource is created (requires temp files to read ouput back into terraform).
* [cloudinit provider](https://registry.terraform.io/providers/hashicorp/cloudinit/latest/docs) - Cloud-Init configuration to override defaults from a resources used image.
* [external shell module](https://registry.terraform.io/modules/Invicton-Labs/shell-resource/external/latest) - Execute an external program on the local instance and expose output for use in terraform (temp files needed to read final output)
* [SSH provider](https://registry.terraform.io/providers/loafoe/ssh/latest/docs) - Execute an external program on a remote instance and expose the output for use in terraform.
* [shell provider](https://registry.terraform.io/providers/scottwinkler/shell/latest/docs) - Execute an external program on the local instance and expose the output for use in terraform.
* [restapi provider](https://registry.terraform.io/providers/Mastercard/restapi/latest/docs) - Use of a rest api within the terraform lifecycle.

## Compatibility

Compatibility table between this provider, the [Terraform Plugin Protocol](https://www.terraform.io/plugin/how-terraform-works#terraform-plugin-protocol)
version it implements, and Terraform:

| Toolbox Provider | Terraform Plugin Protocol | Terraform |   Golang  |
|:-----------------:|:-------------------------:|:---------:|:---------:|
|    `>= 0.1.3`     |            `6`            | `>= 1.3.6`| `>= 1.20` |

## Requirements

* [Terraform](https://www.terraform.io/downloads)
* [Go](https://go.dev/doc/install)
* [GNU Make](https://www.gnu.org/software/make/)
* [golangci-lint](https://golangci-lint.run/usage/install/#local-installation) (optional)

## Documentation, questions and discussions
Official documentation on how to use this provider can be found on the
[Terraform Registry](https://registry.terraform.io/providers/bryan-bar/toolbox/latest/docs).
In case of specific questions or discussions or issues, please raise an issue on github.

We also provide:

* [Support](.github/SUPPORT.md) page for help when using the provider
* [Contributing](.github/CONTRIBUTING.md) guidelines in case you want to help this project

Details can be found querying the [Registry API](https://www.terraform.io/internals/provider-registry-protocol#list-available-versions)
that return all the details about which version are currently available for a particular provider.
[Here](https://registry.terraform.io/v1/providers/bryan-bar/toolbox/versions) are the details for Time (JSON response).


## Development

### Building

1. `git clone` this repository and `cd` into its directory
2. `make` will trigger the Golang build

The provided `GNUmakefile` defines additional commands generally useful during development,
like for running tests, generating documentation, code formatting and linting.
Taking a look at it's content is recommended.

### Testing

In order to test the provider, you can run

* `make test` to run provider tests
* `make testacc` to run provider acceptance tests

It's important to note that acceptance tests (`testacc`) will actually spawn
`terraform` and the provider. Read more about they work on the
[official page](https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests).

### Generating documentation

This provider uses [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs/)
to generate documentation and store it in the `docs/` directory.
Once a release is cut, the Terraform Registry will download the documentation from `docs/`
and associate it with the release version. Read more about how this works on the
[official page](https://www.terraform.io/registry/providers/docs).

Use `make generate` to ensure the documentation is regenerated with any changes.

### Using a development build

If [running tests and acceptance tests](#testing) isn't enough, it's possible to set up a local terraform configuration
to use a development builds of the provider. This can be achieved by leveraging the Terraform CLI
[configuration file development overrides](https://www.terraform.io/cli/config/config-file#development-overrides-for-provider-developers).

First, use `make install` to place a fresh development build of the provider in your
[`${GOBIN}`](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies)
(defaults to `${GOPATH}/bin` or `${HOME}/go/bin` if `${GOPATH}` is not set). Repeat
this every time you make changes to the provider locally.

Then, setup your environment following [these instructions](https://www.terraform.io/plugin/debugging#terraform-cli-development-overrides)
to make your local terraform use your local build.

### Testing GitHub Actions

This project uses [GitHub Actions](https://docs.github.com/en/actions/automating-builds-and-tests) to realize its CI.

Sometimes it might be helpful to locally reproduce the behaviour of those actions,
and for this we use [act](https://github.com/nektos/act). Once installed, you can _simulate_ the actions executed
when opening a PR with:

```shell
# List of workflows for the 'pull_request' action
$ act -l pull_request

# Execute the workflows associated with the `pull_request' action 
$ act pull_request
```

## Releasing

The release process is automated via GitHub Actions, and it's defined in the Workflow
[release.yml](./.github/workflows/release.yml).

Each release is cut by pushing a [semantically versioned](https://semver.org/) tag to the default branch.

## License

[Mozilla Public License v2.0](./LICENSE)
