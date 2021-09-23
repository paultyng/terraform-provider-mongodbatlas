---
layout: "mongodbatlas"
page_title: "MongoDB Atlas: encryption_at_rest"
sidebar_current: "docs-mongodbatlas-resource-encryption_at_rest"
description: |-
    Provides an Encryption At Rest resource.
---

# mongodbatlas_encryption_at_rest

`mongodbatlas_encryption_at_rest` Allows management of encryption at rest for an Atlas project with one of the following providers:

[Amazon Web Services Key Management Service](https://docs.atlas.mongodb.com/security-aws-kms/#security-aws-kms)
[Azure Key Vault](https://docs.atlas.mongodb.com/security-azure-kms/#security-azure-kms)
[Google Cloud KMS](https://docs.atlas.mongodb.com/security-gcp-kms/#security-gcp-kms)

After configuring at least one Encryption at Rest provider for the Atlas project, Project Owners can enable Encryption at Rest for each Atlas cluster for which they require encryption. The Encryption at Rest provider does not have to match the cluster cloud service provider.

Atlas does not automatically rotate user-managed encryption keys. Defer to your preferred Encryption at Rest provider’s documentation and guidance for best practices on key rotation. Atlas automatically creates a 365-day key rotation alert when you configure Encryption at Rest using your Key Management in an Atlas project.

See [Encryption at Rest](https://docs.atlas.mongodb.com/security-kms-encryption/index.html) for more information, including prerequisites and restrictions.

~> **IMPORTANT** Atlas encrypts all cluster storage and snapshot volumes, securing all cluster data on disk: a concept known as encryption at rest, by default.

-> **NOTE:** Groups and projects are synonymous terms. You may find `groupId` in the official documentation.


-> **IMPORTANT NOTE** To disable the encryption at rest with customer key management for a project all existing clusters in the project must first either have encryption at rest for the provider set to none, e.g. `encryption_at_rest_provider = "NONE"`, or be deleted.

## Example Usage

```terraform
resource "mongodbatlas_encryption_at_rest" "test" {
  project_id = "<PROJECT-ID>"

  aws_kms_config {
    enabled                = true
    customer_master_key_id = "5ce83906-6563-46b7-8045-11c20e3a5766"
    region                 = "US_EAST_1"
    role_id                = "60815e2fe01a49138a928ebb"
  }

  azure_key_vault_config {
    enabled             = true
    client_id           = "g54f9e2-89e3-40fd-8188-EXAMPLEID"
    azure_environment   = "AZURE"
    subscription_id     = "0ec944e3-g725-44f9-a147-EXAMPLEID"
    resource_group_name = "ExampleRGName"
    key_vault_name      = "EXAMPLEKeyVault"
    key_identifier      = "https://EXAMPLEKeyVault.vault.azure.net/keys/EXAMPLEKey/d891821e3d364e9eb88fbd3d11807b86"
    secret              = "EXAMPLESECRET"
    tenant_id           = "e8e4b6ba-ff32-4c88-a9af-EXAMPLEID"
  }

  google_cloud_kms_config {
    enabled                 = true
    service_account_key     = "{\"type\": \"service_account\",\"project_id\": \"my-project-common-0\",\"private_key_id\": \"e120598ea4f88249469fcdd75a9a785c1bb3\",\"private_key\": \"-----BEGIN PRIVATE KEY-----\\nMIIEuwIBA(truncated)SfecnS0mT94D9\\n-----END PRIVATE KEY-----\\n\",\"client_email\": \"my-email-kms-0@my-project-common-0.iam.gserviceaccount.com\",\"client_id\": \"10180967717292066\",\"auth_uri\": \"https://accounts.google.com/o/oauth2/auth\",\"token_uri\": \"https://accounts.google.com/o/oauth2/token\",\"auth_provider_x509_cert_url\": \"https://www.googleapis.com/oauth2/v1/certs\",\"client_x509_cert_url\": \"https://www.googleapis.com/robot/v1/metadata/x509/my-email-kms-0%40my-project-common-0.iam.gserviceaccount.com\"}"
    key_version_resource_id = "projects/my-project-common-0/locations/us-east4/keyRings/my-key-ring-0/cryptoKeys/my-key-0/cryptoKeyVersions/1"
  }
}
```

**NOTE**  if using the two resources path for cloud provider access, `cloud_provider_access_setup` and `cloud_provider_access_authorization`, you may need to define a `depends_on` statement for these two resources, because terraform is not able to infer the dependency.

```terraform
resource "mongodbatlas_encryption_at_rest" "default" {
  (...)
  depends_on = [mongodbatlas_cloud_provider_access_setup.<resource_name>, mongodbatlas_cloud_provider_access_authorization.<resource_name>]
}
```

## Example: Configuring encryption at rest using customer key management in Azure and then creating a cluster

The configuration of encryption at rest with customer key management, `mongodbatlas_encryption_at_rest`, needs to be completed before a cluster is created in the project. Force this wait by using an implicit dependency via `project_id` as shown in the example below.

```terraform
resource "mongodbatlas_encryption_at_rest" "example" {
  project_id = "<PROJECT-ID>"

  azure_key_vault_config {
    enabled             = true
    client_id           = "g54f9e2-89e3-40fd-8188-EXAMPLEID"
    azure_environment   = "AZURE"
    subscription_id     = "0ec944e3-g725-44f9-a147-EXAMPLEID"
    resource_group_name = "ExampleRGName"
    key_vault_name      = "EXAMPLEKeyVault"
    key_identifier      = "https://EXAMPLEKeyVault.vault.azure.net/keys/EXAMPLEKey/d891821e3d364e9eb88fbd3d11807b86"
    secret              = "EXAMPLESECRET"
    tenant_id           = "e8e4b6ba-ff32-4c88-a9af-EXAMPLEID"
  }
}

resource "mongodbatlas_cluster" "example_cluster" {
  project_id   = mongodbatlas_encryption_at_rest.example.project_id
  name         = "CLUSTER NAME"
  cluster_type = "REPLICASET"
  replication_specs {
    num_shards = 1
    regions_config {
      region_name     = "REGION"
      electable_nodes = 3
      priority        = 7
      read_only_nodes = 0
    }
  }

  provider_name               = "AZURE"
  provider_instance_size_name = "M10"
  mongo_db_major_version      = "4.4"
  encryption_at_rest_provider = "AZURE"
}
```

## Argument Reference

* `project_id` - (Required) The unique identifier for the project.
* `aws_kms` - (Required) Specifies AWS KMS configuration details and whether Encryption at Rest is enabled for an Atlas project.
* `azure_key_vault` - (Required) Specifies Azure Key Vault configuration details and whether Encryption at Rest is enabled for an Atlas project.
* `google_cloud_kms` - (Required) Specifies GCP KMS configuration details and whether Encryption at Rest is enabled for an Atlas project.

### aws_kms_config
Refer to the example in the [official github repository](https://github.com/mongodb/terraform-provider-mongodbatlas/tree/master/examples) to implement Encryption at Rest
* `enabled` - Specifies whether Encryption at Rest is enabled for an Atlas project, To disable Encryption at Rest, pass only this parameter with a value of false, When you disable Encryption at Rest, Atlas also removes the configuration details.
* `customer_master_key_id` - The AWS customer master key used to encrypt and decrypt the MongoDB master keys.
* `region` - The AWS region in which the AWS customer master key exists: CA_CENTRAL_1, US_EAST_1, US_EAST_2, US_WEST_1, US_WEST_2, SA_EAST_1
* `role_id` - ID of an AWS IAM role authorized to manage an AWS customer master key. To find the ID for an existing IAM role check the `role_id` attribute of the `mongodbatlas_cloud_provider_access` resource.

### azure_key_vault_config
* `enabled` - Specifies whether Encryption at Rest is enabled for an Atlas project. To disable Encryption at Rest, pass only this parameter with a value of false. When you disable Encryption at Rest, Atlas also removes the configuration details.
* `client_id` - The client ID, also known as the application ID, for an Azure application associated with the Azure AD tenant.
* `azure_environment` - The Azure environment where the Azure account credentials reside. Valid values are the following: AZURE, AZURE_CHINA, AZURE_GERMANY
* `subscription_id` - The unique identifier associated with an Azure subscription.
* `resource_group_name` - The name of the Azure Resource group that contains an Azure Key Vault.
* `key_vault_name` - The name of an Azure Key Vault containing your key.
* `key_identifier` - The unique identifier of a key in an Azure Key Vault.
* `secret` - The secret associated with the Azure Key Vault specified by azureKeyVault.tenantID.
* `tenant_id` - The unique identifier for an Azure AD tenant within an Azure subscription.

### google_cloud_kms_config
* `enabled` - Specifies whether Encryption at Rest is enabled for an Atlas project. To disable Encryption at Rest, pass only this parameter with a value of false. When you disable Encryption at Rest, Atlas also removes the configuration details.
* `service_account_key` - String-formatted JSON object containing GCP KMS credentials from your GCP account.
* `key_version_resource_id` - The Key Version Resource ID from your GCP account.


For more information see: [MongoDB Atlas API Reference for Encryption at Rest using Customer Key Management.](https://docs.atlas.mongodb.com/reference/api/encryption-at-rest/)