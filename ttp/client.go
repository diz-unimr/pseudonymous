package ttp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"pseudonymous/config"
	"strings"
)

type GpasClient struct {
	Config config.Gpas
}

func NewGpasClient(cfg config.Gpas) *GpasClient {
	return &GpasClient{Config: cfg}
}

type AddDomainEnvelope struct {
	XMLName xml.Name      `xml:"soap:Envelope"`
	XMLNSs  string        `xml:"xmlns:soap,attr"`
	Psn     string        `xml:"xmlns:psn,attr"`
	Header  string        `xml:"soap:Header"`
	Body    AddDomainBody `xml:"soap:Body"`
}

type AddDomainBody struct {
	XMLName   xml.Name  `xml:"soap:Body"`
	AddDomain AddDomain `xml:"psn:addDomain"`
}

type AddDomain struct {
	DomainDTO DomainDTO `xml:"domainDTO"`
}
type DomainDTO struct {
	Name              string       `xml:"name"`
	Label             string       `xml:"label"`
	CheckDigitClass   string       `xml:"checkDigitClass"`
	Alphabet          string       `xml:"alphabet"`
	ParentDomainNames string       `xml:"parentDomainNames,omitempty"`
	Config            DomainConfig `xml:"config"`
}

type DomainConfig struct {
	PsnLength     int    `xml:"psnLength"`
	PsnPrefix     string `xml:"psnPrefix"`
	PsnsDeletable bool   `xml:"psnsDeletable"`
}

type FaultEnvelope struct {
	XMLName xml.Name  `xml:"Envelope"`
	Body    FaultBody `xml:"Body"`
}

type FaultBody struct {
	XMLName xml.Name `xml:"Body"`
	Fault   Fault    `xml:"Fault"`
}

type Fault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
}

func (c *GpasClient) SetupDomains(project string) error {
	// project parent domain
	domainConfig := createDomainDto(project, "", "")
	if err := c.send(domainConfig); err != nil {
		slog.Error("Failed to create gPAS domain", "domain", domainConfig.Name, "error", err)
		return err
	}

	for domain, prefix := range c.Config.Domains.Config {
		domainConfig = createDomainDto(project, domain, prefix)
		if err := c.send(domainConfig); err != nil {
			slog.Error("Failed to create gPAS domain", "domain", domainConfig.Name, "error", err)
			return err
		}
	}

	return nil
}

func createDomainDto(project string, idType string, prefix string) DomainDTO {

	name := project
	var parent string
	psnPrefix := fmt.Sprintf("PSN-%s-", strings.ToUpper(project))
	if idType != "" {
		name += fmt.Sprintf("-%s", idType)
		psnPrefix += fmt.Sprintf("%s-", strings.ToUpper(prefix))
		parent = project
	}

	return DomainDTO{
		Name:              name,
		Label:             name,
		CheckDigitClass:   "org.emau.icmvc.ganimed.ttp.psn.generator.NoCheckDigits",
		Alphabet:          "org.emau.icmvc.ganimed.ttp.psn.alphabets.Symbol32",
		ParentDomainNames: parent,
		Config: DomainConfig{
			PsnLength:     16,
			PsnPrefix:     psnPrefix,
			PsnsDeletable: false,
		},
	}
}

func (c *GpasClient) send(domainConfig DomainDTO) error {

	soap := AddDomainEnvelope{
		XMLNSs: "http://schemas.xmlsoap.org/soap/envelope/",
		Psn:    "http://psn.ttp.ganimed.icmvc.emau.org/",
		Body: AddDomainBody{
			AddDomain: AddDomain{DomainDTO: domainConfig},
		},
	}

	body, err := xml.MarshalIndent(&soap, " ", "  ")
	if err != nil {
		return err
	}

	// send soap request
	req, err := http.NewRequest(http.MethodPost, c.Config.Url, bytes.NewBufferString(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml")
	if c.Config.Auth != nil && c.Config.Auth.Basic != nil {
		req.SetBasicAuth(c.Config.Auth.Basic.Username, c.Config.Auth.Basic.Password)
	}

	resp, err := http.DefaultClient.Do(req)
	if resp != nil && resp.StatusCode != http.StatusOK {

		if resp.StatusCode == http.StatusInternalServerError && c.Config.Domains.UseExisting {
			// check response body
			defer closeBody(resp.Body)

			respBody, _ := io.ReadAll(resp.Body)
			//response := string(respBody)

			// parse soap response
			var fault FaultEnvelope
			err = xml.Unmarshal(respBody, &fault)
			if err != nil {
				return err
			}

			if fault.Body.Fault.FaultString == fmt.Sprintf("domain %s already exists", domainConfig.Name) {
				slog.Warn("Reusing existing domain", "domain", domainConfig.Name)
				return nil
			}
		}

		err = fmt.Errorf("soap request failed with status code %d", resp.StatusCode)
	}

	return err
}

func closeBody(body io.ReadCloser) {
	_ = body.Close()
}
