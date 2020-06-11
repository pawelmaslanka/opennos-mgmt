package config

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"opennos-mgmt/gnmi"
	"opennos-mgmt/gnmi/modeldata/oc"
	"opennos-mgmt/utils"
	"regexp"
	"strconv"
	"time"

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
	deleteIpv4AddrFromAggIntfC                      // Remove IPv4/CIDRv4 address from LAG interface
	deleteIpv6AddrFromEthIntfC                      // Remove IPv6/CIDRv6 address from Ethernet interface
	deleteIpv6AddrFromAggIntfC                      // Remove IPv6/CIDRv6 address from LAG interface
	deleteEthIntfFromAccessVlanC                    // Remove Ethernet interface from access VLAN
	deleteAggIntfFromAccessVlanC                    // Remove LAG interface from access VLAN
	deleteEthIntfFromNativeVlanC                    // Remove Ethernet interface from native VLAN
	deleteAggIntfFromNativeVlanC                    // Remove LAG interface from native VLAN
	deleteEthIntfFromTrunkVlanC                     // Remove Ethernet interface from trunk VLAN
	deleteVlanC                                     // Delete VLAN
	deleteEthIntfFromAggIntfC                       // Remove Ethernet interface from LAG membership
	deleteAggIntfParamsC                            // Remove LAG parameters
	deleteAggIntfMemberC                            // Remove Ethernet interface from LAG
	deleteAggIntfLacpC                              // Disable LACP protocol for aggregate interface
	deleteAggIntfC                                  // Delete LAG interface
	deleteEthIntfC                                  // Delete Ethernet interface
	deletePortBreakoutC                             // Combine multiple logical ports into single port
	setPortBreakoutC                                // Break out front panel port into multiple logical ports
	setPortBreakoutChanSpeedC                       // Set channel speed on logical ports (lanes)
	setEthIntfC                                     // Create new Ethernet interface
	setDescForEthIntfC                              // Set description of Ethernet interface
	setPortAutoNegForEthIntfC                       // Enable or disable auto-negotiation on port
	setPortMtuForEthIntfC                           // Set MTU on port
	setPortSpeedForEthIntfC                         // Set port speed
	setAggIntfC                                     // Create new LAG interface
	setAggIntfLacpC                                 // Enable LACP protocol for aggregate interface
	setAggIntfParamsC                               // Set LAG parameters
	setAggIntfMemberC                               // Add Ethernet interface to LAG
	setVlanC                                        // Create new VLAN
	setVlanModeForEthIntfC                          // Set VLAN interface mode for Ethernet interface
	setVlanModeForAggIntfC                          // Set VLAN interface mode for LAG interface
	setAccessVlanForEthIntfC                        // Assign Ethernet interface to access VLAN
	setAccessVlanForAggIntfC                        // Assign LAG interface to access VLAN
	setNativeVlanForEthIntfC                        // Assign Ethernet interface to native VLAN
	setNativeVlanForAggIntfC                        // Assign LAG interface to native VLAN
	setTrunkVlanForEthIntfC                         // Assign Ethernet interface to trunk VLAN
	setTrunkVlanForAggIntfC                         // Assign LAG interface to trunk VLAN
	setIpv4AddrForEthIntfC                          // Assign IPv4/CIDRv4 address to Ethernet interface
	setIpv4AddrForAggIntfC                          // Assign IPv4/CIDRv4 address to LAG interface
	setIpv6AddrForEthIntfC                          // Assign IPv6/CIDRv6 address to Ethernet interface
	setIpv6AddrForAggIntfC                          // Assign IPv6/CIDRv6 address to LAG interface
	setLagTypeOfAggIntfC                            // Set the type of LAG
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

const (
	startupConfigFilenameC = "startup-config.json"
)

type cmdByNameT map[string]cmd.CommandI

