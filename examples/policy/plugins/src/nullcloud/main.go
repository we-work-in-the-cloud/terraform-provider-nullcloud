package main

import (
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-policy-plugin-framework/policy-plugin/plugins"
)

func init() {
	plugins.RegisterFunction("valid_crn", validCRN)
}

func main() {
	plugins.Serve()
}

func validCRN(crn, expectedType string) (bool, error) {
	parts := strings.Split(crn, ":")
	if len(parts) != 4 || parts[0] != "crn" || parts[1] != "nullcloud" || parts[2] != expectedType || parts[3] == "" || strings.ContainsFunc(parts[3], unicode.IsSpace) {
		return false, nil
	}
	switch expectedType {
	case "vpc", "subnet", "instance", "loadbalancer", "bucket", "database", "cluster":
		return true, nil
	default:
		return false, nil
	}
}
