package services

import (
	"encoding/json"
	artifactoryUtils "github.com/madotis/jfrog-client-go/artifactory/services/utils"
	"github.com/madotis/jfrog-client-go/auth"
	distributionServiceUtils "github.com/madotis/jfrog-client-go/distribution/services/utils"
	"github.com/madotis/jfrog-client-go/http/jfroghttpclient"
	"github.com/madotis/jfrog-client-go/utils"
	"github.com/madotis/jfrog-client-go/utils/errorutils"
	"github.com/madotis/jfrog-client-go/utils/log"
	"net/http"
)

type CreateReleaseBundleService struct {
	UpdateReleaseBundleService
}

func NewCreateReleaseBundleService(client *jfroghttpclient.JfrogHttpClient) *CreateReleaseBundleService {
	return &CreateReleaseBundleService{UpdateReleaseBundleService{client: client}}
}

func (cb *CreateReleaseBundleService) GetDistDetails() auth.ServiceDetails {
	return cb.DistDetails
}

func (cb *CreateReleaseBundleService) CreateReleaseBundle(createBundleParams CreateReleaseBundleParams) (*utils.Sha256Summary, error) {
	releaseBundleBody, err := distributionServiceUtils.CreateBundleBody(createBundleParams.ReleaseBundleParams, cb.DryRun)
	if err != nil {
		return nil, err
	}

	body := &createReleaseBundleBody{
		Name:              createBundleParams.Name,
		Version:           createBundleParams.Version,
		ReleaseBundleBody: *releaseBundleBody,
	}

	return cb.execCreateReleaseBundle(createBundleParams.GpgPassphrase, body)
}

// In case of an immediate sign- release bundle detailed summary (containing sha256) will be returned.
// In other cases summary will be nil.
func (cb *CreateReleaseBundleService) execCreateReleaseBundle(gpgPassphrase string, releaseBundle *createReleaseBundleBody) (*utils.Sha256Summary, error) {
	var summary *utils.Sha256Summary = nil
	if *releaseBundle.SignImmediately {
		summary = utils.NewSha256Summary()
	}
	httpClientsDetails := cb.DistDetails.CreateHttpClientDetails()
	content, err := json.Marshal(releaseBundle)
	if err != nil {
		return summary, errorutils.CheckError(err)
	}
	dryRunStr := ""
	if releaseBundle.DryRun {
		dryRunStr = "[Dry run] "
	}
	log.Info(dryRunStr + "Creating: " + releaseBundle.Name + "/" + releaseBundle.Version)

	url := cb.DistDetails.GetUrl() + "api/v1/release_bundle"
	distributionServiceUtils.AddGpgPassphraseHeader(gpgPassphrase, &httpClientsDetails.Headers)
	artifactoryUtils.SetContentType("application/json", &httpClientsDetails.Headers)
	resp, body, err := cb.client.SendPost(url, content, &httpClientsDetails)
	if err != nil {
		return summary, err
	}
	if !(resp.StatusCode == http.StatusCreated || (resp.StatusCode == http.StatusOK && releaseBundle.DryRun)) {
		return summary, errorutils.CheckErrorf("Distribution response: " + resp.Status + "\n" + utils.IndentJson(body))
	}
	if summary != nil {
		summary.SetSucceeded(true)
		summary.SetSha256(resp.Header.Get("X-Checksum-Sha256"))
	}

	log.Debug("Distribution response:", resp.Status)
	log.Debug(utils.IndentJson(body))
	return summary, nil
}

type createReleaseBundleBody struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	distributionServiceUtils.ReleaseBundleBody
}

type CreateReleaseBundleParams struct {
	distributionServiceUtils.ReleaseBundleParams
}

func NewCreateReleaseBundleParams(name, version string) CreateReleaseBundleParams {
	return CreateReleaseBundleParams{
		distributionServiceUtils.ReleaseBundleParams{
			Name:    name,
			Version: version,
		},
	}
}