// ConfigMngrT is responisble for management of device configuration
type ConfigMngrT struct {
	configLookupTbl         *configLookupTablesT
	runningConfig           ygot.ValidatedGoStruct
	cmdByName               [maxNumberOfActionsInTransactionC]cmdByNameT
	ethSwitchMgmtClientConn *grpc.ClientConn
	ethSwitchMgmtClient     *mgmt.EthSwitchMgmtClient
	// transactions    [TransactionIdx][maxNumberOfActionsInTransactionC]cmdByNameT
	// transConfigLookupTbl every queued command should remove dependency from here
	// e.g. when LAG is going to be remove, we should remove ports from this LAG, and LAG itself
	transConfigLookupTbl             *configLookupTablesT
	transCmdList                     *list.List
	transConfirmationTimeoutCtx      context.Context
	transConfirmationCancel          context.CancelFunc
	transConfirmationCandidateConfig *ygot.ValidatedGoStruct
	transHasBeenStarted              bool // marks if transaction has been started
}

// NewConfigMngrT creates instance of ConfigMngrT object
func NewConfigMngrT() *ConfigMngrT {
	return &ConfigMngrT{
		configLookupTbl:     newConfigLookupTables(),
		transHasBeenStarted: false,
	}
}

func (this *ConfigMngrT) NewTransaction() error {
	if this.isTransPending() {
		return errors.New("Transaction is already active")
	}
	conn, err := grpc.Dial(fmt.Sprintf(":%d", serv_param.MgmtListeningTcpPortC), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Errorf("Failed to dial into the switch gRPC server: %v", err)
		return err
	}
	ethSwitchMgmtClient := mgmt.NewEthSwitchMgmtClient(conn)
	nilCmd := &cmd.NilCmdT{}
	// TODO: Check if it is still required?
	var i OrdinalNumberT
	for i = 0; i < maxNumberOfActionsInTransactionC; i++ {
		this.cmdByName[i] = make(cmdByNameT, 1)
		this.cmdByName[i][nilCmd.GetName()] = nilCmd
	}

	this.transConfigLookupTbl = this.configLookupTbl.makeCopy()
	this.transCmdList = list.New()
	this.ethSwitchMgmtClientConn = conn
	this.ethSwitchMgmtClient = &ethSwitchMgmtClient
	this.transHasBeenStarted = true
	return nil
}

