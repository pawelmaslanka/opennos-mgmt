package command

// NilCmdT is a stub of Command pattern interface. In itself, it defines the Nil Object Pattern
type NilCmdT struct {
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

func (this *NilCmdT) Equals(cmd CommandI) bool {
	return this.GetName() == cmd.GetName()
}
