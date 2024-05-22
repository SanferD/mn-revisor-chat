package scrapers

const (
	sectionDivXPath                      = "//div[@class='section']"
	sectionsListH2XPath                  = "//h2[@class='chapter_title']"
	tableOfChaptersH2XPath               = "//h2/../table/../h2[not(@class='subd_no')]"
	tableXPath                           = "//table[@id='toc_table']/tbody/tr"
	shortTableXPath                      = "//table[@id='chapters_table']/tbody/tr"
	sectionsTableXPath                   = "//div[@id='chapter_analysis']/table/tbody/tr"
	statuteSectionXPath                  = "//div[@class='section']"
	titleRelativeToRowXPath              = "/td[2]"
	hrefRelativeToRowXPath               = "/td[1]/a"
	titleRelativeToSectionXPath          = "//h1['shn']"
	subdivDivRelativeToSectionXPath      = "//div[@class='subd']"
	paraNodeRelativeToSectionXPath       = "//p"
	subdivNumberRelativeToSubdivDivXPath = "//h2[@class='subd_no']"
	contentRelativeToSubdivPath          = "//p"
)
