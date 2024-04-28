package main

import (
	"context"
	"flag"

	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/facebookincubator/go-belt/tool/logger/implementation/logrus"
	"github.com/xaionaro-go/obsidian2logseq/pkg/obsidian2logseq"
)

func main() {
	obsidianDirIsPages := flag.Bool("obsidian-dir-is-pages", false, "The path to the Obsidian vault is actually the path to 'pages'")
	flag.Parse()
	obsidianVaultPath := flag.Arg(0)
	logseqGraphPath := flag.Arg(1)

	l := logrus.Default()
	ctx := logger.CtxWithLogger(context.Background(), l)

	if obsidianVaultPath == "" || logseqGraphPath == "" {
		panic("Usage: obsidian2logseq <Obsidian Vault> <Logseq Graph>")
	}

	err := obsidian2logseq.ConvertVault(
		obsidianVaultPath,
		logseqGraphPath,
		obsidian2logseq.OptionObsidianDirIsPages(*obsidianDirIsPages),
	)
	if err != nil {
		logger.Panic(ctx, err)
	}
}
