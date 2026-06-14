package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type bucketModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	CRN       types.String `tfsdk:"crn"`
	Region    types.String `tfsdk:"region"`
	CreatedAt types.String `tfsdk:"created_at"`
}

type kubernetesClusterModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Version   types.String `tfsdk:"version"`
	Status    types.String `tfsdk:"status"`
	CRN       types.String `tfsdk:"crn"`
	NodeCount types.Int64  `tfsdk:"node_count"`
	SubnetIDs types.List   `tfsdk:"subnet_ids"`
	CreatedAt types.String `tfsdk:"created_at"`
}

type databaseModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Engine    types.String `tfsdk:"engine"`
	Version   types.String `tfsdk:"version"`
	Status    types.String `tfsdk:"status"`
	CRN       types.String `tfsdk:"crn"`
	Plan      types.String `tfsdk:"plan"`
	SubnetIDs types.List   `tfsdk:"subnet_ids"`
	CreatedAt types.String `tfsdk:"created_at"`
	Endpoint  types.String `tfsdk:"endpoint"`
}

type instanceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	SubnetID  types.String `tfsdk:"subnet_id"`
	Profile   types.String `tfsdk:"profile"`
	Image     types.String `tfsdk:"image"`
	Status    types.String `tfsdk:"status"`
	CRN       types.String `tfsdk:"crn"`
	PrimaryIP types.String `tfsdk:"primary_ip"`
	CreatedAt types.String `tfsdk:"created_at"`
}

type loadBalancerModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	CRN       types.String `tfsdk:"crn"`
	Protocol  types.String `tfsdk:"protocol"`
	Port      types.Int64  `tfsdk:"port"`
	Targets   types.List   `tfsdk:"targets"`
	CreatedAt types.String `tfsdk:"created_at"`
}

type loadBalancerTargetModel struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

type subnetModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	VPCID     types.String `tfsdk:"vpc_id"`
	Zone      types.String `tfsdk:"zone"`
	Status    types.String `tfsdk:"status"`
	CRN       types.String `tfsdk:"crn"`
	CIDRBlock types.String `tfsdk:"cidr_block"`
	CreatedAt types.String `tfsdk:"created_at"`
}

type vpcModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	CRN       types.String `tfsdk:"crn"`
	Region    types.String `tfsdk:"region"`
	CreatedAt types.String `tfsdk:"created_at"`
}
