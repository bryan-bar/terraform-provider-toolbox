---
page_title: "Provider: Toolbox"
description: |-
  The Toolbox provider allows external scripts to be integrated with Terraform.
---

# toolbox Provider

~> **Warning** Please use the
[external provider by HashiCorp](https://registry.terraform.io/providers/hashicorp/external/latest)
if you only need to use the data source.

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

This provider is intended to be used for simple situations where you wish
to integrate Terraform with a system for which a first-class provider
doesn't exist. It is not as powerful as a first-class Terraform provider,
so users of this interface should carefully consider the implications
described on each of the child documentation pages (available from the
navigation bar) for each type of object this provider supports.
