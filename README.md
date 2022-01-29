# tfgen - Terraform boilerplate generator

`tfgen`, short for Terraform Generator, is a tool to generate boilerplate code for Terraform, based on a yaml configuration file. It's useful for creating a set of pre-defined configuration files with common Terraform definitions like backend, provider, variables, etc. The tool was created mainly to be used on [Terraform monorepos](https://github.com/refl3ction/terraform-monorepo-example) that contains multiple environments (different or same AWS accounts for example). This way you can dynamically configure your provider and backend configuration for each module, and also provide common variables.

__Benefits:__

- This is not just a tool, it's a way of doing things.
- Increase your productivity.
- Reduce the risk of making mistakes during copy+paste operations.
- Scale your monorepo following the same pattern across the modules.

## Motivation

`Terragrunt` is a great tool and inspired me a lot to create `tfgen`, but instead of being a wrapper for the Terraform binary, `tfgen` just creates Terraform files from templates and doesn't interact with Terraform at all. Terraform will be used independently on your local environment or in your CI system to deploy the resources.

## Features

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

### Configuration files

`tfgen` works based on yaml configuration files. It searches recursively from the working directory up to the parent directories until it finds the root config file, if it doesn't find the file it will exit with an error. All the files are be merged into the root config file, but the inner configuration have precedence over the outer.

#### Root config

In the root config file, you can set variables that can be used across all environments, and also templates that will be reused.

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

> Note that `.aws_region` and `.aws_account` are variables that you need to provide in the environment specific config, on the other side `tfgen_working_dir` is provided by the tool

#### Environment specific config

In the environment specific config file (non root), you can pass additional configuration, or override configuration from the root config file. You can have multiple specific config files, all of them will be merged into the root one.

```yaml
# infra-live/dev/.tfgen.yaml
---
root_file: false
vars:
  aws_account: 1111111111
  aws_region: us-east-1
template_files:
  _vars.tf: |
    variable "env" {
        type    = string
        default = "dev"
    }

# infra-live/prod/.tfgen.yaml
---
root_file: false
vars:
  aws_account: 2222222222
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

#### tfgen variables

These variables are injected into the templates:

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
```

This execution will create all the files declared in the `.tfgen.yaml` files inside the working directory (the directory where you run the command), executing the templates and passing in all the variables declared in the config files.

The files will look like this:

#### _backend.tf

```hcl
terraform {
  backend "s3" {
    bucket         = "my-state-bucket"
    dynamodb_table = "my-lock-table"
    encrypt        = true
    key            = "dev/s3/dev-tfgen-bucket/terraform.tfstate"
    region         = "us-east-1"
    role_arn       = "arn:aws:iam::1111111111:role/terraformRole"
  }
}
```

#### _provider.tf

```hcl
provider "aws" {
  region = "us-east-1"
}
```

#### _vars.tf

```hcl
variable "env" {
  type    = string
  default = "dev"
}
```

### Next steps

After creating the common Terraform files, probably you'll start writing your `main.tf` file. So at this point, you already know what to do.

```bash
terraform init

terraform plan -out tf.out

terraform apply tf.out
```

Have fun!
