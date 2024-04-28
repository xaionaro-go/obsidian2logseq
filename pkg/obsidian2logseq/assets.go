package obsidian2logseq

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

func assetFilename(
	obsidianVaultName string,
	relPath string,
) string {
	relPath = strings.ReplaceAll(relPath, "Images"+string(filepath.Separator), "")
	hB := sha1.Sum([]byte(path.Dir(relPath)))
	return fmt.Sprintf(
		"obsidian_%s_%s_%s",
		obsidianVaultName, hex.EncodeToString(hB[:])[:8], strings.ReplaceAll(path.Base(relPath), " ", "_"),
	)
}
