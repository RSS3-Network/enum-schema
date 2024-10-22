package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	typeName    string
	lineComment bool
	indent      bool
	output      string
)

var command = &cobra.Command{
	Use:   "enum schema",
	Short: "Generate schema",
	RunE: func(cmd *cobra.Command, args []string) error {

		var (
			dir string
			g   Generator
		)

		if len(args) == 1 && isDirectory(args[0]) {
			dir = args[0]
		} else {
			dir = filepath.Dir(args[0])
		}

		g.parsePackage(args, []string{})

		if err := g.generate(typeName, lineComment, indent); err != nil {
			return err
		}

		tmpFile, err := os.CreateTemp(dir, fmt.Sprintf("%s-schema-enum.json", typeName))
		defer tmpFile.Close()
		if err != nil {
			return err
		}

		_, err = tmpFile.Write(g.buf.Bytes())
		if err != nil {
			os.Remove(tmpFile.Name())
			return err
		}

		return os.Rename(tmpFile.Name(), output)
	},
}

func init() {
	command.PersistentFlags().StringVar(&typeName, "type", "", "type name")
	command.PersistentFlags().BoolVar(&lineComment, "line-comment", false, "line comment")
	command.PersistentFlags().BoolVar(&indent, "indent", false, "indent")
	command.PersistentFlags().StringVar(&output, "output", "schema.json", "output file")
}

func main() {
	if err := command.Execute(); err != nil {
		fmt.Println(err)
	}
}
