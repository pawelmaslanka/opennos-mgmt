package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/abiosoft/ishell"
	log "github.com/golang/glog"
	"github.com/jinzhu/copier"
	"github.com/r3labs/diff"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"opennos-mgmt/gnmi"
	"opennos-mgmt/gnmi/modeldata"
	"opennos-mgmt/gnmi/modeldata/oc"

	vlan "opennos-mgmt/management/vlan"
	"opennos-mgmt/utils/credentials"

	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"

	lib "golibext"
)

var validIfaces = [...]string{
	"eth-1", "eth-1/1", "eth-1/2", "eth-1/3", "eth-1/4",
	"eth-2", "eth-2/1", "eth-2/2", "eth-2/3", "eth-2/4",
	"eth-3", "eth-3/1", "eth-3/2", "eth-3/3", "eth-3/4",
	"eth-4", "eth-4/1", "eth-4/2", "eth-4/3", "eth-4/4",
	"eth-5", "eth-5/1", "eth-5/2", "eth-5/3", "eth-5/4",
	"eth-6", "eth-6/1", "eth-6/2", "eth-6/3", "eth-6/4",
	"eth-7", "eth-7/1", "eth-7/2", "eth-7/3", "eth-7/4",
	"eth-8", "eth-8/1", "eth-8/2", "eth-8/3", "eth-8/4",
	"eth-9", "eth-9/1", "eth-9/2", "eth-9/3", "eth-9/4",
	"eth-10", "eth-10/1", "eth-10/2", "eth-10/3", "eth-10/4",
	"eth-11", "eth-11/1", "eth-11/2", "eth-11/3", "eth-11/4",
	"eth-12", "eth-12/1", "eth-12/2", "eth-12/3", "eth-12/4",
	"eth-13", "eth-13/1", "eth-13/2", "eth-13/3", "eth-13/4",
	"eth-14", "eth-14/1", "eth-14/2", "eth-14/3", "eth-14/4",
	"eth-15", "eth-15/1", "eth-15/2", "eth-15/3", "eth-15/4",
	"eth-16", "eth-16/1", "eth-16/2", "eth-16/3", "eth-16/4",
	"eth-17", "eth-17/1", "eth-17/2", "eth-17/3", "eth-17/4",
	"eth-18", "eth-18/1", "eth-18/2", "eth-18/3", "eth-18/4",
	"eth-19", "eth-19/1", "eth-19/2", "eth-19/3", "eth-19/4",
	"eth-20", "eth-20/1", "eth-20/2", "eth-20/3", "eth-20/4",
	"eth-21", "eth-21/1", "eth-21/2", "eth-21/3", "eth-21/4",
	"eth-22", "eth-22/1", "eth-22/2", "eth-22/3", "eth-22/4",
	"eth-32", "eth-23/1", "eth-23/2", "eth-23/3", "eth-23/4",
	"eth-24", "eth-24/1", "eth-24/2", "eth-24/3", "eth-24/4",
	"eth-25", "eth-25/1", "eth-25/2", "eth-25/3", "eth-25/4",
	"eth-26", "eth-26/1", "eth-26/2", "eth-26/3", "eth-26/4",
	"eth-27", "eth-27/1", "eth-27/2", "eth-27/3", "eth-27/4",
	"eth-28", "eth-28/1", "eth-28/2", "eth-28/3", "eth-28/4",
	"eth-29", "eth-29/1", "eth-29/2", "eth-29/3", "eth-29/4",
	"eth-30", "eth-30/1", "eth-30/2", "eth-30/3", "eth-30/4",
	"eth-31", "eth-31/1", "eth-31/2", "eth-31/3", "eth-31/4",
	"eth-32", "eth-32/1", "eth-32/2", "eth-32/3", "eth-32/4",
}

var gEditIfaceCmdCompleterInvoked bool = false

type Iface struct {
	speed uint32
}

func NewIface() *Iface {
	return &Iface{}
}

type EditIfaceCmdCtx struct {
	completerInvoked bool
}

func NewEditIfaceCmdCtx() *EditIfaceCmdCtx {
	return &EditIfaceCmdCtx{
		completerInvoked: false,
	}
}

var editIfaceCmdCtx *EditIfaceCmdCtx = NewEditIfaceCmdCtx()

var (
	bindAddr   = flag.String("bind_address", ":10161", "Bind to address:port or just :port")
	configFile = flag.String("config", "", "IETF JSON file for target startup config")
)

type server struct {
	*gnmi.Server
}

const (
	maxMasterPortsC             = 32  // DX010: All front panel ports
	maxSlavePortsC              = 128 // DX010: All ports after split
	maxPortsC                   = maxMasterPortsC + maxSlavePortsC
	maxSlavePortsPerMasterPortC = 4
	maxLagInterfacesC           = 1024
	maxVlansC                   = 4096
	maxStpInstancesC            = maxVlansC
	portBaseIdxC                = 0
	lagBaseIdx                  = portBaseIdxC + maxPortsC
	vlanBaseIdx                 = lagBaseIdx + maxLagInterfacesC
	stpBaseIdx                  = vlanBaseIdx + maxVlansC
)

