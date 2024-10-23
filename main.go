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
	example     string
	description string

	trimPrefix string
	addPrefix  string
	transform  string

	templateFile string
)

var command = &cobra.Command{
	Use:   "enum schema",
	Short: "Generate schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			dir = "."
			g   Generator
		)

		if len(args) == 1 && isDirectory(args[0]) {
			dir = args[0]
		} else if len(args) > 0 {
			dir = filepath.Dir(args[0])
		}

		g.parsePackage(args, []string{})

		if err := g.generate(
			typeName, trimPrefix, addPrefix, transform,
			templateFile, lineComment,
		); err != nil {
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
	command.PersistentFlags().BoolVar(&lineComment, "linecomment", false, "line comment")
	command.PersistentFlags().StringVarP(&output, "output", "o", "schema.json", "output file")

	command.PersistentFlags().StringVar(&trimPrefix, "trimprefix", "", "trim prefix")
	command.PersistentFlags().StringVar(&addPrefix, "addprefix", "", "add prefix")
	command.PersistentFlags().StringVar(&transform, "transform", "", "transform")

	command.PersistentFlags().StringVarP(&templateFile, "template", "t", "", "template file")
}

func main() {
	if err := command.Execute(); err != nil {
		fmt.Println(err)
	}
}
