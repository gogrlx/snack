//go:build freebsd

package detect

import (
	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/pkg"
)

func candidates() []managerFactory {
	return []managerFactory{
		func() snack.Manager { return pkg.New() },
	}
}

func allManagers() []managerFactory {
	return candidates()
}
