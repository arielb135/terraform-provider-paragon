# terraform-paragon-provider
Wrap [paragon](https://useparagon.com) APIs with terraform resources - provider is available at https://registry.terraform.io/providers/arielb135/paragon

was written mostly with the help of Claude.ai

> **Important:** This is not an official provider and is not supported by Paragon. This provider is maintained by the community and is not officially supported by HashiCorp. The APIs used by this provider are not officially supported by Paragon and may change at any time.

## Local development

In order to build the provider locally, just run `go install .` and the provider will be installed into your `$GOPATH/bin` directory.

You can also run this command `GOBIN=YOUR_GO_PATH_BIN_DIRECTORY go install .`

In order to test it, first update your `.terraformrc` file to include the following:

```hcl
provider_installation {
  dev_overrides {
      "arielb135/paragon" = "YOUR_GO_PATH_BIN_DIRECTORY"
  }

  direct {}
}
```

Then you may test any terraform code you'd like, and the local binary will be used:

```hcl
terraform {
  required_providers {
    paragon = {
      source  = "arielb135/paragon"
      version = "1.0.1"
    }
  }
}

provider "paragon" {
  username = "your_email"
  password = "your_password"
}

# Read your organization to get it by name.
data "paragon_organization" "my_org" {
  name = "my_paragon_organization"
}
```