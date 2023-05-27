---
page_title: "Provider: Toolbox"
description: |-
  The Toolbox provider allows external scripts to be integrated with Terraform.
---

# toolbox Provider

~> **Warning** Please use the
[external provider by HashiCorp](https://registry.terraform.io/providers/hashicorp/external/latest)
if you only need to use the data source.
Terraform Enterprise does not guarantee availability of any
particular language runtimes or external programs beyond standard shell
utilities, so it is not recommended to use this provider within configurations
that are applied within Terraform Enterprise.

The `Toolbox` provider is a provider that provides an interface between Terraform and external programs.
Using this provider, it is possible to write separate programs that can participate in the Terraform lifecycle by implementing a specific protocol(shell/python/binaries/ansible/cli-tools).
This provider is for last resort and existing providers should be preferred as external programs escape the lifecycle and can depend on the environment.

Uses:
* Allow execution of external programs during Terraform's lifecycle.
* Allow external program data to be returned back into the Terraform lifecycle.

Limitations with existing solutions:
* External Provider - Uses a data source which executes during all stages and should not perform actions which will cause side-effects.
* remote-exec and local-exec provisioner - Commands cannot return data back into the terraform lifecycle so workarounds are commonly used such as exporting a file within the provisioner for further use/state.
* cloud-init - Final state from bootstrap is not returned back into terraform and might be needed such as volumes which can be out of order or renamed by providers unless formatted/mounted.

This provider is intended to be used for simple situations where you wish
to integrate Terraform with a system for which a first-class provider
doesn't exist. It is not as powerful as a first-class Terraform provider,
so users of this interface should carefully consider the implications
described on each of the child documentation pages (available from the
navigation bar) for each type of object this provider supports.
