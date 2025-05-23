// Copyright IBM Corp. 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package vpc

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/IBM/vpc-go-sdk/vpcv1"
)

func DataSourceIBMIsVPCAddressPrefix() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMIsVPCAddressPrefixRead,

		Schema: map[string]*schema.Schema{
			"vpc": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"vpc", "vpc_name"},
				Description:  "The VPC identifier.",
			},
			"vpc_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"vpc", "vpc_name"},
				Description:  "The VPC name.",
			},
			"address_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"address_prefix", "address_prefix_name"},
				Description:  "The address prefix identifier.",
			},
			"address_prefix_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"address_prefix", "address_prefix_name"},
				Description:  "The address prefix name.",
			},
			"cidr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The CIDR block for this prefix.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time that the prefix was created.",
			},
			"has_subnets": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether subnets exist with addresses from this prefix.",
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL for this address prefix.",
			},
			"is_default": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether this is the default prefix for this zone in this VPC. If a default prefix was automatically created when the VPC was created, the prefix is automatically named using a hyphenated list of randomly-selected words, but may be updated with a user-specified name.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user-defined name for this address prefix. Names must be unique within the VPC the address prefix resides in.",
			},
			"zone": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The zone this address prefix resides in.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"href": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The URL for this zone.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The globally unique name for this zone.",
						},
					},
				},
			},
		},
	}
}

func dataSourceIBMIsVPCAddressPrefixRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcClient, err := meta.(conns.ClientSession).VpcV1API()
	if err != nil {
		tfErr := flex.DiscriminatedTerraformErrorf(err, err.Error(), "(Data) ibm_is_vpc_address_prefix", "read", "initialize-client")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	vpc_id := d.Get("vpc").(string)
	address_prefix_id := d.Get("address_prefix").(string)
	address_prefix_name := d.Get("address_prefix_name").(string)
	vpc_name := d.Get("vpc_name").(string)
	var addressPrefix *vpcv1.AddressPrefix
	if vpc_id == "" {
		start := ""
		allrecs := []vpcv1.VPC{}
		for {
			listVpcsOptions := &vpcv1.ListVpcsOptions{}
			if start != "" {
				listVpcsOptions.Start = &start
			}
			vpcs, _, err := vpcClient.ListVpcsWithContext(context, listVpcsOptions)
			if err != nil {
				tfErr := flex.TerraformErrorf(err, fmt.Sprintf("ListVpcsWithContext failed: %s", err.Error()), "(Data) ibm_is_vpc_address_prefix", "read")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return tfErr.GetDiag()
			}
			start = flex.GetNext(vpcs.Next)
			allrecs = append(allrecs, vpcs.Vpcs...)
			if start == "" {
				break
			}
		}
		vpc_found := false
		for _, vpc := range allrecs {
			if *vpc.Name == vpc_name {
				vpc_id = *vpc.ID
				vpc_found = true
				break
			}
		}
		if !vpc_found {
			err = fmt.Errorf("VPC with given name not found %s", vpc_name)
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("ListVpcsWithContext failed: %s", err.Error()), "(Data) ibm_is_vpc_address_prefix", "read")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
	}
	if address_prefix_id != "" {
		getVPCAddressPrefixOptions := &vpcv1.GetVPCAddressPrefixOptions{}

		getVPCAddressPrefixOptions.SetVPCID(vpc_id)
		getVPCAddressPrefixOptions.SetID(address_prefix_id)

		addressPrefix1, _, err := vpcClient.GetVPCAddressPrefixWithContext(context, getVPCAddressPrefixOptions)
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetVPCAddressPrefixWithContext failed: %s", err.Error()), "(Data) ibm_is_vpc_address_prefix", "read")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
		addressPrefix = addressPrefix1

	} else {
		start := ""
		allrecs := []vpcv1.AddressPrefix{}
		listVpcAddressPrefixesOptions := &vpcv1.ListVPCAddressPrefixesOptions{}

		listVpcAddressPrefixesOptions.SetVPCID(vpc_id)
		for {
			if start != "" {
				listVpcAddressPrefixesOptions.Start = &start
			}
			addressPrefixCollection, _, err := vpcClient.ListVPCAddressPrefixesWithContext(context, listVpcAddressPrefixesOptions)
			if err != nil {
				tfErr := flex.TerraformErrorf(err, fmt.Sprintf("ListVPCAddressPrefixesWithContext failed: %s", err.Error()), "(Data) ibm_is_vpc_address_prefix", "read")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return tfErr.GetDiag()
			}
			start = flex.GetNext(addressPrefixCollection.Next)
			allrecs = append(allrecs, addressPrefixCollection.AddressPrefixes...)
			if start == "" {
				break
			}
		}
		address_prefix_found := false
		for _, addressPrefixItem := range allrecs {
			if *addressPrefixItem.Name == address_prefix_name {
				addressPrefix = &addressPrefixItem
				address_prefix_found = true
				break
			}
		}
		if !address_prefix_found {
			err = fmt.Errorf("Address Prefix with given name not found %s", address_prefix_name)
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("ListVPCAddressPrefixesWithContext failed: %s", err.Error()), "(Data) ibm_is_vpc_address_prefix", "read")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
	}
	d.SetId(*addressPrefix.ID)
	if err = d.Set("cidr", addressPrefix.CIDR); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting cidr: %s", err), "(Data) ibm_is_vpc_address_prefix", "read", "set-cidr").GetDiag()
	}

	if err = d.Set("created_at", flex.DateTimeToString(addressPrefix.CreatedAt)); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting created_at: %s", err), "(Data) ibm_is_vpc_address_prefix", "read", "set-created_at").GetDiag()
	}

	if err = d.Set("has_subnets", addressPrefix.HasSubnets); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting has_subnets: %s", err), "(Data) ibm_is_vpc_address_prefix", "read", "set-has_subnets").GetDiag()
	}

	if err = d.Set("href", addressPrefix.Href); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting href: %s", err), "(Data) ibm_is_vpc_address_prefix", "read", "set-href").GetDiag()
	}

	if err = d.Set("is_default", addressPrefix.IsDefault); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting is_default: %s", err), "(Data) ibm_is_vpc_address_prefix", "read", "set-is_default").GetDiag()
	}

	if err = d.Set("name", addressPrefix.Name); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting name: %s", err), "(Data) ibm_is_vpc_address_prefix", "read", "set-name").GetDiag()
	}

	zone := []map[string]interface{}{}
	if addressPrefix.Zone != nil {
		modelMap, err := dataSourceIBMIsVPCAddressPrefixZoneReferenceToMap(addressPrefix.Zone)
		if err != nil {
			return flex.DiscriminatedTerraformErrorf(err, err.Error(), "(Data) ibm_is_vpc_address_prefix", "read", "zone-to-map").GetDiag()
		}
		zone = append(zone, modelMap)
	}
	if err = d.Set("zone", zone); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting zone: %s", err), "(Data) ibm_is_vpc_address_prefix", "read", "set-zone").GetDiag()
	}

	return nil
}

func dataSourceIBMIsVPCAddressPrefixZoneReferenceToMap(model *vpcv1.ZoneReference) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	if model.Href != nil {
		modelMap["href"] = model.Href
	}
	if model.Name != nil {
		modelMap["name"] = model.Name
	}
	return modelMap, nil
}