func (this *ConfigMngrT) Commit() error {
	if !this.isTransPending() {
		return errors.New("Transaction has not been started")
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

func (this *ConfigMngrT) Rollback() error {
	if !this.isTransPending() {
		return errors.New("Transaction has not been started")
	}

	var err error = nil
	for un := this.transCmdList.Back(); un != nil; un = un.Prev() {
		undoCmd := un.Value.(cmd.CommandI)
		log.Infof("Undo command %q", undoCmd.GetName())
		if err = undoCmd.Undo(); err != nil {
			for ex := un.Next(); ex != nil; ex = ex.Next() {
				execCmd := ex.Value.(cmd.CommandI)
				execCmd.Execute()
			}

			break
		}
	}

	this.DiscardOrFinishTrans()
	return err
}

func (this *ConfigMngrT) CommitConfirm() error {
	return this.Commit()
}

func (this *ConfigMngrT) Confirm() error {
	if !this.transHasBeenStarted {
		return errors.New("Transaction has not been started")
	}

	candidateConfig := this.transConfirmationCandidateConfig
	this.configLookupTbl = this.transConfigLookupTbl.makeCopy()
	this.DiscardOrFinishTrans()
	return this.CommitCandidateConfig(candidateConfig)
}

func (this *ConfigMngrT) DiscardOrFinishTrans() error {
	if !this.isTransPending() {
		return errors.New("Transaction has not been started")
	}
	this.ethSwitchMgmtClientConn.Close()
	this.ethSwitchMgmtClient = nil
	this.transConfigLookupTbl = nil
	this.transCmdList.Init()
	// Check context before clean all related data
	this.transConfirmationTimeoutCtx = nil
	this.transConfirmationCancel = nil
	this.transConfirmationCandidateConfig = nil
	this.transHasBeenStarted = false
	return nil
}

func getBreakoutMasterPort(ifname string) (string, bool) {
	rgx := regexp.MustCompile(`eth-|/`)
	tokens := rgx.Split(ifname, -1)
	if len(tokens) == 4 { // breakout mode enable
		masterPort := fmt.Sprintf("eth-%s/%s", tokens[1], tokens[2]) // masterPort
		return masterPort, true
	}

	return "", false
}

func (this *ConfigMngrT) LoadConfig(model *gnmi.Model, config []byte) error {
	var err error
	configModel, err := model.NewConfigStruct(config)
	if err != nil {
		return err
	}

	log.Infof("Dump config model: %+v", configModel)
	device := configModel.(*oc.Device)
	for ifname := range device.Interface {
		masterPort, exists := getBreakoutMasterPort(ifname)
		if exists { // breakout mode enable
			if !isPortSplitted(device, masterPort) {
				log.Infof("Port %s is not splitted", masterPort)
				continue
			}
		}

		intf := device.Interface[ifname]
		if agg := intf.GetAggregation(); agg != nil {
			if agg.GetLagType() == oc.OpenconfigIfAggregate_AggregationType_UNSET {
				return fmt.Errorf("Invalid LAG type on interface %s",
					ifname, err)
			}

			if err = this.configLookupTbl.addNewAggIntfIfItDoesNotExist(ifname); err != nil {
				return err
			}
		} else {
			if err = this.configLookupTbl.addNewEthIntfIfItDoesNotExist(ifname); err != nil {
				return err
			}
		}
	}

	for ethIfname := range this.configLookupTbl.idxByEthIfname {
		masterPort, exists := getBreakoutMasterPort(ethIfname)
		if exists { // breakout mode enable
			if !isPortSplitted(device, masterPort) {
				log.Infof("Port %s is not splitted", masterPort)
				continue
			}
		}

		intf := device.Interface[ethIfname]
		if intf == nil {
			log.Info("Cannot find interface ", ethIfname)
			return fmt.Errorf("Failed to get interface %s info", ethIfname)
		}

		eth := intf.GetEthernet()
		if eth != nil {
			log.Infof("Configuring interface %s as LAG member", ethIfname)
			if err := this.configLookupTbl.parseInterfaceAsLagMember(ethIfname, eth); err != nil {
				return err
			}

			swVlan := eth.GetSwitchedVlan()
			if swVlan != nil {
				if err := this.configLookupTbl.parseVlanForIntf(ethIfname, swVlan); err != nil {
					return err
				}
			}
		}

		subIntf := intf.GetSubinterface(0)
		if subIntf != nil {
			if err := this.configLookupTbl.parseSubinterface(ethIfname, subIntf); err != nil {
				return err
			}
		}
	}

	for aggIfname := range this.configLookupTbl.idxByAggIfname {
		lag := device.Interface[aggIfname]
		if lag == nil {
			return fmt.Errorf("Failed to get LAG %s info", aggIfname)
		}

		agg := lag.GetAggregation()
		if agg != nil {
			swVlan := agg.GetSwitchedVlan()
			if swVlan != nil {
				if err := this.configLookupTbl.parseVlanForAggIntf(aggIfname, swVlan); err != nil {
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

	if err = this.configureDevice(&configModel); err != nil {
		return err
	}

	return this.CommitCandidateConfig(&configModel)
}

func (this *ConfigMngrT) configureDevice(configModel *ygot.ValidatedGoStruct) error {
	device := (*configModel).(*oc.Device)
	var err error
	if err = this.NewTransaction(); err != nil {
		return err
	}

	if err = this.setPortBreakout(device); err != nil {
		return err
	}

	if err = this.setEthIntf(device); err != nil {
		return err
	}

	if err = this.setAggIntf(device); err != nil {
		return err
	}

	if err = this.setAggIntfMember(device); err != nil {
		return err
	}

	if err = this.setVlanEthIntf(device); err != nil {
		return err
	}

	if err = this.setIpv4AddrEthIntf(device); err != nil {
		return err
	}

	if err = this.Commit(); err != nil {
		return err
	}

	return this.DiscardOrFinishTrans()
}

func isPortSplitted(device *oc.Device, ethIfname string) bool {
	// We want to process only not splitted ports
	_, exists := getBreakoutMasterPort(ethIfname)
	if exists { // breakout mode enable
		return false
	}

	comp := device.GetComponent(ethIfname)
	if comp == nil {
		return false
	}

	port := comp.GetPort()
	if port == nil {
		return false
	}

	mode := port.GetBreakoutMode()
	if mode == nil {
		return false
	}

	numChannels := mode.GetNumChannels()
	if numChannels == uint8(cmd.PortBreakoutModeInvalidC) {
		return false
	}

	if numChannels != uint8(cmd.PortBreakoutModeNoneC) {
		return true
	}

	return false
}

func (this *ConfigMngrT) appendCmdToTransaction(idName string, cmdAdd cmd.CommandI, idx OrdinalNumberT, shouldBeMerged bool) error {
	cmds := this.cmdByName[idx]
	for _, command := range cmds {
		if command.Equals(cmdAdd) {
			return fmt.Errorf("Command %q already exists in transaction", command.GetName())
		}
	}

	log.Infof("%s: Command %q and should be in batch - %v", idName, cmdAdd.GetName(), shouldBeMerged)

	if shouldBeMerged {
		if cmd, exists := cmds[idName]; exists {
			log.Infof("Already exists")
			if ok, err := cmd.Append(cmdAdd); err != nil {
				return err
			} else if ok {
				log.Infof("Appended to transaction")
				return nil
			}
		} else {
			log.Infof("Not exists")
		}
	}

	cmds[idName] = cmdAdd
	this.addCmdToListTrans(cmdAdd)
	log.Infof("Added to transaction")

	return nil
}

// TODO: Maybe move it into DiscardOrFinishTrans()
func (this *ConfigMngrT) CommitCandidateConfig(candidateConfig *ygot.ValidatedGoStruct) error {
	// TODO: Consider if we should commit transConfigLookupTable here?
	// TODO: Make deep copy?
	if err := copier.Copy(&this.runningConfig, &candidateConfig); err != nil {
		return err
	}

	return gnmi.SaveConfigFile(this.runningConfig, startupConfigFilenameC)
}

func (this *ConfigMngrT) GetDiffRunningConfigWithCandidateConfig(candidateConfig *ygot.ValidatedGoStruct) (diff.Changelog, error) {
	return diff.Diff(this.runningConfig, *candidateConfig)
}

func (this *ConfigMngrT) isEthIntfAvailable(ifname string) bool {
	if _, exists := this.transConfigLookupTbl.idxByEthIfname[ifname]; exists {
		return true
	}

	return false
}

func (this *ConfigMngrT) addCmdToListTrans(cmd cmd.CommandI) {
	this.transCmdList.PushBack(cmd)
}

func (this *ConfigMngrT) isTransPending() bool {
	return this.transHasBeenStarted
}

func extractCreateEthIntfParams(changelog *diff.Changelog) (*diff.Changelog, error) {
	changes := make([]diff.Change, 0)
	var err error
	for _, ch := range *changelog {
		if !isCreateDiffChange(&ch) {
			continue
		}

		var newChanges []diff.Change
		if isCreateOrDeleteEthIntf(&ch) {
			newEthIntfChanges := make([]diff.Change, 0)
			ifname := ch.Path[cmd.EthIntfIfnamePathItemIdxC]
			fmt.Printf("Creating new Ethernet interface %s\n%T\n", ifname, ch.To)
			isDelete := true
			ethIntf := ch.To.(*oc.Interface_Ethernet)
			if newEthIntfChanges, err = extractAggIdFromEthIntf(ifname, ethIntf, isDelete); err != nil {
				return nil, err
			}

			if len(newEthIntfChanges) > 0 {
				newChanges = append(newChanges, newEthIntfChanges...)
			}

			if newChanges, err = extractVlanRelatedParametersFromEthIntf(ifname, ethIntf, isDelete); err != nil {
				return nil, err
			}

			if len(newEthIntfChanges) > 0 {
				newChanges = append(newChanges, newEthIntfChanges...)
			}
		}

		if isCreateOrDeleteEthSubintfIpv4(&ch) {
			ifname := ch.Path[cmd.Ipv4AddrEthIfnamePathItemIdxC]
			subintfIdx, err := strconv.Atoi(ch.Path[cmd.Ipv4AddrEthSubintfIdxPathItemIdxC])
			if err != nil {
				return nil, err
			}

			if subintfIdx < 0 {
				return nil, fmt.Errorf("Negative value of Ethernet subinterface index")
			}

			fmt.Printf("Creating new Ethernet subinterface IPv4 %s\n%T\n", ifname, ch.To)
			subintf := ch.To.(*oc.Interface_Subinterface_Ipv4)
			if newChanges, err = extractIpParametersFromEthSubintfIpv4(ifname, subintfIdx, subintf, false); err != nil {
				return nil, err
			}
		}

		changes = append(changes, newChanges...)
	}

	*changelog = append(*changelog, changes...)

	return changelog, nil
}

func extractDeleteEthIntfParams(changelog *diff.Changelog) (*diff.Changelog, error) {
	changes := make([]diff.Change, 0)
	var err error
	for _, ch := range *changelog {
		fmt.Printf("Delete change \n%v\n", ch)

		if !isDeleteDiffChange(&ch) {
			continue
		}

		var newChanges []diff.Change
		// On the above check we know that we have 'delete case'
		if isCreateOrDeleteEthIntf(&ch) {
			newEthIntfChanges := make([]diff.Change, 0)
			ifname := ch.Path[cmd.EthIntfIfnamePathItemIdxC]
			fmt.Printf("Removing Ethernet interface %s\n%T\n", ifname, ch.From)
			isDelete := true
			ethIntf := ch.From.(*oc.Interface_Ethernet)
			if newEthIntfChanges, err = extractAggIdFromEthIntf(ifname, ethIntf, isDelete); err != nil {
				return nil, err
			}

			if len(newEthIntfChanges) > 0 {
				newChanges = append(newChanges, newEthIntfChanges...)
			}

			if newChanges, err = extractVlanRelatedParametersFromEthIntf(ifname, ethIntf, isDelete); err != nil {
				return nil, err
			}

			if len(newEthIntfChanges) > 0 {
				newChanges = append(newChanges, newEthIntfChanges...)
			}
		}

		if isCreateOrDeleteEthSubintfIpv4(&ch) {
			ifname := ch.Path[cmd.Ipv4AddrEthIfnamePathItemIdxC]
			subintfIdx, err := strconv.Atoi(ch.Path[cmd.Ipv4AddrEthSubintfIdxPathItemIdxC])
			if err != nil {
				return nil, err
			}

			if subintfIdx < 0 {
				return nil, fmt.Errorf("Negative value of Ethernet subinterface index")
			}

			fmt.Printf("Creating new Ethernet subinterface IPv4 %s\n%T\n", ifname, ch.From)
			subintf := ch.From.(*oc.Interface_Subinterface_Ipv4)
			if newChanges, err = extractIpParametersFromEthSubintfIpv4(ifname, subintfIdx, subintf, true); err != nil {
				return nil, err
			}
		}

		changes = append(changes, newChanges...)
	}

	*changelog = append(*changelog, changes...)

	return changelog, nil
}

func isCreateDiffChange(change *diff.Change) bool {
	if change.Type != diff.CREATE {
		return false
	}

	if change.From != nil {
		return false
	}

	if change.To == nil {
		return false
	}

	return true
}

func isDeleteDiffChange(change *diff.Change) bool {
	if change.Type != diff.DELETE {
		return false
	}

	if change.From == nil {
		return false
	}

	if change.To != nil {
		return false
	}

	return true
}

func (this *ConfigMngrT) CommitChangelog(changelog *diff.Changelog, candidateConfig *ygot.ValidatedGoStruct) error {
	var err error
	if changelog, err = extractCreateEthIntfParams(changelog); err != nil {
		return fmt.Errorf("Failed to extract new Ethernet interface parameters from changelog: %s", err)
	}

	if changelog, err = extractDeleteEthIntfParams(changelog); err != nil {
		return fmt.Errorf("Failed to extract new Ethernet interface parameters from changelog: %s", err)
	}

	diffChangelog := NewDiffChangelogMgmtT(changelog)
	currentDefaultConfigAction := this.getCurrentTransDefaultConfigAction()
	if change, exists := findDisallowedManagementTreeNodeDeleteOperation(diffChangelog); exists {
		return fmt.Errorf("Delete operation on tree node %q is disallowed", change.Path)
	}

	// Stub for marking processed change
	_, err = this.findTransDefaultConfigActionChange(diffChangelog)
	if err != nil {
		return err
	}

	configAction, err := this.findTransConfigActionChange(diffChangelog)
	if err != nil {
		return err
	}

	commitConfirmTimeout, err := this.findTransCommitConfirmTimeoutChange(diffChangelog)
	if err != nil {
		return err
	}

	if configAction == oc.OpenconfigManagement_TRANS_TYPE_UNSET {
		configAction = currentDefaultConfigAction
	}

	if configAction != oc.OpenconfigManagement_TRANS_TYPE_TRANS_DRY_RUN {
		if (configAction == oc.OpenconfigManagement_TRANS_TYPE_TRANS_CONFIRM) && this.isTransPending() {
			this.transConfirmationCancel()
			this.Confirm()
			return nil
		}

		if err = this.NewTransaction(); err != nil {
			log.Errorf("Failed to start new transaction")
			return err
		}
	} else {
		log.Infof("Dry running transaction")
		this.transConfigLookupTbl = this.configLookupTbl.makeCopy()
	}

	if err = this.parseChangelogAndConvertToCommands(diffChangelog); err != nil {
		this.DiscardOrFinishTrans()
		return err
	}

	rawRunningConfig, err := gnmi.ConvertYgotGoStructIntoJsonByteStream(this.runningConfig)
	if err != nil {
		return err
	}

	rawCandidateConfig, err := gnmi.ConvertYgotGoStructIntoJsonByteStream(*candidateConfig)
	if err != nil {
		return err
	}

	configJsonDiff, err := utils.GetJsonDiff(rawRunningConfig, rawCandidateConfig)
	if err != nil {
		return err
	}

	if configAction == oc.OpenconfigManagement_TRANS_TYPE_TRANS_COMMIT_CONFIRM {
		if err := this.CommitConfirm(); err != nil {
			return err
		}

		this.transConfirmationCandidateConfig = candidateConfig
		this.transConfirmationTimeoutCtx, this.transConfirmationCancel = context.WithCancel(context.Background())
		go this.startCountingForConfirmationTimeout(&this.transConfirmationTimeoutCtx, commitConfirmTimeout)

		return fmt.Errorf("\nWaiting %d seconds for confirmation changes\n%s",
			commitConfirmTimeout, configJsonDiff)
	}

	defer this.DiscardOrFinishTrans()
	if configAction != oc.OpenconfigManagement_TRANS_TYPE_TRANS_DRY_RUN {
		if err := this.Commit(); err != nil {
			log.Errorf("Failed to commit changes")
			return err
		}

		this.configLookupTbl = this.transConfigLookupTbl.makeCopy()
		this.DiscardOrFinishTrans()
		log.Infof("Save new config")
		return this.CommitCandidateConfig(candidateConfig)
	}

	// Deferred DiscardOrFinishTrans() will clean transConfigLookupTbl
	this.transConfigLookupTbl = nil
	// It is not really error, we just passing information that we have finished dry running with success
	return fmt.Errorf("\nDry running: requested changes are valid\n%s", configJsonDiff)
}

func (this *ConfigMngrT) parseChangelogAndConvertToCommands(diffChangelog *DiffChangelogMgmtT) error {
	var err error
	for {
		// Deletion section
		if err = this.processDeleteIpv4AddrEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processDeleteAccessVlanEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processDeleteNativeVlanEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processDeleteTrunkVlanEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processDeleteAggIntfMemberFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processDeleteAggIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processDeleteEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		// Set section
		if err = this.processSetPortBreakoutFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processSetPortBreakoutChanSpeedFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processSetEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processSetAggIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processSetAggIntfMemberFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processSetVlanModeEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processSetAccessVlanEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processSetNativeVlanEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processSetTrunkVlanEthIntfFromChangelog(diffChangelog); err != nil {
			return err
		}
		if err = this.processSetIpv4AddrEthIntfFromChangelog(diffChangelog); err != nil {
			return err
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

	return nil
}

func (this *ConfigMngrT) startCountingForConfirmationTimeout(ctx *context.Context, timeout uint16) {
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		if err := this.Rollback(); err != nil {
			log.Errorf("%s", err)
		} else {
			log.Infof("Rollback changes")
		}
	case <-(*ctx).Done():
		log.Infof("Cancelled counting for commit confirmation timeout")
	}
}
