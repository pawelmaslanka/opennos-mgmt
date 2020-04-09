package command

import (
	mgmt "opennos-eth-switch-service/mgmt"

	"github.com/r3labs/diff"
)

const (
	LagIntfMemberIntfPathItemIdxC     = 0
	LagIntfMemberIfnamePathItemIdxC   = 1
	LagIntfMemberEthernetPathItemIdxC = 2
	LagIntfMemberAggIdPathItemIdxC    = 3
	LagIntfMemberPathItemsCountC      = 4

	LagIntfMemberIntfPathItemC     = "Interface"
	LagIntfMemberEthernetPathItemC = "Ethernet"
	LagIntfMemberAggIdPathItemC    = "AggregateId"
)

const (
	lagMemberChangeIdxC = iota
	maxLagMemberChangeIdxC
)

// SetLagIntfMemberCmdT implements command for add Ethernet interface to LAG
type SetLagIntfMemberCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetLagIntfMemberCmdT creates new instance of SetLagIntfMemberCmdT type
func NewSetLagIntfMemberCmdT(vlan *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetLagIntfMemberCmdT {
	changes := make([]*diff.Change, maxLagMemberChangeIdxC)
	changes[lagMemberChangeIdxC] = vlan
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
