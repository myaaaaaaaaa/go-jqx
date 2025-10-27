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

	var sb strings.Builder
	for _, node := range nodes {
		sb.WriteString(node.OutputXML(true))
	}

	return sb.String(), nil
}