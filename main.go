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

	xGoType            string
	xGoTypeImportPath  string
	xGoTypeImportName  string
	xGoTypeSkipPointer bool
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

		if err := g.generate(
			typeName, example, description,
			lineComment, indent,
			xGoType, xGoTypeImportPath, xGoTypeImportName, xGoTypeSkipPointer,
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
	command.PersistentFlags().BoolVar(&lineComment, "line-comment", false, "line comment")
	command.PersistentFlags().BoolVar(&indent, "indent", false, "indent")
	command.PersistentFlags().StringVar(&output, "output", "schema.json", "output file")
	command.PersistentFlags().StringVar(&example, "example", "", "example")
	command.PersistentFlags().StringVar(&description, "description", "", "description")

	command.PersistentFlags().StringVar(&xGoType, "x-go-type", "", "x-go-type")
	command.PersistentFlags().StringVar(&xGoTypeImportPath, "x-go-type-import-path", "", "x-go-type-import-path")
	command.PersistentFlags().StringVar(&xGoTypeImportName, "x-go-type-import-name", "", "x-go-type-import-name")
	command.PersistentFlags().BoolVar(&xGoTypeSkipPointer, "x-go-type-skip-optional-pointer", false, "x-go-type-skip-pointer")
}

func main() {
	if err := command.Execute(); err != nil {
		fmt.Println(err)
	}
}
