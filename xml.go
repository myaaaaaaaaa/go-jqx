package jqx

import (
	"strings"

	"github.com/antchfx/xmlquery"
)

func xmlQueryPath(xmlString, xpath string) (string, error) {
	doc, err := xmlquery.Parse(strings.NewReader(xmlString))
	if err != nil {
		return "", err
	}

	nodes, err := xmlquery.QueryAll(doc, xpath)
	if err != nil {
		return "", err
	}

	if len(nodes) == 0 {
		return "", nil
	}

	return nodes[0].OutputXML(true), nil
}