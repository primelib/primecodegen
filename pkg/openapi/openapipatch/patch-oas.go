package openapipatch

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func FixOAS300Version(doc *libopenapi.DocumentModel[v3.Document]) error {
	doc.Model.Version = "3.0.0"
	return nil
}

func FixOAS301Version(doc *libopenapi.DocumentModel[v3.Document]) error {
	doc.Model.Version = "3.0.1"
	return nil
}

func FixOAS302Version(doc *libopenapi.DocumentModel[v3.Document]) error {
	doc.Model.Version = "3.0.2"
	return nil
}

func FixOAS303Version(doc *libopenapi.DocumentModel[v3.Document]) error {
	doc.Model.Version = "3.0.3"
	return nil
}

func FixOAS304Version(doc *libopenapi.DocumentModel[v3.Document]) error {
	doc.Model.Version = "3.0.4"
	return nil
}

func FixOAS310Version(doc *libopenapi.DocumentModel[v3.Document]) error {
	doc.Model.Version = "3.1.0"
	return nil
}

func FixOAS311Version(doc *libopenapi.DocumentModel[v3.Document]) error {
	doc.Model.Version = "3.1.1"
	return nil
}
