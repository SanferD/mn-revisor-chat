package scrapers

import (
	"code/core"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDataFolder                 = "test_data"
	chaptersTable                  = "chapters_table.html"
	chaptersShortTable             = "chapters_subtable.html"
	sectionsTable                  = "sections_list.html"
	sectionNoSubSections           = "section_no_subsections.html"
	sectionWithSubSections         = "section_with_subsections.html"
	sectionWithRepealedSubSections = "section_with_repealed_subsections.html"
)

type pageKindTest struct {
	testKind core.MNRevisorPageKind
	fileName string
}

var pageKindTests []pageKindTest = []pageKindTest{
	{testKind: core.StatutesChaptersTable, fileName: chaptersTable},
	{testKind: core.StatutesChaptersShortTable, fileName: chaptersShortTable},
	{testKind: core.StatutesSectionsTable, fileName: sectionsTable},
	{testKind: core.Statutes, fileName: sectionNoSubSections},
	{testKind: core.Statutes, fileName: sectionWithSubSections},
	{testKind: core.Statutes, fileName: sectionWithRepealedSubSections},
}

type extractURLsTest struct {
	fileName string
	urls     []string
}

var extractURLsTests = []extractURLsTest{
	{fileName: chaptersTable, urls: chaptersTableURLs},
	{fileName: chaptersShortTable, urls: chaptersShortTableURLs},
	{fileName: sectionsTable, urls: sectionsTableURLs},
}

type extractStatuteTest struct {
	fileName string
	statute  core.Statute
}

var extractStatuteTests = []extractStatuteTest{
	{fileName: sectionWithSubSections, statute: sectionWithSubSectionsStatute},
	{fileName: sectionWithRepealedSubSections, statute: sectionWithRepealedSubSectionsStatute},
	{fileName: sectionNoSubSections, statute: sectionWithNoSubSectionsStatute},
}

func TestScrapers(t *testing.T) {
	scraper, err := InitializeScraper()
	assert.NoError(t, err)

	t.Run("testing GetPageKind", func(t *testing.T) {
		for _, test := range pageKindTests {
			contents, err := readContents(test.fileName)
			assert.NoError(t, err, "error on reading text file contents: %v", err)
			pageKind, err := scraper.GetPageKind(contents)
			if assert.NoError(t, err, "error on inferring page kind: %v", err) {
				assert.Equal(t, test.testKind, pageKind, "expected test kind is not equal to actual test kind")
			}
		}
	})

	t.Run("testing extract urls", func(t *testing.T) {
		for _, test := range extractURLsTests {
			contents, err := readContents(test.fileName)
			assert.NoError(t, err, "error on reading text file contents: %v", err)
			pageKind, err := scraper.GetPageKind(contents)
			assert.NoError(t, err, "error on get page kind: %v", err)
			contents, _ = readContents(test.fileName)
			urls, err := scraper.ExtractURLs(contents, pageKind)
			if assert.NoError(t, err, "error on extract urls from chapters table: %v", err) {
				assert.ElementsMatch(t, test.urls, urls, "missing some urls")
			}
		}
	})

	t.Run("testing Statutes", func(t *testing.T) {
		for _, test := range extractStatuteTests {
			contents, err := readContents(test.fileName)
			assert.NoError(t, err, "error on reading text file contents: %v", err)
			statute, err := scraper.ExtractStatute(contents)
			if assert.NoError(t, err, "error on extracting statute: %v") {
				assert.Equal(t, test.statute, statute, "statutes are not equal")
			}
		}
	})
}

func readContents(fileName string) (io.Reader, error) {
	relativeFilePath := "./" + testDataFolder + "/" + fileName
	file, err := os.Open(relativeFilePath)
	if err != nil {
		return nil, fmt.Errorf("error on opening file: %v", err)
	}
	return file, nil
}
