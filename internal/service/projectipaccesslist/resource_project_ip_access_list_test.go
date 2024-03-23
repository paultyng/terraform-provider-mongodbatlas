package projectipaccesslist_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
)

const (
	resourceName   = "mongodbatlas_project_ip_access_list.test"
	dataSourceName = "data.mongodbatlas_project_ip_access_list.test"
)

func TestAccProjectIPAccesslist_settingIPAddress(t *testing.T) {
	var (
		projectID        = acc.ProjectIDExecution(t)
		ipAddress        = acc.RandomIP(179, 154, 226)
		comment          = fmt.Sprintf("TestAcc for ipAddress (%s)", ipAddress)
		updatedIPAddress = acc.RandomIP(179, 154, 228)
		updatedComment   = fmt.Sprintf("TestAcc for ipAddress updated (%s)", updatedIPAddress)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             acc.CheckDestroyProjectIPAccessList,
		Steps: []resource.TestStep{
			{
				Config: configWithIPAddress(projectID, ipAddress, comment),
				Check:  resource.ComposeTestCheckFunc(commonChecks(ipAddress, "", "", comment)...),
			},
			{
				Config: configWithIPAddress(projectID, updatedIPAddress, updatedComment),
				Check:  resource.ComposeTestCheckFunc(commonChecks(updatedIPAddress, "", "", updatedComment)...),
			},
		},
	})
}

func TestAccProjectIPAccessList_settingCIDRBlock(t *testing.T) {
	var (
		projectID        = acc.ProjectIDExecution(t)
		cidrBlock        = acc.RandomIP(179, 154, 226) + "/32"
		comment          = fmt.Sprintf("TestAcc for cidrBlock (%s)", cidrBlock)
		updatedCIDRBlock = acc.RandomIP(179, 154, 228) + "/32"
		updatedComment   = fmt.Sprintf("TestAcc for cidrBlock updated (%s)", updatedCIDRBlock)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             acc.CheckDestroyProjectIPAccessList,
		Steps: []resource.TestStep{
			{
				Config: configWithCIDRBlock(projectID, cidrBlock, comment),
				Check:  resource.ComposeTestCheckFunc(commonChecks("", cidrBlock, "", comment)...),
			},
			{
				Config: configWithCIDRBlock(projectID, updatedCIDRBlock, updatedComment),
				Check:  resource.ComposeTestCheckFunc(commonChecks("", updatedCIDRBlock, "", updatedComment)...),
			},
		},
	})
}

func TestAccProjectIPAccessList_settingAWSSecurityGroup(t *testing.T) {
	var (
		vpcID            = os.Getenv("AWS_VPC_ID")
		vpcCIDRBlock     = os.Getenv("AWS_VPC_CIDR_BLOCK")
		awsAccountID     = os.Getenv("AWS_ACCOUNT_ID")
		awsRegion        = os.Getenv("AWS_REGION")
		orgID            = os.Getenv("MONGODB_ATLAS_ORG_ID")
		awsSGroup        = os.Getenv("AWS_SECURITY_GROUP_1")
		updatedAWSSgroup = os.Getenv("AWS_SECURITY_GROUP_2")
		providerName     = "AWS"
		projectName      = acc.RandomProjectName()
		comment          = fmt.Sprintf("TestAcc for awsSecurityGroup (%s)", awsSGroup)
		updatedComment   = fmt.Sprintf("TestAcc for awsSecurityGroup updated (%s)", updatedAWSSgroup)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckPeeringEnvAWS(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             acc.CheckDestroyProjectIPAccessList,
		Steps: []resource.TestStep{
			{
				Config: acc.ConfigProjectIPAccessListWithAWSSecurityGroup(orgID, projectName, providerName, vpcID, awsAccountID, vpcCIDRBlock, awsRegion, awsSGroup, comment),
				Check:  resource.ComposeTestCheckFunc(commonChecks("", "", awsSGroup, comment)...),
			},
			{
				Config: acc.ConfigProjectIPAccessListWithAWSSecurityGroup(orgID, projectName, providerName, vpcID, awsAccountID, vpcCIDRBlock, awsRegion, updatedAWSSgroup, updatedComment),
				Check:  resource.ComposeTestCheckFunc(commonChecks("", "", updatedAWSSgroup, updatedComment)...),
			},
		},
	})
}

