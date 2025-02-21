package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/golang/glog"
	"github.com/urfave/cli"

	"github.com/itomsawyer/gospy/pkg/proc"
	"github.com/itomsawyer/gospy/pkg/term"
)

func validPC(pc string) error {
	if pc != "current" && pc != "start" && pc != "caller" {
		return fmt.Errorf("Invalid pc type: %s", pc)
	}
	return nil
}

func main() {
	var bin string
	var pid int
	var refresh int
	var nonblocking bool
	var pcType string
	binFlag := cli.StringFlag{
		Name:        "bin",
		Usage:       "external binary with debug info",
		Destination: &bin,
	}
	pidFlag := cli.IntFlag{
		Name:        "pid",
		Usage:       "target go process id to spy",
		Required:    true,
		Destination: &pid,
	}
	refreshFlag := cli.IntFlag{
		Name:        "refresh",
		Usage:       "refresh interval in seconds",
		Value:       2,
		Destination: &refresh,
	}
	nonblockingFlag := cli.BoolFlag{
		Name:        "non-blocking",
		Usage:       "Don't suspend target process",
		Destination: &nonblocking,
	}
	pcFlag := cli.StringFlag{
		Name:        "pc",
		Usage:       "The program counter type: start, caller, current",
		Value:       "start",
		Destination: &pcType,
	}
	app := cli.NewApp()
	app.Name = "gospy"
	app.Usage = "inspect goroutines in non-invasive fashion"
	app.Commands = []cli.Command{
		{
			Name:    "summary",
			Aliases: []string{"s"},
			Usage:   "Dump go process internal summary",
			Flags:   []cli.Flag{binFlag, pidFlag, nonblockingFlag, pcFlag},
			Action: func(c *cli.Context) error {
				if err := validPC(pcType); err != nil {
					return err
				}
				p, err := proc.New(pid, bin)
				if err != nil {
					return err
				}
				sum, err := p.Summary(!nonblocking)
				if err != nil {
					return err
				}
				fmt.Println(sum)
				gs := sum.Gs
				sort.Slice(gs, func(i, j int) bool {
					return gs[i].ID < gs[j].ID
				})
				fmt.Print("goroutines:\n\n")
				for _, g := range gs {
					status := g.Status.String()
					if g.Waiting() {
						status = "waiting for " + g.WaitReason.String()
					}
					if g.M == nil {
						fmt.Printf("%d - %s: %s \n", g.ID, status, g.GetLocation(pcType).String())
					} else {
						fmt.Printf("%d(M%d)- %s: %s \n", g.ID, g.M.ID, status, g.GetLocation(pcType).String())
					}
				}
				return nil
			},
		},
		{
			Name:    "top",
			Aliases: []string{"t"},
			Usage:   "top like interface of executing functions",
			Flags:   []cli.Flag{binFlag, pidFlag, refreshFlag, nonblockingFlag, pcFlag},
			Action: func(c *cli.Context) error {
				if err := validPC(pcType); err != nil {
					return err
				}
				p, err := proc.New(pid, bin)
				if err != nil {
					return err
				}

				t := term.NewTerm(p, refresh, nonblocking, pcType)
				if err := t.Display(); err != nil {
					return err
				}
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		glog.Error(err)
		os.Exit(1)
	}
}
