package src

import (
	"errors"
	"fmt"
	"io/fs"
	// "path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SQL2Interface struct {
	Config   *Config
	Sql      SQL
	Combiner map[string][]Combiner
}

type SQL struct {
	FileName  string   `json:"file_name"`
	TableName string   `json:"table_name"`
	Columns   []Column `json:"columns"`
}

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ConvertedStructure struct {
	StructureDefinition map[string]string
	StructureNames      map[string][]string
}

// NewSQL2Interface initializes a new SQL2Interface instance with the provided configuration directory, source, and target.
// It loads the configuration and combiner settings, and returns a pointer to the new instance.
//
// confDir: The directory path where the configuration file is located.
// source: The source directory or file path for SQL files to be converted.
// target: The target directory where the converted interface files will be saved.
//
// Returns: A pointer to a new SQL2Interface instance.
func NewSQL2Interface(confDir string) *SQL2Interface {
	s2i := &SQL2Interface{}
	s2i.LoadConfig(confDir)
	s2i.LoadCombiner()
	return s2i
}

/* CONFIG */

// LoadConfig initializes and loads the configuration from the provided directory path.
//
// Parameters:
// confDir (string): The directory path where the configuration file is located.
//
// Returns:
// This function does not return any value. However, it sets the Config field of the SQL2Interface instance to the loaded configuration.
// If an error occurs during the loading process, it panics with the error message.
func (s2i *SQL2Interface) LoadConfig(confDir string) {
	// Load configuration from file
	conf, loadConfError := LoadConfig(confDir)

	if loadConfError != nil {
		panic(loadConfError)
	}

	s2i.Config = conf
}



// Convert processes a single SQL file and converts it into an interface or struct based on the configuration settings.
// It checks if the file should be ignored, retrieves the file content, parses the SQL, adds the parsed data to the combiner,
// and creates and saves the interface or struct files based on the configuration settings.
//
// Parameters:
// fileName (string): The name of the SQL file to be converted. If the source directory is a file, this parameter should be an empty string.
//
// Return:
// This function does not return any value. However, it prints error messages if any errors occur during the conversion process.
func (s2i *SQL2Interface) Convert(files []fs.DirEntry) {
	fmt.Println(files)
	var output ConvertedStructure
	output.StructureDefinition = make(map[string]string)
	output.StructureNames = make(map[string][]string)

	for _, file := range files {
		fileName := file.Name()
		fmt.Printf("=> attempting to convert %v\n", fileName)
		ignoreFiles := s2i.Config.IgnoreFiles
		sourceDir := s2i.Config.Input

		if IsFileIgnored(fileName, ignoreFiles) {
			continue
		}

		fileContent, getContentErr := GetFileContent(sourceDir, fileName)

		if getContentErr != nil {
			fmt.Println(getContentErr)
			continue
		}

		parsedData, err := s2i.ParseSQL("typescript", fileName, fileContent)

		if err != nil {
			fmt.Println(err)
			continue
		}

		addedToCombiner, index := s2i.AddToCombiner("typescript", parsedData)

		if addedToCombiner && index != -1 {
			convertSingleTable := s2i.ConvertSingleTable("typescript", index)
	
			if !convertSingleTable {
				fmt.Println("  => conversion will be skipped for type 'typescript' since convert_single_tables is set to false for this file...")
				continue
			}
		}

		output.StructureDefinition["typescript"] = output.StructureDefinition["typescript"] + "\n\n" + CreateInterface(parsedData)
		output.StructureNames["typescript"] = append(output.StructureNames["typescript"], parsedData.TableName)

		//convert to go struct
		parsedData, err = s2i.ParseSQL("go", fileName, fileContent)

		if err != nil {
			fmt.Println(err)
			continue
		}

		addedToCombiner, index = s2i.AddToCombiner("go", parsedData)

		if addedToCombiner && index != -1 {
			convertSingleTable := s2i.ConvertSingleTable("go", index)
	
			if !convertSingleTable {
				fmt.Println("  => conversion will be skipped since convert_single_tables is set to false for this file...")
				continue
			}
		}

		output.StructureDefinition["go"] = output.StructureDefinition["go"] + "\n\n" + CreateStruct(parsedData)
		output.StructureNames["go"] = append(output.StructureNames["go"], parsedData.TableName)

	}

	//handle conmbiners
	s2i.CombinerToStructure(&output)



	//get options for output
	targetDirTs, tsDirExists := s2i.Config.Output["typescript"]["output_dir"]
	targetFileNameTs, tsFileNameExists := s2i.Config.Output["typescript"]["output_file"]
	exportTypesTs, tsExportTypesExists := s2i.Config.Output["typescript"]["export_types"]

	targetDirGo, goDirExists := s2i.Config.Output["go"]["output_dir"]
	goPackageName, packageNameExists := s2i.Config.Output["go"]["package_name"]
	targetFileNameGo, goFileNameExists := s2i.Config.Output["go"]["output_file"]

	if tsDirExists && strings.TrimSpace(targetDirTs) != "" && tsFileNameExists && strings.TrimSpace(targetFileNameTs) != "" {
		content := output.StructureDefinition["typescript"]

		if exportTypesTs == "true" && tsExportTypesExists {
			s2i.AddInterfaceExports(&content, output.StructureNames["typescript"])
		}
		fmt.Println(content)
		SaveFile(targetDirTs, targetFileNameTs, content)
	}

	if goDirExists && strings.TrimSpace(targetDirGo) != "" && goFileNameExists && strings.TrimSpace(targetFileNameGo) != "" {
		content := output.StructureDefinition["go"]
		if goPackageName != "" && packageNameExists {
			content = fmt.Sprintf("package %v\n%v", goPackageName, content)
		}
		SaveFile(targetDirGo, targetFileNameGo, content)
	}
}



