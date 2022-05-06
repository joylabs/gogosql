package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/kr/pretty"
	"gopkg.in/yaml.v2"
)

type GoTypeConfig struct {
	Name      string
	GoPackage *string `yaml:",omitempty"`
}

type ColumnConfig struct {
	Name     string
	Type     GoTypeConfig
	Nullable bool    `yaml:",omitempty"`
	GoField  *string `yaml:",omitempty"`
	GoArg    *string `yaml:",omitempty"`
	GoColumn *string `yaml:",omitempty"`
}

type ColumnReference string

type TableConfig struct {
	ColumnIndexes map[ColumnReference]int `yaml:"-"`
	Name          string
	GoStruct      *string `yaml:",omitempty"`
	GoTable       *string `yaml:",omitempty"`
	GoTableVar    *string `yaml:",omitempty"`
	GoColumns     *string `yaml:",omitempty"`
	Columns       []*ColumnConfig
	PrimaryKey    []*ColumnReference
}

type SQLConfig struct {
	GoPackage *string `yaml:",omitempty"`
	Tables    []*TableConfig
}

type GoTableConfig struct {
	TableConfig
}

type GoPackageConfig struct {
	PackageName   string
	Imports       map[string]bool
	Tables        []GoTableConfig
	TablesVarName string
}

func (cfg *ColumnConfig) FillPackageConfig(gocfg *GoPackageConfig) {
	if cfg.Type.GoPackage != nil {
		gocfg.Imports[*cfg.Type.GoPackage] = true
	}
}

func (cfg *ColumnConfig) FillDefaults() {
	if cfg.GoField == nil {
		s := strcase.ToCamel(cfg.Name)
		cfg.GoField = &s
	}

	if cfg.GoArg == nil {
		s := strcase.ToLowerCamel(*cfg.GoField)
		cfg.GoArg = &s
	}

	if cfg.GoColumn == nil {
		s := strcase.ToCamel(cfg.Name)
		cfg.GoColumn = &s
	}
}

func (cfg *TableConfig) PKColumns() []*ColumnConfig {
	r := []*ColumnConfig{}
	for _, ref := range cfg.PrimaryKey {
		i := cfg.ColumnIndexes[*ref]
		c := cfg.Columns[i]
		r = append(r, c)
	}
	return r
}
func (cfg *TableConfig) Validate() error {
	cfg.ColumnIndexes = map[ColumnReference]int{}

	for i, col := range cfg.Columns {
		n := col.Name
		cfg.ColumnIndexes[ColumnReference(n)] = i
	}

	if len(cfg.PrimaryKey) == 0 {
		return fmt.Errorf("each table must have at least one column to define the primary key, violating table: %v", cfg.Name)
	}

	for _, ref := range cfg.PrimaryKey {
		_, ok := cfg.ColumnIndexes[*ref]
		if !ok {
			return fmt.Errorf("PrimaryKey definition expects column named: %v, but no column with that name is defined", *ref)
		}
	}

	return nil
}

func (cfg *TableConfig) FillPackageConfig(gocfg *GoPackageConfig) {
	gocfg.Tables = append(gocfg.Tables, GoTableConfig{*cfg})

	for _, col := range cfg.Columns {
		col.FillPackageConfig(gocfg)
	}
}

func (cfg *TableConfig) FillDefaults() {
	if cfg.GoStruct == nil {
		s := strcase.ToCamel(pl.Singular(cfg.Name))
		cfg.GoStruct = &s
	}

	if cfg.GoTableVar == nil {
		s := strcase.ToCamel(cfg.Name)
		cfg.GoTableVar = &s
	}

	if cfg.GoTable == nil {
		s := "GoGo" + *cfg.GoTableVar + "Table"
		cfg.GoTable = &s
	}

	if cfg.GoColumns == nil {
		s := "GoGo" + *cfg.GoTableVar + "Columns"
		cfg.GoColumns = &s
	}

	for _, col := range cfg.Columns {
		col.FillDefaults()
	}
}

func (cfg *SQLConfig) Validate() error {
	for _, tbl := range cfg.Tables {
		err := tbl.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (cfg *SQLConfig) FillPackageConfig(gocfg *GoPackageConfig) {
	gocfg.PackageName = *cfg.GoPackage
	gocfg.TablesVarName = "Tables"

	for _, tbl := range cfg.Tables {
		tbl.FillPackageConfig(gocfg)
	}
}

func (cfg *SQLConfig) FillDefaults() {
	if cfg.GoPackage == nil {
		s := "gogosql"
		cfg.GoPackage = &s
	}

	for _, tbl := range cfg.Tables {
		tbl.FillDefaults()
	}
}

func BuildPackageConfig(sql *SQLConfig) GoPackageConfig {
	gocfg := GoPackageConfig{Imports: map[string]bool{}}
	sql.FillPackageConfig(&gocfg)
	return gocfg
}

var (
	pl           *pluralize.Client
	templatePath string
)

func main() {
	var isVerbose bool
	flag.BoolVar(&isVerbose, "verbose", false, "verbose to include more logging")

	flag.Parse()

	args := flag.Args()
	log.Printf("args: %v", args)
	if len(args) != 1 {
		panic(fmt.Errorf("expected 1 arg to config file, found: %v args", len(args)))
	}

	templatePath = "templates/gogosql.go.tmpl"
	pl = pluralize.NewClient()

	yamlBytes, err := os.ReadFile(args[0])
	if err != nil {
		log.Fatal(err)
	}

	cfg := &SQLConfig{}
	err = yaml.Unmarshal(yamlBytes, cfg)
	if err != nil {
		log.Fatal(err)
	}
	if isVerbose {
		fmt.Println(">>> Raw Config: ")
		pretty.Println(cfg)
	}

	err = cfg.Validate()
	if err != nil {
		log.Printf(">>> Error: %v\n", err)
		log.Println()

		log.Println("Parsed Config: ")
		pretty.Println(cfg)
		return
	}

	cfg.FillDefaults()
	if isVerbose {
		fmt.Println(">>> Filled Config: ")
		pretty.Println(cfg)
	}

	gocfg := BuildPackageConfig(cfg)

	if isVerbose {
		fmt.Println(">>> Package Config: ")
		pretty.Println(gocfg)
	}

	templateName := path.Base(templatePath)
	temp, err := template.New(templateName).Funcs(template.FuncMap{
		"Deref": func(i *ColumnReference) ColumnReference { return *i },
	}).ParseFiles(templatePath)
	if err != nil {
		log.Printf(">>> Error: %v\n", err)
		log.Println()

		log.Println("Package Config: ")
		pretty.Println(gocfg)
		os.Exit(-1)
	}

	file1, err := ioutil.TempFile("", templateName)
	if err != nil {
		log.Fatalln("Unable to create a tempory file")
	}
	defer file1.Close()

	err = temp.Execute(file1, gocfg)
	if err != nil {
		log.Fatalln(err)
	}

	path := file1.Name()
	file2, err := os.Open(path)
	if err != nil {
		log.Fatalf("Unable to reopen temporary file for reading")
	}
	defer file2.Close()

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "", file2, parser.AllErrors)
	if err != nil {
		log.Printf("Generated go code has an error: %s", path)
		log.Println(err)

		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			log.Println(err)
			os.Exit(-1)
		}
		log.Println("\n" + string(bytes))
		os.Exit(-1)
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read a tempory file before moving to generated code location: %s\n", path)
	}

	log.Println(string(bytes))
}
