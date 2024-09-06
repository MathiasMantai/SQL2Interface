# SQL2Interface
This program converts SQL CREATE statements into Typescript interfaces

# Compile
```
go build -o s2i.exe
```


# Basic Usage
In order to use this program you have to specify an input and an output directoy via the i and o flags

## Example

In the directory <b>sql</b> we have a file called users.sql with the following create statement:
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

After executing the program with the following command:
```bash
.\s2i.exe -i '.\sql' -o '.\output'
```

We will get a file Users.ts in the directory <b>output</b> which will look like this:
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

# Config 
A yaml file called s2iconfig.yaml can be used to configure certain aspects of this program

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

# Combining multiple Tables into a single Interface
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

