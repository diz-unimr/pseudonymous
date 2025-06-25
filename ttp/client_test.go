package ttp

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"pseudonymous/config"
	"testing"
)

func TestSetupDomains(t *testing.T) {

	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		res.WriteHeader(http.StatusOK)
	}))
	defer s.Close()

	client := GpasClient{config: config.Gpas{
		Url: s.URL,
		Domains: map[string]string{
			"foo": "bar",
			"bla": "blubb",
		},
	}}
	project := "test"

	err := client.SetupDomains(project)

	assert.Nil(t, err)
}
