package config

import (
	"fmt"
	lib "golibext"
	"opennos-mgmt/gnmi/modeldata/oc"
)

type countT uint16

func convertInterfaceIntoString(value interface{}) (string, error) {
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

func convertInterfaceIntoVlanId(vlanId interface{}) (lib.VidT, error) {
	var vid lib.VidT
	switch v := vlanId.(type) {
	case *oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_Uint16:
		vid = lib.VidT(v.Uint16)
	case oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_Uint16:
		vid = lib.VidT(v.Uint16)
	case *uint16:
		vid = lib.VidT(*v)
	case uint16:
		vid = lib.VidT(v)
	default:
		return 0, fmt.Errorf("Cannot convert %v to any of [uint16, Interface_Ethernet_SwitchedVlan_TrunkVlans_Union], unsupported union type, got: %T", v, v)
	}

	return vid, nil
}
