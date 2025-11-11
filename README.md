# tfgen - Terraform boilerplate generator

[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://GitHub.com/Naereen/StrapDown.js/graphs/commit-activity)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/0xDones/tfgen.svg)](https://github.com/0xDones/tfgen)
[![GitHub stars](https://img.shields.io/github/stars/0xDones/tfgen.svg?style=social&label=Star)](https://github.com/0xDones/tfgen/stargazers/)

Terragrunt alternative to keep your Terraform code consistent and DRY

## Overview

### What is tfgen

`tfgen` is useful for maintaining and scaling a [Terraform Monorepo](https://github.com/0xDones/terraform-monorepo-example), in which you provision resources in a multi-environment/account setup. It is designed to create consistent Terraform definitions, like backend (with dynamic key), provider, and variables for each environment/account, as defined in a set of YAML configuration files.

### Why tfgen

[Terragrunt](https://github.com/gruntwork-io/terragrunt) - a thin wrapper for Terraform that provides extra tools for working with multiple Terraform modules - is a great tool and inspired me a lot to create `tfgen`, but instead of being a wrapper for the Terraform binary, `tfgen` just creates Terraform files from templates and doesn't interact with Terraform at all. Terraform will be used independently on your local environment or in your CI system to deploy the resources.

- This is not just a tool, it's a way of doing things
- Keep your Terraform configuration consistent across the environments
- Reduce the risk of making mistakes while copying+pasting your backend, provider, and other common Terraform definitions
- Increase your productivity
- Scale your mono repo following the same pattern across the modules

### Features

- Builtin functionality to provide the remote state key dynamically
- YAML file configuration
- Templates are parsed using `Go templates`
- Automatic semantic versioning on every push to main

## Versioning

This project uses semantic versioning with automatic version bumping based on commit messages. When you push to the `main` branch, the version is automatically bumped based on your commit message format:

- **Major version bump** (`x.0.0`): Use `feat!:` or include `BREAKING CHANGE:` in your commit message
  - Example: `feat!: redesign configuration format`
  - Example: `BREAKING CHANGE: remove deprecated command`
- **Minor version bump** (`0.x.0`): Use `feat:` prefix for new features
  - Example: `feat: add support for new cloud provider`
  - Example: `feat(templates): add new template engine`
- **Patch version bump** (`0.0.x`): Use `fix:` prefix for bug fixes or any other commit type
  - Example: `fix: correct state key generation`
  - Example: `docs: update installation guide`

The version bump workflow automatically creates a git tag, which triggers the release workflow to build binaries for multiple platforms.

### Required GitHub Secret Setup

For the automatic versioning workflow to trigger the release workflow, you need to set up a Personal Access Token (PAT):

1. Create a Personal Access Token with the following permissions:
   - `contents: write` - to push tags
   - `workflows: write` - to trigger other workflows
2. Add the token as a repository secret named `PAT_TOKEN`

**Why is this needed?** GitHub Actions' default `GITHUB_TOKEN` does not trigger other workflows for security reasons (to prevent recursive workflow runs). Using a PAT allows the version bump workflow to trigger the release workflow.

## Getting Started

### Prereqs

- Docker or Go

### Installation

```bash
git clone --depth 1 git@github.com:0xDones/tfgen.git
cd tfgen

# Using Docker
docker run --rm -v $PWD:/src -w /src -e GOOS=darwin -e GOARCH=amd64 golang:alpine go build -o bin/tfgen

# Using Go
go build -o bin/tfgen

mv bin/tfgen /usr/local/bin
```

__Note:__ when building using Docker, change `GOOS=darwin` to `GOOS=linux` or `GOOS=windows` based on your system

## Usage

### Basic Usage

```bash
$ tfgen help
tfgen is a devtool to keep your Terraform code consistent and DRY

Usage:
  tfgen [command]

Available Commands:
  clean       clean templates from the target directory
  completion  Generate the autocompletion script for the specified shell
  exec        Execute the templates in the given target directory
  help        Help about any command

Flags:
  -h, --help      help for tfgen
  -v, --verbose   verbose output

Use "tfgen [command] --help" for more information about a command.
```

### Configuration files

The configuration files are written in YAML and have the following structure:

```yaml
---
root_file: bool
vars:
  var1: value1
  var2: value2
template_files:
  template1.tf: |
    template content
  template2.tf: |
    template content
```

#### How config files are parsed

__tfgen__ will recursively look for all `.tfgen.yaml` files from the target directory up to the parent directories until it finds the __root config file__, if it doesn't find the file it will exit with an error. All the other files found on the way up are merged into the root config file, and the __inner config file has precedence over the outer__.

We have two types of configuration files:

1. Root config
2. Environment specific config

#### Root config

In the root config file, you can set variables and templates that can be reused across all environments. You need at least 1 root config file.

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
        key            = "{{ .Vars.tfgen_state_key }}/terraform.tfstate"
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

> Note that `aws_region`, `aws_account`, and `env` are variables that you need to provide in the environment-specific config. `tfgen_state_key` is provided by the `tfgen`, it will be explained below.

#### Environment specific config

In the environment-specific config file (non-root), you can pass additional configuration, or override configuration from the root config file. You can have multiple specific config files, all of them will be merged into the root one.

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

### Provided Variables

These variables are automatically injected into the templates:

- `tfgen_state_key`: The path from the root config file to the target directory

## Practical Example

### Repository Structure

The [terraform-monorepo-example](https://github.com/0xDones/terraform-monorepo-example) repository can be used as an example of how to structure your repository to leverage `tfgen` and also follow Terraform best practices.

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

Inside our `infra-live` folder, we have two environments, dev and prod. They are deployed in different aws accounts, and each one has a different role that needs to be assumed in the provider configuration. Instead of copying the files back and forth every time we need to create a new module, we'll let `tfgen` create it for us based on our configuration defined on the `.tfgen.yaml` config files.

### Running the `exec` command

Let's create the common files to start writing our Terraform module

```bash
# If you didn't clone the example repo yet
git clone git@github.com:0xDones/terraform-monorepo-example.git
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

After creating the common Terraform files, you'll probably start writing your `main.tf` file. So at this point, you already know what to do.

```bash
terraform init

terraform plan -out tf.out

terraform apply tf.out
```

## Related

- [terraform-monorepo-example](https://github.com/0xDones/terraform-monorepo-example) - Example repo used in the tutorial
- [Terragrunt](https://github.com/gruntwork-io/terragrunt) - Tool that inspired me to create `tfgen`

Have fun!
