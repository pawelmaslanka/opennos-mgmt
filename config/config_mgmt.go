package config

import (
	"errors"
	"fmt"
	"opennos-mgmt/gnmi"
	"opennos-mgmt/gnmi/modeldata/oc"

	cmd "opennos-mgmt/config/command"

	log "github.com/golang/glog"
	"github.com/jinzhu/copier"
	"github.com/openconfig/ygot/ygot"
	"github.com/r3labs/diff"
	"google.golang.org/grpc"

	mgmt "opennos-eth-switch-service/mgmt"
	serv_param "opennos-eth-switch-service/serv-param"
)

// OrdinalNumberT underlying type of ordinal number for action performed in transaction
type OrdinalNumberT uint16

// The following constants define ordering numbers of actions in transaction
const (
	unorderedActionInTransactionC        OrdinalNumberT = iota
	deleteOrRemoveIpv4FromEthIntfC                      // Remove IPv4/CIDRv4 address from Ethernet interface
	deleteOrRemoveIpv4FromLagIntfC                      // Remove IPv4/CIDRv4 address from LAG interface
	deleteOrRemoveIpv6FromEthIntfC                      // Remove IPv6/CIDRv6 address from Ethernet interface
	deleteOrRemoveIpv6FromLagIntfC                      // Remove IPv6/CIDRv6 address from LAG interface
	deleteOrRemoveEthIntfFromAccessVlanC                // Remove Ethernet interface from access VLAN
	deleteOrRemoveLagIntfFromAccessVlanC                // Remove LAG interface from access VLAN
	deleteOrRemoveEthIntfFromNativeVlanC                // Remove Ethernet interface from native VLAN
	deleteOrRemoveLagIntfFromNativeVlanC                // Remove LAG interface from native VLAN
	deleteOrRemoveEthIntfFromTrunkVlanC                 // Remove Ethernet interface from trunk VLAN
	deleteOrRemoveLagIntfFromTrunkVlanC                 // Remove LAG interface from trunk VLAN
	deleteOrRemoveEthIntfFromLagIntfC                   // Remove Ethernet interface from LAG membership
	deleteOrRemoveLagIntfC                              // Delete LAG interface
	deleteOrRemovePortBreakoutC                         // Combine multiple logical ports into single port
	setOrAddPortBreakoutC                               // Break out front panel port into multiple logical ports
	setOrAddPortBreakoutChanSpeedC                      // Set channel speed on logical ports (lanes)
	setOrAddDescOnEthIntfC                              // Set description of Ethernet interface
	setOrAddPortAutoNegOnEthIntfC                       // Enable or disable auto-negotiation on port
	setOrAddPortMtuOnEthIntfC                           // Set MTU on port
	setOrAddPortSpeedOnEthIntfC                         // Set port speed
	setOrAddLagIntfC                                    // Add LAG interface
	setOrAddIpv4OnEthIntfC                              // Assign IPv4/CIDRv4 address to Ethernet interface
	setOrAddIpv4OnLagIntfC                              // Assign IPv4/CIDRv4 address to LAG interface
	setOrAddIpv6OnEthIntfC                              // Assign IPv6/CIDRv6 address to Ethernet interface
	setOrAddIpv6OnLagIntfC                              // Assign IPv6/CIDRv6 address to LAG interface
	setOrAddVlanIntfModeOfLagIntfC                      // Set VLAN interface mode
	setOrAddAccessVlanOnEthIntfC                        // Assign Ethernet interface to access VLAN
	setOrAddAccessVlanOnLagIntfC                        // Assign LAG interface to access VLAN
	setOrAddNativeVlanOnEthIntfC                        // Assign Ethernet interface to native VLAN
	setOrAddNativeVlanOnLagIntfC                        // Assign LAG interface to native VLAN
	setOrAddTrunkVlanOnEthIntfC                         // Assign Ethernet interface to trunk VLAN
	setOrAddTrunkVlanOnLagIntfC                         // Assign LAG interface to trunk VLAN
	setOrAddLagTypeOfLagIntfC                           // Set the type of LAG
	setOrAddLacpIntervalC                               // Set the period between LACP messages
	setOrAddLacpModeC                                   // Set LACP activity - active or passive
	maxNumberOfActionsInTransactionC                    // Defines maximum number of possible actions in transaction
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

type cmdByIfnameT map[string]cmd.CommandI

type ConfigMngrT struct {
	configLookupTbl         *configLookupTablesT
	runningConfig           ygot.ValidatedGoStruct
	cmdByIfname             [maxNumberOfActionsInTransactionC]cmdByIfnameT
	ethSwitchMgmtClientConn *grpc.ClientConn
	ethSwitchMgmtClient     *mgmt.EthSwitchMgmtClient
	// transactions    [TransactionIdx][maxNumberOfActionsInTransactionC]cmdByIfnameT
	// transConfigLookupTbl every queued command should remove dependency from here
	// e.g. when LAG is going to be remove, we should remove ports from this LAG, and LAG itself
	transConfigLookupTbl *configLookupTablesT
	transHasBeenStarted  bool // marks if transaction has been started
}

func NewConfigMngrT() *ConfigMngrT {
	return &ConfigMngrT{
		configLookupTbl:     newConfigLookupTables(),
		transHasBeenStarted: false,
	}
}

func (this *ConfigMngrT) NewTransaction() error {
	if this.transHasBeenStarted {
		return errors.New("Transaction is already active")
	}
	conn, err := grpc.Dial(fmt.Sprintf(":%d", serv_param.MgmtListeningTcpPortC), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Errorf("Failed to dial into the switch gRPC server: %v", err)
		return err
	}
	ethSwitchMgmtClient := mgmt.NewEthSwitchMgmtClient(conn)
	nilCmd := &cmd.NilCmdT{}
	var i OrdinalNumberT
	for i = 0; i < maxNumberOfActionsInTransactionC; i++ {
		this.cmdByIfname[i] = make(cmdByIfnameT, 1)
		this.cmdByIfname[i][nilCmd.GetName()] = nilCmd
	}

	this.transConfigLookupTbl = newConfigLookupTables()
	copier.Copy(&this.transConfigLookupTbl, &this.configLookupTbl)
	this.ethSwitchMgmtClientConn = conn
	this.ethSwitchMgmtClient = &ethSwitchMgmtClient
	this.transHasBeenStarted = true
	return nil
}

func (this *ConfigMngrT) Commit() error {
	if !this.transHasBeenStarted {
		return errors.New("Transaction not has been started")
	}

	var i OrdinalNumberT
	for i = 0; i < maxNumberOfActionsInTransactionC; i++ {
		for _, command := range this.cmdByIfname[i] {
			log.Infof("Execute command %q", command.GetName())
			if err := command.Execute(); err != nil {
				for i > 0 {
					i--
					command.Undo()
				}
				this.DiscardOrFinishTrans()
				return err
			}
		}
	}

	return nil
}

func (this *ConfigMngrT) Rollback() error {
	if !this.transHasBeenStarted {
		return errors.New("Transaction not has been started")
	}

	i := maxNumberOfActionsInTransactionC
	for i > 0 {
		i--
		for _, command := range this.cmdByIfname[i] {
			if err := command.Undo(); err != nil {
				for i < maxNumberOfActionsInTransactionC {
					i++
					command.Execute()
				}
				return err
			}
		}
	}

	this.DiscardOrFinishTrans()
	return nil
}

func (this *ConfigMngrT) CommitConfirm() error {
	if !this.transHasBeenStarted {
		return errors.New("Transaction not has been started")
	}
	this.DiscardOrFinishTrans()
	return nil
}

func (this *ConfigMngrT) Confirm() error {
	if !this.transHasBeenStarted {
		return errors.New("Transaction not has been started")
	}
	// TODO: Implement logic
	return nil
}

func (this *ConfigMngrT) DiscardOrFinishTrans() error {
	if !this.transHasBeenStarted {
		return errors.New("Transaction not has been started")
	}
	this.ethSwitchMgmtClientConn.Close()
	this.ethSwitchMgmtClient = nil
	this.transConfigLookupTbl = nil
	this.transHasBeenStarted = false
	return nil
}

func (this *ConfigMngrT) LoadConfig(model *gnmi.Model, config []byte) error {
	configModel, err := model.NewConfigStruct(config)
	if err != nil {
		return err
	}

	log.Infof("Dump config model: %+v", configModel)
	device := configModel.(*oc.Device)
	for intfName, _ := range device.Interface {
		if err := this.configLookupTbl.addNewInterfaceIfItDoesNotExist(intfName); err != nil {
			return err
		}
	}

	for intfName, _ := range this.configLookupTbl.idxByIntfName {
		intf := device.Interface[intfName]
		if intf == nil {
			log.Info("Cannot find interface ", intfName)
			return fmt.Errorf("Failed to get interface %s info", intfName)
		}

		eth := intf.GetEthernet()
		if eth != nil {
			log.Infof("Configuring interface %s as LAG member", intfName)
			if err := this.configLookupTbl.parseInterfaceAsLagMember(intfName, eth); err != nil {
				return err
			}

			swVlan := eth.GetSwitchedVlan()
			if swVlan != nil {
				if err := this.configLookupTbl.parseVlanForIntf(intfName, swVlan); err != nil {
					return err
				}
			}
		}

		subIntf := intf.GetSubinterface(0)
		if subIntf != nil {
			if err := this.configLookupTbl.parseSubinterface(intfName, subIntf); err != nil {
				return err
			}
		}
	}

	for lagName, _ := range this.configLookupTbl.idxByLagName {
		lag := device.Interface[lagName]
		if lag == nil {
			return fmt.Errorf("Failed to get LAG %s info", lagName)
		}

		agg := lag.GetAggregation()
		if agg != nil {
			swVlan := agg.GetSwitchedVlan()
			if swVlan != nil {
				if err := this.configLookupTbl.parseVlanForLagIntf(lagName, swVlan); err != nil {
					return err
				}
			}
		}
	}

	this.configLookupTbl.dump()
	// TODO: Check if there isn't inconsistency in VLANs between ethernet
	//       interface and aggregate ethernet interfaces

	log.Infof("There are loaded %d interfaces and %d LAGs",
		this.configLookupTbl.idxOfLastAddedIntf, this.configLookupTbl.idxOfLastAddedLag)

	return this.CommitCandidateConfig(&configModel)
}

func (this *ConfigMngrT) IsChangedPortBreakout(change *diff.Change) bool {
	if len(change.Path) < cmd.PortBreakoutPathItemsCountC {
		return false
	}

	if (change.Path[cmd.PortBreakoutCompPathItemIdxC] != cmd.PortBreakoutCompPathItemC) || (change.Path[cmd.PortBreakoutPortPathItemIdxC] != cmd.PortBreakoutPortPathItemC) || (change.Path[cmd.PortBreakoutModePathItemIdxC] != cmd.PortBreakoutModePathItemC) || ((change.Path[cmd.PortBreakoutNumChanPathItemIdxC] != cmd.PortBreakoutNumChanPathItemC) && (change.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] != cmd.PortBreakoutChanSpeedPathItemC)) {
		return false
	}

	return true
}

