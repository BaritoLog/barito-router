package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func fetchProfileByClusterName(client *http.Client, marketUrl, path, clusterName string) (profile *Profile, err error) {

	address := fmt.Sprintf("%s/%s", marketUrl, path)
	q := url.Values{}
	q.Add("cluster_name", clusterName)

	req, _ := http.NewRequest("GET", address, nil)
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		profile, err = NewProfileFromBytes(body)
	}

	return
}
