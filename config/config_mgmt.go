package config

import (
	"container/list"
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
	unorderedActionInTransactionC    OrdinalNumberT = iota
	deleteIpv4AddrFromEthIntfC                      // Remove IPv4/CIDRv4 address from Ethernet interface
	deleteIpv4AddrFromLagIntfC                      // Remove IPv4/CIDRv4 address from LAG interface
	deleteIpv6AddrFromEthIntfC                      // Remove IPv6/CIDRv6 address from Ethernet interface
	deleteIpv6AddrFromLagIntfC                      // Remove IPv6/CIDRv6 address from LAG interface
	deleteEthIntfFromAccessVlanC                    // Remove Ethernet interface from access VLAN
	deleteLagIntfFromAccessVlanC                    // Remove LAG interface from access VLAN
	deleteEthIntfFromNativeVlanC                    // Remove Ethernet interface from native VLAN
	deleteLagIntfFromNativeVlanC                    // Remove LAG interface from native VLAN
	deleteEthIntfFromTrunkVlanC                     // Remove Ethernet interface from trunk VLAN
	deleteEthIntfFromLagIntfC                       // Remove Ethernet interface from LAG membership
	deleteLagIntfC                                  // Delete LAG interface
	deletePortBreakoutC                             // Combine multiple logical ports into single port
	setPortBreakoutC                                // Break out front panel port into multiple logical ports
	setPortBreakoutChanSpeedC                       // Set channel speed on logical ports (lanes)
	setDescForEthIntfC                              // Set description of Ethernet interface
	setPortAutoNegForEthIntfC                       // Enable or disable auto-negotiation on port
	setPortMtuForEthIntfC                           // Set MTU on port
	setPortSpeedForEthIntfC                         // Set port speed
	setLagIntfC                                     // Add LAG interface
	setIpv4AddrForEthIntfC                          // Assign IPv4/CIDRv4 address to Ethernet interface
	setIpv4AddrForLagIntfC                          // Assign IPv4/CIDRv4 address to LAG interface
	setIpv6AddrForEthIntfC                          // Assign IPv6/CIDRv6 address to Ethernet interface
	setIpv6AddrForLagIntfC                          // Assign IPv6/CIDRv6 address to LAG interface
	setVlanIntfModeForEthIntfC                      // Set VLAN interface mode for Ethernet interface
	setVlanIntfModeForLagIntfC                      // Set VLAN interface mode for LAG interface
	setAccessVlanForEthIntfC                        // Assign Ethernet interface to access VLAN
	setAccessVlanForLagIntfC                        // Assign LAG interface to access VLAN
	setNativeVlanForEthIntfC                        // Assign Ethernet interface to native VLAN
	setNativeVlanForLagIntfC                        // Assign LAG interface to native VLAN
	setTrunkVlanForEthIntfC                         // Assign Ethernet interface to trunk VLAN
	setTrunkVlanForLagIntfC                         // Assign LAG interface to trunk VLAN
	setLagTypeOfLagIntfC                            // Set the type of LAG
	setLacpIntervalC                                // Set the period between LACP messages
	setLacpModeC                                    // Set LACP activity - active or passive
	maxNumberOfActionsInTransactionC                // Defines maximum number of possible actions in transaction
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
	transCmdList         *list.List
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

	this.transConfigLookupTbl = this.configLookupTbl.makeCopy()
	this.transCmdList = list.New()
	this.ethSwitchMgmtClientConn = conn
	this.ethSwitchMgmtClient = &ethSwitchMgmtClient
	this.transHasBeenStarted = true
	return nil
}

