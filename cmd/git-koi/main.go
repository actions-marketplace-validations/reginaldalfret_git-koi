package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/reginaldalfret/git-koi/pkg/behavior"
	"github.com/reginaldalfret/git-koi/pkg/coat"
	"github.com/reginaldalfret/git-koi/pkg/collect"
	"github.com/reginaldalfret/git-koi/pkg/history"
	"github.com/reginaldalfret/git-koi/pkg/koi"
	"github.com/reginaldalfret/git-koi/pkg/motion"
	"github.com/reginaldalfret/git-koi/pkg/output"
	"github.com/reginaldalfret/git-koi/pkg/render"
)

const version = "1.0.2"

const usage = `Usage: git-koi collect LOGIN [--output FILE]
       git-koi render SNAPSHOT --output-dir DIR [--body BODY]
       git-koi --help | --version
Collect LOGIN's public GitHub snapshot as JSON through ` + "`gh api`" + `: public owned and
contributed repositories, and daily contributions from the account's creation
through yesterday (UTC), the last complete day.
Writes to standard output, or replaces FILE only after collection succeeds.
Requires an authenticated GitHub CLI; in GitHub Actions, set GH_TOKEN to github.token.
Render a collected SNAPSHOT into DIR: koi.gif swims the behavior chosen for its
observation date, koi.png is the still koi, and preview.html shows both on light and
dark backgrounds. BODY is white (the default), black, asagi, tea, gold, or platinum.
The files are replaced only after all of them render.
Errors go to stderr and exit 1; invalid usage exits 2.
Examples: git-koi collect octocat --output snapshot.json
          git-koi render snapshot.json --output-dir koi --body asagi`

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	switch args[0] {
	case "--help":
		fmt.Println(usage)
		os.Exit(0)
	case "--version":
		fmt.Printf("git-koi %s (Go)\n", version)
		os.Exit(0)
	case "collect":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(2)
		}
		if args[1] == "--help" {
			fmt.Println(usage)
			os.Exit(0)
		}
		login := args[1]
		var outputPath string
		if len(args) >= 4 && args[2] == "--output" {
			outputPath = args[3]
		} else if len(args) > 2 {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(2)
		}
		if err := runCollect(login, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "git-koi: %v\n", err)
			os.Exit(1)
		}
	case "render":
		if len(args) < 4 {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(2)
		}
		if args[1] == "--help" {
			fmt.Println(usage)
			os.Exit(0)
		}
		snapshotPath := args[1]
		var outputDir string
		body := coat.BodyWhite

		for i := 2; i < len(args); i++ {
			if args[i] == "--output-dir" && i+1 < len(args) {
				outputDir = args[i+1]
				i++
			} else if args[i] == "--body" && i+1 < len(args) {
				parsedBody, ok := coat.ParseKoiBody(args[i+1])
				if !ok {
					fmt.Fprintf(os.Stderr, "git-koi: unknown body %s\n", args[i+1])
					os.Exit(2)
				}
				body = parsedBody
				i++
			} else {
				fmt.Fprintln(os.Stderr, usage)
				os.Exit(2)
			}
		}

		if outputDir == "" {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(2)
		}

		if err := runRender(snapshotPath, outputDir, body); err != nil {
			fmt.Fprintf(os.Stderr, "git-koi: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
}

func runCollect(login, outputPath string) error {
	now := time.Now().UTC()
	today := history.NewDate(now.Year(), now.Month(), now.Day())

	client := collect.GhCli{}
	snapshot, err := collect.CollectAccountSnapshot(client, login, today)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if outputPath == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	return replaceFile(outputPath, data)
}

func runRender(snapshotPath, outputDir string, body coat.KoiBody) error {
	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		return fmt.Errorf("%s: %w", snapshotPath, err)
	}

	var snapshot history.AccountSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return fmt.Errorf("%s: %w", snapshotPath, err)
	}

	dev, err := koi.DeriveKoiDevelopment(&snapshot)
	if err != nil {
		return err
	}

	b := behavior.SelectBehavior(dev)

	frameCount := motion.FrameCount(b)
	frames := make([]*output.KoiFrame, frameCount)
	for i := uint32(0); i < frameCount; i++ {
		frames[i] = render.RenderSwimFrame(dev, body, b, i)
	}

	gifBytes, err := output.EncodeGIF(frames)
	if err != nil {
		return err
	}

	stillFrame := render.RenderKoi(dev, body)
	pngBytes, err := output.EncodePNG(stillFrame)
	if err != nil {
		return err
	}

	caption := fmt.Sprintf("%s for %s, %s body", b.String(), dev.ObservedOn.String(), body.String())
	previewHTML := output.EncodePreviewHTML(caption)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	gifPath := filepath.Join(outputDir, "koi.gif")
	pngPath := filepath.Join(outputDir, "koi.png")
	htmlPath := filepath.Join(outputDir, "preview.html")

	if err := replaceFile(gifPath, gifBytes); err != nil {
		return err
	}
	if err := replaceFile(pngPath, pngBytes); err != nil {
		return err
	}
	if err := replaceFile(htmlPath, []byte(previewHTML)); err != nil {
		return err
	}

	fmt.Printf("%s: wrote %s, %s, %s\n", caption, gifPath, pngPath, htmlPath)
	return nil
}

func replaceFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	tmpPath := filepath.Join(dir, base+".tmp")

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}
