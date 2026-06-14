package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func reportNotFoundDataSource(resp *datasource.ReadResponse, resourceType, id string) {
	resp.Diagnostics.AddError(
		fmt.Sprintf("%s not found", resourceType),
		fmt.Sprintf("%s with ID %q not found", resourceType, id),
	)
}

func listOfStrings(values []string) types.List {
	elements := make([]attr.Value, len(values))
	for i, v := range values {
		elements[i] = types.StringValue(v)
	}
	listValue, _ := types.ListValue(types.StringType, elements)
	return listValue
}
