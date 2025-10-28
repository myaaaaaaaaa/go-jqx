package jqx

import (
	"strings"

	"github.com/antchfx/xmlquery"
)

func xmlQueryPath(xmlString, xpath string) ([]string, error) {
	doc, err := xmlquery.Parse(strings.NewReader(xmlString))
	if err != nil {
		return nil, err
	}

	nodes, err := xmlquery.QueryAll(doc, xpath)
	if err != nil {
		return nil, err
	}

	var rt []string
	for _, node := range nodes {
		rt = append(rt, node.OutputXML(true))
	}
	return rt, nil
}