type configLookupTablesT struct {
	idxOfLastAddedIntf lib.IdxT
	idxByIntfName      map[string]lib.IdxT
	intfNameByIdx      map[lib.IdxT]string
	idxOfLastAddedLag  lib.IdxT
	idxByLagName       map[string]lib.IdxT
	lagNameByIdx       map[lib.IdxT]string
	idxOfLastAddedVlan lib.IdxT
	idxByVlanName      map[string]lib.IdxT
	vlanNameByIdx      map[lib.IdxT]string

	// L3 interface can have assigned many IPv4 addresses
	ipv4ByIntf map[lib.IdxT]*lib.StringSet
	// L3 LAG can have assigned many IPv4 addresses
	ipv4ByLag map[lib.IdxT]*lib.StringSet
	// L3 VLAN can have assigned many IPv4 addresses
	ipv4ByVlan map[lib.VidT]*lib.StringSet
	intfByIpv4 map[string]lib.IdxT
	lagByIpv4  map[string]lib.IdxT
	vlanByIpv4 map[string]lib.VidT
	ipv6ByIntf map[lib.IdxT]*lib.StringSet
	ipv6ByLag  map[lib.IdxT]*lib.StringSet
	ipv6ByVlan map[lib.VidT]*lib.StringSet
	// L3 interface can have assigned many IPv6 addresses
	intfByIpv6 map[string]lib.IdxT
	// L3 LAG can have assigned many IPv6 addresses
	lagByIpv6 map[string]lib.IdxT
	// L3 VLAN interface can have assigned many IPv6 addresses
	vlanByIpv6 map[string]lib.VidT
	lagByIntf  map[lib.IdxT]lib.IdxT
	// LAG can have many interface members
	intfByLag        map[lib.IdxT]*lib.IdxTSet
	stpByIntf        *lib.IdxTSet
	vlanAccessByIntf map[lib.IdxT]lib.VidT
	// There can be many ports in specific VLAN ID for access mode
	intfByVlanAccess map[lib.VidT]*lib.IdxTSet
	vlanAccessByLag  map[lib.IdxT]lib.VidT
	// There can be many LAGs in specific VLAN ID for access mode
	lagByVlanAccess  map[lib.VidT]*lib.IdxTSet
	vlanNativeByIntf map[lib.IdxT]lib.VidT
	// There can be many ports in specific VLAN ID for native tag
	intfByVlanNative map[lib.VidT]*lib.IdxTSet
	vlanNativeByLag  map[lib.IdxT]lib.VidT
	// There can be many LAGs in specific VLAN ID for native tag
	lagByVlanNative map[lib.VidT]*lib.IdxTSet
	vlanTrunkByIntf map[lib.IdxT]*lib.VidTSet
	// There can be many ports in VLAN trunk
	intfByVlanTrunk map[lib.VidT]*lib.IdxTSet
	vlanTrunkByLag  map[lib.IdxT]*lib.VidTSet
	// There can be many LAGs in VLAN trunk
	lagByVlanTrunk map[lib.VidT]*lib.IdxTSet
}

func (t *configLookupTablesT) dump() {
	log.Infoln("Dump internal state of config lookup tables")
	log.Infoln("========================================")
	intfs := make([]string, 0)
	for intfName, _ := range t.idxByIntfName {
		intfs = append(intfs, intfName)
	}
	sort.Strings(intfs)
	log.Infof("There are %d LAG interfaces", len(intfs))
	log.Infoln("Print list of interfaces:")
	for _, intfName := range intfs {
		log.Infoln(intfName)
	}

	log.Infoln("========================================")
	lags := make([]string, 0)
	for lagName, _ := range t.idxByLagName {
		lags = append(lags, lagName)
	}
	sort.Strings(lags)
	log.Infof("There are %d LAG interfaces", len(lags))
	log.Infoln("Print list of LAG interfaces:")
	for _, lagName := range lags {
		log.Infoln(lagName)
	}

	log.Infoln("========================================")
	log.Infoln("Print membership of interfaces in access VLAN:")
	for _, intfName := range intfs {
		if accessVid, exists := t.vlanAccessByIntf[t.idxByIntfName[intfName]]; exists {
			log.Infof("Access VLAN on interface %s: %d", intfName, accessVid)
		} else {
			log.Infof("There isn't access VLAN on interface %s", intfName)
			log.Infoln("----------------------------------------")
			continue
		}
	}

	log.Infoln("========================================")
	log.Infoln("Print membership of interfaces in trunk VLANs:")
	for _, intfName := range intfs {
		idx := t.idxByIntfName[intfName]
		vids, exists := t.vlanTrunkByIntf[idx]
		if !exists {
			log.Infof("There aren't any trunk VLANs on interface %s", intfName)
			log.Infoln("----------------------------------------")
			continue
		}

		vlans := make([]int, 0)
		for _, vid := range vids.VidTs() {
			vlans = append(vlans, int(vid))
		}
		sort.Ints(vlans)
		log.Infof("There are %d VLANs on interface %s:", len(vlans), intfName)
		for _, vid := range vlans {
			log.Infoln(vid)
		}

		if nativeVid, exists := t.vlanNativeByIntf[idx]; exists {
			log.Infof("Native VLAN on interface %s: %d", intfName, nativeVid)
		}
	}

	log.Infoln("========================================")
	log.Infoln("Print membership of LAG in access VLAN:")
	for _, lagName := range lags {
		if accessVid, exists := t.vlanAccessByLag[t.idxByLagName[lagName]]; exists {
			log.Infof("Access VLAN on LAG %s: %d", lagName, accessVid)
		} else {
			log.Infof("There isn't access VLAN on LAG %s", lagName)
			log.Infoln("----------------------------------------")
			continue
		}
	}

	log.Infoln("========================================")
	log.Infoln("Print membership of LAG interfaces in trunk VLANs:")
	for _, lagName := range lags {
		idx := t.idxByLagName[lagName]
		vids, exists := t.vlanTrunkByLag[idx]
		if !exists {
			log.Infof("There aren't any trunk VLANs on LAG %s", lagName)
			log.Infoln("----------------------------------------")
			continue
		}

		vlans := make([]int, 0)
		for _, vid := range vids.VidTs() {
			vlans = append(vlans, int(vid))
		}
		sort.Ints(vlans)
		log.Infof("There are %d VLANs on LAG %s:", len(vlans), lagName)
		for _, vid := range vlans {
			log.Infoln(vid)
		}

		if nativeVid, exists := t.vlanNativeByLag[idx]; exists {
			log.Infof("Native VLAN on LAG %s: %d", lagName, nativeVid)
		}
	}

	log.Infoln("========================================")
	log.Infoln("Print IPv4 addresses on interfaces:")
	for _, intfName := range intfs {
		idx := t.idxByIntfName[intfName]
		ipAddrs, exists := t.ipv4ByIntf[idx]
		if !exists {
			log.Infof("There aren't any IPv4 addresses on interface %s", intfName)
			log.Infoln("----------------------------------------")
			continue
		}

		addrs := make([]string, 0)
		for _, ip := range ipAddrs.Strings() {
			addrs = append(addrs, ip)
		}
		sort.Strings(addrs)
		log.Infof("There are %d IPv4 addresses on interface %s:", len(addrs), intfName)
		for _, ip := range addrs {
			log.Infoln(ip)
		}
	}

	log.Infoln("========================================")
	log.Infoln("Print IPv6 addresses on interfaces:")
	for _, intfName := range intfs {
		idx := t.idxByIntfName[intfName]
		ipAddrs, exists := t.ipv6ByIntf[idx]
		if !exists {
			log.Infof("There aren't any IPv6 addresses on interface %s", intfName)
			log.Infoln("----------------------------------------")
			continue
		}

		addrs := make([]string, 0)
		for _, ip := range ipAddrs.Strings() {
			addrs = append(addrs, ip)
		}
		sort.Strings(addrs)
		log.Infof("There are %d IPv6 addresses on interface %s:", len(addrs), intfName)
		for _, ip := range addrs {
			log.Infoln(ip)
		}
	}
}

