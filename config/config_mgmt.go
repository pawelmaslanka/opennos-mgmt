package config

import (
	"fmt"
	"opennos-mgmt/gnmi"
	"opennos-mgmt/gnmi/modeldata/oc"

	cmd "opennos-mgmt/config/command"

	log "github.com/golang/glog"
	"github.com/r3labs/diff"
	"google.golang.org/grpc"

	mgmt "opennos-eth-switch-service/mgmt"
	serv_param "opennos-eth-switch-service/serv-param"
)

// OrdinalNumberT underlying type of ordinal number for action performed in transaction
type OrdinalNumberT uint16

// The following constants define ordering numbers of actions in transaction
const (
	UnorderedActionInTransactionC    OrdinalNumberT = iota
	WithdrawIpv4FromEthIntfC                        // Remove IPv4/CIDRv4 address from Ethernet interface
	WithdrawIpv4FromLagIntfC                        // Remove IPv4/CIDRv4 address from LAG interface
	WithdrawIpv6FromEthIntfC                        // Remove IPv6/CIDRv6 address from Ethernet interface
	WithdrawIpv6FromLagIntfC                        // Remove IPv6/CIDRv6 address from LAG interface
	WithdrawEthIntfFromAccessVlanC                  // Remove Ethernet interface from access VLAN
	WithdrawLagIntfFromAccessVlanC                  // Remove LAG interface from access VLAN
	WithdrawEthIntfFromNativeVlanC                  // Remove Ethernet interface from native VLAN
	WithdrawFromLagIntfNativeVlanC                  // Remove LAG interface from native VLAN
	WithdrawEthIntfFromTrunkVlanC                   // Remove Ethernet interface from trunk VLAN
	WithdrawLagIntfFromTrunkVlanC                   // Remove LAG interface from trunk VLAN
	WithdrawPortBreakoutC                           // Combine multiple logical ports into single port
	SetPortBreakoutC                                // Break out front panel port into multiple logical ports
	SetPortBreakoutSpeedC                           // Set speed on logical ports (lanes)
	SetDescOnEthIntfC                               // Set description of Ethernet interface
	SetPortAutoNegOnEthIntfC                        // Enable or disable auto-negotiation on port
	SetPortMtuOnEthIntfC                            // Set MTU on port
	SetPortSpeedOnEthIntfC                          // Set port speed
	SetIpv4OnEthIntfC                               // Assign IPv4/CIDRv4 address to Ethernet interface
	SetIpv4OnLagIntfC                               // Assign IPv4/CIDRv4 address to LAG interface
	SetIpv6OnEthIntfC                               // Assign IPv6/CIDRv6 address to Ethernet interface
	SetIpv6OnLagIntfC                               // Assign IPv6/CIDRv6 address to LAG interface
	SetVlanIntfModeOfLagIntfC                       // Set VLAN interface mode
	SetAccessVlanOnEthIntfC                         // Assign Ethernet interface to access VLAN
	SetAccessVlanOnLagIntfC                         // Assign LAG interface to access VLAN
	SetNativeVlanOnEthIntfC                         // Assign Ethernet interface to native VLAN
	SetNativeVlanOnLagIntfC                         // Assign LAG interface to native VLAN
	SetTrunkVlanOnEthIntfC                          // Assign Ethernet interface to trunk VLAN
	SetTrunkVlanOnLagIntfC                          // Assign LAG interface to trunk VLAN
	SetLagTypeOfLagIntfC                            // Set the type of LAG
	SetLacpIntervalC                                // Set the period between LACP messages
	SetLacpModeC                                    // Set LACP activity - active or passive
	MaxNumberOfActionsInTransactionC                // Defines maximum number of possible actions in transaction
)

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

type CfgMngrT struct {
	cfgLookupTbl *configLookupTablesT
}

func NewCfgMngrT() *CfgMngrT {
	return &CfgMngrT{
		cfgLookupTbl: newConfigLookupTables(),
	}
}

