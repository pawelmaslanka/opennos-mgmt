package command

import (
	"encoding/json"
	"fmt"

	log "github.com/golang/glog"

	mgmt "opennos-eth-switch-service/mgmt"

	"github.com/r3labs/diff"
)

// CommandI defines interface of the Command Design Pattern
type CommandI interface {
	// Execute runs action according to specific operation delivered by derived command
	Execute() error
	// Undo withdraws action performed by derived command executed in Execute() method
	Undo() error
	// GetName returns name of derived command
	GetName() string
	// EqualTo checks if 'this' command is equal to another 'cmd'
	Equals(other CommandI) bool
	// Append extracts internal data of 'other' and attach them to 'this'. After that, 'other'
	// should not be executed as a separate command due to its internal data are cleaned.
	// Returns tuple:
	// true, nil - if 'other' has been successfully appended;
	// false, nil - if 'other' has not been appended, because particular command does not support this capability;
	// false, error - if there was an error during appending 'other'
	Append(other CommandI) (bool, error)
}

// commandT is desired to embed in derivation type of Command pattern interface for use common
// data by all specific commands
type commandT struct {
	ethSwitchMgmt   *mgmt.EthSwitchMgmtClient
	name            string
	changes         []*diff.Change
	hasBeenExecuted bool
}

func newCommandT(name string, changes []*diff.Change, ethSwitchMgmt *mgmt.EthSwitchMgmtClient) *commandT {
	return &commandT{
		ethSwitchMgmt:   ethSwitchMgmt,
		name:            name,
		changes:         changes,
		hasBeenExecuted: false,
	}
}

// finalize is responsible for create or restore memento in order to use proper data during execute
// or undo command. Should be call before return from Execute() or Undo() method
func (cmd *commandT) finalize() {
	for _, ch := range cmd.changes {
		ch.From, ch.To = ch.To, ch.From
		if ch.Type == diff.DELETE {
			ch.Type = diff.CREATE
		} else if ch.Type == diff.CREATE {
			ch.Type = diff.DELETE
		} // else ch.Type == diff.UPDATE
	}

	cmd.hasBeenExecuted = !cmd.hasBeenExecuted
}

func (cmd *commandT) isAbleOnlyToUndo() bool {
	return cmd.hasBeenExecuted
}

func (cmd *commandT) createErrorAccordingToExecutionState() error {
	var fmtStr string
	if cmd.hasBeenExecuted {
		fmtStr = "Cannot execute command %q, because it has not been undo yet"
	} else {
		fmtStr = "Cannot undo command %q, because it has not been executed yet"
	}

	return fmt.Errorf(fmtStr, cmd.name)
}

func (cmd *commandT) dumpInternalData() {
	log.Infof("Dump internal data of command \n=== %q ===", cmd.name)
	log.Infoln("Change(s) to apply:")
	indent := "    "
	for _, ch := range cmd.changes {
		if jsonDump, err := json.MarshalIndent(ch, "", indent); err != nil {
			log.Infof("Failed to dump internal data of command %q: %s", cmd.name, err)
		} else {
			log.Infof("\n%s", string(jsonDump))
		}
	}
	log.Infof("Has been already executed: %v", cmd.hasBeenExecuted)
}

func (this *commandT) equals(other *commandT) bool {
	if this.name != other.name {
		return false
	} else if len(this.changes) != len(other.changes) {
		return false
	}

	changelog, err := diff.Diff(this.changes, other.changes)
	if err != nil {
		log.Errorf("Failed to get diff of two config objects: %s", err)
		return false
	}

	if len(changelog) != 0 {
		return false
	}

	return true
}

func (this *commandT) append(cmd CommandI) (bool, error) {
	other, err := getCommandT(cmd)
	if err != nil {
		return false, err
	}

	if this.name != other.name {
		return false, fmt.Errorf("Requested command to append %q is not the same to %q",
			other.name, this.name)
	}

	this.changes = append(this.changes, other.changes...)
	other.erase()
	return true, nil
}

func (this *commandT) erase() {
	this.ethSwitchMgmt = nil
	this.name = ""
	this.changes = nil
}

func getCommandT(cmd CommandI) (*commandT, error) {
	var cmdT *commandT
	switch v := cmd.(type) {
	case *NilCmdT:
		cmdT = NilCmdT(*v).commandT
	case *SetAggIntfCmdT:
		cmdT = SetAggIntfCmdT(*v).commandT
	case *DeleteAggIntfCmdT:
		cmdT = DeleteAggIntfCmdT(*v).commandT
	case *SetAggIntfMemberCmdT:
		cmdT = SetAggIntfMemberCmdT(*v).commandT
	case *DeleteAggIntfMemberCmdT:
		cmdT = DeleteAggIntfMemberCmdT(*v).commandT
	case *SetIpv4AddrEthIntfCmdT:
		cmdT = SetIpv4AddrEthIntfCmdT(*v).commandT
	case *DeleteIpv4AddrEthIntfCmdT:
		cmdT = DeleteIpv4AddrEthIntfCmdT(*v).commandT
	case *SetPortBreakoutCmdT:
		cmdT = SetPortBreakoutCmdT(*v).commandT
	case *SetPortBreakoutChanSpeedCmdT:
		cmdT = SetPortBreakoutChanSpeedCmdT(*v).commandT
	case *SetVlanModeEthIntfCmdT:
		cmdT = SetVlanModeEthIntfCmdT(*v).commandT
	case *SetAccessVlanEthIntfCmdT:
		cmdT = SetAccessVlanEthIntfCmdT(*v).commandT
	case *DeleteAccessVlanEthIntfCmdT:
		cmdT = DeleteAccessVlanEthIntfCmdT(*v).commandT
	case *SetNativeVlanEthIntfCmdT:
		cmdT = SetNativeVlanEthIntfCmdT(*v).commandT
	case *DeleteNativeVlanEthIntfCmdT:
		cmdT = DeleteNativeVlanEthIntfCmdT(*v).commandT
	case *SetTrunkVlanEthIntfCmdT:
		cmdT = SetTrunkVlanEthIntfCmdT(*v).commandT
	case *DeleteTrunkVlanEthIntfCmdT:
		cmdT = DeleteTrunkVlanEthIntfCmdT(*v).commandT
	default:
		return nil, fmt.Errorf("Cannot convert %v to any of known command, got: %T", v, v)
	}

	return cmdT, nil
}
