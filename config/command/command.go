package command

import (
	"encoding/json"
	"fmt"
	"log"

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
	Equals(cmd CommandI) bool
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
	log.Printf("Dump internal data of command %q", cmd.name)
	log.Println("Change(s) to apply:")
	indent := "    "
	for _, ch := range cmd.changes {
		if jsonDump, err := json.MarshalIndent(ch, "", indent); err != nil {
			log.Printf("Failed to dump internal data of command %q: %s", cmd.name, err)
		} else {
			log.Printf("%s", string(jsonDump))
		}
	}
	log.Printf("Has been already executed: %v", cmd.hasBeenExecuted)
}

func (this *commandT) equals(other *commandT) bool {
	if this.name != other.name {
		return false
	} else if len(this.changes) != len(other.changes) {
		return false
	}

	for i, change := range this.changes {
		if change.Type != other.changes[i].Type {
			return false
		} else if len(change.Path) != len(other.changes[i].Path) {
			return false
		} else if change.From != other.changes[i].From {
			return false
		} else if change.To != other.changes[i].To {
			return false
		}

		for j, item := range change.Path {
			if item != other.changes[i].Path[j] {
				return false
			}
		}
	}

	return true
}
