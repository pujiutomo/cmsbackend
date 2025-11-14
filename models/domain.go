package models

type Domain struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	MetaTitle   string `json:"meta_title"`
	MetaDesc    string `json:"meta_desc"`
	MetaKeyword string `json:"meta_keyword"`
	MetaIco     string `json:"meta_ico"`
	Modul       string `json:"modul"`
}
