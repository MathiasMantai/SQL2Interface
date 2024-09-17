# SQL2Interface
This program converts SQL CREATE statements into Typescript interfaces

# Compile
```
go build -o s2i.exe
```


# Basic Usage
In order to use this program you have to specify an input and an output directoy via the i and o flags

## Example

In the input directory (specified in the config file) we have a file called users.sql with the following create statement:
```sql
CREATE TABLE users (
    id INT PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
)
```

We will get a file (which should be secified in the config file) with the following:
```ts
interface Users {
    Id: Number, 
    Name: String,
    Email: String, 
    Password: String, 
    Created_at: String, 
    Updated_at: String
}
```

```go
type Users struct {
	Id int
	Name string
	Email string
	Password string
}
```

# Config 
A yaml file called s2iconfig.yaml can be used to configure certain aspects of this program

# Specify input and output
The input and output locations have to be specified. For output, the directories can be different

## Example
```yaml
input: "C:\\Users\\<user>\\Desktop\\SQL2Interface\\sql"
output:
  typescript: 
    output_dir: "C:\\Users\\<user>\\Desktop\\SQL2Interface\\output"
    output_file: "Types.ts"
    export_types: true
  go: 
    output_dir: "C:\\Users\\<user>\\Desktop\\SQL2Interface\\output_go"
    output_file: "types.go"
    export_types: true
    package_name: "main"
```

# Ignore files and columns
Specific files or columns per file can be ignored

## Example

```yaml
ignore_files:
 - product_prices.sql
 - warehouses.sql
ignore_columns:
  users.sql:
    - created_at
    - updated_at
```

This will completely ignore the files product_prices.sql and warehouses.sql. From the file users.sql the columns created_at and updated_at will be ignored

# Combining multiple Tables into a single Interface/Struct
Multiple Tables can be combined into a single interface.
<b>Note:</b> This works together with ignore_files and ignore_columns.
Make sure the files that you want to combined are not included in ignore_files.
If you want to ignore speciifc columns from tables before you combine them, you can use ignore_columns for this

## Example

```yaml
ignore_files:
 - warehouses.sql
ignore_columns:
  product_prices.sql:
    - rowid
    - fk_product
combine_tables:
  Products:
    name: 'Products'
    tables: ['product.sql', 'product_price.sql']
    convert_single_tables: false
```

This will combine the tables from product.sql and product_price.sql into an interface Products.
The columns rowid and fk_product will be ignored from product_prices.sql for this conversion.

