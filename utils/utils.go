// Package utils implements utilities for opennos-mgmt.
package utils

import (
	"flag"
	"fmt"
	"opennos-mgmt/gnmi/modeldata/oc"

	"github.com/golang/protobuf/proto"
	"github.com/kylelemons/godebug/pretty"
)

var (
	usePretty = flag.Bool("pretty", false, "Shows PROTOs using Pretty package instead of PROTO Text Marshal")
)

// PrintProto prints a Proto in a structured way.
func PrintProto(m proto.Message) {
	if *usePretty {
		pretty.Print(m)
		return
	}
	fmt.Println(proto.MarshalTextString(m))
}

// ConvertGoInterfaceIntoString converts Go interface{} into string value
func ConvertGoInterfaceIntoString(value interface{}) (string, error) {
	var str string
	switch v := value.(type) {
	case *string:
		str = *v
	case string:
		str = v
	default:
		return "", fmt.Errorf("Unexpected interface type to conversion: %v", value)
	}

	return str, nil
}

// ConvertGoInterfaceIntoUint16 converts Go interface{} into uint16 value
func ConvertGoInterfaceIntoUint16(vlanId interface{}) (uint16, error) {
	var vid uint16
	switch v := vlanId.(type) {
	case *oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_Uint16:
		vid = v.Uint16
	case oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_Uint16:
		vid = v.Uint16
	case *uint16:
		vid = *v
	case uint16:
		vid = v
	default:
		return 0, fmt.Errorf("Cannot convert %v to any of [uint16, Interface_Ethernet_SwitchedVlan_TrunkVlans_Union], unsupported union type, got: %T", v, v)
	}

	return vid, nil
}

// ConvertGoInterfaceIntoUint8 converts Go interface{} into uint8 value
func ConvertGoInterfaceIntoUint8(value interface{}) (uint8, error) {
	var rv uint8
	switch v := value.(type) {
	case *uint8:
		rv = *v
	case uint8:
		rv = v
	default:
		return 0, fmt.Errorf("Cannot convert %v to any of [uint16, Interface_Ethernet_SwitchedVlan_TrunkVlans_Union], unsupported union type, got: %T", v, v)
	}

	return rv, nil
}
