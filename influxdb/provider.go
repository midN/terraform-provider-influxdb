package influxdb

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/influxdata/influxdb/client"
)

var quoteReplacer = strings.NewReplacer(`"`, `\"`)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"influxdb_database":         resourceDatabase(),
			"influxdb_user":             resourceUser(),
			"influxdb_continuous_query": resourceContinuousQuery(),
		},

		Schema: map[string]*schema.Schema{
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc(
					"INFLUXDB_URL", "http://localhost:8086/",
				),
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFLUXDB_USERNAME", ""),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFLUXDB_PASSWORD", ""),
			},
			"skip_ssl_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFLUXDB_SKIP_SSL_VERIFY", "0"),
			},
		},

		ConfigureFunc: configure,
	}
}

func configure(d *schema.ResourceData) (interface{}, error) {
	url, err := url.Parse(d.Get("url").(string))
	if err != nil {
		return nil, fmt.Errorf("invalid InfluxDB URL: %s", err)
	}

	config := client.Config{
		URL:       *url,
		Username:  d.Get("username").(string),
		Password:  d.Get("password").(string),
		UnsafeSsl: d.Get("skip_ssl_verify").(bool),
	}

	conn, err := client.NewClient(config)
	if err != nil {
		return nil, err
	}

	_, _, err = conn.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging server: %s", err)
	}

	return conn, nil
}

func quoteIdentifier(ident string) string {
	return fmt.Sprintf(`%q`, quoteReplacer.Replace(ident))
}

func exec(conn *client.Client, query string) error {
	resp, err := conn.Query(client.Query{
		Command: query,
	})
	if err != nil {
		return err
	}
	if resp.Err != nil {
		return resp.Err
	}
	return nil
}
