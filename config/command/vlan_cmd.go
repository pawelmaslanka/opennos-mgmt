package command

import (
	"context"
	"fmt"
	mgmt "opennos-eth-switch-service/mgmt"
	"opennos-eth-switch-service/mgmt/interfaces"
	"opennos-eth-switch-service/mgmt/vlan"
	"opennos-mgmt/utils"
	"time"

	"github.com/r3labs/diff"
)

const (
	VlanEthIntfPathItemIdxC             = 0
	VlanEthIfnamePathItemIdxC           = 1
	VlanEthEthernetPathItemIdxC         = 2
	VlanEthSwVlanPathItemIdxC           = 3
	VlanEthVlanModePathItemIdxC         = 4
	VlanEthAccessVlanPathItemIdxC       = 4
	VlanEthNativeVlanPathItemIdxC       = 4
	VlanEthTrunkVlanPathItemIdxC        = 4
	VlanModeEthPathItemsCountC          = 5
	AccessVlanEthPathItemsCountC        = 5
	NativeVlanEthPathItemsCountC        = 5
	TrunkVlanEthIdxPathItemIdxC         = 5
	TrunkVlanEthPathItemsCountC         = 6
	TrunkVlanEthPathItemsCountIfUpdateC = 7

	VlanEthIntfPathItemC               = "Interface"
	VlanEthEthernetPathItemC           = "Ethernet"
	VlanEthSwVlanPathItemC             = "SwitchedVlan"
	VlanEthVlanModePathItemC           = "InterfaceMode"
	VlanEthAccessVlanPathItemC         = "AccessVlan"
	VlanEthNativeVlanPathItemC         = "NativeVlan"
	VlanEthTrunkVlanPathItemC          = "TrunkVlans"
	TrunkVlanEthValTypeUint16PathItemC = "Uint16"
	TrunkVlanEthValTypeStringPathItemC = "String"
)

const (
	vlanChangeIdxC = iota
	maxChangeVlanIdxC
)