func NewConfigLookupTables() *configLookupTablesT {
	return &configLookupTablesT{
		idxOfLastAddedIntf: 0,
		idxByIntfName:      make(map[string]lib.IdxT, maxPortsC),
		intfNameByIdx:      make(map[lib.IdxT]string, maxPortsC),
		idxOfLastAddedLag:  0,
		idxByLagName:       make(map[string]lib.IdxT),
		lagNameByIdx:       make(map[lib.IdxT]string),
		idxOfLastAddedVlan: 0,
		idxByVlanName:      make(map[string]lib.IdxT),
		vlanNameByIdx:      make(map[lib.IdxT]string),
		ipv4ByIntf:         make(map[lib.IdxT]*lib.StringSet),
		ipv4ByLag:          make(map[lib.IdxT]*lib.StringSet),
		ipv4ByVlan:         make(map[lib.VidT]*lib.StringSet),
		intfByIpv4:         make(map[string]lib.IdxT),
		lagByIpv4:          make(map[string]lib.IdxT),
		vlanByIpv4:         make(map[string]lib.VidT),
		ipv6ByIntf:         make(map[lib.IdxT]*lib.StringSet),
		ipv6ByLag:          make(map[lib.IdxT]*lib.StringSet),
		ipv6ByVlan:         make(map[lib.VidT]*lib.StringSet),
		intfByIpv6:         make(map[string]lib.IdxT),
		lagByIpv6:          make(map[string]lib.IdxT),
		vlanByIpv6:         make(map[string]lib.VidT),
		lagByIntf:          make(map[lib.IdxT]lib.IdxT),
		intfByLag:          make(map[lib.IdxT]*lib.IdxTSet),
		stpByIntf:          lib.NewIdxTSet(),
		vlanAccessByIntf:   make(map[lib.IdxT]lib.VidT),
		vlanAccessByLag:    make(map[lib.IdxT]lib.VidT),
		intfByVlanAccess:   make(map[lib.VidT]*lib.IdxTSet),
		lagByVlanAccess:    make(map[lib.VidT]*lib.IdxTSet),
		vlanNativeByIntf:   make(map[lib.IdxT]lib.VidT),
		vlanNativeByLag:    make(map[lib.IdxT]lib.VidT),
		intfByVlanNative:   make(map[lib.VidT]*lib.IdxTSet),
		lagByVlanNative:    make(map[lib.VidT]*lib.IdxTSet),
		vlanTrunkByIntf:    make(map[lib.IdxT]*lib.VidTSet),
		vlanTrunkByLag:     make(map[lib.IdxT]*lib.VidTSet),
		intfByVlanTrunk:    make(map[lib.VidT]*lib.IdxTSet),
		lagByVlanTrunk:     make(map[lib.VidT]*lib.IdxTSet),
	}
}

var gCurrentConfig ygot.ValidatedGoStruct
var doOnce sync.Once

type portBreakoutModeT uint8

const (
	kComponentPortBreakoutModePathElemSize                    = 5
	kComponentPortBreakoutModeIfnameElemIdx                   = 1
	kCPathElemIdx                                             = 0
	kCPathElem                                                = "Component"
	kCPPathElemIdx                                            = 2
	kCPPathElem                                               = "Port"
	kCPBMPathElemIdx                                          = 3
	kCPBMPathElem                                             = "BreakoutMode"
	kCPBMPNCPathElemIdx                                       = 4
	kCPBMPNCPathElem                                          = "NumChannels"
	kCPBMPCSPathElemIdx                                       = 4
	kCPBMPCSPathElem                                          = "ChannelSpeed"
	kDisabledPortBreakout                                     = 1
	kEnabledPortBreakout                                      = 4
	kPortBreakoutMode4x                     portBreakoutModeT = 4
	kPortBreakoutMode2x                     portBreakoutModeT = 2
	kPortBreakoutModeNone                   portBreakoutModeT = 1
	kPortBreakoutModeInvalid                portBreakoutModeT = 0
)

type ethSpeedT oc.E_OpenconfigIfEthernet_ETHERNET_SPEED

type configMngrT struct {
}

func NewConfigMngrT() *configMngrT {
	return &configMngrT{}
}

func (c *configMngrT) isChangedPortBreakoutChannelSpeed(change *diff.Change) bool {
	if len(change.Path) < kComponentPortBreakoutModePathElemSize {
		return false
	}

	if (change.Path[kCPathElemIdx] != kCPathElem) || (change.Path[kCPPathElemIdx] != kCPPathElem) || (change.Path[kCPBMPathElemIdx] != kCPBMPathElem) || (change.Path[kCPBMPCSPathElemIdx] != kCPBMPCSPathElem) {
		return false
	}

	return true
}

func (c *configMngrT) isChangedPortBreakoutNumChannels(change *diff.Change) bool {
	if len(change.Path) < kComponentPortBreakoutModePathElemSize {
		return false
	}

	if (change.Path[kCPathElemIdx] != kCPathElem) || (change.Path[kCPPathElemIdx] != kCPPathElem) || (change.Path[kCPBMPathElemIdx] != kCPBMPathElem) || (change.Path[kCPBMPNCPathElemIdx] != kCPBMPNCPathElem) {
		return false
	}

	return true
}

func (c *configLookupTablesT) isChangedBreakoutMode(change *diff.Change) bool {
	if len(change.Path) < kComponentPortBreakoutModePathElemSize {
		return false
	}

	if (change.Path[kCPathElemIdx] != kCPathElem) || (change.Path[kCPPathElemIdx] != kCPPathElem) || (change.Path[kCPBMPathElemIdx] != kCPBMPathElem) || ((change.Path[kCPBMPNCPathElemIdx] != kCPBMPNCPathElem) && (change.Path[kCPBMPCSPathElemIdx] != kCPBMPCSPathElem)) {
		return false
	}

	return true
}

func (t *configLookupTablesT) isIntfAvailable(ifname string) bool {
	if _, exists := t.idxByIntfName[ifname]; exists {
		return true
	}

	return false
}

