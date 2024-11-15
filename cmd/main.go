package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"text/template"

	"github.com/justmumu/burptomap"
	"github.com/urfave/cli/v2"
)

//go:embed run_sqlmap.sh
var SQLMAP_RUNNER string

var EXCLUDED_EXTENTIONS = []string{"js", "css", "gif", "jpg", "png"}

func main() {
	app := &cli.App{
		Name:  "BurpToMap",
		Usage: "Simple Burp Request to SQLMap compatible request conversion",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "burp-file",
				Aliases: []string{"bf"},
				Usage:   "Load burp requests from xml `FILE`",
			},
			&cli.StringFlag{
				Name:     "reqs-dir",
				Aliases:  []string{"rd"},
				Usage:    "Modified requests' export `DIRECTORY`",
				Value:    "./requests",
				Required: false,
			},
			&cli.StringFlag{
				Name:    "sqlmap-dir",
				Aliases: []string{"sd"},
				Usage:   "SQLMap output directory `DIRECTORY`",
				Value:   "./outputs",
			},
		},
		Action: run,
		Authors: []*cli.Author{{
			Name: "justmumu (https://github.com/justmumu)",
		}},
		HelpName: "burptomap",
		CustomAppHelpTemplate: `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   Pass SQLMap commands after "-" sign. (Ex: burptomap --bf test.xml - --header "Cookie: test=1;" --force-ssl --batch)
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
`,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(cCtx *cli.Context) error {
	burpFile := cCtx.String("burp-file")
	requestsFolder := cCtx.String("reqs-dir")
	outputsFolder := cCtx.String("sqlmap-dir")

	// Clear and create requests and outputs folders
	for _, folder := range []string{requestsFolder, outputsFolder} {
		fInfo, err := os.Stat(folder)
		if err == nil {
			if !fInfo.IsDir() {
				return fmt.Errorf("%s is a file path not directory", folder)
			}
			if err := os.RemoveAll(folder); err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		os.Mkdir(folder, os.ModePerm)
	}

	if err := exportRequests(burpFile, requestsFolder); err != nil {
		return err
	}

	if err := runSqlmap(cCtx); err != nil {
		return err
	}

	return nil
}

func exportRequests(burpFile string, requestsFolder string) error {
	// Read Burpfile
	root, err := burptomap.UnmarshalFile(burpFile)
	if err != nil {
		return err
	}

	// Eliminate the same requests
	eliminatedItems, err := burptomap.Eliminate(root)
	if err != nil {
		return err
	}

	// Mark them with asteriks
	counter := 0
	for _, item := range eliminatedItems {
		// Pass excluded extentions
		if slices.Contains(EXCLUDED_EXTENTIONS, item.Extension) {
			continue
		}

		iCount, marked, err := burptomap.MarkAllInjectionPoints(item.Request.Value)
		if err != nil {
			return err
		}

		if iCount < 1 {
			continue
		}

		counter += 1

		os.WriteFile(filepath.Join(requestsFolder, fmt.Sprintf("request-%d.req", counter)), []byte(marked), 0644)
	}
	return nil
}

func runSqlmap(cCtx *cli.Context) error {
	requestsFolder := cCtx.String("reqs-dir")
	outputsFolder := cCtx.String("sqlmap-dir")
	sqlMapCommands := ""
	if cCtx.Args().Len() > 0 {
		for i, arg := range cCtx.Args().Slice()[1:] {
			if regexp.MustCompile(`\s`).MatchString(arg) {
				sqlMapCommands += fmt.Sprintf("\"%s\"", arg)
			} else {
				sqlMapCommands += arg
			}
			if i+2 < cCtx.Args().Len() {
				sqlMapCommands += " "
			}
		}
	}

	templateInput := map[string]string{
		"REQS_FOLDER":     requestsFolder,
		"OUTPUTS_FOLDER":  outputsFolder,
		"SQLMAP_COMMANDS": sqlMapCommands,
	}

	shScriptTmpl, err := template.New("script").Parse(SQLMAP_RUNNER)
	if err != nil {
		return err
	}

	script := bytes.NewBuffer(nil)
	shScriptTmpl.Execute(script, templateInput)

	scriptFile, err := os.CreateTemp("/tmp", "sqlmap.*.sh")
	if err != nil {
		return err
	}
	defer os.Remove(scriptFile.Name())

	if err := scriptFile.Chmod(0744); err != nil {
		return err
	}

	if _, err := scriptFile.Write(script.Bytes()); err != nil {
		return err
	}

	if err := scriptFile.Close(); err != nil {
		return err
	}

	fmt.Println(script.String())

	c := exec.Cmd{
		Path:   scriptFile.Name(),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := c.Start(); err != nil {
		return err
	}

	if err = c.Wait(); err != nil {
		return err
	}

	return nil
}
