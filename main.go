package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

//Define structs for program
type Data struct {
	Records []Record `xml:"doc"`
}

type Record struct {
	Fields []Field `xml:"field"`
}

type Field struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type GeneTissueExpression struct {
	NCBIGene string
	Organ    string
	Variance string
	RPKM     string
}

func main() {
	//Define default flag values and enable input from command line
	inputXML := flag.String("in", "", "Path of XML file to open, leave empty for STDIN")
	outputFile := flag.String("out", "", "Path of file to write, leave empty for STDOUT")
	geneID := flag.String("geneID", "", "NCBI GeneID to extract from XML")
	flag.Parse()

	data, err := parseXMLFileToData(*inputXML)
	if err != nil {
		log.Fatal("Could not parse XML file", err)
	}

	var gtes []GeneTissueExpression
	//Find data relevant for selected geneID
	if len(*geneID) > 0 {
		geneIDs := strings.Split(*geneID, ",")
		gtes = data.extractSingleGene(geneIDs)
	} else {
		gtes = data.extractAllGenes()
	}

	writeGeneTissueExpressionFile(*outputFile, gtes)
}

func parseXMLFileToData(file string) (*Data, error) {
	expressionFile := os.Stdin
	var err error
	// Open XML file for reading
	if len(file) > 0 {
		expressionFile, err = os.Open(file)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Fprintln(os.Stderr, "Expression file has opened successfully!")
		}

	}
	defer expressionFile.Close()

	// Create XML decoder
	xmlDecoder := xml.NewDecoder(expressionFile)

	//Parse XML file into the data struct variable
	var data Data
	err = xmlDecoder.Decode(&data)
	if err != nil {
		return &data, err
	}
	return &data, nil
}

func (data *Data) extractSingleGene(geneIDs []string) []GeneTissueExpression {
	geneSet := map[string]bool{}
	for _, geneID := range geneIDs {
		geneSet[geneID] = true
	}
	gtes := []GeneTissueExpression{}
	for _, record := range data.Records {
		for _, field := range record.Fields {
			if field.Name == "gene" {
				_, ok := geneSet[field.Value]
				if ok {
					gtes = append(gtes, record.extractGeneTissueExpression())

				}
			}
		}
	}
	return gtes
}

func (data *Data) extractAllGenes() []GeneTissueExpression {
	gtes := []GeneTissueExpression{}
	for _, record := range data.Records {
		gte := record.extractGeneTissueExpression()
		if gte.NCBIGene != "" {
			gtes = append(gtes, gte)
		}
	}
	return gtes
}

func (record Record) extractGeneTissueExpression() GeneTissueExpression {
	gte := GeneTissueExpression{}
	for _, field := range record.Fields {
		if field.Name == "gene" {
			gte.NCBIGene = field.Value
		}
		if field.Name == "source_name" {
			gte.Organ = field.Value
		}
		if field.Name == "var" {
			gte.Variance = field.Value
		}
		if field.Name == "full_rpkm" {
			gte.RPKM = field.Value
		}
	}
	return gte
}

func writeGeneTissueExpressionFile(outputFile string, gtes []GeneTissueExpression) {
	out := os.Stdout
	var err error
	if len(outputFile) > 0 {
		out, err = os.Create(outputFile)
		if err != nil {
			log.Fatal("Could not create outputfile: ", outputFile, "\n", err)
		}
	}
	defer out.Close()
	for _, gte := range gtes {
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", gte.NCBIGene, gte.Organ, gte.RPKM, gte.Variance)
	}
}

