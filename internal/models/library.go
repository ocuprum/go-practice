package models

import (
	"encoding/xml"
)

type Author struct {
	Name string
	Year int `xml:"born"`
}

type Book struct {
	Name string
	Authors []Author
	Year int `json:"published"`
	Read bool
	Comments map[string]string `json:"Comments,omitempty" xml:"-"`
}

type BookList struct {
	XMLName xml.Name `xml:"myshelf"`
	Items   []Book  `xml:"item"`
}

var Books = []Book{
	{Name: "Mama Pijama", Authors: []Author{{Name: "Frank Galagher", Year: 500}}, Year: 1900, Read: true, 
	Comments: map[string]string{
		"Vova": "Cool book", 
		"Sasha": "Awful",
	}},
	{Name: "FlowerGirl", Authors: []Author{{Name: "Frank", Year: 400},
										   {Name: "Lisy", Year: 2000}, 
										   {Name: "Aimee", Year: 1000}}, Year: 2000, Read: true},
	{Name: "FlowerBoy", Authors: []Author{{Name: "Frank", Year: 400},
										   {Name: "Lisy", Year: 2000}, 
										   {Name: "Aimee", Year: 1000}}, Year: 2000, Read: true}}

var BookListS = BookList{Items: Books}
