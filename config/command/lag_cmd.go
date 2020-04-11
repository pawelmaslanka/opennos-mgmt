package command

import (
	mgmt "opennos-eth-switch-service/mgmt"

	"github.com/r3labs/diff"
)

const (
	LagIntfInterfacePathItemIdxC      = 0
	LagIntfIfnamePathItemIdxC         = 1
	LagIntfNamePathItemIdxC           = 2
	LagIntfPathItemsCountC            = 3
	LagIntfMemberEthernetPathItemIdxC = 2
	LagIntfMemberAggIdPathItemIdxC    = 3
	LagIntfMemberPathItemsCountC      = 4

	LagIntfInterfacePathItemC      = "Interface"
	LagIntfMemberEthernetPathItemC = "Ethernet"
	LagIntfMemberAggIdPathItemC    = "AggregateId"
	LagIntfNamePathItemC           = "Name"
)

const (
	lagChangeIdxC = iota
	maxLagChangeIdxC
)

// DeleteLagIntfCmdT implements command for LAG interface
type DeleteLagIntfCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteLagIntfCmdT creates new instance of DeleteLagIntfCmdT type
func NewDeleteLagIntfCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteLagIntfCmdT {
	changes := make([]*diff.Change, maxLagChangeIdxC)
	changes[lagChangeIdxC] = vlan
	return &DeleteLagIntfCmdT{
		commandT: newCommandT("delete lag interface", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and deletes LAG interface
func (this *DeleteLagIntfCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	return this.doDeleteLagIntfCmd(shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *DeleteLagIntfCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	return this.doDeleteLagIntfCmd(shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *DeleteLagIntfCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteLagIntfCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteLagIntfCmdT)
	return this.equals(otherCmd.commandT)
}

func (this *commandT) doDeleteLagIntfCmd(shouldBeAbleOnlyToUndo bool) error {
	if this.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return this.createErrorAccordingToExecutionState()
	}

	this.dumpInternalData()
	// TODO: Make implementations

	this.finalize()
	return nil
}

// SetLagIntfMemberCmdT implements command for add Ethernet interface to LAG
type SetLagIntfMemberCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetLagIntfMemberCmdT creates new instance of SetLagIntfMemberCmdT type
func NewSetLagIntfMemberCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetLagIntfMemberCmdT {
	changes := make([]*diff.Change, maxLagChangeIdxC)
	changes[lagChangeIdxC] = vlan
	return &SetLagIntfMemberCmdT{
		commandT: newCommandT("set lag interface member", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and adds Ethernet interface to LAG
func (this *SetLagIntfMemberCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	return this.doSetLagIntfMemberCmd(shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *SetLagIntfMemberCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	return this.doSetLagIntfMemberCmd(shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *SetLagIntfMemberCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetLagIntfMemberCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetLagIntfMemberCmdT)
	return this.equals(otherCmd.commandT)
}

func (this *commandT) doSetLagIntfMemberCmd(shouldBeAbleOnlyToUndo bool) error {
	if this.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return this.createErrorAccordingToExecutionState()
	}

	this.dumpInternalData()
	// TODO: Make implementations

	this.finalize()
	return nil
}

// DeleteLagIntfMemberCmdT implements command for remove Ethernet interface from LAG
type DeleteLagIntfMemberCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteLagIntfMemberCmdT creates new instance of DeleteLagIntfMemberCmdT type
func NewDeleteLagIntfMemberCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteLagIntfMemberCmdT {
	changes := make([]*diff.Change, maxLagChangeIdxC)
	changes[lagChangeIdxC] = vlan
	return &DeleteLagIntfMemberCmdT{
		commandT: newCommandT("delete lag interface member", changes, ethSwitchMgmt),
	}
}

// Execute implements the same method from CommandI interface and removes Ethernet interface from LAG
func (this *DeleteLagIntfMemberCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	return this.doDeleteLagIntfMemberCmd(shouldBeAbleOnlyToUndo)
}

// Undo implements the same method from CommandI interface and withdraws changes performed by
// previously execution of Execute() method
func (this *DeleteLagIntfMemberCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	return this.doDeleteLagIntfMemberCmd(shouldBeAbleOnlyToUndo)
}

// GetName implements the same method from CommandI interface and returns name of command
func (this *DeleteLagIntfMemberCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteLagIntfMemberCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteLagIntfMemberCmdT)
	return this.equals(otherCmd.commandT)
}

func (this *commandT) doDeleteLagIntfMemberCmd(shouldBeAbleOnlyToUndo bool) error {
	if this.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return this.createErrorAccordingToExecutionState()
	}

	this.dumpInternalData()
	// TODO: Make implementations

	this.finalize()
	return nil
}