func (this *ConfigMngrT) IsChangedPortBreakoutChanSpeed(change *diff.Change) bool {
	if len(change.Path) < cmd.PortBreakoutPathItemsCountC {
		return false
	}

	if (change.Path[cmd.PortBreakoutCompPathItemIdxC] != cmd.PortBreakoutCompPathItemC) || (change.Path[cmd.PortBreakoutPortPathItemIdxC] != cmd.PortBreakoutPortPathItemC) || (change.Path[cmd.PortBreakoutModePathItemIdxC] != cmd.PortBreakoutModePathItemC) || ((change.Path[cmd.PortBreakoutNumChanPathItemIdxC] != cmd.PortBreakoutNumChanPathItemC) && (change.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] != cmd.PortBreakoutChanSpeedPathItemC)) {
		return false
	}

	return true
}

func (this *ConfigMngrT) IsIntfAvailable(ifname string) bool {
	if _, exists := this.configLookupTbl.idxByIntfName[ifname]; exists {
		return true
	}

	return false
}

func (this *ConfigMngrT) ValidatePortBreakoutChanging(changedItem *DiffChangeMgmtT, changelog *DiffChangelogMgmtT) error {
	ifname := changedItem.Change.Path[cmd.PortBreakoutIfnamePathItemIdxC]
	if !this.IsIntfAvailable(ifname) {
		return fmt.Errorf("Port %s is unrecognized", ifname)
	}

	var numChannels cmd.PortBreakoutModeT = cmd.PortBreakoutModeInvalidC
	var channelSpeed oc.E_OpenconfigIfEthernet_ETHERNET_SPEED = oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
	var numChannelsChangeItem *DiffChangeMgmtT
	var channelSpeedChangeItem *DiffChangeMgmtT
	var err error

	if changedItem.Change.Path[cmd.PortBreakoutNumChanPathItemIdxC] == cmd.PortBreakoutNumChanPathItemC {
		channelSpeed, err = this.getPortBreakoutChannelSpeedFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}

		numChannels = cmd.PortBreakoutModeT(changedItem.Change.To.(uint8))
		if !this.isValidPortBreakoutNumChannels(numChannels) {
			return fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
		}

		channelSpeedChangeItem, err = this.getPortBreakoutChannelSpeedChangeItemFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}
		numChannelsChangeItem = changedItem
	} else if changedItem.Change.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] == cmd.PortBreakoutChanSpeedPathItemC {
		numChannels, err = this.getPortBreakoutNumChannelsFromChangelog(ifname, changelog)
		if err != nil {
			return this.validatePortBreakoutChannSpeedChanging(changedItem, changelog)
		}

		channelSpeed = changedItem.Change.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
		if !this.isValidPortBreakoutChannelSpeed(numChannels, channelSpeed) {
			return fmt.Errorf("Speed channel (%d) is invalid", channelSpeed)
		}

		numChannelsChangeItem, err = this.getPortBreakoutNumChannelsChangeItemFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}
		channelSpeedChangeItem = changedItem
	} else {
		return fmt.Errorf("Unable to get port breakout changing")
	}

	log.Infof("Requested changing port %s breakout into mode %d with speed %d", ifname, numChannels, channelSpeed)
	setPortBreakoutCmd := cmd.NewSetPortBreakoutCmdT(numChannelsChangeItem.Change, channelSpeedChangeItem.Change, this.ethSwitchMgmtClient)
	if numChannels == cmd.PortBreakoutModeNoneC {
		for i := 1; i <= 4; i++ {
			slavePort := fmt.Sprintf("%s.%d", ifname, i)
			log.Infof("Composed slave port: %s", slavePort)
			if err := this.configLookupTbl.checkDependenciesForDeletePortBreakout(slavePort); err != nil {
				return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
					setPortBreakoutCmd.GetName(), slavePort, err)
			}
		}
	} else {
		if err := this.configLookupTbl.checkDependenciesForDeletePortBreakout(ifname); err != nil {
			return fmt.Errorf("Cannot %q because there are dependencies from interface %s:\n%s",
				setPortBreakoutCmd.GetName(), ifname, err)
		}
	}

	if this.transHasBeenStarted {
		setPortBreakoutCmd := cmd.NewSetPortBreakoutCmdT(numChannelsChangeItem.Change, channelSpeedChangeItem.Change, this.ethSwitchMgmtClient)
		if err = this.appendSetPortBreakoutCmdToTransaction(ifname, setPortBreakoutCmd); err != nil {
			return err
		}

		numChannelsChangeItem.MarkAsProcessed()
		channelSpeedChangeItem.MarkAsProcessed()
	}

	return nil
}

// TODO: Maybe move it into DiscardOrFinishTrans()
func (this *ConfigMngrT) CommitCandidateConfig(candidateConfig *ygot.ValidatedGoStruct) error {
	return copier.Copy(&this.runningConfig, candidateConfig)
}

func (this *ConfigMngrT) GetDiffRunningConfigWithCandidateConfig(candidateConfig *ygot.ValidatedGoStruct) (diff.Changelog, error) {
	return diff.Diff(this.runningConfig, *candidateConfig)
}
