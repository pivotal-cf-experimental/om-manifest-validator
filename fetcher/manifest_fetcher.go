package fetcher

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/pivotal-cf-experimental/om-manifest-validator/bosh"

	"gopkg.in/yaml.v2"
)

type Products []Product

type Product struct {
	Type string `yaml:"type"`
	GUID string `yaml:"guid"`
}

type Environment struct {
	URL      string
	Username string
	Password string
}

func (e Environment) GetStagedProductManifest(name string) (*bosh.Manifest, error) {
	guid, err := e.GetProductGUID(name)
	if err != nil {
		return nil, err
	}

	return e.GetStagedProductManifestByGUID(guid)
}

func (e Environment) GetProductGUID(name string) (string, error) {
	var productGUID string

	client, err := e.oauthClient()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("GET", e.URL+"/api/v0/staged/products", nil)
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		b, _ := httputil.DumpResponse(res, true)
		return "", errors.New("error getting manifest: " + string(b))
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	ps := &Products{}
	yaml.Unmarshal(b, ps)

	for _, p := range *ps {
		if p.Type == name {
			productGUID = p.GUID
			break
		}
	}
	if productGUID == "" {
		return "", fmt.Errorf("could not find a product named %s", name)
	}

	return productGUID, nil
}

func (e Environment) GetStagedProductManifestByGUID(guid string) (*bosh.Manifest, error) {
	client, err := e.oauthClient()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", e.URL+"/api/v0/staged/products/"+guid+"/manifest", nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		b, _ := httputil.DumpResponse(res, true)
		return nil, errors.New("error getting manifest: " + string(b))
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	r := &bosh.StagedManifestResponse{}
	yaml.Unmarshal(b, r)

	return r.Manifest, nil
}

func (e Environment) oauthClient() (*http.Client, error) {
	return NewOAuthHTTPClient(e.URL+"/uaa", e.Username, e.Password)
}