func (mngr *CfgMngrT) LoadConfig(model *gnmi.Model, config []byte) error {
	configModel, err := model.NewConfigStruct(config)
	if err != nil {
		return err
	}

	log.Infof("Dump config model: %+v", configModel)
	device := configModel.(*oc.Device)
	for intfName, _ := range device.Interface {
		if err := mngr.cfgLookupTbl.addNewInterfaceIfItDoesNotExist(intfName); err != nil {
			return err
		}
	}

	for intfName, _ := range mngr.cfgLookupTbl.idxByIntfName {
		intf := device.Interface[intfName]
		if intf == nil {
			log.Info("Cannot find interface ", intfName)
			return fmt.Errorf("Failed to get interface %s info", intfName)
		}

		eth := intf.GetEthernet()
		if eth != nil {
			log.Infof("Configuring interface %s as LAG member", intfName)
			if err := mngr.cfgLookupTbl.parseInterfaceAsLagMember(intfName, eth); err != nil {
				return err
			}

			swVlan := eth.GetSwitchedVlan()
			if swVlan != nil {
				if err := mngr.cfgLookupTbl.parseVlanForIntf(intfName, swVlan); err != nil {
					return err
				}
			}
		}

		subIntf := intf.GetSubinterface(0)
		if subIntf != nil {
			if err := mngr.cfgLookupTbl.parseSubinterface(intfName, subIntf); err != nil {
				return err
			}
		}
	}

	for lagName, _ := range mngr.cfgLookupTbl.idxByLagName {
		lag := device.Interface[lagName]
		if lag == nil {
			return fmt.Errorf("Failed to get LAG %s info", lagName)
		}

		agg := lag.GetAggregation()
		if agg != nil {
			swVlan := agg.GetSwitchedVlan()
			if swVlan != nil {
				if err := mngr.cfgLookupTbl.parseVlanForLagIntf(lagName, swVlan); err != nil {
					return err
				}
			}
		}
	}

	mngr.cfgLookupTbl.dump()
	// TODO: Check if there isn't inconsistency in VLANs between ethernet
	//       interface and aggregate ethernet interfaces

	log.Infof("There are loaded %d interfaces and %d LAGs",
		mngr.cfgLookupTbl.idxOfLastAddedIntf, mngr.cfgLookupTbl.idxOfLastAddedLag)

	return nil
}

func (mngr *CfgMngrT) IsChangedBreakoutMode(change *diff.Change) bool {
	if len(change.Path) < cmd.PortBreakoutPathItemsCountC {
		return false
	}

	if (change.Path[cmd.PortBreakoutCompPathItemIdxC] != cmd.PortBreakoutCompPathItemC) || (change.Path[cmd.PortBreakoutPortPathItemIdxC] != cmd.PortBreakoutPortPathItemC) || (change.Path[cmd.PortBreakoutModePathItemIdxC] != cmd.PortBreakoutModePathItemC) || ((change.Path[cmd.PortBreakoutNumChanPathItemIdxC] != cmd.PortBreakoutNumChanPathItemC) && (change.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] != cmd.PortBreakoutChanSpeedPathItemC)) {
		return false
	}

	return true
}

func (mngr *CfgMngrT) IsIntfAvailable(ifname string) bool {
	if _, exists := mngr.cfgLookupTbl.idxByIntfName[ifname]; exists {
		return true
	}

	return false
}

