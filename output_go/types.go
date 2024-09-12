package main


type Product struct {
	Rowid int
	Name string
}

type Product_info struct {
	Infotext string
}

type Product_price struct {
	Price string
}

type Users struct {
	Id int
	Name string
	Email string
	Password string
}

type Products struct {
	Rowid int
	Name string
	Infotext string
	Price string
}