func (s2i *SQL2Interface) AddInterfaceExports(content *string, interfaceNames []string) {
	AddTabToSlice(&interfaceNames)
	*content = fmt.Sprintf("%v\n\nexport{\n%v\n}", *content, strings.Join(interfaceNames, ",\n"))
}



// ParseSQL parses a raw SQL string into a SQL struct and returns it along with any encountered error.
// It extracts the table name and column details from the raw SQL string and populates the SQL struct accordingly.
//
// Parameters:
// - fileName (string): The name of the SQL file being parsed.
// - rawSQL (string): The raw SQL string to be parsed.
//
// Return:
// - SQL: A SQL struct containing the parsed table name and column details.
// - error: An error encountered during the parsing process, or nil if no error occurred.
func (s2i *SQL2Interface) ParseSQL(definitionType string, fileName string, rawSQL string) (SQL, error) {
	var sql SQL

	rawSQL = strings.ToUpper(rawSQL)

	chunks := strings.Split(rawSQL, "(")

	rawTableName := strings.TrimSpace(chunks[0])
	rawColumns := (strings.Split(chunks[1], ")"))[0]
	columns, parseColumnsError := s2i.ParseRowColumnDefinitions(definitionType, fileName, rawColumns)

	if parseColumnsError != nil {
		return sql, parseColumnsError
	}
	sql.FileName = fileName
	sql.TableName = s2i.ParseRawTableName(rawTableName)
	sql.Columns = columns

	return sql, nil
}


// ParseRawTableName extracts and processes the table name from a raw SQL table definition.
// It removes SQL keywords like "CREATE", "TABLE", "IF", "NOT", and "EXISTS", and converts the remaining string to PascalCase.
//
// Parameters:
// - rawTableDefinition (string): The raw SQL table definition string.
//
// Return:
// - string: The processed table name in PascalCase.
func (s2i *SQL2Interface) ParseRawTableName(rawTableDefinition string) string {
	replacer := strings.NewReplacer("CREATE", "", "TABLE", "", "IF", "", "NOT", "", "EXISTS", "")
	caser := cases.Title(language.Und, cases.NoLower)
	return caser.String(strings.ToLower(strings.TrimSpace(replacer.Replace(rawTableDefinition))))
}


// ParseRowColumnDefinitions parses a raw SQL column definitions string into a slice of Column structs.
// It extracts column names and types from the raw string, applies type mapping, and ignores specified columns.
//
// Parameters:
// - fileName (string): The name of the SQL file being parsed. Used for logging and ignoring columns.
// - rawColumnDefinitions (string): The raw SQL column definitions string to be parsed.
//
// Return:
// - []Column: A slice of Column structs containing the parsed column names and types.
// - error: An error encountered during the parsing process, or nil if no error occurred.
func (s2i *SQL2Interface) ParseRowColumnDefinitions(definitionType string, fileName string, rawColumnDefinitions string) ([]Column, error) {
	var columns []Column

	columnDefinitions := strings.Split(rawColumnDefinitions, ",")

	for _, columnDefinition := range columnDefinitions {

		if strings.TrimSpace(columnDefinition) == "" {
			continue
		}

		columnDefinition = strings.TrimSpace(columnDefinition)
		chunks := strings.Split(columnDefinition, " ")

		caser := cases.Title(language.Und, cases.NoLower)
		columnName := caser.String(strings.ToLower(strings.TrimSpace(chunks[0])))
		columnType := strings.ToLower(strings.TrimSpace(TypeMapper(definitionType, chunks[1])))

		if definitionType == "typescript" {
			columnType = caser.String(columnType)
		}

		if IsColumnIgnored(fileName, columnName, s2i.Config.IgnoreColumns) {
			continue
		}

		columns = append(columns, Column{
			Name: columnName,
			Type: columnType,
		})
	}

	return columns, nil
}


