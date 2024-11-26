package fhir

import (
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"pseudonymous/config"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {

	// arrange
	c := config.Pseudonymizer{
		Retry: config.Retry{
			Count:   3,
			Timeout: 5,
			Wait:    5,
			MaxWait: 15,
		},
		Auth: &config.Auth{
			Basic: &config.Basic{
				Username: "foo",
				Password: "bar",
			},
		},
	}

	// act
	client := NewClient(c)

	// assert client config reflects pseudonymizer config
	assert.EqualValues(t, resty.User{
		Username: "foo",
		Password: "bar",
	}, *client.rest.UserInfo)
	assert.Equal(t, 3, client.rest.RetryCount)
	assert.Equal(t, 5*time.Second, client.rest.GetClient().Timeout)
	assert.Equal(t, 5*time.Second, client.rest.RetryWaitTime)
	assert.Equal(t, 15*time.Second, client.rest.RetryMaxWaitTime)
}