func (c *configMngrT) isValidPortBreakoutNumChannels(numChannels portBreakoutModeT) bool {
	if numChannels == kPortBreakoutModeNone || numChannels == kPortBreakoutMode4x {
		return true
	}

	return false
}

func (c *configMngrT) isValidPortBreakoutChannelSpeed(numChannels portBreakoutModeT,
	channelSpeed oc.E_OpenconfigIfEthernet_ETHERNET_SPEED) bool {
	log.Infof("Split (%d), speed (%d)", numChannels, channelSpeed)
	switch channelSpeed {
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB:
		if numChannels == kPortBreakoutMode4x {
			return true
		}
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_100GB:
		fallthrough
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_40GB:
		if numChannels == kPortBreakoutModeNone {
			return true
		}
	}

	return false
}

func (c *configMngrT) getPortBreakoutNumChannelsFromChangelog(ifname string, changelog *diff.Changelog) (portBreakoutModeT, error) {
	var err error = nil
	numChannels := kPortBreakoutModeInvalid
	for _, change := range *changelog {
		if c.isChangedPortBreakoutNumChannels(&change) {
			log.Infof("Found changing number of channels request too:\n%+v", change)
			if change.Path[kComponentPortBreakoutModeIfnameElemIdx] == ifname {
				numChannels = portBreakoutModeT(change.To.(uint8))
				break
			}
		}
	}

	if !c.isValidPortBreakoutNumChannels(numChannels) {
		err = fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
	}

	return numChannels, err
}

func (c *configMngrT) getPortBreakoutChannelSpeedFromChangelog(ifname string, changelog *diff.Changelog) (oc.E_OpenconfigIfEthernet_ETHERNET_SPEED, error) {
	var err error = nil
	channelSpeed := oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
	for _, change := range *changelog {
		if c.isChangedPortBreakoutChannelSpeed(&change) {
			log.Infof("Found channel speed request too:\n%+v", change)
			if change.Path[kComponentPortBreakoutModeIfnameElemIdx] == ifname {
				channelSpeed = change.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
				break
			}
		}
	}

	if channelSpeed == oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET {
		err = fmt.Errorf("Could not found set channel speed request")
	}

	return channelSpeed, err
}

func (c *configMngrT) checkingPortBreakoutModeDependency(changedItem *diff.Change, changelog *diff.Changelog,
	lookupTables *configLookupTablesT) error {
	// TODO: First of all find all action related to interface(s) which is/are going to be removed.
	ifname := changedItem.Path[kComponentPortBreakoutModeIfnameElemIdx]
	if !lookupTables.isIntfAvailable(ifname) {
		return fmt.Errorf("Port %s is unrecognized", ifname)
	}

	var numChannels portBreakoutModeT = kPortBreakoutModeInvalid
	var channelSpeed oc.E_OpenconfigIfEthernet_ETHERNET_SPEED = oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
	var err error

	if changedItem.Path[kCPBMPNCPathElemIdx] == kCPBMPNCPathElem {
		channelSpeed, err = c.getPortBreakoutChannelSpeedFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}

		numChannels = portBreakoutModeT(changedItem.To.(uint8))
		if !c.isValidPortBreakoutNumChannels(numChannels) {
			return fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
		}
	} else if changedItem.Path[kCPBMPCSPathElemIdx] == kCPBMPCSPathElem {
		numChannels, err = c.getPortBreakoutNumChannelsFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}

		channelSpeed = changedItem.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
		if !c.isValidPortBreakoutChannelSpeed(numChannels, channelSpeed) {
			return fmt.Errorf("Speed channel (%d) is invalid", channelSpeed)
		}
	} else {
		return fmt.Errorf("Unable to get port breakout changing")
	}

	log.Infof("Requested changing port %s breakout into %d mode with speed %d", ifname, numChannels, channelSpeed)

	if numChannels == kPortBreakoutModeNone {
		// Check if there won't be any dependenies from master port
	} else {
		slavePorts := make([]string, kPortBreakoutMode4x)
		for i := 1; i <= 4; i++ {
			slavePorts[i-1] = fmt.Sprintf("%s.%d", ifname, i)
			log.Infof("Composed slave port: %s", slavePorts[i-1])
		}
		// Check if there won't be any dependenies from slave ports
	}

	return nil
}

var gnmiCallback gnmi.ConfigCallback = func(newConfig ygot.ValidatedGoStruct, cbUserData interface{}) error {
	doOnce.Do(func() {
		copier.Copy(&gCurrentConfig, &newConfig)
	})

	configMngr := NewConfigMngrT()
	lookupTables := cbUserData.(*configLookupTablesT)
	changelog, err := diff.Diff(gCurrentConfig, newConfig)
	if err != nil {
		log.Errorf("Failed to get diff of two config objects: %s", err)
		return err
	}

	log.Infof("Changlog (%d):\n%#v\n", len(changelog), changelog)
	jsonDump, err := json.MarshalIndent(changelog, "", "    ")
	if err != nil {
		log.Errorf("Failed to JSON dump: %s", err)
	}

	log.Infof("Dump JSON: %s", string(jsonDump))
	if len(changelog) > 0 {
		log.Infof("Configuration has been changed")
		for _, changedItem := range changelog {
			log.Infof("Change item: %#v", changedItem)
			if len(changedItem.Path) > 4 {
				if lookupTables.isChangedBreakoutMode(&changedItem) {
					if err := configMngr.checkingPortBreakoutModeDependency(&changedItem, &changelog, lookupTables); err != nil {
						log.Errorf("%s", err)
						return err
					}
				} else if "NativeVlan" == changedItem.Path[4] {
					port := make([]string, 1)
					port[0] = changedItem.Path[1]
					// TODO: Uncomment if build is dedicated for target device
					// if err := vlan.SetNativeVlan(port, changedItem.To.(uint16)); err != nil {
					// 	log.Errorf("Failed to set native VLAN")
					// 	return err
					// }

					log.Infof("Native VLAN has been changed to %d on port %s",
						changedItem.To, changedItem.Path[1])
				}
			} else if len(changedItem.Path) > 2 {
				if "Mtu" == changedItem.Path[2] {
					log.Infof("Changing MTU to %d on port %s", changedItem.To, changedItem.Path[1])
				}
			}
		}

		log.Infof("Save new config")
		copier.Copy(&gCurrentConfig, &newConfig)
	}
	//modify updated
	return nil
}

