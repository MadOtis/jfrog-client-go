package utils

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/madotis/jfrog-client-go/utils"
	"github.com/madotis/jfrog-client-go/utils/io/content"
	"github.com/madotis/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
)

func TestMatchingDelete(t *testing.T) {
	var actual string
	actual, _ = WildcardToDirsPath("s/*/path/", "s/a/path/b.zip")
	assertDeletePattern(t, "s/a/path/", actual)
	actual, _ = WildcardToDirsPath("s/*/path/", "s/a/b/c/path/b.zip")
	assertDeletePattern(t, "s/a/b/c/path/", actual)
	actual, _ = WildcardToDirsPath("s/a/*/", "s/a/b/path/b.zip")
	assertDeletePattern(t, "s/a/b/", actual)
	actual, _ = WildcardToDirsPath("s/*/path/*/", "s/a/path/a/b.zip")
	assertDeletePattern(t, "s/a/path/a/", actual)
	actual, _ = WildcardToDirsPath("s/*/path/*/", "s/a/a/path/a/b/c/d/b.zip")
	assertDeletePattern(t, "s/a/a/path/a/", actual)
	actual, _ = WildcardToDirsPath("s/*/", "s/a/a/path/a/b/c/d/b.zip")
	assertDeletePattern(t, "s/a/", actual)
	actual, _ = WildcardToDirsPath("s/*/a/*/", "s/a/a/path/k/b/c/d/b.zip")
	assertDeletePattern(t, "s/a/a/path/", actual)
	actual, _ = WildcardToDirsPath("s/*/a/*/*/", "s/a/a/path/k/b/c/d/b.zip")
	assertDeletePattern(t, "s/a/a/path/k/", actual)
	actual, _ = WildcardToDirsPath("s/*/*l*/*/*/", "s/a/l/path/k/b/c/d/b.zip")
	assertDeletePattern(t, "s/a/l/path/k/", actual)
	actual, _ = WildcardToDirsPath("s/*/a*/", "s/a/a/path/k/b/c/d/b.zip")
	assertDeletePattern(t, "s/a/a/", actual)
	actual, _ = WildcardToDirsPath("s/a*/", "s/a/a/path/k/b/c/d/b.zip")
	assertDeletePattern(t, "s/a/", actual)
	actual, _ = WildcardToDirsPath("s/*/", "s/a/a/path/k/b/c/d/b.zip")
	assertDeletePattern(t, "s/a/", actual)
	actual, _ = WildcardToDirsPath("s/*/*path*/", "s/a/h/path/k/b/c/d/b.zip")
	assertDeletePattern(t, "s/a/h/path/", actual)
	actual, _ = WildcardToDirsPath("a/b/*********/*******/", "a/b/c/d/e.zip")
	assertDeletePattern(t, "a/b/c/d/", actual)
	_, err := WildcardToDirsPath("s/*/a/*/*", "s/a/a/path/k/b/c/d/b.zip")
	assertDeletePatternErr(t, "delete pattern must end with \"/\"", err.Error())
}

func assertDeletePattern(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Error("Wrong matching expected: `" + expected + "` Got `" + actual + "`")
	}
}

func assertDeletePatternErr(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Error("Wrong err message expected: `" + expected + "` Got `" + actual + "`")
	}
}

func TestWriteCandidateDirsToBeDeleted(t *testing.T) {
	testPath := getBaseTestDir(t)
	{
		var bufferFiles []*content.ContentReader
		for i := 1; i <= 3; i++ {
			bufferFiles = append(bufferFiles, content.NewContentReader(filepath.Join(testPath, "buffer_file_ascending_order_"+strconv.Itoa(i)+".json"), content.DefaultKey))
		}
		resultWriter, err := content.NewContentWriter(content.DefaultKey, true, false)
		assert.NoError(t, err)
		artifactNotToBeDeleteReader := content.NewContentReader(filepath.Join(testPath, "artifact_file_1.json"), content.DefaultKey)
		assert.NoError(t, WriteCandidateDirsToBeDeleted(bufferFiles, artifactNotToBeDeleteReader, resultWriter))
		assert.NoError(t, resultWriter.Close())
		result, err := fileutils.JsonEqual(filepath.Join(testPath, "candidate_dirs_to_be_deleted_results.json"), resultWriter.GetFilePath())
		assert.NoError(t, err)
		assert.True(t, result)
		assert.NoError(t, resultWriter.RemoveOutputFilePath())
	}
	// Fixes issue https://github.com/jfrog/jfrog-cli/issues/808
	{
		resultWriter, err := content.NewContentWriter(content.DefaultKey, true, false)
		assert.NoError(t, err)
		var bufferFiles []*content.ContentReader
		bufferFiles = append(bufferFiles, content.NewContentReader(filepath.Join(testPath, "buffer_file_ascending_order_4.json"), content.DefaultKey))
		artifactNotToBeDeleteReader := content.NewContentReader(filepath.Join(testPath, "artifact_file_2.json"), content.DefaultKey)
		assert.NoError(t, WriteCandidateDirsToBeDeleted(bufferFiles, artifactNotToBeDeleteReader, resultWriter))
		assert.NoError(t, resultWriter.Close())
		assert.True(t, resultWriter.IsEmpty())
	}
}

func TestFilterCandidateToBeDeleted(t *testing.T) {
	testPath := getBaseTestDir(t)
	resultWriter, err := content.NewContentWriter(content.DefaultKey, true, false)
	assert.NoError(t, err)
	deleteCandidates := content.NewContentReader(filepath.Join(testPath, "prebuffer_file.json"), content.DefaultKey)
	assert.NoError(t, err)
	oldMaxSize := utils.MaxBufferSize
	defer func() { utils.MaxBufferSize = oldMaxSize }()
	utils.MaxBufferSize = 3
	sortedFiles, err := FilterCandidateToBeDeleted(deleteCandidates, resultWriter, "folder")
	assert.Len(t, sortedFiles, 3)
	assert.NoError(t, err)
	for i, val := range sortedFiles {
		assert.Equal(t, 1, len(val.GetFilesPaths()))
		result, err := fileutils.JsonEqual(val.GetFilesPaths()[0], filepath.Join(testPath, "buffer_file_ascending_order_"+strconv.Itoa(i+1)+".json"))
		assert.NoError(t, err)
		assert.True(t, result)
		assert.NoError(t, val.Close())
	}
	assert.NoError(t, resultWriter.Close())
	result, err := fileutils.JsonEqual(resultWriter.GetFilePath(), filepath.Join(testPath, "candidate_artifact_to_be_deleted_results.json"))
	assert.NoError(t, err)
	assert.True(t, result)
	assert.NoError(t, resultWriter.RemoveOutputFilePath())
}
