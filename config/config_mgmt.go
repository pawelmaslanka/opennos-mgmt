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

type cmdByIfnameT map[string]cmd.CommandI

type ConfigMngrT struct {
	configLookupTbl         *configLookupTablesT
	runningConfig           ygot.ValidatedGoStruct
	cmdByIfname             [MaxNumberOfActionsInTransactionC]cmdByIfnameT
	ethSwitchMgmtClientConn *grpc.ClientConn
	ethSwitchMgmtClient     *mgmt.EthSwitchMgmtClient
	// transactions    [TransactionIdx][MaxNumberOfActionsInTransactionC]cmdByIfnameT
	// transConfigLookupTbl every queued command should remove dependency from here
	// e.g. when LAG is going to be remove, we should remove ports from this LAG, and LAG itself
	transConfigLookupTbl *configLookupTablesT
	transHasBeenStarted  bool
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
	for i = 0; i < MaxNumberOfActionsInTransactionC; i++ {
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
	for i = 0; i < MaxNumberOfActionsInTransactionC; i++ {
		for _, command := range this.cmdByIfname[i] {
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

	i := MaxNumberOfActionsInTransactionC
	for i > 0 {
		i--
		for _, command := range this.cmdByIfname[i] {
			if err := command.Undo(); err != nil {
				for i < MaxNumberOfActionsInTransactionC {
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

func (this *ConfigMngrT) ValidatePortBreakoutChanging(changedItem *diff.Change, changelog *diff.Changelog) error {
	ifname := changedItem.Path[cmd.PortBreakoutIfnamePathItemIdxC]
	if !this.IsIntfAvailable(ifname) {
		return fmt.Errorf("Port %s is unrecognized", ifname)
	}

	var numChannels cmd.PortBreakoutModeT = cmd.PortBreakoutModeInvalidC
	var channelSpeed oc.E_OpenconfigIfEthernet_ETHERNET_SPEED = oc.OpenconfigIfEthernet_ETHERNET_SPEED_UNSET
	var numChannelsChangeItem *diff.Change
	var channelSpeedChangeItem *diff.Change
	var err error

	if changedItem.Path[cmd.PortBreakoutNumChanPathItemIdxC] == cmd.PortBreakoutNumChanPathItemC {
		channelSpeed, err = this.getPortBreakoutChannelSpeedFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}

		numChannels = cmd.PortBreakoutModeT(changedItem.To.(uint8))
		if !this.isValidPortBreakoutNumChannels(numChannels) {
			return fmt.Errorf("Number of channels (%d) to breakout is invalid", numChannels)
		}

		channelSpeedChangeItem, err = this.getPortBreakoutChannelSpeedChangeItemFromChangelog(ifname, changelog)
		if err != nil {
			return err
		}
		numChannelsChangeItem = changedItem
	} else if changedItem.Path[cmd.PortBreakoutChanSpeedPathItemIdxC] == cmd.PortBreakoutChanSpeedPathItemC {
		numChannels, err = this.getPortBreakoutNumChannelsFromChangelog(ifname, changelog)
		if err != nil {
			return this.validatePortBreakoutChannSpeedChanging(changedItem, changelog)
		}

		channelSpeed = changedItem.To.(oc.E_OpenconfigIfEthernet_ETHERNET_SPEED)
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

	log.Infof("Requested changing port %s breakout into %d mode with speed %d", ifname, numChannels, channelSpeed)
	setPortBreakoutCmd := cmd.NewSetPortBreakoutCmdT(numChannelsChangeItem, channelSpeedChangeItem, this.ethSwitchMgmtClient)
	if numChannels == cmd.PortBreakoutModeNoneC {
		// Check if there won't be any dependencies from slave port
		slavePorts := make([]string, cmd.PortBreakoutMode4xC)
		for i := 1; i <= 4; i++ {
			slavePorts[i-1] = fmt.Sprintf("%s.%d", ifname, i)
			log.Infof("Composed slave port: %s", slavePorts[i-1])
			slaveIfname := slavePorts[i-1]
			idx := this.transConfigLookupTbl.idxByIntfName[slaveIfname]
			// TODO: Go through by all dependencies like ther ordinal number
			if ip4, exists := this.transConfigLookupTbl.ipv4ByIntf[idx]; exists {
				return fmt.Errorf("Cannot %q because there is dependency from IPv4 %s", setPortBreakoutCmd.GetName(), ip4.Strings()[0])
			}
		}
	} else {
		// Check if there won't be any dependenies from master ports
	}

	// TODO: Remove this code: Execute/Undo
	if err = setPortBreakoutCmd.Execute(); err != nil {
		return fmt.Errorf("Failed to execute set port breakout request: %s", err)
	}
	if err = setPortBreakoutCmd.Undo(); err != nil {
		return fmt.Errorf("Failed to withdraw port breakout request: %s", err)
	}

	if this.transHasBeenStarted {
		setPortBreakoutCmd := cmd.NewSetPortBreakoutCmdT(numChannelsChangeItem, channelSpeedChangeItem, this.ethSwitchMgmtClient)
		return this.appendSetPortBreakoutCmdToTransaction(ifname, setPortBreakoutCmd)
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
