package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-clix/cli"
	neat "github.com/itaysk/kubectl-neat/cmd"
)

func main() {
	log.SetFlags(0)

	cmd := cli.Command{
		Use:   "kubectl-neat-diff [file1] [file2]",
		Short: "Remove fields from kubectl diff that carry low / no information (`KUBE_NEAT_DIFF_OPTS` is diff options, default is `-uN`)",
		Args:  cli.ArgsExact(2),
	}

	var diffOpts []string
	diffOpts = strings.Fields(os.Getenv("KUBE_NEAT_DIFF_OPTS"))
	if len(diffOpts) == 0 {
		diffOpts = []string{"-uN"}
	}

	cmd.Run = func(cmd *cli.Command, args []string) error {
		if err := neatifyDir(args[0]); err != nil {
			return err
		}
		if err := neatifyDir(args[1]); err != nil {
			return err
		}

		diffOpts = append(diffOpts, args[0])
		diffOpts = append(diffOpts, args[1])
		c := exec.Command("diff", diffOpts...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	err := cmd.Execute()

	// diff command always return exit 1, when diff is succeeded
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			// this is just an exit code error, no worries
			// do nothing

		default: //couldn't run diff
			log.Fatalln("Error:", err)
		}
	}
}

func neatifyDir(dir string) error {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		filename := filepath.Join(dir, fi.Name())
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		n, err := neat.NeatYAMLOrJSON(data)
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(filename, []byte(n), fi.Mode()); err != nil {
			return err
		}
	}

	return nil
}
