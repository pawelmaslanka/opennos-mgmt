package command

import "fmt"

// NilCmdT is a stub of Command pattern interface. In itself, it defines the Nil Object Pattern
type NilCmdT struct {
	*commandT
}

// Execute is a stub method of Command pattern interface
func (c *NilCmdT) Execute() error {
	return nil
}

// Undo is a stub method of Command pattern interface
func (c *NilCmdT) Undo() error {
	return nil
}

// GetName is a stub method of Command pattern interface
func (c *NilCmdT) GetName() string {
	return "nil"
}

func (this *NilCmdT) Equals(other CommandI) bool {
	return this.GetName() == other.GetName()
}

// Append is not supported
func (this *NilCmdT) Append(other CommandI) (bool, error) {
	return false, fmt.Errorf("Unsupported")
}
