package ovh

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

type IpLoadbalancingHttpFarmBackendProbe struct {
	Match    string `json:"match,omitempty"`
	Port     int    `json:"port,omitempty"`
	Interval int    `json:"interval,omitempty"`
	Negate   bool   `json:"negate,omitempty"`
	Pattern  string `json:"pattern,omitempty"`
	ForceSsl bool   `json:"forceSsl,omitempty"`
	URL      string `json:"url,omitempty"`
	Method   string `json:"method,omitempty"`
	Type     string `json:"type,omitempty"`
}

type IpLoadbalancingHttpFarm struct {
	FarmId         int                                  `json:"farmId,omitempty"`
	Zone           string                               `json:"zone,omitempty"`
	VrackNetworkId int                                  `json:"vrackNetworkId,omitempty"`
	Port           int                                  `json:"port,omitempty"`
	Stickiness     string                               `json:"stickiness,omitempty"`
	Balance        string                               `json:"balance,omitempty"`
	Probe          *IpLoadbalancingHttpFarmBackendProbe `json:"probe,omitempty"`
	DisplayName    string                               `json:"displayName,omitempty"`
}

func resourceIpLoadbalancingHttpFarm() *schema.Resource {
	return &schema.Resource{
		Create: resourceIpLoadbalancingHttpFarmCreate,
		Read:   resourceIpLoadbalancingHttpFarmRead,
		Update: resourceIpLoadbalancingHttpFarmUpdate,
		Delete: resourceIpLoadbalancingHttpFarmDelete,

		Schema: map[string]*schema.Schema{
			"service_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"balance": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					err := validateStringEnum(v.(string), []string{"first", "leastconn", "roundrobin", "source"})
					if err != nil {
						errors = append(errors, err)
					}
					return
				},
			},
			"display_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"port": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
			},
			"stickiness": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					err := validateStringEnum(v.(string), []string{"sourceIp"})
					if err != nil {
						errors = append(errors, err)
					}
					return
				},
			},
			"vrack_network_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
			},
			"zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"probe": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"match": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								err := validateStringEnum(v.(string), []string{"contains", "default", "internal", "matches", "status"})
								if err != nil {
									errors = append(errors, err)
								}
								return
							},
						},
						"port": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
						},
						"interval": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  30,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								value := v.(int)
								if value < 30 || value > 3600 {
									errors = append(errors, fmt.Errorf("Probe interval not in 30..3600 range"))
								}
								return
							},
						},
						"negate": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
						},
						"pattern": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"force_ssl": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
						},
						"url": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"method": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								err := validateStringEnum(v.(string), []string{"GET", "HEAD", "OPTIONS", "internal"})
								if err != nil {
									errors = append(errors, err)
								}
								return
							},
						},
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								err := validateStringEnum(v.(string), []string{"http", "internal", "mysql", "oko", "pgsql", "smtp", "http"})
								if err != nil {
									errors = append(errors, err)
								}
								return
							},
						},
					},
				},
			},
		},
	}
}

func resourceIpLoadbalancingHttpFarmCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	probe := &IpLoadbalancingHttpFarmBackendProbe{}
	probeSet := d.Get("probe").(*schema.Set)
	if probeSet.Len() > 0 {
		probeData := probeSet.List()[0].(map[string]interface{})
		probe.Match = probeData["match"].(string)
		probe.Port = probeData["port"].(int)
		probe.Interval = probeData["interval"].(int)
		probe.Negate = probeData["negate"].(bool)
		probe.Pattern = probeData["pattern"].(string)
		probe.ForceSsl = probeData["force_ssl"].(bool)
		probe.URL = probeData["url"].(string)
		probe.Method = probeData["method"].(string)
		probe.Type = probeData["type"].(string)
	}

	farm := &IpLoadbalancingHttpFarm{
		Zone:           d.Get("zone").(string),
		VrackNetworkId: d.Get("vrack_network_id").(int),
		Port:           d.Get("port").(int),
		Stickiness:     d.Get("stickiness").(string),
		Balance:        d.Get("balance").(string),
		Probe:          probe,
		DisplayName:    d.Get("display_name").(string),
	}

	service := d.Get("service_name").(string)
	resp := &IpLoadbalancingHttpFarm{}
	endpoint := fmt.Sprintf("/ipLoadbalancing/%s/http/farm", service)

	err := config.OVHClient.Post(endpoint, farm, resp)
	if err != nil {
		return fmt.Errorf("calling POST %s :\n\t %s", endpoint, err.Error())
	}

	d.SetId(fmt.Sprintf("%d", resp.FarmId))

	return nil
}

func resourceIpLoadbalancingHttpFarmRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	service := d.Get("service_name").(string)
	r := &IpLoadbalancingHttpFarm{}
	endpoint := fmt.Sprintf("/ipLoadbalancing/%s/http/farm/%s", service, d.Id())

	err := config.OVHClient.Get(endpoint, &r)
	if err != nil {
		return fmt.Errorf("calling %s:\n\t %s", endpoint, err.Error())
	}

	d.Set("display_name", r.DisplayName)

	return nil
}

func resourceIpLoadbalancingHttpFarmUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	service := d.Get("service_name").(string)
	endpoint := fmt.Sprintf("/ipLoadbalancing/%s/http/farm/%s", service, d.Id())

	probe := &IpLoadbalancingHttpFarmBackendProbe{}
	probeSet := d.Get("probe").(*schema.Set)
	if probeSet.Len() > 0 {
		probeData := probeSet.List()[0].(map[string]interface{})
		probe.Match = probeData["match"].(string)
		probe.Port = probeData["port"].(int)
		probe.Interval = probeData["interval"].(int)
		probe.Negate = probeData["negate"].(bool)
		probe.Pattern = probeData["pattern"].(string)
		probe.ForceSsl = probeData["force_ssl"].(bool)
		probe.URL = probeData["url"].(string)
		probe.Method = probeData["method"].(string)
		probe.Type = probeData["type"].(string)
	}

	farm := &IpLoadbalancingHttpFarm{
		VrackNetworkId: d.Get("vrack_network_id").(int),
		Port:           d.Get("port").(int),
		Stickiness:     d.Get("stickiness").(string),
		Balance:        d.Get("balance").(string),
		Probe:          probe,
		DisplayName:    d.Get("display_name").(string),
	}

	err := config.OVHClient.Put(endpoint, farm, nil)
	if err != nil {
		return fmt.Errorf("calling %s:\n\t %s", endpoint, err.Error())
	}

	return nil
}

func resourceIpLoadbalancingHttpFarmDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	service := d.Get("service_name").(string)
	r := &IpLoadbalancingHttpFarm{}
	endpoint := fmt.Sprintf("/ipLoadbalancing/%s/http/farm/%s", service, d.Id())

	err := config.OVHClient.Delete(endpoint, &r)
	if err != nil {
		return fmt.Errorf("Error calling %s: %s \n", endpoint, err.Error())
	}

	return nil
}
