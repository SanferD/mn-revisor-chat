package scrapers

const (
	contentRelativeToSubdivXPath         = "//p"
	hrefRelativeToRowXPath               = "/td[1]/a"
	paraNodeRelativeToSectionXPath       = "//p"
	sectionDivXPath                      = "//div[@class='section']"
	sectionsListH2XPath                  = "//h2[@class='chapter_title']"
	sectionsTableXPath                   = "//div[@id='chapter_analysis']/table/tbody/tr"
	shortTableXPath                      = "//table[@id='chapters_table']/tbody/tr"
	statuteSectionXPath                  = "//div[@class='section']"
	subdivDivRelativeToSectionXPath      = "//div[@class='subd']"
	subdivNumberRelativeToSubdivDivXPath = "//h2[@class='subd_no']"
	tableBodyRelativeToTableXPath        = "//tbody"
	tableCellRelativeToTableRowXPath     = "//td"
	tableOfChaptersH2XPath               = "//h2/../table/../h2[not(@class='subd_no')]"
	tableRelativeToSubdivisionXPath      = "//table"
	tableRowRelativeToTableBodyXPath     = "//tr"
	tableXPath                           = "//table[@id='toc_table']/tbody/tr"
	titleRelativeToRowXPath              = "/td[2]"
	titleRelativeToSectionXPath          = "//h1['shn']"
)