// TypeMapper maps SQL column types to their corresponding TypeScript types.
// It takes a string representing an SQL column type as input and returns a string representing the corresponding TypeScript type.
//
// Parameters:
// - colType (string): The SQL column type to be mapped.
//
// Return:
// - string: The corresponding typescript or go type.
func TypeMapper(definitionType string, colType string) string {

	if strings.Contains(strings.ToUpper(colType), "VARCHAR") {
		colType = "VARCHAR"
	}

	if definitionType == "typescript" {
		switch strings.ToUpper(colType) {
		case "VARCHAR", "TEXT", "DATE", "DATETIME", "TIMESTAMP", "TIME", "YEAR", "ENUM":
			return "String"
		case "INT", "INTEGER", "SMALLINT", "TINYINT", "MEDIUMINT", "BIGINT", "DECIMAL", "NUMERIC", "FLOAT", "DOUBLE", "SERIAL":
			return "Number"
		case "BOOLEAN", "BOOL":
			return "Boolean"
		default:
			return "String"
		}
	} else if definitionType == "go" {
		switch strings.ToUpper(colType) {
		case "VARCHAR", "TEXT", "DATE", "DATETIME", "TIMESTAMP", "TIME", "YEAR", "ENUM":
			return "string"
		case "INT", "INTEGER", "SERIAL":
			return "int"
		case "BIGINT":
			return "int64"
		case "MEDIUMINT", "SMALLINT", "TINYINT":
			return "int32"
		case "FLOAT":
			return "float"
		case "DOUBLE":
			return "float64"
		case "BOOLEAN", "BOOL":
			return "bool"
		default:
			return "string"
		}
	}

	return "string"
}


/* CREATE STRUCTURES */

// CreateInterface generates a TypeScript interface based on the provided SQL table definition.
// The function iterates through the columns of the SQL table and constructs the interface fields.
//
// Parameters:
// - sql: A SQL struct containing the table name and column details.
//
// Return:
// - string: A string representing the TypeScript interface definition.
//
// The interface definition follows the format:
//
//	interface TableName {
//	    columnName: columnType,
//	    ...
//	}
func CreateInterface(sql SQL) string {
	interfaceFields := ""
	length := len(sql.Columns)
	for i, column := range sql.Columns {
		interfaceFields += fmt.Sprintf("\t%v: %v", column.Name, column.Type)

		if i < length-1 {
			interfaceFields += ", "
		}

		interfaceFields += "\r\n"
	}

	return fmt.Sprintf("interface %v {\n%v}", sql.TableName, interfaceFields)
}


// CreateStruct generates a Go struct based on the provided SQL table definition.
// It iterates through the columns of the SQL table and constructs the struct fields.
//
// Parameters:
// - sql: A SQL struct containing the table name and column details.
//
// Return:
// - string: A string representing the Go struct definition.
//
// The struct definition follows the format:
//
//	type TableName struct {
//	    columnName columnType
//	    ...
//	}
func CreateStruct(sql SQL) string {
	structFields := ""
	for _, column := range sql.Columns {
		structFields += fmt.Sprintf("\t%v %v", column.Name, column.Type)

		structFields += "\r\n"
	}
	return fmt.Sprintf("type %v struct {\n%v}", sql.TableName, structFields)
}

/* COMBINER */

type Combiner struct {
	Tables              []string `json:"tables"`
	Amount              int      `json:"amount"`
	InterfaceName       string   `json:"interface_name"`
	TableDefinitions    []SQL    `json:"table_definitions"`
	ConvertSingleTables bool     `json:"convert_single_tables"`
}


// LoadCombiner initializes and loads the combiner configuration from the SQL2Interface instance.
// It iterates through the combine_tables configuration and populates the Combiner slice with the parsed data.
// If the combine_tables configuration is not found or is empty, it prints a message and returns without any further action.
func (s2i *SQL2Interface) LoadCombiner() {
	combinerConf := s2i.Config.CombineTables

	// Handling of errors and edge cases
	if combinerConf == nil {
		fmt.Println("=> no combine_tables configuration found. skipping...")
		return
	}

	if len(combinerConf) == 0 {
		fmt.Println("=> no combine_tables configuration found. skipping...")
		return
	}

	s2i.Combiner = make(map[string][]Combiner)

	// Iterating through the combine_tables configuration and populating the Combiner slice
	for _, singleCombinerConf := range combinerConf {
		s2i.Combiner["typescript"] = append(s2i.Combiner["typescript"], Combiner{
			Tables:        singleCombinerConf.Tables,
			Amount:        len(singleCombinerConf.Tables),
			InterfaceName: singleCombinerConf.Name,
			ConvertSingleTables: singleCombinerConf.ConvertSingleTables,
		})
		s2i.Combiner["go"] = append(s2i.Combiner["go"], Combiner{
			Tables:        singleCombinerConf.Tables,
			Amount:        len(singleCombinerConf.Tables),
			InterfaceName: singleCombinerConf.Name,
			ConvertSingleTables: singleCombinerConf.ConvertSingleTables,
		})
	}
}

