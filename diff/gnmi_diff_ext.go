package diff

const (
	kBreakoutModeComponentPathElemSize = 5
)

func (ch *Change) isChangedBreakoutMode() bool {
	if len(ch.Path) < kBreakoutModeComponentPathElemSize {
		return false
	}

	ch.
	return true
}
