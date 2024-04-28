package obsidian2logseq

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type config struct {
	ObsidianDirIsPages bool
}

type Option interface {
	apply(*config)
}

type OptionObsidianDirIsPages bool

func (opt OptionObsidianDirIsPages) apply(cfg *config) {
	cfg.ObsidianDirIsPages = bool(opt)
}

type Options []Option

func (s Options) Config() config {
	cfg := config{}
	for _, opt := range s {
		opt.apply(&cfg)
	}
	return cfg
}

func ConvertVault(
	obsidianVaultPath string,
	logseqGraphPath string,
	opts ...Option,
) error {
	cfg := Options(opts).Config()

	obsidianVaultName := getVaultName(cfg, obsidianVaultPath)
	obsidianPagesPath := getObsidianPagesPath(cfg, obsidianVaultPath)

	logseqAssetsPath := path.Join(logseqGraphPath, "assets")
	logseqPagesPath := path.Join(logseqGraphPath, "pages")

	return filepath.Walk(obsidianPagesPath,
		func(srcPath string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("cannot walk through directories: %v %v %w", srcPath, info, err)
			}

			if info.IsDir() {
				if info.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}

			srcRelPath, err := getRelativePath(obsidianPagesPath, srcPath)
			if err != nil {
				return fmt.Errorf("unable to get a relative path of '%s': %w", srcPath, err)
			}

			switch {
			case strings.HasSuffix(srcRelPath, `.md`):

				b, err := os.ReadFile(srcPath)
				if err != nil {
					return fmt.Errorf("cannot read file '%s': %w", srcPath, err)
				}

				bC, err := ConvertMarkdown(obsidianPagesPath, srcPath, b, opts...)
				if err != nil {
					return fmt.Errorf("unable to convert the content of file '%s': %w", srcPath, err)
				}
				dstRelDir := strings.ReplaceAll(path.Dir(srcRelPath), " ", "_")
				dstDir := path.Join(logseqPagesPath, dstRelDir)
				err = os.MkdirAll(dstDir, 0750)
				if err != nil {
					return fmt.Errorf("unable to create dir '%s': %w", dstDir, err)
				}
				dstFilename := strings.Join(strings.Split(srcRelPath, string(filepath.Separator)), "___")
				dstPath := path.Join(dstDir, dstFilename)
				err = os.WriteFile(dstPath, bC, 0640)
				if err != nil {
					return fmt.Errorf("unable to write to file '%s'/'%s' (src rel path: '%s'): %w", dstDir, dstFilename, srcRelPath, err)
				}
			default:
				return func() error {
					srcFile, err := os.Open(srcPath)
					if err != nil {
						return fmt.Errorf("unable to open file '%s' for reading: %w", srcPath, err)
					}
					defer srcFile.Close()

					err = os.MkdirAll(logseqAssetsPath, 0750)
					if err != nil {
						return fmt.Errorf("unable to create dir '%s': %w", logseqAssetsPath, err)
					}
					dstPath := path.Join(
						logseqAssetsPath,
						assetFilename(obsidianVaultName, srcRelPath),
					)
					dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, 0640)
					if err != nil {
						return fmt.Errorf("unable to open file '%s' for writing: %w", dstPath, err)
					}
					defer dstFile.Close()

					_, err = io.Copy(dstFile, srcFile)
					if err != nil {
						return fmt.Errorf("unable to copy data from '%s' to '%s': %w", srcPath, dstPath, err)
					}
					return nil
				}()
			}

			return nil
		})
}
