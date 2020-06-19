package command

import (
	"fmt"
	mgmt "opennos-eth-switch-service/mgmt"

	"github.com/r3labs/diff"
)

const (
	LacpInterfacePathItemIdxC = 0
	LacpIfnamePathItemIdxC    = 1
	LacpNamePathItemIdxC      = 2
	LacpPathItemsCountC       = 3

	LacpInterfacePathItemC = "Interface"
	LacpNamePathItemC      = "Name"
)

const (
	lacpChangeIdxC = iota
	maxLacpChangeIdxC
)

const (
	lacpMemberChangeIdxC = iota
	maxLacpMemberChangeIdxC
)

// SetLacpCmdT implements command for creating LACP
type SetLacpCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetLacpCmdT creates new instance of SetLacpCmdT type
func NewSetLacpCmdT(change *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetLacpCmdT {
	changes := make([]*diff.Change, maxLacpChangeIdxC)
	changes[lacpChangeIdxC] = change
	return &SetLacpCmdT{
		commandT: newCommandT("set lacp", changes, ethSwitchMgmt),
	}
}

// Execute implements the same mlacpod from CommandI interface and creates LACP
func (this *SetLacpCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doLacpCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same mlacpod from CommandI interface and withdraws changes performed by
// previously execution of Execute() mlacpod
func (this *SetLacpCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doLacpCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same mlacpod from CommandI interface and returns name of command
func (this *SetLacpCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetLacpCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetLacpCmdT)
	return this.equals(otherCmd.commandT)
}

// Append is not supported
func (this *SetLacpCmdT) Append(cmd CommandI) (bool, error) {
	return false, fmt.Errorf("Unsupported")
}

// SetLacpCmdT implements command for removing LACP
type DeleteLacpCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteLacpCmdT creates new instance of DeleteLacpCmdT type
func NewDeleteLacpCmdT(change *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteLacpCmdT {
	changes := make([]*diff.Change, maxLacpChangeIdxC)
	changes[lacpChangeIdxC] = change
	return &DeleteLacpCmdT{
		commandT: newCommandT("delete lacp", changes, ethSwitchMgmt),
	}
}

// Execute implements the same mlacpod from CommandI interface and creates LACP
func (this *DeleteLacpCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doLacpCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same mlacpod from CommandI interface and withdraws changes performed by
// previously execution of Execute() mlacpod
func (this *DeleteLacpCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doLacpCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same mlacpod from CommandI interface and returns name of command
func (this *DeleteLacpCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteLacpCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteLacpCmdT)
	return this.equals(otherCmd.commandT)
}

// Append is not supported
func (this *DeleteLacpCmdT) Append(cmd CommandI) (bool, error) {
	return false, fmt.Errorf("Unsupported")
}

// SetLacpMemberCmdT implements command for add members to LACP
type SetLacpMemberCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewSetLacpCmdT creates new instance of SetLacpMemberCmdT type
func NewSetLacpMemberCmdT(change *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *SetLacpMemberCmdT {
	changes := make([]*diff.Change, maxLacpMemberChangeIdxC)
	changes[lacpMemberChangeIdxC] = change
	return &SetLacpMemberCmdT{
		commandT: newCommandT("set lacp member", changes, ethSwitchMgmt),
	}
}

// Execute implements the same mlacpod from CommandI interface and creates LACP
func (this *SetLacpMemberCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doLacpMemberCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same mlacpod from CommandI interface and withdraws changes performed by
// previously execution of Execute() mlacpod
func (this *SetLacpMemberCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doLacpMemberCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same mlacpod from CommandI interface and returns name of command
func (this *SetLacpMemberCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *SetLacpMemberCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*SetLacpMemberCmdT)
	return this.equals(otherCmd.commandT)
}

// Append is not supported
func (this *SetLacpMemberCmdT) Append(cmd CommandI) (bool, error) {
	return false, fmt.Errorf("Unsupported")
}

// SetLacpMemberCmdT implements command for removing LACP
type DeleteLacpMemberCmdT struct {
	*commandT // commandT is embedded as a pointer because its state will be modify
}

// NewDeleteLacpMemberCmdT creates new instance of DeleteLacpCmdT type
func NewDeleteLacpMemberCmdT(change *diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *DeleteLacpMemberCmdT {
	changes := make([]*diff.Change, maxLacpMemberChangeIdxC)
	changes[lacpMemberChangeIdxC] = change
	return &DeleteLacpMemberCmdT{
		commandT: newCommandT("delete lacp member", changes, ethSwitchMgmt),
	}
}

// Execute implements the same mlacpod from CommandI interface and creates LACP
func (this *DeleteLacpMemberCmdT) Execute() error {
	shouldBeAbleOnlyToUndo := false
	isGoingToBeDeleted := false
	return doLacpCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// Undo implements the same mlacpod from CommandI interface and withdraws changes performed by
// previously execution of Execute() mlacpod
func (this *DeleteLacpMemberCmdT) Undo() error {
	shouldBeAbleOnlyToUndo := true
	isGoingToBeDeleted := true
	return doLacpCmd(this.commandT, isGoingToBeDeleted, shouldBeAbleOnlyToUndo)
}

// GetName implements the same mlacpod from CommandI interface and returns name of command
func (this *DeleteLacpMemberCmdT) GetName() string {
	return this.name
}

// Equals checks if 'this' command and 'other' command are the same... do the same thing
func (this *DeleteLacpMemberCmdT) Equals(other CommandI) bool {
	otherCmd := other.(*DeleteLacpMemberCmdT)
	return this.equals(otherCmd.commandT)
}

// Append is not supported
func (this *DeleteLacpMemberCmdT) Append(cmd CommandI) (bool, error) {
	return false, fmt.Errorf("Unsupported")
}

func doLacpCmd(cmd *commandT, isDelete bool, shouldBeAbleOnlyToUndo bool) error {
	if cmd.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return cmd.createErrorAccordingToExecutionState()
	}

	cmd.dumpInternalData()

	// var err error
	// var ifname string
	// if isDelete {
	// 	ifname, err = utils.ConvertGoInterfaceIntoString(cmd.changes[0].From)
	// } else {
	// 	ifname, err = utils.ConvertGoInterfaceIntoString(cmd.changes[0].To)
	// }
	// if err != nil {
	// 	return err
	// }

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	// if isDelete {
	// 	_, err = (*cmd.ethSwitchMgmt).DeleteLacp(ctx, &interfaces.DeleteLacpRequest{
	// 		Lacp: &interfaces.Lacp{
	// 			Ifname: ifname,
	// 		},
	// 	})
	// } else {
	// 	_, err = (*cmd.ethSwitchMgmt).CreateLacp(ctx, &interfaces.CreateLacpRequest{
	// 		Lacp: &interfaces.Lacp{
	// 			Ifname: ifname,
	// 		},
	// 	})
	// }
	// if err != nil {
	// 	return err
	// }

	// cmd.finalize()
	return nil
}

func doLacpMemberCmd(cmd *commandT, isDelete bool, shouldBeAbleOnlyToUndo bool) error {
	if cmd.isAbleOnlyToUndo() != shouldBeAbleOnlyToUndo {
		return cmd.createErrorAccordingToExecutionState()
	}

	cmd.dumpInternalData()

	// var err error
	// var ifname string
	// if isDelete {
	// 	ifname, err = utils.ConvertGoInterfaceIntoString(cmd.changes[0].From)
	// } else {
	// 	ifname, err = utils.ConvertGoInterfaceIntoString(cmd.changes[0].To)
	// }
	// if err != nil {
	// 	return err
	// }

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	// if isDelete {
	// 	_, err = (*cmd.ethSwitchMgmt).DeleteLacp(ctx, &interfaces.DeleteLacpRequest{
	// 		Lacp: &interfaces.Lacp{
	// 			Ifname: ifname,
	// 		},
	// 	})
	// } else {
	// 	_, err = (*cmd.ethSwitchMgmt).CreateLacp(ctx, &interfaces.CreateLacpRequest{
	// 		Lacp: &interfaces.Lacp{
	// 			Ifname: ifname,
	// 		},
	// 	})
	// }
	// if err != nil {
	// 	return err
	// }

	// cmd.finalize()
	return nil
}
