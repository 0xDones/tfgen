# tfgen - Terraform boilerplate generator

tfgen, short for Terraform Generator, is a tool to generate boilerplate code for Terraform, based on a yaml configuration file. It's useful for creating a set of pre-defined configuration files with common Terraform definitions like backend, provider, variables, etc. The tool was created mainly to be used on Terraform monorepos that contains multiple environments (different or same AWS accounts for example). This way you can dynamically configure your provider and backend configuration for each module, and also provide common variables.

## Motivation

`Terragrunt` is a great tool and inspired me a lot to create `tfgen`, but instead of being a wrapper for the Terraform binary, `tfgen` just creates Terraform files from templates and doesn't interact with Terraform at all. Terraform will be used independently to locally or in your CI system to deploy the resources.

## Feautures

- Create Terraform "common" files from templates in the selected working directory
- Fill the state key dynamically based on the relative path from the root config file to your working directory
- Use go template to generate the files, passing variables to the template dynamically

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

__Note:__ change `GOOS=darwin` to `linux` or `windows` based on your system

## Usage

Let's assume we have the following structure:

```md
.
├── README.md
├── infra-live
│   ├── dev
│   │   ├── networking
│   │   └── s3
│   └── prod
│       ├── networking
│       └── s3
└── modules
    └── my-custom-module
        └── main.tf
```

> You can check a complete monorepo example [here](https://github.com/refl3ction/terraform-monorepo-example)

Inside our `infra-live` folder, we have two environments, dev and prod. They are deployed in different aws accounts, and each one have a different role that needs to be assumed in the provider configuration. Instead of copying the files back and forth every time we need to create a new module, we'll let `tfgen` create it for us.

Let's create our config files.

### Configuration file

First we need to create our root config file. Let's create it inside our `infra-live` folder.

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
        key            = "{{ .tfgen_working_dir }}/terraform.tfstate"
        region         = "{{ .aws_region }}"
        role_arn       = "arn:aws:iam::{{ .aws_account }}:role/terraformRole"
      }
    }
  _provider.tf: |
    provider "aws" {
      region = "{{ .aws_region }}"
    }
```

> We are covering all the possible variables in this example.

Now, let's create a config file for the dev and another for the prod environment.

```yaml
# infra-live/dev/.tfgen
---
root_file: false
vars:
  aws_accouint: 1111111111
  aws_region: us-east-1
template_files:
  _vars.tf: |
    variable "env" {
        type    = string
        default = "dev"
    }

# infra-live/prod/.tfgen
---
root_file: false
vars:
  aws_accouint: 2222222222
  aws_region: us-east-2
template_files:
  _vars.tf: |
    variable "env" {
        type    = string
        default = "prod"
    }
  example.tf: |
    # I'll just be created on modules inside the prod folder
```

### Running

Now we simply run tfgen

```bash
tfgen init .
```

## tfgen variables

These variables are injected into the templates

- `tfgen_working_dir`: The path from the root config file to the working directory