// AddToCombiner checks if the SQL table definition's file name is in the list of tables to combine.
// If it is, the function appends the table definition to the corresponding combiner and returns true along with the index of the combiner.
// If the file name is not found in any of the combiners, the function returns false and -1.
//
// Parameters:
// - definition (SQL): The SQL table definition to be added to the combiner.
//
// Return:
// - bool: Indicates whether the table definition was added to a combiner (true) or not (false).
// - int: The index of the combiner in the SQL2Interface's Combiner slice. If no combiner was found, it returns -1.
func (s2i *SQL2Interface) AddToCombiner(definitionType string, definition SQL) (bool, int) {
	for i, combiner := range s2i.Combiner[definitionType] {
		if ValueInSlice(interface{}(definition.FileName), StringToInterfaceSlice(combiner.Tables)) {
			s2i.Combiner[definitionType][i].TableDefinitions = append(s2i.Combiner[definitionType][i].TableDefinitions, definition)
			return true, i
		}
	}

	return false, -1
}

// ConvertSingleTable checks if the single table conversion is enabled for a combiner.
//
// Parameters:
// - index (int): The index of the combiner in the SQL2Interface's Combiner slice.
//
// Return:
// - bool: Indicates whether single table conversion is enabled for the combiner (true) or not (false).
func (s2i *SQL2Interface) ConvertSingleTable(definitionType string, index int) bool {
	return s2i.Combiner[definitionType][index].ConvertSingleTables
}

// CombineTables combines multiple SQL table definitions into a single SQL table definition.
// It takes an interface name and a variable number of SQL table definitions as input.
// The function returns a slice of Column structs representing the combined SQL table definition.
//
// Parameters:
// - interfaceName (string): The name of the interface that will be created for the combined table.
// - interfaceDefinitions (SQL)...: A variable number of SQL table definitions to be combined.
//
// Return:
// - []Column: A slice of Column structs representing the combined SQL table definition.
func CombineTables(interfaceName string, interfaceDefinitions ...SQL) []Column {
	var newSql []Column

	// Iterate through each SQL table definition and append its columns to the newSql slice.
	for _, definition := range interfaceDefinitions {
		newSql = append(newSql, definition.Columns...)
	}

	// Return the combined SQL table definition.
	return newSql
}


func (s2i *SQL2Interface) CombinerToStructure(output *ConvertedStructure) error {
	fmt.Println("=> attempting to convert combined tables to interfaces...")

	for outputType, combinersPerType := range s2i.Combiner {
		for _, singleCombiner := range combinersPerType {
			structureName := singleCombiner.InterfaceName
			tableDefinitions := singleCombiner.TableDefinitions
			combinedColumns := CombineTables(structureName, tableDefinitions...)

			newSQL := SQL{
				TableName: structureName,
                Columns:   combinedColumns,
			}
			if outputType == "typescript" {
				(*output).StructureDefinition["typescript"] = output.StructureDefinition["typescript"] + "\n\n" + CreateInterface(newSQL)
				output.StructureNames["typescript"] = append(output.StructureNames["typescript"], structureName)
			} else if outputType == "go" {
				(*output).StructureDefinition["go"] = output.StructureDefinition["go"] + "\n\n" + CreateStruct(newSQL)
				(*output).StructureNames["go"] = append(output.StructureNames["go"], structureName)		
			}
		}
	}

	return nil
}

/* VALIDATE */
func ValidateCreateStatement(statement string) error {
	statement = strings.ToUpper(statement)

	if !strings.Contains(statement, "CREATE") || !strings.Contains(statement, "TABLE") {
		return errors.New("invalid CREATE statement found")
	}

	return nil
}

func ValidateInputOutput(input string, output string) error {
	if input == "" || output == "" {
		return errors.New("input and output file paths cannot be empty")
	}

	return nil
}

/* MAIN */

// main method to run the program
// checks whether the source directory is a file and will only convert the file if true.
func (s2i *SQL2Interface) Run() {
	sourceDir := s2i.Config.Input
	files, err := GetFiles(sourceDir)

	if err != nil {
		fmt.Println(err)
	}

	s2i.Convert(files)
}
