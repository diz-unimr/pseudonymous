package ttp

import (
	"encoding/xml"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"pseudonymous/config"
	"regexp"
	"testing"
)

func TestSetupDomains(t *testing.T) {

	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		res.WriteHeader(http.StatusOK)
	}))
	defer s.Close()

	client := GpasClient{Config: config.Gpas{
		Url: s.URL,
		Domains: config.Domains{
			Config: map[string]string{
				"foo": "bar",
				"bla": "blubb",
			},
		},
	}}
	project := "test"

	err := client.SetupDomains(project)

	assert.Nil(t, err)
}

func TestSetupDomainsExist(t *testing.T) {

	project := "test"

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer closeBody(r.Body)

		reqBody, _ := io.ReadAll(r.Body)

		// workaround for missing namespace in unmarshalled soap envelope
		re := regexp.MustCompile(`<name>(.*)</name>`)
		name := re.FindStringSubmatch(string(reqBody))[1]

		w.WriteHeader(http.StatusInternalServerError)

		res := FaultEnvelope{
			XMLName: xml.Name{
				Local: "Envelope",
			},
			Body: FaultBody{
				XMLName: xml.Name{
					Local: "Body",
				},
				Fault: Fault{
					FaultString: fmt.Sprintf("domain %s already exists", name),
				},
			},
		}

		resBody, _ := xml.Marshal(res)

		_, _ = w.Write(resBody)
	}))
	defer s.Close()

	client := GpasClient{Config: config.Gpas{
		Url: s.URL,
		Domains: config.Domains{
			UseExisting: true,
			Config: map[string]string{
				"foo": "bar",
				"bla": "blubb",
			},
		},
	}}

	err := client.SetupDomains(project)

	assert.Nil(t, err)
}
