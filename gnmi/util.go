package gnmi

import (
	"encoding/json"
	"fmt"
	"opennos-mgmt/gnmi/modeldata"
	"opennos-mgmt/gnmi/modeldata/oc"
	"os"
	"reflect"
	"strconv"

	log "github.com/golang/glog"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/experimental/ygotutils"
	"github.com/openconfig/ygot/ygot"
	cpb "google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/openconfig/gnmi/proto/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// getChildNode gets a node's child with corresponding schema specified by path
// element. If not found and createIfNotExist is set as true, an empty node is
// created and returned.
func getChildNode(node map[string]interface{}, schema *yang.Entry, elem *pb.PathElem, createIfNotExist bool) (interface{}, *yang.Entry) {
	var nextSchema *yang.Entry
	var ok bool

	if nextSchema, ok = schema.Dir[elem.Name]; !ok {
		return nil, nil
	}

	var nextNode interface{}
	if elem.GetKey() == nil {
		if nextNode, ok = node[elem.Name]; !ok {
			if createIfNotExist {
				node[elem.Name] = make(map[string]interface{})
				nextNode = node[elem.Name]
			}
		}
		return nextNode, nextSchema
	}

	nextNode = getKeyedListEntry(node, elem, createIfNotExist)
	return nextNode, nextSchema
}

// getKeyedListEntry finds the keyed list entry in node by the name and key of
// path elem. If entry is not found and createIfNotExist is true, an empty entry
// will be created (the list will be created if necessary).
func getKeyedListEntry(node map[string]interface{}, elem *pb.PathElem, createIfNotExist bool) map[string]interface{} {
	curNode, ok := node[elem.Name]
	if !ok {
		if !createIfNotExist {
			return nil
		}

		// Create a keyed list as node child and initialize an entry.
		m := make(map[string]interface{})
		for k, v := range elem.Key {
			m[k] = v
			if vAsNum, err := strconv.ParseFloat(v, 64); err == nil {
				m[k] = vAsNum
			}
		}
		node[elem.Name] = []interface{}{m}
		return m
	}

	// Search entry in keyed list.
	keyedList, ok := curNode.([]interface{})
	if !ok {
		return nil
	}
	for _, n := range keyedList {
		m, ok := n.(map[string]interface{})
		if !ok {
			log.Errorf("wrong keyed list entry type: %T", n)
			return nil
		}
		keyMatching := true
		// must be exactly match
		for k, v := range elem.Key {
			attrVal, ok := m[k]
			if !ok {
				return nil
			}
			if v != fmt.Sprintf("%v", attrVal) {
				keyMatching = false
				break
			}
		}
		if keyMatching {
			return m
		}
	}
	if !createIfNotExist {
		return nil
	}

	// Create an entry in keyed list.
	m := make(map[string]interface{})
	for k, v := range elem.Key {
		m[k] = v
		if vAsNum, err := strconv.ParseFloat(v, 64); err == nil {
			m[k] = vAsNum
		}
	}
	node[elem.Name] = append(keyedList, m)
	return m
}

func SaveConfigFile(config ygot.ValidatedGoStruct, filename string) error {
	model := NewModel(modeldata.ModelData,
		reflect.TypeOf((*oc.Device)(nil)),
		oc.SchemaTree["Device"],
		oc.Unmarshal,
		oc.ΛEnum)
	nilPath := &gnmi.Path{}
	node, stat := ygotutils.GetNode(model.schemaTreeRoot, config, nilPath)
	if isNil(node) || stat.GetCode() != int32(cpb.Code_OK) {
		return status.Errorf(codes.NotFound, "root path not found")
	}

	nodeStruct, _ := node.(ygot.GoStruct)
	// Return IETF JSON by default.
	jsonEncoder := func() (map[string]interface{}, error) {
		// AppendModuleName determines whether the module name is appended to elements
		return ygot.ConstructIETFJSON(nodeStruct, &ygot.RFC7951JSONConfig{AppendModuleName: false})
	}
	jsonType := "IETF"
	jsonTree, err := jsonEncoder()
	if err != nil {
		msg := fmt.Sprintf("error in constructing %s JSON tree from requested node: %v", jsonType, err)
		log.Error(msg)
		return status.Error(codes.Internal, msg)
	}

	jsonDump, err := json.MarshalIndent(jsonTree, "", "  ")
	if err != nil {
		msg := fmt.Sprintf("error in marshaling %s JSON tree to bytes: %v", jsonType, err)
		log.Error(msg)
		return status.Error(codes.Internal, msg)
	}

	// If the file doesn't exist, create it, or append to the file
	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(jsonDump)
	return err
}
