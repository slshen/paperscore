package cmd

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/slshen/paperscore/pkg/game"
	"github.com/spf13/cobra"
)

func fmtCommand() *cobra.Command {
	var inplace bool
	c := &cobra.Command{
		Use:   "fmt",
		Short: "Format a game file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			newname := name
			g, err := game.ReadGameFile(name)
			if err != nil {
				return err
			}
			var out io.Writer
			if inplace {
				if path.Ext(name) == ".yaml" {
					newname = name[0:len(name)-5] + ".gm"
				}
				f, err := os.CreateTemp(path.Dir(name), fmt.Sprintf("%s*", path.Base(newname)))
				if err != nil {
					return err
				}
				out = f
			} else {
				out = cmd.OutOrStdout()
			}
			g.File.Write(out)
			if inplace {
				f := out.(*os.File)
				if err := f.Close(); err != nil {
					return err
				}
				if err := os.Rename(f.Name(), newname); err != nil {
					return err
				}
				if name != newname {
					return os.Remove(name)
				}
			}
			return nil
		},
	}
	c.Flags().BoolVarP(&inplace, "inplace", "i", false, "Convert file in place")
	return c
}
