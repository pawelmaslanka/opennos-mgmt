/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package vlan_mgmt

import (
	pb "bcm-eth-switch-mgmt/grpc_services/vlan"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

const (
	kAddress = "localhost:50056"
)

type EthSwitchVlanMgmt struct {
	clientConn     *grpc.ClientConn
	vlanMgmtClient pb.VlanMgmtClient
}

var gCtx EthSwitchVlanMgmt

func ConnectWithGrpcService() error {
	// Set up a connection to the server.
	conn, err := grpc.Dial(kAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to dial into gRPC server: %v", err)
		return err
	}

	c := pb.NewVlanMgmtClient(conn)
	gCtx.clientConn = conn
	gCtx.vlanMgmtClient = c
	return nil
}

func CloseConnWithGrpcService() error {
	gCtx.clientConn.Close()
	return nil
}

// SetNativeVlan sets native VLAN on given interfaces. Interface can be given 
// as physial front panel port or logical LAG interface. This function is responsible 
// for parse LAG interface and pass its port members.
func SetNativeVlan(ifnames []string, vid uint16) error {
	ports := make([]*pb.Port, len(ifnames))
	for i := 0; i < len(ifnames); i++ {
		// TODO: Extract LAG members
		// if strings.Contains(ifname, "ae") {
		// }
		ports[i] = &pb.Port{ Name: ifnames[i] }
	}

	if err := ConnectWithGrpcService(); err != nil {
		log.Fatalf("Failed to connect with VLAN management gRPC service server: %v", err)
		return err
	}
	defer CloseConnWithGrpcService()
	log.Printf("Set native VLAN %d on %d ports", vid, len(ifnames))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := gCtx.vlanMgmtClient.SetNativeVlan(ctx, &pb.NativeVlan{
		Ports: ports,
		Vid:   uint32(vid),
	})

	if err != nil {
		log.Fatalf("Failed to set native VLAN for interface: %v", err)
		return err
	}

	log.Printf("SetNativeVlan() result: %s", r.GetResult())
	return nil
}
