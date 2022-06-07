# tfgen - Terraform boilerplate generator

[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://GitHub.com/Naereen/StrapDown.js/graphs/commit-activity)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/refl3ction/tfgen.svg)](https://github.com/refl3ction/tfgen)
[![GitHub stars](https://img.shields.io/github/stars/refl3ction/tfgen.svg?style=social&label=Star)](https://github.com/refl3ction/tfgen/stargazers/)

Terragrunt alternative to keep your Terraform code consistent and DRY

## Overview

### What is tfgen

`tfgen` is useful for maintaining and scaling a [Terraform Monorepo](https://github.com/refl3ction/terraform-monorepo-example), in which you provision resources in a multi environment/account setup. It is designed to create consistent Terraform definitions, like backend (with dynamic key), provider, and variables for each environment/account, as defined in a set of yaml configuration files.

### Why tfgen

[Terragrunt](https://github.com/gruntwork-io/terragrunt) - a thin wrapper for Terraform that provides extra tools for working with multiple Terraform modules - is a great tool and inspired me a lot to create `tfgen`, but instead of being a wrapper for the Terraform binary, `tfgen` just creates Terraform files from templates and doesn't interact with Terraform at all. Terraform will be used independently on your local environment or in your CI system to deploy the resources.

- This is not just a tool, it's a way of doing things
- Keep your Terraform configuration consistent across the environments
- Reduce the risk of making mistakes while copying+pasting your backend, provider and other common Terraform definitions
- Increase your productivity
- Scale your monorepo following the same pattern across the modules

### Features

- Builtin functionallity to provide the remote state key dynamically
- YAML file configuration
- Templates are parsed using `Go templates`
- Support for a config directory containing templates and YAML configuration
- Manage dependencies for providers/modules in a mappting of key values to source/version etc
- Some utility functions to make it easier to use those depenencies in templates

## Getting Started

### Prereqs

- Docker or Go

### Installation

```bash
git clone --depth 1 git@github.com:refl3ction/tfgen.git
cd tfgen

# Using Docker
docker run --rm -v $PWD:/src -w /src -e GOOS=darwin -e GOARCH=amd64 golang:alpine go build

# Using Go
go build

mv tfgen /usr/local/bin
```

__Note:__ when building using Docker, change `GOOS=darwin` to `GOOS=linux` or `GOOS=windows` based on your system

## Usage

### Basic Usage

```bash
# executing the templates
$ tfgen exec <target dir>

$ tfgen help
tfgen is a devtool to keep your Terraform code consistent and DRY

Usage:
  tfgen [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  exec        Execute the templates in the given target directory.
  help        Help about any command

Flags:
  -h, --help   help for tfgen

Use "tfgen [command] --help" for more information about a command.
```

### Quick Tutorial

Before we start, let's clone our [terraform-monorepo-example](https://github.com/refl3ction/terraform-monorepo-example) repository. All the examples bellow are based on the structure of this repo:

```md
.
├── infra-live
│   ├── dev
│   │   ├── networking
│   │   ├── s3
│   │   ├── security
│   │   ├── stacks
│   │   └── .tfgen.yaml     # Environment specific config
│   ├── prod
│   │   ├── networking
│   │   ├── s3
│   │   ├── security
│   │   ├── stacks
│   │   └── .tfgen.yaml     # Environment specific config
│   └── .tfgen.yaml         # Root config file
└── modules
    └── my-custom-module
```

Inside our `infra-live` folder, we have two environments, dev and prod. They are deployed in different aws accounts, and each one have a different role that needs to be assumed in the provider configuration. Instead of copying the files back and forth every time we need to create a new module, we'll let `tfgen` create it for us based on our `.tfgen.yaml` config files.

Alternatively you can also use a directory to hold your config YAML and template files. You can mix and match plain YAML file or configuration directories but you can only use one per directory. It could look something would look like this:

```md
.
├── infra-live
│   ├── dev
│   │   |── .tfgen.d          # Environment specific config
│   │   │   │── .tfgen.yaml
│   │   │   ├── _data.tf
│   │   │   └── _vars.tf  
│   │   ├── networking
│   │   ├── s3
│   │   ├── security
│   │   ├── stacks
│   │   └── databases
│   ├── prod
│   │   |── .tfgen.d          # Environment specific config
│   │   │   │── .tfgen.yaml
│   │   │   ├── _data.tf
│   │   │   └── _vars.tf
│   │   ├── networking
│   │   ├── s3
│   │   ├── security
│   │   ├── stacks
│   │   └── databases
│   └── .tfgen.yaml         # Root config file (could also be a dir)
└── modules
    └── my-custom-module
```

### Configuration files

#### How config files are parsed

__tfgen__ will recursively look for all `.tfgen.yaml` files or `.tfgen.d` directories from the working directory up to the parent directories until it finds the root config file, if it doesn't find the file it will exit with an error. All the other files found on the way up are merged into the root config file, and the inner configuration have precedence over the outer.

We have two types of configuration files:

1. Root config
2. Environment specific config

#### Root config

In the root config file, you can set variables and templates that can be reused across all environments. You need at least 1 root config file.

If you choose to use a directory you can specify some templates in the config file and some as separate files but if you define the same one both ways the separate file will take precedence.

```yaml
# infra-live/.tfgen.yaml
---
root_file: true
vars:
  company: acme
template_files:
  _backend.tf: |
    terraform {
      backend "s3" {
        bucket         = "my-state-bucket"
        dynamodb_table = "my-lock-table"
        encrypt        = true
        key            = "{{ .Vars.tfgen_working_dir }}/terraform.tfstate"
        region         = "{{ .Vars.aws_region }}"
        role_arn       = "arn:aws:iam::{{ .Vars.aws_account_id }}:role/terraformRole"
      }
    }
  _provider.tf: |
    provider "aws" {
      region = "{{ .Vars.aws_region }}"
      allowed_account_ids = [
        "{{ .Vars.aws_account_id }}"
      ]
    }
  _vars.tf: |
    variable "env" {
      type    = string
      default = "{{ .Vars.env }}"
    }

```

> Note that `aws_region`, `aws_account` and `env` are variables that you need to provide in the environment specific config. `tfgen_working_dir` is provided by the `tfgen`, it will be explained below.

#### Environment specific config

In the environment specific config file (non root), you can pass additional configuration, or override configuration from the root config file. You can have multiple specific config files, all of them will be merged into the root one.

If you choose to use a directory you can specify some templates in the config file and some as separate files but if you define the same one both ways the separate file will take precedence.

```yaml
# infra-live/dev/.tfgen.yaml
---
root_file: false
vars:
  aws_account_id: 111111111111
  aws_region: us-east-1
  env: dev

# infra-live/prod/.tfgen.yaml
---
root_file: false
vars:
  aws_account_id: 222222222222
  aws_region: us-east-2
  env: prod
template_files:
  additional.tf: |
    # I'll just be created on modules inside the prod folder
```

#### Additional YAML config features:
The `.tfgen.yaml` files specify which type of config they are (`root_file` is true or false). They also may contain dependencies, and variables.

For example, your root might look something like this:

```yaml
---
root_file: true
deps:
  terraform_version: ">= 1.2.1"
  default_providers:
	- aws
  default_remote_states:
    - some_project_you_reference_everywhere
  required_providers:
	aws:
	  source: "hashicorp/aws"
	  version: "~> 4.17.0"
	tls:
	  source: "hashicorp/tls"
	  version: "~> 3.4.0"
	null:
	  source: "hashicorp/null"
	  version: "~> 3.1.1"
	modules:
      vpc:
  		source: "registry.terraform.io/terraform-aws-modules/vpc/aws"
  		version: "~> 3.14.0"
  	  rds_aurora:
  		source: "registry.terraform.io/terraform-aws-modules/rds-aurora/aws"
  		version: "~> 7.1.0"

```

Somewhere else you might add this:


```yaml
---
root_file: false
deps:
  extra_providers:
    - tls
    - null
  extra_remote_states:
    - networking
    - security
```

`default_providers` can be overridden in other YAML, but it will replace them all just like with `vars`, `extra_providers` are added on so that you can pick and choose which ones are needed for each part of your code. In the tfenv source, there are some functions defined for convenience so you your template for versions.tf could look like this (and it would render all your providers depending on context):

```hcl
terraform {
  required_version = "{{ .Deps.TerraformVersion}}"
  required_providers {
    {{- range $key := concat $.Deps.DefaultProviders $.Deps.ExtraProviders }}
    {{ $.Deps.RenderRequiredProvider $key | indent 4 | trim }}
    {{- end }}
  }
}
```

`default_remote_states` and `extra_remote_states` also work in a similar way but they just hold keys (the varitions of what reference to a remote state might look like make it hard to fully make a dictionary of the whole thing) but in your template someplace you can say things like this (`UsesRemoteState` is another convenience function that checks the union of the two lists) and define as many as you need:

```hcl
{{- if .Deps.UsesRemoteState "my_project" }}{{ "\n" }}
data "terraform_remote_state" "my_project" {
  backend = "s3"
  config = {
    bucket         = "{{ .Vars.state_s3_bucket }}"
    dynamodb_table = "{{ .Vars.dynamodb_table }}"
    key            = "core"
    region         = "{{ .Vars.state_aws_region }}"
  }
}
{{- end }}

```

There is also the ability to rewrite the names of the templates when they are rendered in case you like to have one name for the template type to be recognized by IDE and another for once its rendered as plain terraform:

```yaml
template_name_rewrite:
  pattern: "^(.+)(\\.tfgen.tf)$"
  replacement: "_$1.tf"
```

### Internal variables

These variables are automatically injected into the templates:

- `tfgen_working_dir`: The path from the root config file to the working directory

### Running

Let's create the common files to start writing our Terraform module

```bash
# If you didn't clone the example repo yet
git clone git@github.com:refl3ction/terraform-monorepo-example.git
cd terraform-monorepo-example

# Create a folder for our new module
mkdir -p infra-live/dev/s3/dev-tfgen-bucket
cd infra-live/dev/s3/dev-tfgen-bucket

# Generate the files
tfgen exec .

# Checking the result (See Output section)
cat _backend.tf _provider.tf _vars.tf
```

This execution will create all the files inside the working directory, executing the templates and passing in all the variables declared in the config files.

### Output

This will be the content of the files created by `tfgen`:

#### _backend.tf

```hcl
terraform {
  backend "s3" {
    bucket         = "my-state-bucket"
    dynamodb_table = "my-lock-table"
    encrypt        = true
    key            = "dev/s3/dev-tfgen-bucket/terraform.tfstate"
    region         = "us-east-1"
    role_arn       = "arn:aws:iam::111111111111:role/terraformRole"
  }
}
```

#### _provider.tf

```hcl
provider "aws" {
  region = "us-east-1"
  allowed_account_ids = [
    "111111111111"
  ]
}
```

#### _vars.tf

```hcl
variable "env" {
  type    = string
  default = "dev"
}
```

## Next steps

After creating the common Terraform files, probably you'll start writing your `main.tf` file. So at this point, you already know what to do.

```bash
terraform init

terraform plan -out tf.out

terraform apply tf.out
```

There is also a fantastic plugin for IntelliJ Idea which lets you highlight go templates.

- [intellij-idea-plugin](https://plugins.jetbrains.com/plugin/10581-go-template) - Direct link to the template

`Settings | Editor | File Types` has the place to add your template type. The hidden gem: once started, there is a popup for the type of the underlying template. So for example you can create `*.tfgen.tf` and then tell it the underlying type is *Terraform* and then _voila_, language sensitive rendering!

## Related

- [terraform-monorepo-example](https://github.com/refl3ction/terraform-monorepo-example) - Example repo used in the tutorial
- [Terragrunt](https://github.com/gruntwork-io/terragrunt) - Tool that inspired me to create `tfgen`

Have fun!
