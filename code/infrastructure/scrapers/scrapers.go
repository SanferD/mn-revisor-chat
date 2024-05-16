package scrapers

import (
	"code/core"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const (
	tableOfChaptersHeadingContents  = "Table of Chapters"
	subTableOfChaptersHeadingPrefix = "Table of Chapters, "
	subdivisionPrefix               = "Subdivision "
	subdPrefix                      = "Subd. "
	repealedPrefix                  = "[Repealed,"
)

type Scraper struct{}

func InitializeScraper() (*Scraper, error) {
	return &Scraper{}, nil
}

func (scraper *Scraper) GetPageKind(contents io.Reader) (core.MNRevisorPageKind, error) {
	doc, err := htmlquery.Parse(contents)
	if err != nil {
		return core.MNRevisorPageKindError, fmt.Errorf("error on parsing html: %v", err)
	}
	// chapters table/subtable
	tocHeading := htmlquery.FindOne(doc, tableOfChaptersH2XPath)
	if tocHeading != nil {
		headingStr := strings.TrimSpace(htmlquery.InnerText(tocHeading))
		if headingStr == tableOfChaptersHeadingContents { // heading == Table of Chapters
			return core.StatutesChaptersTable, nil
		} else if strings.HasPrefix(headingStr, subTableOfChaptersHeadingPrefix) { // heading == Table of Chapters, 1 - 2A
			return core.StatutesChaptersShortTable, nil
		} else { // did not forsee this
			return core.MNRevisorPageKindError, errors.New("could not determine page kind from page header")
		}
	}
	// sections list
	sectionsListHeading := htmlquery.FindOne(doc, sectionsListH2XPath)
	if sectionsListHeading != nil {
		return core.StatutesSectionsTable, nil
	}
	// statutes
	statutesHeading := htmlquery.FindOne(doc, sectionDivXPath)
	if statutesHeading != nil {
		return core.Statutes, nil
	}
	return core.MNRevisorPageKindError, errors.New("cound not determine page kind")
}

func (scraper *Scraper) ExtractURLs(contents io.Reader, pageKind core.MNRevisorPageKind) ([]string, error) {
	var xpath, identifier string
	switch pageKind {
	case core.StatutesChaptersTable:
		xpath = tableXPath
		identifier = "toc_table"
	case core.StatutesChaptersShortTable:
		xpath = shortTableXPath
		identifier = "chapters_table"
	case core.StatutesSectionsTable:
		xpath = sectionsTableXPath
		identifier = "chapters_analysis"
	default:
		return nil, errors.New("error on extracting urls")
	}
	return scraper.extractURLsFromTableXPath(contents, xpath, identifier)
}

func (scraper *Scraper) extractURLsFromTableXPath(contents io.Reader, xpath string, errIdentifier string) ([]string, error) {
	doc, err := htmlquery.Parse(contents)
	if err != nil {
		return nil, fmt.Errorf("error on parsing html: %v", err)
	}
	rowNodes := htmlquery.Find(doc, xpath)
	if len(rowNodes) == 0 {
		return nil, fmt.Errorf("could not find '%s' table rows for chapters table", errIdentifier)
	}

	var urls = make([]string, 0)
	for _, rowNode := range rowNodes {

		// confirm that the chapter is valid and isn't repealed, etc
		titleNode := htmlquery.FindOne(rowNode, titleRelativeToRowXPath)
		titleContents := strings.TrimSpace(htmlquery.InnerText(titleNode))
		isTitleContentsAllCaps := strings.ToUpper(titleContents) == titleContents
		if !isTitleContentsAllCaps { // not a valid statute (likely repealed?)
			continue
		}

		// find href attribute for node and if found, append it to urls slice
		aNode := htmlquery.FindOne(rowNode, hrefRelativeToRowXPath)
		url := htmlquery.SelectAttr(aNode, "href")
		if len(url) == 0 {
			return nil, fmt.Errorf("could not find 'href' attribute, %v", htmlquery.InnerText(rowNode))
		} else {
			urls = append(urls, url)
		}
	}
	return urls, nil
}

func (scraper *Scraper) ExtractStatute(contents io.Reader) (*core.Statute, error) {
	doc, err := htmlquery.Parse(contents)
	if err != nil {
		return nil, fmt.Errorf("error on parsing html: %v", err)
	}
	sectionNode := htmlquery.FindOne(doc, sectionDivXPath)
	if sectionNode == nil {
		return nil, fmt.Errorf("error could not find 'section' div")
	}
	title := htmlquery.FindOne(sectionNode, titleRelativeToSectionXPath)
	if title == nil {
		return nil, fmt.Errorf("error could not find statute title")
	}
	subdivisionDivs := htmlquery.Find(sectionNode, subdivDivRelativeToSectionXPath)
	var subdivisions []core.Subdivision
	if subdivisionDivs == nil {
		paraNode := htmlquery.FindOne(sectionNode, paraNodeRelativeToSectionXPath)
		if paraNode == nil {
			return nil, fmt.Errorf("error could not find subdivisionss")
		}
		var subdivision = core.Subdivision{
			Number:  -1,
			Heading: "",
			Content: htmlquery.InnerText(paraNode),
		}
		subdivisions = []core.Subdivision{subdivision}
	} else {
		subdivisions, err = scraper.extractSubdivisions(subdivisionDivs)
		if err != nil {
			return nil, err
		}
	}

	titleStr := htmlquery.InnerText(title)
	parts := strings.SplitN(titleStr, " ", 2)
	parts2 := strings.SplitN(parts[0], ".", 2)
	statute := core.Statute{
		Chapter:      parts2[0],
		Section:      parts2[1],
		Title:        parts[1],
		Subdivisions: subdivisions,
	}
	return &statute, nil
}

func (*Scraper) extractSubdivisions(subdivisionDivs []*html.Node) ([]core.Subdivision, error) {
	var subdivisions = make([]core.Subdivision, 0)
	for _, subd := range subdivisionDivs {

		subdNoNode := htmlquery.FindOne(subd, subdivNumberRelativeToSubdivDivXPath)
		if subdNoNode == nil {
			return nil, fmt.Errorf("could not find subdivision headers")
		}
		subdNoText := htmlquery.InnerText(subdNoNode)
		var subdNumTitle string
		if strings.HasPrefix(subdNoText, subdivisionPrefix) {
			subdNumTitle = subdNoText[len(subdivisionPrefix):]
		} else if strings.HasPrefix(subdNoText, subdPrefix) {
			subdNumTitle = subdNoText[len(subdPrefix):]
		} else {
			return nil, fmt.Errorf("could not determine the subdivision format")
		}

		subdNumTitleParts := strings.SplitN(subdNumTitle, ".", 2)
		if len(subdNumTitleParts) < 2 {
			return nil, fmt.Errorf("did not correctly parse subdivision number")
		}
		subdivNum, err := strconv.Atoi(subdNumTitleParts[0])
		if err != nil {
			return nil, fmt.Errorf("could not determine subdivision number")
		}
		heading := strings.TrimSpace(subdNumTitleParts[1])

		contentNode := htmlquery.FindOne(subd, contentRelativeToSubdivPath)
		if contentNode == nil {
			return nil, fmt.Errorf("could not determine subdivision body")
		}
		content := htmlquery.InnerText(contentNode)

		if len(heading) == 0 {
			if !strings.HasPrefix(content, repealedPrefix) {
				return nil, fmt.Errorf("could not verify subdivision was repealed")
			}
			continue
		}

		subdivision := core.Subdivision{
			Number:  subdivNum,
			Heading: heading,
			Content: content,
		}
		subdivisions = append(subdivisions, subdivision)
	}
	return subdivisions, nil
}