func TestAccProjectIPAccessList_settingMultiple(t *testing.T) {
	var (
		resourceFmt      = "mongodbatlas_project_ip_access_list.test_%d"
		orgID            = os.Getenv("MONGODB_ATLAS_ORG_ID")
		projectName      = acc.RandomProjectName()
		ipWhiteListCount = 20
		accessList       = []map[string]string{}
		checks           = []resource.TestCheckFunc{}
	)

	for i := 0; i < ipWhiteListCount; i++ {
		entry := make(map[string]string)
		entryName := ""
		ipAddr := ""

		if i%2 == 0 {
			entryName = "cidr_block"
			entry["cidr_block"] = acc.RandomIP(byte(i), 2, 3) + "/32"
			ipAddr = entry["cidr_block"]
		} else {
			entryName = "ip_address"
			entry["ip_address"] = acc.RandomIP(byte(i), 2, 3)
			ipAddr = entry["ip_address"]
		}
		entry["comment"] = fmt.Sprintf("TestAcc for %s (%s)", entryName, ipAddr)

		accessList = append(accessList, entry)
		checks = append(checks, acc.CheckProjectIPAccessListExists(fmt.Sprintf(resourceFmt, i)))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             acc.CheckDestroyProjectIPAccessList,
		Steps: []resource.TestStep{
			{
				Config: acc.ConfigProjectIPAccessListWithMultiple(projectName, orgID, accessList, false),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
			{
				Config: acc.ConfigProjectIPAccessListWithMultiple(projectName, orgID, accessList, true),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccProjectIPAccessList_importBasic(t *testing.T) {
	var (
		projectID = acc.ProjectIDExecution(t)
		ipAddress = acc.RandomIP(179, 154, 226)
		comment   = fmt.Sprintf("TestAcc for ipaddres (%s)", ipAddress)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             acc.CheckDestroyProjectIPAccessList,
		Steps: []resource.TestStep{
			{
				Config: configWithIPAddress(projectID, ipAddress, comment),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: acc.ImportStateProjecIPAccessListtIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccProjectIPAccessList_importIncorrectId(t *testing.T) {
	var (
		projectID = acc.ProjectIDExecution(t)
		ipAddress = acc.RandomIP(179, 154, 226)
		comment   = fmt.Sprintf("TestAcc for ipaddres (%s)", ipAddress)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             acc.CheckDestroyProjectIPAccessList,
		Steps: []resource.TestStep{
			{
				Config: configWithIPAddress(projectID, ipAddress, comment),
			},
			{
				ResourceName:  resourceName,
				ImportState:   true,
				ImportStateId: "incorrect_id_without_project_id_and_dash",
				ExpectError:   regexp.MustCompile("import format error"),
			},
		},
	})
}

func commonChecks(ipAddress, cidrBlock, awsSGroup, comment string) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		acc.CheckProjectIPAccessListExists(resourceName),
		acc.CheckProjectIPAccessListExists(dataSourceName),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(dataSourceName, "project_id"),
		resource.TestCheckResourceAttr(resourceName, "comment", comment),
		resource.TestCheckResourceAttr(dataSourceName, "comment", comment),
	}
	if ipAddress != "" {
		checks = append(checks,
			resource.TestCheckResourceAttr(resourceName, "ip_address", ipAddress),
			resource.TestCheckResourceAttr(dataSourceName, "ip_address", ipAddress))
	}
	if cidrBlock != "" {
		checks = append(checks,
			resource.TestCheckResourceAttr(resourceName, "cidr_block", cidrBlock),
			resource.TestCheckResourceAttr(dataSourceName, "cidr_block", cidrBlock))
	}
	if awsSGroup != "" {
		checks = append(checks,
			resource.TestCheckResourceAttr(resourceName, "aws_security_group", awsSGroup),
			resource.TestCheckResourceAttr(dataSourceName, "aws_security_group", awsSGroup))
	}
	return checks
}

func configWithIPAddress(projectID, ipAddress, comment string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_project_ip_access_list" "test" {
			project_id = %[1]q
			ip_address = %[2]q
			comment    = %[3]q
		}

		data "mongodbatlas_project_ip_access_list" "test" {
			project_id = mongodbatlas_project_ip_access_list.test.project_id
			ip_address = mongodbatlas_project_ip_access_list.test.ip_address
		}
	`, projectID, ipAddress, comment)
}

func configWithCIDRBlock(projectID, cidrBlock, comment string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_project_ip_access_list" "test" {
			project_id = %[1]q
			cidr_block = %[2]q
			comment    = %[3]q
		}

		data "mongodbatlas_project_ip_access_list" "test" {
			project_id = mongodbatlas_project_ip_access_list.test.project_id
			cidr_block = mongodbatlas_project_ip_access_list.test.cidr_block
		}
	`, projectID, cidrBlock, comment)
}
