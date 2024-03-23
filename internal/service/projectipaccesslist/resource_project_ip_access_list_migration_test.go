package projectipaccesslist_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/mig"
)

func TestMigProjectIPAccessList_settingIPAddress(t *testing.T) {
	var (
		orgID       = os.Getenv("MONGODB_ATLAS_ORG_ID")
		projectName = acc.RandomProjectName()
		ipAddress   = acc.RandomIP(179, 154, 226)
		comment     = fmt.Sprintf("TestAcc for ipAddress (%s)", ipAddress)
		config      = acc.ConfigProjectIPAccessListWithIPAddress(orgID, projectName, ipAddress, comment)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { mig.PreCheckBasic(t) },
		CheckDestroy: acc.CheckDestroyProjectIPAccessList,
		Steps: []resource.TestStep{
			{
				ExternalProviders: mig.ExternalProviders(),
				Config:            config,
				Check:             resource.ComposeTestCheckFunc(commonChecks(ipAddress, "", "", comment)...),
			},
			mig.TestStepCheckEmptyPlan(config),
		},
	})
}

func TestMigProjectIPAccessList_settingCIDRBlock(t *testing.T) {
	var (
		orgID       = os.Getenv("MONGODB_ATLAS_ORG_ID")
		projectName = acc.RandomProjectName()
		cidrBlock   = acc.RandomIP(179, 154, 226) + "/32"
		comment     = fmt.Sprintf("TestAcc for cidrBlock (%s)", cidrBlock)
		config      = acc.ConfigProjectIPAccessListWithCIDRBlock(orgID, projectName, cidrBlock, comment)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { mig.PreCheckBasic(t) },
		CheckDestroy: acc.CheckDestroyProjectIPAccessList,
		Steps: []resource.TestStep{
			{
				ExternalProviders: mig.ExternalProviders(),
				Config:            config,
				Check:             resource.ComposeTestCheckFunc(commonChecks("", cidrBlock, "", comment)...),
			},
			mig.TestStepCheckEmptyPlan(config),
		},
	})
}

func TestMigProjectIPAccessList_settingAWSSecurityGroup(t *testing.T) {
	var (
		orgID        = os.Getenv("MONGODB_ATLAS_ORG_ID")
		awsSGroup    = os.Getenv("AWS_SECURITY_GROUP_1")
		vpcID        = os.Getenv("AWS_VPC_ID")
		vpcCIDRBlock = os.Getenv("AWS_VPC_CIDR_BLOCK")
		awsAccountID = os.Getenv("AWS_ACCOUNT_ID")
		awsRegion    = os.Getenv("AWS_REGION")
		providerName = "AWS"
		projectName  = acc.RandomProjectName()
		comment      = fmt.Sprintf("TestAcc for awsSecurityGroup (%s)", awsSGroup)
		config       = acc.ConfigProjectIPAccessListWithAWSSecurityGroup(orgID, projectName, providerName, vpcID, awsAccountID, vpcCIDRBlock, awsRegion, awsSGroup, comment)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { mig.PreCheckPeeringEnvAWS(t) },
		CheckDestroy: acc.CheckDestroyProjectIPAccessList,
		Steps: []resource.TestStep{
			{
				ExternalProviders: mig.ExternalProviders(),
				Config:            config,
				Check:             resource.ComposeTestCheckFunc(commonChecks("", "", awsSGroup, comment)...),
			},
			mig.TestStepCheckEmptyPlan(config),
		},
	})
}
