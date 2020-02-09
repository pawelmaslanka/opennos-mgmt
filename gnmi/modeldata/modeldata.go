/* Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package modeldata contains the following model data in gnmi proto struct:
//	openconfig-interfaces 2.4.3,
//	openconfig-if-aggregate 2.4.2,
//	openconfig-if-ethernet 2.7.2,
//	openconfig-if-ip 3.0.0,
//	openconfig-if-ip-ext 2.3.1,
//	openconfig-lacp 1.1.1,
//	openconfig-lldp 0.2.1,
//	openconfig-platform-transceiver 0.7.0,
//  openconfig-spanning-tree 0.3.1.
package modeldata

import (
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

const (
	// OpenconfigInterfacesModel is the openconfig YANG model for interfaces.
	OpenconfigInterfacesModel = "openconfig-interfaces"
	// OpenconfigInterfaceAggregateModel is the openconfig YANG model for aggregate interface.
	OpenconfigInterfaceAggregateModel = "openconfig-if-aggregate"
	// OpenconfigInterfaceEthernetModel is the openconfig YANG model for ethernet interface.
	OpenconfigInterfaceEthernetModel = "openconfig-if-ethernet"
	// OpenconfigInterfaceIPModel is the openconfig YANG model for IP on interface.
	OpenconfigInterfaceIPModel = "openconfig-if-ip"
	// OpenconfigInterfaceIPExtModel is the openconfig YANG model for extended IP on interface.
	OpenconfigInterfaceIPExtModel = "openconfig-if-ip-ext"
	// OpenconfigLACPModel is the openconfig YANG model for LACP.
	OpenconfigLACPModel = "openconfig-lacp"
	// OpenconfigLLDPModel is the openconfig YANG model for LLDP.
	OpenconfigLLDPModel = "openconfig-lldp"
	// OpenconfigPlatformTransceiverModel is the openconfig YANG model for platform transceiver.
	OpenconfigPlatformTransceiverModel = "openconfig-platform-transceiver"
	// OpenconfigSTPModel is the openconfig YANG model for STP.
	OpenconfigSTPModel = "openconfig-spanning-tree"
)

var (
	// ModelData is a list of supported models.
	ModelData = []*pb.ModelData{{
		Name:         OpenconfigInterfacesModel,
		Organization: "OpenConfig working group",
		Version:      "2.4.3",
	}, {
		Name:         OpenconfigInterfaceAggregateModel,
		Organization: "OpenConfig working group",
		Version:      "2.4.2",
	}, {
		Name:         OpenconfigInterfaceEthernetModel,
		Organization: "OpenConfig working group",
		Version:      "2.7.2",
	}, {
		Name:         OpenconfigInterfaceIPModel,
		Organization: "OpenConfig working group",
		Version:      "3.0.0",
	}, {
		Name:         OpenconfigInterfaceIPExtModel,
		Organization: "OpenConfig working group",
		Version:      "2.3.1",
	}, {
		Name:         OpenconfigLACPModel,
		Organization: "OpenConfig working group",
		Version:      "1.1.1",
	}, {
		Name:         OpenconfigLLDPModel,
		Organization: "OpenConfig working group",
		Version:      "0.2.1",
	}, {
		Name:         OpenconfigPlatformTransceiverModel,
		Organization: "OpenConfig working group",
		Version:      "0.7.0",
	}, {
		Name:         OpenconfigSTPModel,
		Organization: "OpenConfig working group",
		Version:      "0.3.1",
	}}
)