func (mngr *CfgMngrT) CheckingPortBreakoutModeDependency(changedItem *diff.Change, changelog *diff.Changelog) error {
	// TODO: First of all, find all action related to interface(s) which is/are going to be removed.
	ifname := changedItem.Path[cmd.PortBreakoutIfnamePathItemIdxC]
	if !mngr.IsIntfAvailable(ifname) {
		return fmt.Errorf("Port %s is unrecognized", ifname)
	}

	var numChannels cmd.PortBreakoutModeT = cmd.PortBreakoutModeInvalidC
	var channelSpeed oc.E_OpenconfigIfEthernet_ETHERNET_SPEED = oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
	var numChannelsChangeItem *diff.Change
	var channelSpeedChangeItem *diff.Change
	var err error

	if changedItem.Path[cmd.PortBreakoutNumChanPathItemIdxC] == cmd.PortBreakoutNumChanPathItemC {
		channelSpeed, err = mngr.getPortBreakoutChannelSpeedFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}

		numChannels = cmd.PortBreakoutModeT(changedItem.To.(uint8))
		if !mngr.isValidPortBreakoutNumChannels(numChannels) {
			return fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
		}

		channelSpeedChangeItem, err = mngr.getPortBreakoutChannelSpeedChangeItemFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}
		numChannelsChangeItem = changedItem
	} else if changedItem.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] == cmd.PortBreakoutChanSpeedPathItemC {
		numChannels, err = mngr.getPortBreakoutNumChannelsFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}

		channelSpeed = changedItem.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
		if !mngr.isValidPortBreakoutChannelSpeed(numChannels, channelSpeed) {
			return fmt.Errorf("Speed channel (%d) is invalid", channelSpeed)
		}

		numChannelsChangeItem, err = mngr.getPortBreakoutNumChannelsChangeItemFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}
		channelSpeedChangeItem = changedItem
	} else {
		return fmt.Errorf("Unable to get port breakout changing")
	}

	log.Infof("Requested changing port %s breakout into %d mode with speed %d", ifname, numChannels, channelSpeed)
	conn, err := grpc.Dial(fmt.Sprintf(":%d", serv_param.MgmtListeningTcpPortC), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to dial into the switch gRPC server: %v", err)
		return err
	}
	defer conn.Close()
	ethSwitchMgmtClient := mgmt.NewEthSwitchMgmtClient(conn)
	setPortBreakoutCmd := cmd.NewSetPortBreakoutCmdT(numChannelsChangeItem, channelSpeedChangeItem, &ethSwitchMgmtClient)
	if err = setPortBreakoutCmd.Execute(); err != nil {
		return fmt.Errorf("Failed to execute set port breakout request: %s", err)
	}

	if err = setPortBreakoutCmd.Undo(); err != nil {
		return fmt.Errorf("Failed to withdraw port breakout request: %s", err)
	}
	if numChannels == cmd.PortBreakoutModeNoneC {
		// Check if there won't be any dependenies from master port
	} else {
		slavePorts := make([]string, cmd.PortBreakoutMode4xC)
		for i := 1; i <= 4; i++ {
			slavePorts[i-1] = fmt.Sprintf("%s.%d", ifname, i)
			log.Infof("Composed slave port: %s", slavePorts[i-1])
		}
		// Check if there won't be any dependenies from slave ports
	}

	return nil
}

