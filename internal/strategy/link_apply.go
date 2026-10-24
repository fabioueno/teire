package strategy

import "github.com/fabioueno/teire/internal/fsx"

// ApplyLink carries out a link change. A replace first moves the old target
// into backups, a folder named for this apply.
func ApplyLink(c LinkChange, backups string) error {
	switch c.Action {
	case LinkCreate:
		return fsx.Symlink(c.Source, c.Path)
	case LinkReplace:
		if _, err := fsx.Backup(c.Path, backups); err != nil {
			return err
		}

		return fsx.Symlink(c.Source, c.Path)
	case LinkRemove:
		return fsx.RemoveLink(c.Path)
	case LinkUnchanged, LinkConflict, LinkRelease:
	}

	return nil
}
