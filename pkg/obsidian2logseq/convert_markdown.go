package obsidian2logseq

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type fix struct {
	Name string
	Expr *regexp.Regexp
	To   string
}

var fixes = []fix{
	{
		Name: "numerated lists",
		Expr: regexp.MustCompile(`(?m)^([ ]*)([0-9]+\.)`),
		To:   `$1- $2`,
	},
	{
		Name: "titles",
		Expr: regexp.MustCompile(`(?m)^([ ]*)(#+ .*)`),
		To:   `$1- $2`,
	},
	/*
		{
			Name: "everything is a point: shift existing points to the right",
			Expr: regexp.MustCompile(`(?m)^([ ]*)- `),
			To:   `$1- `,
		},

		{
			Name: "everything is a point: add dashes to the root points",
			Expr: regexp.MustCompile(`(?m)^([ ]*)([^\ \-\|])`),
			To:   `$1- $2`,
		},

		{
			Name: "everything is a point: add dashes to the root points",
			Expr: regexp.MustCompile(`(?m)^([ ]*)\|`),
			To:   `$1  |`,
		},
	*/
}

var chapterExpr = regexp.MustCompile("(?m)^([ ]*)(#+) (.*)((\n[^#].+)*)")
var contentLinkExpr = regexp.MustCompile(`!\[\[([^\]]+)\]\]`)

func ConvertMarkdown(
	vaultDir string,
	filePath string,
	in []byte,
	opts ...Option,
) ([]byte, error) {
	cfg := Options(opts).Config()
	relPath, err := getRelativePath(vaultDir, filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get a relative path of '%s': %w", filePath, err)
	}
	obsidianVaultName := getVaultName(cfg, vaultDir)

	result := bytes.ReplaceAll(append(in, []byte("\n#")...), []byte("\t"), []byte("  "))

	result = chapterExpr.ReplaceAllFunc(result, func(b []byte) []byte {
		matches := chapterExpr.FindAllSubmatch(b, 1)
		spaces := matches[0][1]
		_ = spaces
		hashes := matches[0][2]
		title := matches[0][3]
		content := matches[0][4]
		//panic(fmt.Sprintf("<%s><%s><%s><%s><%s>", spaces, hashes, title, content, tail))

		var output, indent bytes.Buffer
		for i := 0; i < len(hashes)-1; i++ {
			indent.WriteString("  ")
		}
		output.Write(indent.Bytes())
		output.Write(hashes)
		indent.WriteString("  ")
		output.WriteRune(' ')
		output.Write(title)
		lnWithIdent := append([]byte("\n"), indent.Bytes()...)
		content = bytes.ReplaceAll(content, []byte("\n"), lnWithIdent)
		/*if bytes.Contains(result, []byte("The American Immigration Council")) {
			fmt.Printf("<%s<%s<%s<%s>>>>(%s)", spaces, hashes, title, content)
		}*/
		output.Write(content)
		return output.Bytes()
	})

	for _, fix := range fixes {
		result = fix.Expr.ReplaceAll(result, []byte(fix.To))
	}

	result = contentLinkExpr.ReplaceAllFunc(result, func(s []byte) []byte {
		assetSrcFilename := s[3 : len(s)-2]
		pathPaths := []string{".."}
		depth := len(strings.Split(path.Dir(relPath), string(filepath.Separator)))
		for i := 0; i < depth; i++ {
			pathPaths = append(pathPaths, "..")
		}
		pathPaths = append(pathPaths, "assets")
		assetDir := path.Join(pathPaths...)
		assetRelSrcPath := path.Join(path.Dir(relPath), string(assetSrcFilename))
		assetFilename := assetFilename(obsidianVaultName, assetRelSrcPath)
		resultPath := path.Join(assetDir, assetFilename)
		return []byte(`![` + string(assetSrcFilename) + `](` + string(resultPath) + `)`)
	})

	return result, nil
}

func getVaultName(
	cfg config,
	obsidianVaultPath string,
) string {
	if cfg.ObsidianDirIsPages {
		return "unknown"
	}

	return path.Base(obsidianVaultPath)
}

func getObsidianPagesPath(
	cfg config,
	obsidianVaultPath string,
) string {
	if cfg.ObsidianDirIsPages {
		return obsidianVaultPath
	}

	return path.Join(obsidianVaultPath, "pages")
}

func getRelativePath(
	obsidianPagesPath string,
	obsidianFilePath string,
) (string, error) {
	if !strings.HasPrefix(obsidianFilePath, obsidianPagesPath) {
		return "", fmt.Errorf("got a path outside of the vault path '%s': '%s'", obsidianPagesPath, obsidianFilePath)
	}

	return filepath.Rel(obsidianPagesPath, obsidianFilePath)
}
