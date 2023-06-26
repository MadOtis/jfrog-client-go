package services

import (
	"encoding/json"
	"github.com/madotis/jfrog-client-go/artifactory/services/utils"
	"github.com/madotis/jfrog-client-go/auth"
	"github.com/madotis/jfrog-client-go/http/jfroghttpclient"
	clientutils "github.com/madotis/jfrog-client-go/utils"
	"github.com/madotis/jfrog-client-go/utils/errorutils"
	"github.com/madotis/jfrog-client-go/utils/log"
	"net/http"
)

package _import

type ImportService struct {
	client     *jfroghttpclient.JfrogHttpClient
	artDetails auth.ServiceDetails
	// If true, the import will only print the parameters
	DryRun bool
}

func NewImportService(artDetails auth.ServiceDetails, client *jfroghttpclient.JfrogHttpClient) *ImportService {
	return &ImportService{artDetails: artDetails, client: client}
}

func (drs *ImportService) GetJfrogHttpClient() *jfroghttpclient.JfrogHttpClient {
	return drs.client
}

func (drs *ImportService) Import(importParams ImportParams) error {
	httpClientsDetails := drs.artDetails.CreateHttpClientDetails()
	requestContent, err := json.Marshal(ImportBody(importParams))
	if err != nil {
		return errorutils.CheckError(err)
	}

	importMessage := "Running full system import..."
	if drs.DryRun {
		log.Info("[Dry run] " + importMessage)
		log.Info("Import parameters: \n" + clientutils.IndentJson(requestContent))
		return nil
	}
	log.Info(importMessage)

	utils.SetContentType("application/json", &httpClientsDetails.Headers)
	resp, body, err := drs.client.SendPost(drs.artDetails.GetUrl()+"api/import/system", requestContent, &httpClientsDetails)
	if err != nil {
		return err
	}
	if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK); err != nil {
		return err
	}
	log.Info(string(body))
	log.Debug("Artifactory response:", resp.Status)
	return nil
}

type ImportParams struct {
	// Mandatory:
	// A path to a directory on the local file system of Artifactory server
	ImportPath string

	// Optional:
	// If true, repository metadata is included in import (Maven 2 metadata is unaffected by this setting)
	IncludeMetadata *bool
	// If true, prints more verbose logging
	Verbose *bool
	// If true, includes Maven 2 repository metadata and checksum files as part of the import
	M2 *bool
	// If true, repository binaries are excluded from the import
	ExcludeContent *bool
}

type ImportBody struct {
	ImportPath      string `json:"importPath,omitempty"`
	IncludeMetadata *bool  `json:"includeMetadata,omitempty"`
	Verbose         *bool  `json:"verbose,omitempty"`
	M2              *bool  `json:"m2,omitempty"`
	ExcludeContent  *bool  `json:"excludeContent,omitempty"`
}

func NewImportParams(importPath string) ImportParams {
	return ImportParams{ImportPath: importPath}
}
