package brew

import "github.com/gogrlx/snack"

// Compile-time interface checks.
var (
	_ snack.Manager        = (*Brew)(nil)
	_ snack.VersionQuerier = (*Brew)(nil)
	_ snack.Cleaner        = (*Brew)(nil)
	_ snack.FileOwner      = (*Brew)(nil)
	_ snack.NameNormalizer = (*Brew)(nil)
)