// No exported data
func (mngr *CfgMngrT) getPortBreakoutChannelSpeedFromChangelog(ifname string, changelog *diff.Changelog) (oc.E_OpenconfigIfEthernet_ETHERNET_SPEED, error) {
	var err error = nil
	channelSpeed := oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
	for _, change := range *changelog {
		if mngr.isChangedPortBreakoutChannelSpeed(&change) {
			log.Infof("Found channel speed request too:\n%+v", change)
			if change.Path[cmd.PortBreakoutIfnamePathItemIdxC] == ifname {
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

func (mngr *CfgMngrT) getPortBreakoutChannelSpeedChangeItemFromChangelog(ifname string, changelog *diff.Changelog) (*diff.Change, error) {
	var err error = nil
	var changeItem *diff.Change
	channelSpeed := oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
	for _, change := range *changelog {
		if mngr.isChangedPortBreakoutChannelSpeed(&change) {
			log.Infof("Found channel speed request too:\n%+v", change)
			if change.Path[cmd.PortBreakoutIfnamePathItemIdxC] == ifname {
				channelSpeed = change.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
				changeItem = &change
				break
			}
		}
	}

	if channelSpeed == oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET {
		err = fmt.Errorf("Could not found set channel speed request")
	}

	return changeItem, err
}

func (mngr *CfgMngrT) isChangedPortBreakoutChannelSpeed(change *diff.Change) bool {
	if len(change.Path) < cmd.PortBreakoutPathItemsCountC {
		return false
	}

	if (change.Path[cmd.PortBreakoutCompPathItemIdxC] != cmd.PortBreakoutCompPathItemC) || (change.Path[cmd.PortBreakoutPortPathItemIdxC] != cmd.PortBreakoutPortPathItemC) || (change.Path[cmd.PortBreakoutModePathItemIdxC] != cmd.PortBreakoutModePathItemC) || (change.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] != cmd.PortBreakoutChanSpeedPathItemC) {
		return false
	}

	return true
}

func (mngr *CfgMngrT) isChangedPortBreakoutNumChannels(change *diff.Change) bool {
	if len(change.Path) < cmd.PortBreakoutPathItemsCountC {
		return false
	}

	if (change.Path[cmd.PortBreakoutCompPathItemIdxC] != cmd.PortBreakoutCompPathItemC) || (change.Path[cmd.PortBreakoutPortPathItemIdxC] != cmd.PortBreakoutPortPathItemC) || (change.Path[cmd.PortBreakoutModePathItemIdxC] != cmd.PortBreakoutModePathItemC) || (change.Path[cmd.PortBreakoutNumChanPathItemIdxC] != cmd.PortBreakoutNumChanPathItemC) {
		return false
	}

	return true
}

func (mngr *CfgMngrT) isValidPortBreakoutNumChannels(numChannels cmd.PortBreakoutModeT) bool {
	if numChannels == cmd.PortBreakoutModeNoneC || numChannels == cmd.PortBreakoutMode4xC {
		return true
	}

	return false
}

func (mngr *CfgMngrT) isValidPortBreakoutChannelSpeed(numChannels cmd.PortBreakoutModeT,
	channelSpeed oc.E_OpenconfigIfEthernet_ETHERNET_SPEED) bool {
	log.Infof("Split (%d), speed (%d)", numChannels, channelSpeed)
	switch channelSpeed {
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB:
		if numChannels == cmd.PortBreakoutMode4xC {
			return true
		}
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_100GB:
		fallthrough
	case oc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_40GB:
		if numChannels == cmd.PortBreakoutModeNoneC {
			return true
		}
	}

	return false
}

func (mngr *CfgMngrT) getPortBreakoutNumChannelsFromChangelog(ifname string, changelog *diff.Changelog) (cmd.PortBreakoutModeT, error) {
	var err error = nil
	numChannels := cmd.PortBreakoutModeInvalidC
	for _, change := range *changelog {
		if mngr.isChangedPortBreakoutNumChannels(&change) {
			log.Infof("Found changing number of channels request too:\n%+v", change)
			if change.Path[cmd.PortBreakoutIfnamePathItemIdxC] == ifname {
				numChannels = cmd.PortBreakoutModeT(change.To.(uint8))
				break
			}
		}
	}

	if !mngr.isValidPortBreakoutNumChannels(numChannels) {
		err = fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
	}

	return numChannels, err
}

func (mngr *CfgMngrT) getPortBreakoutNumChannelsChangeItemFromChangelog(ifname string, changelog *diff.Changelog) (*diff.Change, error) {
	var err error = nil
	var changeItem *diff.Change
	numChannels := cmd.PortBreakoutModeInvalidC
	for _, change := range *changelog {
		if mngr.isChangedPortBreakoutNumChannels(&change) {
			log.Infof("Found changing number of channels request too:\n%+v", change)
			if change.Path[cmd.PortBreakoutIfnamePathItemIdxC] == ifname {
				numChannels = cmd.PortBreakoutModeT(change.To.(uint8))
				changeItem = &change
				break
			}
		}
	}

	if !mngr.isValidPortBreakoutNumChannels(numChannels) {
		err = fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
	}

	return changeItem, err
}
