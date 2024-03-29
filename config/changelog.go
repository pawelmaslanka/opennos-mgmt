package config

import "github.com/r3labs/diff"

type countT uint16

// TODO: Kazda komenda po przeprocesowaniu diff.Change ustawi na true pole 'processed'
type DiffChangeMgmtT struct {
	Change    *diff.Change
	processed bool
}

func NewDiffChangeMgmtT(change *diff.Change) *DiffChangeMgmtT {
	return &DiffChangeMgmtT{
		Change:    change,
		processed: false,
	}
}

func (this *DiffChangeMgmtT) MarkAsProcessed() {
	this.processed = true
}

func (this *DiffChangeMgmtT) IsProcessed() bool {
	return this.processed
}

type DiffChangelogMgmtT struct {
	Changes []*DiffChangeMgmtT
}

func NewDiffChangelogMgmtT(changelog *diff.Changelog) *DiffChangelogMgmtT {
	changes := *changelog
	var diffChangelog DiffChangelogMgmtT
	diffChangelog.Changes = make([]*DiffChangeMgmtT, len(changes))
	for i := 0; i < len(changes); i++ {
		diffChangelog.Changes[i] = NewDiffChangeMgmtT(&changes[i])
	}

	return &diffChangelog
}

func (this *DiffChangelogMgmtT) isProcessed() bool {
	var cnt countT = 0
	for _, change := range this.Changes {
		if change.IsProcessed() {
			cnt++
		}
	}

	return countT(len(this.Changes)) == cnt
}