func (table *configLookupTablesT) addNewInterfaceIfItDoesNotExist(intfName string) error {
	if strings.Contains(intfName, "ae") {
		if _, exists := table.idxByLagName[intfName]; !exists {
			table.idxByLagName[intfName] = table.idxOfLastAddedLag
			table.lagNameByIdx[table.idxOfLastAddedLag] = intfName
			table.idxOfLastAddedLag++
			log.Infof("Saved LAG %s", intfName)
		}
	} else if strings.Contains(intfName, "eth") {
		if _, exists := table.idxByIntfName[intfName]; !exists {
			table.idxByIntfName[intfName] = table.idxOfLastAddedIntf
			table.intfNameByIdx[table.idxOfLastAddedIntf] = intfName
			table.idxOfLastAddedIntf++
			log.Infof("Saved interface %s", intfName)
		}
	} else {
		err := fmt.Errorf("Unrecognized type of interface %s", intfName)
		return err
	}

	return nil
}

func (table *configLookupTablesT) setNativeVlanOnPort(intfName string, vid lib.VidT) {
	// TODO: Add asserts for checking if interface exists in map
	table.vlanNativeByIntf[table.idxByIntfName[intfName]] = vid
	if _, exists := table.intfByVlanNative[vid]; !exists {
		table.intfByVlanNative[vid] = lib.NewIdxTSet()
	}

	table.intfByVlanNative[vid].Add(table.idxByIntfName[intfName])
	log.Infof("Set native VLAN %d on interface %s", vid, intfName)
}

func (table *configLookupTablesT) setAccessVlanOnPort(intfName string, vid lib.VidT) {
	// TODO: Add asserts for checking if LAG exists in map
	table.vlanAccessByIntf[table.idxByIntfName[intfName]] = vid
	if _, exists := table.intfByVlanAccess[vid]; !exists {
		table.intfByVlanAccess[vid] = lib.NewIdxTSet()
	}

	table.intfByVlanAccess[vid].Add(table.idxByIntfName[intfName])
	log.Infof("Set access VLAN %d on port %s", vid, intfName)
}

func (table *configLookupTablesT) setTrunkVlansOnPort(intfName string, vids []lib.VidT) {
	// TODO: Add asserts for checking if interface exists in map
	for _, vid := range vids {
		if _, exists := table.vlanTrunkByIntf[table.idxByIntfName[intfName]]; !exists {
			table.vlanTrunkByIntf[table.idxByIntfName[intfName]] = lib.NewVidTSet()
		}

		table.vlanTrunkByIntf[table.idxByIntfName[intfName]].Add(vid)
		if _, exists := table.intfByVlanTrunk[vid]; !exists {
			table.intfByVlanTrunk[vid] = lib.NewIdxTSet()
		}

		table.intfByVlanTrunk[vid].Add(table.idxByIntfName[intfName])
		log.Infof("Set trunk VLAN %d on interface %s", vid, intfName)
	}
}

func (table *configLookupTablesT) setAccessVlanOnLag(lagName string, vid lib.VidT) {
	// TODO: Add asserts for checking if LAG exists in map
	table.vlanAccessByLag[table.idxByLagName[lagName]] = vid
	if _, exists := table.lagByVlanAccess[vid]; !exists {
		table.lagByVlanAccess[vid] = lib.NewIdxTSet()
	}

	table.lagByVlanAccess[vid].Add(table.idxByLagName[lagName])
	log.Infof("Set access VLAN %d on LAG %s", vid, lagName)
}

func (table *configLookupTablesT) setNativeVlanOnLag(lagName string, vid lib.VidT) {
	// TODO: Add asserts for checking if LAG exists in map
	table.vlanNativeByLag[table.idxByLagName[lagName]] = vid
	if _, exists := table.lagByVlanNative[vid]; !exists {
		table.lagByVlanNative[vid] = lib.NewIdxTSet()
	}

	table.lagByVlanNative[vid].Add(table.idxByLagName[lagName])
	log.Infof("Set native VLAN %d on LAG %s", vid, lagName)
}

func (table *configLookupTablesT) setTrunkVlansOnLag(lagName string, vids []lib.VidT) {
	// TODO: Add asserts for checking if LAG exists in map
	for _, vid := range vids {
		if _, exists := table.vlanTrunkByLag[table.idxByLagName[lagName]]; !exists {
			table.vlanTrunkByLag[table.idxByLagName[lagName]] = lib.NewVidTSet()
		}

		table.vlanTrunkByLag[table.idxByLagName[lagName]].Add(vid)
		if _, exists := table.lagByVlanTrunk[vid]; !exists {
			table.lagByVlanTrunk[vid] = lib.NewIdxTSet()
		}

		table.lagByVlanTrunk[vid].Add(table.idxByLagName[lagName])
		log.Infof("Set trunk VLAN %d on LAG %s", vid, lagName)
	}
}

