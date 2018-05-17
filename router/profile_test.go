package router

import (
	"fmt"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestProfile_New(t *testing.T) {
	wantId := "some-id"
	wantName := "some-name"
	wantConsul := "some-consul"

	jsonBody := fmt.Sprintf(`{
		"id": "%s",
		"name": "%s",
		"consul": "%s"
	}`, wantId, wantName, wantConsul)
	profile, err := NewProfileFromBytes([]byte(jsonBody))

	FatalIfError(t, err)
	FatalIf(t, profile.Id != wantId, "%s != %s", profile.Id, wantId)
	FatalIf(t, profile.Name != wantName, "%s != %s", profile.Name, wantName)
	FatalIf(t, profile.Consul != wantConsul, "%s != %s", profile.Consul, wantConsul)
}

func TestProfile_New_InvalidJson(t *testing.T) {
	_, err := NewProfileFromBytes([]byte("invalid-json"))
	FatalIfWrongError(t, err, "invalid character 'i' looking for beginning of value")
}