// SetVlanModeEthIntfCmdT implements command for set VLAN mode for Ethernet Interface
type SetVlanModeEthIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetNativeVlanEthIntfCmdT creates new instance of SetVlanModeEthIntfCmdT type
func NewSetVlanModeEthIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetVlanModeEthIntfCmdT {
	changes := make([]*diff.Change, maxChangeVlanIdxC)
	changes[vlanChangeIdxC] = vlan
	return &SetVlanModeEthIntfCmdT{
		commandT: newCommandT("set vlan mode for ethernet interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and set VLAN mode for Ethernet interface
func (this *SetVlanModeEthIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	return this.doSetVlanModeCmd(shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetVlanModeEthIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	return this.doSetVlanModeCmd(shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetVlanModeEthIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetVlanModeEthIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetVlanModeEthIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// Append is not supported
func (this *SetVlanModeEthIntfCmdT) Append(cmd CommandI) (bool, error) {
	return false, fmt.Errorf("Unsupported")
}

func (this *commandT) doSetVlanModeCmd(shouldBeAbleOnlyToUndo bool) error {
	if this.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return this.createErrorAccordingToExecutionState()
	}

	this.dumpInternalData()

	// TODO: Not implemented yet

	this.finalize()
	return nil
}

// SetVlanCmdT implements command for create new VLAN
type SetVlanCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetVlanCmdT creates new instance of SetVlanCmdT type
func NewSetVlanCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetVlanCmdT {
	changes := make([]*diff.Change, maxChangeVlanIdxC)
	changes[vlanChangeIdxC] = vlan
	fmt.Printf("[CREATE] Changes:\n%v\n", changes[vlanChangeIdxC])
	return &SetVlanCmdT{
		commandT: newCommandT("set vlan", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and creates new VLAN
func (this *SetVlanCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doCreateOrDeleteVlanCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetVlanCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doCreateOrDeleteVlanCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetVlanCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetVlanCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetVlanCmdT)
	return this.equals(otherCmd.commandT)
}

// Append extracts internal data of 'other' and attach them to 'this'
func (this *SetVlanCmdT) Append(other CommandI) (bool, error) {
	return this.append(other)
}

// DeleteVlanCmdT implements command for delete VLAN
type DeleteVlanCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteVlanCmdT creates new instance of DeleteAccessVlanCmdT type
func NewDeleteVlanCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteVlanCmdT {
	changes := make([]*diff.Change, maxChangeVlanIdxC)
	changes[vlanChangeIdxC] = vlan
	return &DeleteVlanCmdT{
		commandT: newCommandT("delete vlan", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and deletes VLAN
func (this *DeleteVlanCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := true
	return doCreateOrDeleteVlanCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *DeleteVlanCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := false
	return doCreateOrDeleteVlanCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *DeleteVlanCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteVlanCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteVlanCmdT)
	return this.equals(otherCmd.commandT)
}

// Append extracts internal data of 'other' and attach them to 'this'
func (this *DeleteVlanCmdT) Append(other CommandI) (bool, error) {
	return this.append(other)
}

// SetAccessVlanEthIntfCmdT implements command for set access VLAN for Ethernet Interface
type SetAccessVlanEthIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetAccessVlanEthIntfCmdT creates new instance of SetAccessVlanEthIntfCmdT type
func NewSetAccessVlanEthIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetAccessVlanEthIntfCmdT {
	changes := make([]*diff.Change, maxChangeVlanIdxC)
	changes[vlanChangeIdxC] = vlan
	return &SetAccessVlanEthIntfCmdT{
		commandT: newCommandT("set access vlan for ethernet interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and set access VLAN for Ethernet interface
func (this *SetAccessVlanEthIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doVlanEthIntfCmd(this.commandT, vlan.Vlan_ACCESS, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetAccessVlanEthIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doVlanEthIntfCmd(this.commandT, vlan.Vlan_ACCESS, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetAccessVlanEthIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetAccessVlanEthIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetAccessVlanEthIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// Append extracts internal data of 'other' and attach them to 'this'
func (this *SetAccessVlanEthIntfCmdT) Append(other CommandI) (bool, error) {
	return this.append(other)
}

// DeleteAccessVlanEthIntfCmdT implements command for delete access VLAN from Ethernet Interface
type DeleteAccessVlanEthIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteAccessVlanEthIntfCmdT creates new instance of DeleteAccessVlanEthIntfCmdT type
func NewDeleteAccessVlanEthIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteAccessVlanEthIntfCmdT {
	changes := make([]*diff.Change, maxChangeVlanIdxC)
	changes[vlanChangeIdxC] = vlan
	return &DeleteAccessVlanEthIntfCmdT{
		commandT: newCommandT("delete access vlan from ethernet interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and deletes access VLAN from Ethernet interface
func (this *DeleteAccessVlanEthIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := true
	return doVlanEthIntfCmd(this.commandT, vlan.Vlan_ACCESS, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *DeleteAccessVlanEthIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := false
	return doVlanEthIntfCmd(this.commandT, vlan.Vlan_ACCESS, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *DeleteAccessVlanEthIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteAccessVlanEthIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteAccessVlanEthIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// Append extracts internal data of 'other' and attach them to 'this'
func (this *DeleteAccessVlanEthIntfCmdT) Append(other CommandI) (bool, error) {
	return this.append(other)
}

// SetNativeVlanEthIntfCmdT implements command for set native VLAN for Ethernet Interface
type SetNativeVlanEthIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetNativeVlanEthIntfCmdT creates new instance of SetNativeVlanEthIntfCmdT type
func NewSetNativeVlanEthIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetNativeVlanEthIntfCmdT {
	changes := make([]*diff.Change, maxChangeVlanIdxC)
	changes[vlanChangeIdxC] = vlan
	return &SetNativeVlanEthIntfCmdT{
		commandT: newCommandT("set native vlan for ethernet interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and deletes native VLAN from Ethernet interface
func (this *SetNativeVlanEthIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doVlanEthIntfCmd(this.commandT, vlan.Vlan_NATIVE, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetNativeVlanEthIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doVlanEthIntfCmd(this.commandT, vlan.Vlan_NATIVE, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetNativeVlanEthIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetNativeVlanEthIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetNativeVlanEthIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// Append extracts internal data of 'other' and attach them to 'this'
func (this *SetNativeVlanEthIntfCmdT) Append(other CommandI) (bool, error) {
	return this.append(other)
}

// DeleteNativeVlanEthIntfCmdT implements command for delete native VLAN from Ethernet Interface
type DeleteNativeVlanEthIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteNativeVlanEthIntfCmdT create new instance of DeleteNativeVlanEthIntfCmdT type
func NewDeleteNativeVlanEthIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteNativeVlanEthIntfCmdT {
	changes := make([]*diff.Change, maxChangeVlanIdxC)
	changes[vlanChangeIdxC] = vlan
	return &DeleteNativeVlanEthIntfCmdT{
		commandT: newCommandT("delete native vlan from ethernet interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and deletes native VLAN from Ethernet interface
func (this *DeleteNativeVlanEthIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := true
	return doVlanEthIntfCmd(this.commandT, vlan.Vlan_NATIVE, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *DeleteNativeVlanEthIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := false
	return doVlanEthIntfCmd(this.commandT, vlan.Vlan_NATIVE, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *DeleteNativeVlanEthIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteNativeVlanEthIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteNativeVlanEthIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// Append extracts internal data of 'other' and attach them to 'this'
func (this *DeleteNativeVlanEthIntfCmdT) Append(other CommandI) (bool, error) {
	return this.append(other)
}

// SetTrunkVlanEthIntfCmdT implements command for set trunk VLAN from Ethernet Interface
type SetTrunkVlanEthIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetTrunkVlanEthIntfCmdT creates new instance of SetTrunkVlanEthIntfCmdT type
func NewSetTrunkVlanEthIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetTrunkVlanEthIntfCmdT {
	changes := make([]*diff.Change, maxChangeVlanIdxC)
	changes[vlanChangeIdxC] = vlan
	return &SetTrunkVlanEthIntfCmdT{
		commandT: newCommandT("set trunk vlan for ethernet interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and set trunk VLAN for Ethernet interface
func (this *SetTrunkVlanEthIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doVlanEthIntfCmd(this.commandT, vlan.Vlan_TRUNK, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetTrunkVlanEthIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doVlanEthIntfCmd(this.commandT, vlan.Vlan_TRUNK, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetTrunkVlanEthIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetTrunkVlanEthIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetTrunkVlanEthIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// Append extracts internal data of 'other' and attach them to 'this'
func (this *SetTrunkVlanEthIntfCmdT) Append(other CommandI) (bool, error) {
	return this.append(other)
}

// DeleteTrunkVlanEthIntfCmdT implements command for delete trunk VLAN from Ethernet Interface
type DeleteTrunkVlanEthIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteTrunkVlanEthIntfCmdT creates new instance of DeleteTrunkVlanEthIntfCmdT type
func NewDeleteTrunkVlanEthIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteTrunkVlanEthIntfCmdT {
	changes := make([]*diff.Change, maxChangeVlanIdxC)
	changes[vlanChangeIdxC] = vlan
	return &DeleteTrunkVlanEthIntfCmdT{
		commandT: newCommandT("delete trunk vlan from ethernet interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and deletes trunk VLAN from Ethernet interface
func (this *DeleteTrunkVlanEthIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := true
	return doVlanEthIntfCmd(this.commandT, vlan.Vlan_TRUNK, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *DeleteTrunkVlanEthIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := false
	return doVlanEthIntfCmd(this.commandT, vlan.Vlan_TRUNK, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *DeleteTrunkVlanEthIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteTrunkVlanEthIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteTrunkVlanEthIntfCmdT)
	return this.equals(otherCmd.commandT)
}

// Append extracts internal data of 'other' and attach them to 'this'
func (this *DeleteTrunkVlanEthIntfCmdT) Append(other CommandI) (bool, error) {
	return this.append(other)
}

func doCreateOrDeleteVlanCmd(cmd *commandT, toBeDelete bool, shouldBeAbleOnlyToUndo bool) error {
	if cmd.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return cmd.createErrorAccordingToExecutionState()
	}

	cmd.dumpInternalData()

	var err error
	var vid uint16
	if toBeDelete {
		vid, err = utils.ConvertGoInterfaceIntoUint16(cmd.changes[0].From)
	} else {
		vid, err = utils.ConvertGoInterfaceIntoUint16(cmd.changes[0].To)
	}
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if toBeDelete {
		_, err = (*cmd.ethSwitchMgmt).DeleteVlan(ctx, &vlan.DeleteVlanRequest{
			Vlan: &vlan.Vlan{
				Vid: uint32(vid),
			},
		})
	} else {
		_, err = (*cmd.ethSwitchMgmt).CreateVlan(ctx, &vlan.CreateVlanRequest{
			Vlan: &vlan.Vlan{
				Vid: uint32(vid),
			},
		})
	}
	if err != nil {
		return err
	}

	cmd.finalize()
	return nil
}

func doVlanEthIntfCmd(cmd *commandT, mode vlan.Vlan_Mode, isDelete bool, shouldBeAbleOnlyToUndo bool) error {
	if cmd.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return cmd.createErrorAccordingToExecutionState()
	}

	cmd.dumpInternalData()

	var err error
	var vid uint16
	if isDelete {
		vid, err = utils.ConvertGoInterfaceIntoUint16(cmd.changes[0].From)
	} else {
		vid, err = utils.ConvertGoInterfaceIntoUint16(cmd.changes[0].To)
	}
	if err != nil {
		return err
	}

	ethIntfs := make([]*interfaces.EthernetIntf, len(cmd.changes))
	for i, change := range cmd.changes {
		ethIntfs[i] = &interfaces.EthernetIntf{
			Ifname: change.Path[VlanEthIfnamePathItemIdxC],
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if isDelete {
		_, err = (*cmd.ethSwitchMgmt).RemoveEthernetIntfFromVlan(ctx, &vlan.RemoveEthernetIntfFromVlanRequest{
			Vlan: &vlan.Vlan{
				Vid:  uint32(vid),
				Mode: mode,
			},
			EthIntfs: ethIntfs,
		})
	} else {
		_, err = (*cmd.ethSwitchMgmt).AddEthernetIntfToVlan(ctx, &vlan.AddEthernetIntfToVlanRequest{
			Vlan: &vlan.Vlan{
				Vid:  uint32(vid),
				Mode: mode,
			},
			EthIntfs: ethIntfs,
		})
	}
	if err != nil {
		return err
	}

	cmd.finalize()
	return nil
}