func (this *ConfigMngrT) CommitBackup() error {
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

func (this *ConfigMngrT) Commit() error {
	if !this.transHasBeenStarted {
		return errors.New("Transaction not has been started")
	}

	for ex := this.transCmdList.Front(); ex != nil; ex = ex.Next() {
		execCmd := ex.Value.(cmd.CommandI)
		log.Infof("Execute command %q", execCmd.GetName())
		if err := execCmd.Execute(); err != nil {
			for un := ex.Prev(); un != nil; un = un.Prev() {
				undoCmd := un.Value.(cmd.CommandI)
				undoCmd.Undo()
			}
			this.DiscardOrFinishTrans()
			return err
		}
	}

	return nil
}

func (this *ConfigMngrT) RollbackBackup() error {
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

func (this *ConfigMngrT) Rollback() error {
	if !this.transHasBeenStarted {
		return errors.New("Transaction not has been started")
	}

	for un := this.transCmdList.Back(); un != nil; un = un.Prev() {
		undoCmd := un.Value.(cmd.CommandI)
		log.Infof("Undo command %q", undoCmd.GetName())
		if err := undoCmd.Undo(); err != nil {
			for ex := un.Next(); ex != nil; ex = ex.Next() {
				execCmd := ex.Value.(cmd.CommandI)
				execCmd.Execute()
			}
			this.DiscardOrFinishTrans()
			return err
		}
	}

	this.DiscardOrFinishTrans()
	return nil
}

func (this *ConfigMngrT) CommitConfirm() error {
	return this.Commit()
}

func (this *ConfigMngrT) Confirm() error {
	if !this.transHasBeenStarted {
		return errors.New("Transaction not has been started")
	}

	this.configLookupTbl = this.transConfigLookupTbl.makeCopy()

	this.DiscardOrFinishTrans()
	return nil
}

func (this *ConfigMngrT) DiscardOrFinishTrans() error {
	if !this.transHasBeenStarted {
		return errors.New("Transaction not has been started")
	}
	this.ethSwitchMgmtClientConn.Close()
	this.ethSwitchMgmtClient = nil
	this.transConfigLookupTbl = nil
	this.transCmdList.Init()
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

	for intfName, _ := range this.configLookupTbl.idxByEthName {
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

func (this *ConfigMngrT) appendCmdToTransactionByIfname(ifname string, cmdAdd cmd.CommandI, idx OrdinalNumberT) error {
	cmds := this.cmdByIfname[idx]
	for _, command := range cmds {
		if command.Equals(cmdAdd) {
			return fmt.Errorf("Command %q already exists in transaction", command.GetName())
		}
	}

	log.Infof("Append command %q to transaction", cmdAdd.GetName())

	cmds[ifname] = cmdAdd
	this.addCmdToListTrans(cmdAdd)
	return nil
}

func (this *ConfigMngrT) IsChangedIpv4AddrEth(change *diff.Change) bool {
	if len(change.Path) < cmd.Ipv4AddrEthPathItemsCountC {
		return false
	}

	if (change.Path[cmd.Ipv4AddrEthIntfPathItemIdxC] != cmd.Ipv4AddrEthIntfPathItemC) || (change.Path[cmd.Ipv4AddrEthSubintfPathItemIdxC] != cmd.Ipv4AddrEthSubintfPathItemC) || (change.Path[cmd.Ipv4AddrEthSubintfIpv4PathItemIdxC] != cmd.Ipv4AddrEthSubintfIpv4PathItemC) || (change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPathItemIdxC] != cmd.Ipv4AddrEthSubintfIpv4AddrPathItemC) || ((change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPartIpPathItemIdxC] != cmd.Ipv4AddrEthSubintfIpv4AddrPartIpPathItemC) && (change.Path[cmd.Ipv4AddrEthSubintfIpv4AddrPartPrfxLenPathItemIdxC] != cmd.Ipv4AddrEthSubintfIpv4AddrPartPrfxLenPathItemC)) {
		return false
	}

	return true
}

// TODO: Maybe move it into DiscardOrFinishTrans()
func (this *ConfigMngrT) CommitCandidateConfig(candidateConfig *ygot.ValidatedGoStruct) error {
	// TODO: Consider if we should commit transConfigLookupTable here?
	// var configData []byte
	// configData, err := json.Marshal(*candidateConfig)
	// if err != nil {
	// 	return err
	// }

	// log.Infof("%s", configData)

	// err = oc.Unmarshal(configData, this.runningConfig)
	// // err = json.Unmarshal(configData, &this.runningConfig)
	// if err != nil {
	// 	return err
	// }

	// log.Infof("%v", this.runningConfig)

	// return nil

	// model := gnmi.NewModel(modeldata.ModelData,
	// 	reflect.TypeOf((*oc.Device)(nil)),
	// 	oc.SchemaTree["Device"],
	// 	oc.Unmarshal,
	// 	oc.ΛEnum)
	// this.runningConfig, err = model.NewConfigStruct(configData)
	// return err
	// TODO: Make deep copy
	return copier.Copy(&this.runningConfig, &candidateConfig)
}

func (this *ConfigMngrT) GetDiffRunningConfigWithCandidateConfig(candidateConfig *ygot.ValidatedGoStruct) (diff.Changelog, error) {
	return diff.Diff(this.runningConfig, *candidateConfig)
}

func (this *ConfigMngrT) isEthIntfAvailable(ifname string) bool {
	if _, exists := this.configLookupTbl.idxByEthName[ifname]; exists {
		return true
	}

	return false
}

func (this *ConfigMngrT) addCmdToListTrans(cmd cmd.CommandI) {
	this.transCmdList.PushBack(cmd)
}

func (this *ConfigMngrT) CommitChangelog(changelog *diff.Changelog, dryRun bool) error {
	var err error
	if !dryRun {
		defer this.DiscardOrFinishTrans()
		if err = this.NewTransaction(); err != nil {
			log.Errorf("Failed to start new transaction")
			return err
		}
	} else {
		this.transConfigLookupTbl = this.configLookupTbl.makeCopy()
	}

	countChanges := len(*changelog)
	var cnt int
	diffChangelog := NewDiffChangelogMgmtT(changelog)
	for {
		// Deletion section
		if cnt, err = this.processDeleteIpv4AddrEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		countChanges -= cnt
		if countChanges <= 0 {
			break
		}
		if cnt, err = this.processDeleteAccessVlanEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		countChanges -= cnt
		if countChanges <= 0 {
			break
		}
		if cnt, err = this.processDeleteNativeVlanEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		countChanges -= cnt
		if countChanges <= 0 {
			break
		}
		if cnt, err = this.processDeleteTrunkVlanEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		countChanges -= cnt
		if countChanges <= 0 {
			break
		}
		// Set section
		if cnt, err = this.processSetPortBreakoutFromChangelog(diffChangelog); err != nil {
			return err
		}
		countChanges -= cnt
		if countChanges <= 0 {
			break
		}
		if cnt, err = this.processSetPortBreakoutChanSpeedFromChangelog(diffChangelog); err != nil {
			return err
		}
		countChanges -= cnt
		if countChanges <= 0 {
			break
		}
		if cnt, err = this.processSetIpv4AddrEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		countChanges -= cnt
		if countChanges <= 0 {
			break
		}
		if cnt, err = this.processSetNativeVlanEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		countChanges -= cnt
		if countChanges <= 0 {
			break
		}
		// if len(changedItem.Change.Path) > 4 {
		// 	if "NativeVlan" == changedItem.Change.Path[4] {
		// 		port := make([]string, 1)
		// 		port[0] = changedItem.Change.Path[1]
		// 		// TODO: Uncomment if build is dedicated for target device
		// 		// if err := vlan.SetNativeVlan(port, changedItem.To.(uint16)); err != nil {
		// 		// 	log.Errorf("Failed to set native VLAN")
		// 		// 	return err
		// 		// }
		// 		log.Infof("Native VLAN has been changed to %d on port %s",
		// 			changedItem.Change.To, changedItem.Change.Path[1])
		// 	}
		// } else if len(changedItem.Change.Path) > 2 {
		// 	if "Mtu" == changedItem.Change.Path[2] {
		// 		log.Infof("Changing MTU to %d on port %s", changedItem.Change.To, changedItem.Change.Path[1])
		// 	}
		// }
		break
	}

	if !dryRun {
		if err := this.Commit(); err != nil {
			log.Errorf("Failed to commit changes")
			return err
		}
		if err := this.Confirm(); err != nil {
			log.Errorf("Failed to confirm committed changes")
			return err
		}
	} else {
		this.transConfigLookupTbl = nil
	}

	return nil
}