func (t *configLookupTablesT) saveIpv4AddressForInterface(intfName string, ip string) error {
	intfIdx := t.idxByIntfName[intfName]
	if _, exists := t.intfByIpv4[ip]; exists {
		return fmt.Errorf("Failed to assign IPv4 address %s to interface %s because it is already in use",
			ip, intfName)
	}

	t.intfByIpv4[ip] = intfIdx
	if _, exists := t.ipv4ByIntf[t.idxByIntfName[intfName]]; !exists {
		t.ipv4ByIntf[t.idxByIntfName[intfName]] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv4ByIntf[t.idxByIntfName[intfName]].Add(ip)
	log.Infof("Saved IPv4 %s for interface %s", ip, intfName)
	return nil
}

func (t *configLookupTablesT) saveIpv6AddressForInterface(intfName string, ip string) error {
	intfIdx := t.idxByIntfName[intfName]
	if _, exists := t.intfByIpv6[ip]; exists {
		return fmt.Errorf("Failed to assign IPv6 address %s to interface %s because it is already in use",
			ip, intfName)
	}

	t.intfByIpv6[ip] = intfIdx
	if _, exists := t.ipv6ByIntf[intfIdx]; !exists {
		t.ipv6ByIntf[intfIdx] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv6ByIntf[intfIdx].Add(ip)
	log.Infof("Saved IPv6 %s for interface %s", ip, intfName)
	return nil
}

func (t *configLookupTablesT) saveIpv4AddressForLag(lagName string, ip string) error {
	lagIdx := t.idxByLagName[lagName]
	if _, exists := t.lagByIpv4[ip]; exists {
		return fmt.Errorf("Failed to assign IPv4 address %s to LAG %s because it is already in use",
			ip, lagName)
	}

	t.lagByIpv4[ip] = lagIdx
	if _, exists := t.ipv4ByLag[lagIdx]; !exists {
		t.ipv4ByLag[lagIdx] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv4ByLag[lagIdx].Add(ip)
	log.Infof("Saved IPv4 %s for LAG %s", ip, lagName)
	return nil
}

func (t *configLookupTablesT) saveIpv6AddressForLag(lagName string, ip string) error {
	lagIdx := t.idxByLagName[lagName]
	if _, exists := t.lagByIpv6[ip]; exists {
		return fmt.Errorf("Failed to assign IPv6 address %s to interface %s because it is already in use",
			ip, lagName)
	}
	t.lagByIpv6[ip] = lagIdx

	if _, exists := t.ipv6ByLag[lagIdx]; !exists {
		t.ipv6ByLag[lagIdx] = lib.NewStringSet()
	}
	// TODO: Check if IP is valid
	t.ipv6ByLag[lagIdx].Add(ip)
	log.Infof("Saved IPv6 %s for interface %s", ip, lagName)
	return nil
}

func (t *configLookupTablesT) parseInterfaceAsLagMember(intfName string, eth *oc.Interface_Ethernet) error {
	lagName := eth.GetAggregateId()
	if len(lagName) == 0 {
		return nil
	}
	lagIdx, exists := t.idxByLagName[lagName]
	if !exists {
		return fmt.Errorf("Invalid LAG %s on interface %s: LAG not exists", lagName, intfName)
	}

	intfIdx := t.idxByIntfName[intfName]
	if lag, exists := t.lagByIntf[intfIdx]; exists {
		if lag == lagIdx {
			return fmt.Errorf("Interface %s exists in another LAG %s", intfName, t.lagNameByIdx[lag])
		}
	}
	t.lagByIntf[intfIdx] = lagIdx

	if _, exists = t.intfByLag[lagIdx]; !exists {
		t.intfByLag[lagIdx] = lib.NewIdxTSet()
	}
	t.intfByLag[lagIdx].Add(intfIdx)

	log.Infof("Added interface %s as member of LAG %s", intfName, lagName)
	return nil
}

func (t *configLookupTablesT) parseSubinterface(intfName string, subIntf *oc.Interface_Subinterface) error {
	ipv4 := subIntf.GetIpv4()
	if ipv4 != nil {
		for _, addr := range ipv4.Address {
			ip := fmt.Sprintf("%s/%d", addr.GetIp(), addr.GetPrefixLength())
			if err := t.saveIpv4AddressForInterface(intfName, ip); err != nil {
				return err
			}
		}
	}

	ipv6 := subIntf.GetIpv6()
	if ipv6 != nil {
		for _, addr := range ipv6.Address {
			ip := fmt.Sprintf("%s/%d", addr.GetIp(), addr.GetPrefixLength())
			if err := t.saveIpv6AddressForInterface(intfName, ip); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *configLookupTablesT) parseVlanForIntf(intfName string, swVlan *oc.Interface_Ethernet_SwitchedVlan) error {
	intfMode := swVlan.GetInterfaceMode()
	if intfMode == oc.OpenconfigVlan_VlanModeType_ACCESS {
		vid := lib.VidT(swVlan.GetAccessVlan())
		if vid != 0 {
			t.setAccessVlanOnPort(intfName, vid)
			log.Infof("Set access VLAN %d for interface %s", vid, intfName)
		} else {
			return fmt.Errorf("Failed to parse VLAN on interface %s in access mode", intfName)
		}
	} else if intfMode == oc.OpenconfigVlan_VlanModeType_TRUNK {
		nativeVid := lib.VidT(swVlan.GetNativeVlan())
		if nativeVid != 0 {
			t.setNativeVlanOnPort(intfName, nativeVid)
			log.Infof("Set native VLAN %d for interface %s", nativeVid, intfName)
		}

		trunkVlans := swVlan.GetTrunkVlans()
		if trunkVlans != nil {
			vlans := make([]lib.VidT, 0)
			for _, v := range trunkVlans {
				switch t := v.(type) {
				case *oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_String:
					var lower, upper lib.VidT
					n, err := fmt.Sscanf(t.String, "%d..%d", &lower, &upper)
					if n != 2 || err != nil {
						return fmt.Errorf("Failed to parse lower and upper bound of trunk VLAN rane: %s", err)
					}

					if lower >= maxVlansC || upper >= maxVlansC {
						return fmt.Errorf("Out of range lowwer and upper bound of trunk VLANs (%d, %d)", lower, upper)
					}

					for ; lower <= upper; lower++ {
						vlans = append(vlans, lower)
					}
				case *oc.Interface_Ethernet_SwitchedVlan_TrunkVlans_Union_Uint16:
					vlans = append(vlans, lib.VidT(t.Uint16))
				default:
					return fmt.Errorf("Cannot convert %v to Interface_Ethernet_SwitchedVlan_TrunkVlans_Union, unknown union type, got: %T, want any of [string, uint16]", v, v)
				}
			}

			t.setTrunkVlansOnPort(intfName, vlans)
		}

		if nativeVid == 0 && trunkVlans == nil {
			return fmt.Errorf("Failed to parse VLANs on interface %s in trunk mode", intfName)
		}
	}

	return nil
}

func (t *configLookupTablesT) parseVlanForLagIntf(lagName string, swVlan *oc.Interface_Aggregation_SwitchedVlan) error {
	intfMode := swVlan.GetInterfaceMode()
	if intfMode == oc.OpenconfigVlan_VlanModeType_ACCESS {
		vid := lib.VidT(swVlan.GetAccessVlan())
		if vid != 0 {
			t.setAccessVlanOnLag(lagName, vid)
			log.Infof("Set access VLAN %d for LAG %s", vid, lagName)
		} else {
			return fmt.Errorf("Failed to parse VLAN on interface %s in access mode", lagName)
		}
	} else if intfMode == oc.OpenconfigVlan_VlanModeType_TRUNK {
		nativeVid := lib.VidT(swVlan.GetNativeVlan())
		if nativeVid != 0 {
			t.setNativeVlanOnLag(lagName, nativeVid)
			log.Infof("Set native VLAN %d for LAG %s", lagName, nativeVid)
		}

		trunkVlans := swVlan.GetTrunkVlans()
		if trunkVlans != nil {
			vlans := make([]lib.VidT, 0)
			for _, v := range trunkVlans {
				switch t := v.(type) {
				case *oc.Interface_Aggregation_SwitchedVlan_TrunkVlans_Union_String:
					var lower, upper lib.VidT
					n, err := fmt.Sscanf(t.String, "%d..%d", &lower, &upper)
					if n != 2 || err != nil {
						return fmt.Errorf("Failed to parse lower and upper bound of trunk VLAN rane: %s", err)
					}

					if lower >= maxVlansC || upper >= maxVlansC {
						return fmt.Errorf("Out of range lowwer and upper bound of trunk VLANs (%d, %d)", lower, upper)
					}

					for ; lower <= upper; lower++ {
						vlans = append(vlans, lower)
					}
				case *oc.Interface_Aggregation_SwitchedVlan_TrunkVlans_Union_Uint16:
					vlans = append(vlans, lib.VidT(t.Uint16))
				default:
					return fmt.Errorf("Cannot convert %v to Interface_Aggregation_SwitchedVlan_TrunkVlans_Union, unknown union type, got: %T, want any of [string, uint16]", v, v)
				}
			}

			t.setTrunkVlansOnLag(lagName, vlans)
		}

		if nativeVid == 0 && trunkVlans == nil {
			return fmt.Errorf("Failed to parse VLANs on interface %s in trunk mode", lagName)
		}
	}

	return nil
}

func newServer(model *gnmi.Model, config []byte) (*server, error) {
	configLookupTables := NewConfigLookupTables()
	configModel, err := model.NewConfigStruct(config)
	if err != nil {
		return nil, err
	}

	log.Infof("Dump config model: %+v", configModel)
	device := configModel.(*oc.Device)
	for intfName, _ := range device.Interface {
		if err := configLookupTables.addNewInterfaceIfItDoesNotExist(intfName); err != nil {
			return nil, err
		}
	}

	for intfName, _ := range configLookupTables.idxByIntfName {
		intf := device.Interface[intfName]
		if intf == nil {
			log.Info("Cannot find interface ", intfName)
			return nil, fmt.Errorf("Failed to get interface %s info", intfName)
		}

		eth := intf.GetEthernet()
		if eth != nil {
			log.Infof("Configuring interface %s as LAG member", intfName)
			if err := configLookupTables.parseInterfaceAsLagMember(intfName, eth); err != nil {
				return nil, err
			}

			swVlan := eth.GetSwitchedVlan()
			if swVlan != nil {
				if err := configLookupTables.parseVlanForIntf(intfName, swVlan); err != nil {
					return nil, err
				}
			}
		}

		subIntf := intf.GetSubinterface(0)
		if subIntf != nil {
			if err := configLookupTables.parseSubinterface(intfName, subIntf); err != nil {
				return nil, err
			}
		}
	}

	for lagName, _ := range configLookupTables.idxByLagName {
		lag := device.Interface[lagName]
		if lag == nil {
			return nil, fmt.Errorf("Failed to get LAG %s info", lagName)
		}

		agg := lag.GetAggregation()
		if agg != nil {
			swVlan := agg.GetSwitchedVlan()
			if swVlan != nil {
				if err := configLookupTables.parseVlanForLagIntf(lagName, swVlan); err != nil {
					return nil, err
				}
			}
		}
	}

	configLookupTables.dump()

	// TODO: Check if there isn't inconsistency in VLANs between ethernet
	//       interface and aggregate ethernet interfaces

	log.Infof("There are loaded %d interfaces and %d LAGs",
		configLookupTables.idxOfLastAddedIntf, configLookupTables.idxOfLastAddedLag)

	s, err := gnmi.NewServer(model, config, gnmiCallback, configLookupTables)
	if err != nil {
		return nil, err
	}
	return &server{Server: s}, nil
}

// Get overrides the Get func of gnmi.Target to provide user auth.
func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	msg, ok := credentials.AuthorizeUser(ctx)
	if !ok {
		log.Infof("denied a Get request: %v", msg)
		return nil, status.Error(codes.PermissionDenied, msg)
	}
	log.Infof("allowed a Get request: %v", msg)
	return s.Server.Get(ctx, req)
}

// Set overrides the Set func of gnmi.Target to provide user auth.
func (s *server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	msg, ok := credentials.AuthorizeUser(ctx)
	if !ok {
		log.Infof("denied a Set request: %v", msg)
		return nil, status.Error(codes.PermissionDenied, msg)
	}
	log.Infof("allowed a Set request: %v", msg)
	return s.Server.Set(ctx, req)
}

func gNMIServerRun() {
	model := gnmi.NewModel(modeldata.ModelData,
		reflect.TypeOf((*oc.Device)(nil)),
		oc.SchemaTree["Device"],
		oc.Unmarshal,
		oc.ΛEnum)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Supported models:\n")
		for _, m := range model.SupportedModels() {
			fmt.Fprintf(os.Stderr, "  %s\n", m)
		}
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	opts := credentials.ServerCredentials()
	g := grpc.NewServer(opts...)

	var configData []byte
	if *configFile != "" {
		var err error
		configData, err = ioutil.ReadFile(*configFile)
		if err != nil {
			log.Exitf("error in reading config file: %v", err)
		}
	}
	s, err := newServer(model, configData)
	if err != nil {
		log.Exitf("error in creating gnmi target: %v", err)
	}
	pb.RegisterGNMIServer(g, s)
	reflection.Register(g)

	log.Infof("starting to listen on %s", *bindAddr)
	listen, err := net.Listen("tcp", *bindAddr)
	if err != nil {
		log.Exitf("failed to listen: %v", err)
	}

	log.Info("starting to serve")
	if err := g.Serve(listen); err != nil {
		log.Exitf("failed to serve: %v", err)
	}
}

func main() {
	go gNMIServerRun()
	shell := ishell.New()

	// display info.
	shell.Println("Welcome to OpenNOS CLI")

	//Consider the unicode characters supported by the users font
	//shell.SetMultiChoicePrompt(" >>"," - ")
	//shell.SetChecklistOptions("[ ] ","[X] ")

	// var vlans []string = make([]string, 1)
	// var editableIfaces = map[string]*Iface{}
	var editableIfaces = map[string]*Iface{
		"eth-1": NewIface(), "eth-2": NewIface(), "eth-3": NewIface(), "eth-4": NewIface(), "eth-5": NewIface(),
		"eth-6": NewIface(), "eth-7": NewIface(), "eth-8": NewIface(), "eth-9": NewIface(), "eth-10": NewIface(),
		"eth-11": NewIface(), "eth-12": NewIface(), "eth-13": NewIface(), "eth-14": NewIface(), "eth-15": NewIface(),
		"eth-16": NewIface(), "eth-17": NewIface(), "eth-18": NewIface(), "eth-19": NewIface(), "eth-20": NewIface(),
		"eth-21": NewIface(), "eth-22": NewIface(), "eth-23": NewIface(), "eth-24": NewIface(), "eth-25": NewIface(),
		"eth-26": NewIface(), "eth-27": NewIface(), "eth-28": NewIface(), "eth-29": NewIface(), "eth-30": NewIface(),
		"eth-31": NewIface(), "eth-32": NewIface(),
	}
	// var editableIfaces []string = []string{
	// 	"eth-1", "eth-2", "eth-3", "eth-4", "eth-5", "eth-6", "eth-7", "eth-8", "eth-9", "eth-10",
	// 	"eth-11", "eth-12", "eth-13", "eth-14", "eth-15", "eth-16", "eth-17", "eth-18", "eth-19", "eth-20",
	// 	"eth-21", "eth-22", "eth-23", "eth-24", "eth-25", "eth-26", "eth-27", "eth-28", "eth-29", "eth-30",
	// 	"eth-31", "eth-32",
	// }
	editCmd := &ishell.Cmd{
		Name:     "edit",
		Help:     "edit <interface> | <aggregate> | <vlan>",
		LongHelp: `Edit`,
	}
	editIfaceCmd := &ishell.Cmd{
		Name: "interface",
		Help: "edit interface <interface name>",
		Completer: func(args []string) []string {
			if editIfaceCmdCtx.completerInvoked {
				if len(args) > 1 {
					return nil
				}

				if len(args) > 0 {
					return []string{"help"}
					// return nil
				}
				// log.Println(args)
				// return nil
				// return []string{"\nPress enter to edit..."}
			}
			ifnames := make([]string, len(editableIfaces))
			i := 0
			for ifname := range editableIfaces {
				ifnames[i] = ifname
				i++
			}

			editIfaceCmdCtx.completerInvoked = true
			return ifnames
		},
		Func: func(c *ishell.Context) {
			log.Infof("Choosed interface %s", c.Args[0])
			if len(c.Args) == 0 {
				c.Err(errors.New("Missing interface name"))
				return
			}

			if len(c.Args) == 2 {
				if c.Args[1] != "help" {
					c.Err(errors.New("Invalid argument"))
					return
				}

				log.Infof("Enter to edit interface %s", c.Args[0])
				return
			}

			if len(c.Args) > 1 {
				c.Err(errors.New("Too many arguments"))
				return
			}

			foundIface := false
			for _, iface := range validIfaces {
				if c.Args[0] == iface {
					foundIface = true
					break
				}
			}

			if !foundIface {
				c.Err(errors.New("Invalid argument"))
				return
			}

			// editableIfaces[c.Args[0]] = NewIface()
			editIfaceCmdCtx.completerInvoked = false
			c.SetPrompt(fmt.Sprintf("[edit interface %s]# ", c.Args[0]))
			// editableIfaces[c.Args[2]] = NewIface()
			// vlans = append(vlans, c.Args...)
			port := make([]string, 1)
			port[0] = c.Args[0]
			if err := vlan.SetNativeVlan(port, 2); err != nil {
				c.Err(errors.New("Failed to set native VLAN"))
				return
			}
		},
	}

	executePromptCmd := &ishell.Cmd{
		Name: "execute_prompt",
		Help: "Press Enter to execute command",
		// Func: func(c *ishell.Context) {
		// 	log.Println("Press Enter to execute command")
		// },
	}
	// editIfaceCmd.AddCmd(&ishell.Cmd{
	// 	Name: "add_face_to_vlan",
	// 	Help: "add_face_to_vlan",
	// 	Func: func(c *ishell.Context) {
	// 		if len(c.Args) == 0 {
	// 			c.Err(errors.New("missing interface name"))
	// 			return
	// 		}
	// 		vlans = append(vlans, c.Args...)
	// 	},
	// })
	editIfaceCmd.AddCmd(executePromptCmd)
	editCmd.AddCmd(editIfaceCmd)
	// editCmd.AddCmd(executePromptCmd)

	addCmd := &ishell.Cmd{
		Name: "add",
		Help: "add",
		LongHelp: `Try dynamic autocomplete by adding and removing words.
	Then view the autocomplete by tabbing after "words" subcommand.

	This is an example of a long help.`,
	}
	addIfaceCmd := &ishell.Cmd{
		Name: "interfaces",
		Help: "Specify network interfaces to add into a VLAN",
		LongHelp: `Try dynamic autocomplete by adding and removing words.
	Then view the autocomplete by tabbing after "words" subcommand.

	This is an example of a long help.`,
	}
	addIfaceToCmd := &ishell.Cmd{
		Name: "to",
		Help: "to",
		LongHelp: `Try dynamic autocomplete by adding and removing words.
	Then view the autocomplete by tabbing after "words" subcommand.

	This is an example of a long help.`,
	}
	addIfaceToVlanCmd := &ishell.Cmd{
		Name:     "vlan",
		Help:     "<VLAN-ID> <IFACE-ID> [<IFACE-ID>...]",
		LongHelp: `add interfaces to vlan <VLAN ID> <PORT NAME>`,
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			defaultInput := "vlan-1 eth-1 eth-2"
			if len(c.Args) > 0 {
				defaultInput = strings.Join(c.Args, " ")
			}

			c.Print("input: ")
			read := c.ReadLineWithDefault(defaultInput)

			if read == defaultInput {
				c.Println("you left the default input intact")
			} else {
				c.Printf("you modified input to '%s'", read)
				c.Println()
			}
		},
	}

	addIfaceToCmd.AddCmd(addIfaceToVlanCmd)
	addIfaceCmd.AddCmd(addIfaceToCmd)
	addIfaceCmd.AddCmd(addIfaceCmd)
	addCmd.AddCmd(addIfaceCmd)
	shell.AddCmd(addCmd)

	shell.AddCmd(editCmd)

	// when started with "exit" as first argument, assume non-interactive execution
	if len(os.Args) > 1 && os.Args[1] == "exit" {
		shell.Process(os.Args[2:]...)
	} else {
		// start shell
		shell.Run()
		// teardown
		shell.Close()
	}
}
