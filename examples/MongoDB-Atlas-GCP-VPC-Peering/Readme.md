# Example - GCP and MongoDB Atlas VPC Peering

This project aims to provide an example of using GCP and MongoDB Atlas together.


## Dependencies

* Terraform v0.15
* GCP Account
* A MongoDB Atlas account 

```
Terraform v0.15.3
on darwin_amd64
+ provider registry.terraform.io/hashicorp/google v3.74.0
+ provider registry.terraform.io/mongodb/mongodbatlas v0.9.1
```

## Usage

**1\. Ensure your GCP credentials are set up.**

1. Fetch the Json key from GCP for your project following GCP [documentation](https://cloud.google.com/iam/docs/creating-managing-service-account-keys).
2. Copy the `json` file to the root of the terrform configuration as `service-account.json`.


**2\. TFVARS**

Now create **terraform.tfvars** file with all the variable values and make sure **not to commit it**.

**3\. Review the Terraform plan.**

Execute the below command and ensure you are happy with the plan.

``` bash
$ terraform plan
```
This project currently does the below deployments:

- MongoDB Atlas GCP cluster - M10
- MongoDB Atlas Network Container
- MongoDB Atlas and GCP VPC peering, Routes Entry and IP Access Whitelisting

**4\. Execute the Terraform apply.**

Now execute the plan to provision the resources.

``` bash
$ terraform apply
```

**5\. Destroy the resources.**

Once you are finished your testing, ensure you destroy the resources to avoid unnecessary GCP and Atlas charges.

``` bash
$ terraform destroy
```
