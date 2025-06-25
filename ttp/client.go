package ttp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"pseudonymous/config"
	"strings"
)

type GpasClient struct {
	config config.Gpas
}

func NewGpasClient(cfg config.Gpas) *GpasClient {
	return &GpasClient{config: cfg}
}

type SoapEnvelope struct {
	XMLName xml.Name `xml:"soap:Envelope"`
	XMLNSs  string   `xml:"xmlns:soap,attr"`
	Psn     string   `xml:"xmlns:psn,attr"`
	Header  string   `xml:"soap:Header"`
	Body    SoapBody `xml:"soap:Body"`
}

type SoapBody struct {
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

func (c *GpasClient) SetupDomains(project string) error {
	// project parent domain
	domainConfig := createDomainDto(project, "", "")
	if err := c.send(domainConfig); err != nil {
		slog.Error("Failed to create gPAS domain", "domain", domainConfig.Name, "error", err)
		return err
	}

	for domain, prefix := range c.config.Domains {
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

	soap := SoapEnvelope{
		XMLNSs: "http://schemas.xmlsoap.org/soap/envelope/",
		Psn:    "http://psn.ttp.ganimed.icmvc.emau.org/",
		Body: SoapBody{
			AddDomain: AddDomain{DomainDTO: domainConfig},
		},
	}

	body, err := xml.MarshalIndent(&soap, " ", "  ")
	if err != nil {
		return err
	}

	// send soap request
	req, err := http.NewRequest(http.MethodPost, c.config.Url, bytes.NewBufferString(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml")
	if c.config.Auth != nil && c.config.Auth.Basic != nil {
		req.SetBasicAuth(c.config.Auth.Basic.Username, c.config.Auth.Basic.Password)
	}

	resp, err := http.DefaultClient.Do(req)
	if resp != nil && resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("soap request failed with status code %d", resp.StatusCode)
	}

	return err
}
