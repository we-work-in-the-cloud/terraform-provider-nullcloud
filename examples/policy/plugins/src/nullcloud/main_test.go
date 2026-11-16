package main

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-policy-plugin-framework/policy-plugin/plugins"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

func TestValidCRN(t *testing.T) {
	for _, typ := range []string{"vpc", "subnet", "instance", "loadbalancer", "bucket", "database", "cluster"} {
		t.Run(typ, func(t *testing.T) {
			got, err := plugins.CallFunction("valid_crn", cty.StringVal("crn:nullcloud:"+typ+":id-1"), cty.StringVal(typ))
			if err != nil || !got.RawEquals(cty.True) {
				t.Fatalf("got %s, %v; want true", got.GoString(), err)
			}
		})
	}
	for _, tc := range []struct{ name, crn, typ string }{
		{"mismatch", "crn:nullcloud:vpc:id", "bucket"},
		{"unsupported", "crn:nullcloud:other:id", "other"},
		{"unsupported_expected", "crn:nullcloud:bucket:id", "other"},
		{"empty_expected", "crn:nullcloud:bucket:id", ""},
		{"empty_id", "crn:nullcloud:bucket:", "bucket"},
		{"space", "crn:nullcloud:bucket:id x", "bucket"},
		{"tab", "crn:nullcloud:bucket:id\tx", "bucket"},
		{"newline", "crn:nullcloud:bucket:id\nx", "bucket"},
		{"unicode_space", "crn:nullcloud:bucket:id\u00a0x", "bucket"},
		{"wrong_prefix", "arn:nullcloud:bucket:id", "bucket"},
		{"wrong_cloud", "crn:other:bucket:id", "bucket"},
		{"case", "CRN:nullcloud:bucket:id", "bucket"},
		{"type_case", "crn:nullcloud:Bucket:id", "bucket"},
		{"expected_case", "crn:nullcloud:bucket:id", "Bucket"},
		{"extra", "crn:nullcloud:bucket:id:extra", "bucket"},
		{"missing", "crn:nullcloud:bucket", "bucket"},
		{"empty", "", "bucket"},
		{"leading_space", " crn:nullcloud:bucket:id", "bucket"},
		{"trailing_space", "crn:nullcloud:bucket:id ", "bucket"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := plugins.CallFunction("valid_crn", cty.StringVal(tc.crn), cty.StringVal(tc.typ))
			if err != nil || !got.RawEquals(cty.False) {
				t.Fatalf("got %s, %v; want false", got.GoString(), err)
			}
		})
	}
}

func TestValidCRNArgumentDiagnostics(t *testing.T) {
	for _, bad := range []cty.Value{cty.NumberIntVal(1), cty.True, cty.EmptyObjectVal, cty.EmptyTupleVal, cty.NullVal(cty.String)} {
		for index := range 2 {
			args := []cty.Value{cty.StringVal("crn:nullcloud:bucket:id"), cty.StringVal("bucket")}
			args[index] = bad
			_, err := plugins.CallFunction("valid_crn", args...)
			var argErr function.ArgError
			if !errors.As(err, &argErr) || argErr.Index != index {
				t.Errorf("argument %d = %s: got %v; want argument diagnostic", index, bad.GoString(), err)
			}
		}
	}
}
