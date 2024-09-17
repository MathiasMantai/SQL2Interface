package src

import (
	"errors"
	"fmt"
	"io/fs"
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

		parsedDataTs, err := s2i.ParseSQL("typescript", fileName, fileContent)
		s2i.AddArbitraryFields(&parsedDataTs, "typescript")
		if err != nil {
			fmt.Println("x> error parsing sql data for type 'typescript: " + err.Error())
			continue
		}

		addedToCombinerTs, indexTs := s2i.AddToCombiner("typescript", parsedDataTs)

		//convert to go struct
		parsedDataGo, err := s2i.ParseSQL("go", fileName, fileContent)
		s2i.AddArbitraryFields(&parsedDataGo, "go")
		if err != nil {
			fmt.Println(fmt.Errorf("x> error parsing sql data for type 'go': %v", err.Error()))
			continue
		}

		addedToCombinerGo, indexGo := s2i.AddToCombiner("go", parsedDataGo)

		if addedToCombinerGo && indexGo != -1 && addedToCombinerTs && indexTs != -1 {
			convertSingleTable := s2i.ConvertSingleTable("typescript", indexTs)

			if !convertSingleTable {
				fmt.Println("  => conversion will be skipped since convert_single_tables is set to false for this file...")
				continue
			}
		}

		//add converted structures to outpui
		output.StructureDefinition["typescript"] = output.StructureDefinition["typescript"] + "\n\n" + CreateInterface(parsedDataTs)
		output.StructureNames["typescript"] = append(output.StructureNames["typescript"], parsedDataTs.TableName)

		output.StructureDefinition["go"] = output.StructureDefinition["go"] + "\n\n" + CreateStruct(parsedDataGo)
		output.StructureNames["go"] = append(output.StructureNames["go"], parsedDataGo.TableName)

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

// AddInterfaceExports adds export statements for the given interface names to the content string.
// It first adds a tab to each interface name in the slice.
// Then it formats the content string with the export statements for the interface names.
//
// Parameters:
// - content: A pointer to a string representing the content where the export statements will be added.
// - interfaceNames: A slice of strings representing the names of the interfaces to be exported.
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
			return colType
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
		case "FLOAT", "DECIMAL":
			return "float32"
		case "DOUBLE":
			return "float64"
		case "BOOLEAN", "BOOL":
			return "bool"
		default:
			return colType
		}
	}

	return colType
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
			Tables:              singleCombinerConf.Tables,
			Amount:              len(singleCombinerConf.Tables),
			InterfaceName:       singleCombinerConf.Name,
			ConvertSingleTables: singleCombinerConf.ConvertSingleTables,
		})
		s2i.Combiner["go"] = append(s2i.Combiner["go"], Combiner{
			Tables:              singleCombinerConf.Tables,
			Amount:              len(singleCombinerConf.Tables),
			InterfaceName:       singleCombinerConf.Name,
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
	inCombiner := false
	combinerIndex := -1
	for i, combiner := range s2i.Combiner[definitionType] {
		if ValueInSlice(interface{}(definition.FileName), StringToInterfaceSlice(combiner.Tables)) {
			s2i.Combiner[definitionType][i].TableDefinitions = append(s2i.Combiner[definitionType][i].TableDefinitions, definition)
			inCombiner = true
			combinerIndex = i
		}
	}

	return inCombiner, combinerIndex
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

// CombinerToStructure converts combined SQL table definitions to interfaces or structs.
// It iterates through the combiners for each output type (TypeScript and Go) and combines the table definitions.
// For each combiner, it creates a new SQL table definition with the combined columns and appends it to the output structure.
// If the output type is TypeScript, it adds the interface definition to the TypeScript section of the output structure.
// If the output type is Go, it adds the struct definition to the Go section of the output structure.
// Finally, it returns nil to indicate successful execution.
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
	fmt.Println("    => conversion successful")
	return nil
}

/* VALIDATE */
// ValidateCreateStatement checks if the given SQL statement is a valid CREATE TABLE statement.
//
// Parameters:
// - statement: A string representing the SQL statement to be validated.
//
// Return:
// - error: An error indicating an invalid CREATE TABLE statement if the statement is not valid.
//          Returns nil if the statement is valid.
func ValidateCreateStatement(statement string) error {
	statement = strings.ToUpper(statement)

	// Check if the statement contains both "CREATE" and "TABLE" keywords.
	if !strings.Contains(statement, "CREATE") || !strings.Contains(statement, "TABLE") {
		return errors.New("invalid CREATE statement found")
	}

	return nil
}

// ValidateInputOutput checks if the input and output file paths are not empty.
//
// Parameters:
// - input: A string representing the input file path.
// - output: A string representing the output file path.
//
// Return:
//   - error: An error indicating that either the input or output file path is empty.
//     Returns nil if both paths are not empty.
func ValidateInputOutput(input string, output string) error {
	if input == "" || output == "" {
		return errors.New("input and output file paths cannot be empty")
	}

	return nil
}

// AddArbitraryFields adds arbitrary fields to the SQL table definition based on the configuration.
// It iterates through the arbitrary fields specified in the configuration and adds them to the SQL table definition.
//
// Parameters:
// - sql: A pointer to the SQL struct representing the table definition to which arbitrary fields will be added.
//
// The function checks if the file name of the SQL table matches the file name specified in the configuration.
// If a match is found, it prints a message indicating the addition of the arbitrary field and appends the field to the SQL table definition.
func (s2i *SQL2Interface) AddArbitraryFields(sql *SQL, definitionType string) {
	for fileName, fields := range s2i.Config.ArbitraryFields {
		for _, value := range fields {
			if fileName == sql.FileName {
				var newCol Column
				newCol.Name = value.Name

				if definitionType == "typescript" {
					newCol.Type = value.TypeTs
				} else {
					newCol.Type = value.TypeGo
				}
				if strings.TrimSpace(newCol.Type) == "" || strings.TrimSpace(newCol.Name) == "" {
					continue
				}
				fmt.Printf("  => adding arbitrary field: %v : %v for type %v\n", newCol.Name, newCol.Type, definitionType)
				sql.Columns = append(sql.Columns, newCol)
			}
		}
	}
}

/* MAIN */

// Run starts the conversion process for SQL files to TypeScript and Go interfaces/structs.
// It retrieves the list of SQL files from the input directory specified in the configuration.
// If any error occurs during file retrieval, it prints the error message.
// Finally, it calls the Convert function to perform the actual conversion.
func (s2i *SQL2Interface) Run() {
	sourceDir := s2i.Config.Input
	files, err := GetFiles(sourceDir)

	if err != nil {
		fmt.Println(err)
	}

	s2i.Convert(files)
}
