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

# Ignore files and columns
A s2iconfig.yaml can be created to ignore specific files or columns within a file

## Example

This example will ignore the files product_prices.sql and warehouses.sql and will ignore the columns created_at and updated_at from users.sql

```
ignore_files:
 - product_prices.sql
 - warehouses.sql
ignore_columns:
  users.sql:
    - created_at
    - updated_at
```
