package freeipa

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	ipa "github.com/tehwalris/go-freeipa/freeipa"
)

func resourceFreeIPAService() *schema.Resource {
	return &schema.Resource{
		Create: resourceFreeIPAServiceCreate,
		Read:   resourceFreeIPAServiceRead,
		Update: resourceFreeIPAServiceUpdate,
		Delete: resourceFreeIPAServiceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceFreeIPAServiceImport,
		},

		Schema: map[string]*schema.Schema{
			"service": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"host": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"force": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceFreeIPAServiceCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO][freeipa] Creating Service: %s", d.Id())
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	service := d.Get("service").(string)
	host := d.Get("host").(string)
	force := d.Get("force").(bool)

	krbcanonicalname := fmt.Sprintf("%s/%s", service, host)

	optArgs := ipa.ServiceAddOptionalArgs{
		Force: &force,
	}

	_, err = client.ServiceAdd(
		&ipa.ServiceAddArgs{
			Krbcanonicalname: krbcanonicalname,
		},
		&optArgs,
	)
	log.Printf("[INFO][freeipa] Service %s created: %v", krbcanonicalname, err)
	if err != nil {
		return err
	}

	d.SetId(krbcanonicalname)

	// FIXME: When using a LB in front of a FreeIPA cluster, sometime the record
	// is not replicated on the server where the read is done, so we have to
	// retry to not have "Error: NotFound (4001)".
	// Maybe we should use resource.StateChangeConf instead...
	// sleepDelay := 1 * time.Second
	// for {
	// 	err := resourceFreeIPAServiceRead(d, meta)
	// 	if err == nil {
	// 		return nil
	// 	}
	// 	time.Sleep(sleepDelay)
	// 	sleepDelay = sleepDelay * 2
	// }
	return nil
}

func resourceFreeIPAServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Updating Service: %s", d.Id())
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	service := d.Get("service").(string)
	host := d.Get("host").(string)

	krbcanonicalname := fmt.Sprintf("%s/%s", service, host)

	_, err = client.ServiceMod(
		&ipa.ServiceModArgs{
			Krbcanonicalname: krbcanonicalname,
		},
		&ipa.ServiceModOptionalArgs{},
	)
	if err != nil {
		return err
	}

	return resourceFreeIPAServiceRead(d, meta)
}

func resourceFreeIPAServiceRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Refreshing Service: %s", d.Id())
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	service := d.Get("service").(string)
	host := d.Get("host").(string)

	krbcanonicalname := fmt.Sprintf("%s/%s", service, host)

	_, err = client.ServiceShow(
		&ipa.ServiceShowArgs{
			Krbcanonicalname: krbcanonicalname,
		},
		&ipa.ServiceShowOptionalArgs{},
	)
	if err != nil {
		return err
	}

	return nil
}

func resourceFreeIPAServiceDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Deleting Service: %s", d.Id())
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	service := d.Get("service").(string)
	host := d.Get("host").(string)

	krbcanonicalname := fmt.Sprintf("%s/%s", service, host)

	_, err = client.ServiceDel(
		&ipa.ServiceDelArgs{
			Krbcanonicalname: []string{krbcanonicalname},
		},
		&ipa.ServiceDelOptionalArgs{},
	)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func resourceFreeIPAServiceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())
	krbcanonicalname := strings.Split(d.Id(), "/")
	d.Set("service", krbcanonicalname[0])
	d.Set("host", krbcanonicalname[1])

	err := resourceFreeIPAServiceRead(d, meta)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	return []*schema.ResourceData{d}, nil
}
