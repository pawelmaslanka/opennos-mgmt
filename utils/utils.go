// Package utils implements utilities for opennos-mgmt.
package utils

import (
	"encoding/json"
	"flag"
	"fmt"
	"opennos-mgmt/gnmi/modeldata/oc"

	"github.com/golang/protobuf/proto"
	"github.com/kylelemons/godebug/pretty"

	lib "golibext"

	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
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
func ConvertGoInterfaceIntoString(valueToConvert interface{}) (string, error) {
	var value string
	switch v := valueToConvert.(type) {
	case *string:
		value = *v
	case string:
		value = v
	default:
		return "", fmt.Errorf("Unexpected interface type to conversion: %v", v)
	}

	return value, nil
}

// ConvertGoInterfaceIntoUint64 converts Go interface{} into int64 value
func ConvertGoInterfaceIntoInt64(valueToConvert interface{}) (int64, error) {
	var value int64
	switch v := valueToConvert.(type) {
	case *oc.E_OpenconfigIfAggregate_AggregationType:
		value = int64(*v)
	case oc.E_OpenconfigIfAggregate_AggregationType:
		value = int64(v)
	case *int64:
		value = *v
	case int64:
		value = v
	default:
		return 0, fmt.Errorf("Cannot convert %v to any of [int64, oc.E_OpenconfigIfAggregate_AggregationType], unsupported type, got: %T", v, v)
	}

	return value, nil
}

// ConvertGoInterfaceIntoUint16 converts Go interface{} into uint16 value
func ConvertGoInterfaceIntoUint16(valueToConvert interface{}) (uint16, error) {
	var value uint16
	switch v := valueToConvert.(type) {
	case *oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_Uint16:
		value = v.Uint16
	case oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_Uint16:
		value = v.Uint16
	case *lib.VidT:
		value = uint16(*v)
	case lib.VidT:
		value = uint16(v)
	case *uint16:
		value = *v
	case uint16:
		value = v
	default:
		return 0, fmt.Errorf("Cannot convert %v to any of [uint16, vidT, Interface_Ethernet_SwitchedVlan_TrunkVlans_Union], unsupported type, got: %T", v, v)
	}

	return value, nil
}

// ConvertGoInterfaceIntoUint8 converts Go interface{} into uint8 value
func ConvertGoInterfaceIntoUint8(value interface{}) (uint8, error) {
	var rv uint8
	switch v := value.(type) {
	case *oc.E_OpenconfigIfEthernet_ETHERNET_SPEED:
		rv = uint8(*v)
	case oc.E_OpenconfigIfEthernet_ETHERNET_SPEED:
		rv = uint8(v)
	case *uint8:
		rv = *v
	case uint8:
		rv = v
	default:
		return 0, fmt.Errorf("Cannot convert %v to any of [uint8, E_OpenconfigIfEthernet_ETHERNET_SPEED], unsupported union type, got: %T", v, v)
	}

	return rv, nil
}

func GetJsonDiff(a []byte, b []byte) (string, error) {
	differ := diff.New()
	d, err := differ.Compare(a, b)
	if err != nil {
		return "", fmt.Errorf("Failed to unmarshal file: %s", err.Error())
	}

	var diffString string
	// Output the result
	if d.Modified() {
		format := "ascii"
		if format == "ascii" {
			var aJson map[string]interface{}
			json.Unmarshal(a, &aJson)

			config := formatter.AsciiFormatterConfig{
				ShowArrayIndex: false,
				Coloring:       false,
			}

			formatter := formatter.NewAsciiFormatter(aJson, config)
			diffString, err = formatter.Format(d)
			if err != nil {
				// No error can occur
			}
		} else if format == "delta" {
			formatter := formatter.NewDeltaFormatter()
			diffString, err = formatter.Format(d)
			if err != nil {
				// No error can occur
			}
		}
	}

	return diffString, nil
}
