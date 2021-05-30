package config

import (
	"context"

	"github.com/digitalocean/godo"
)

func DropletList(ctx context.Context, client *godo.Client) ([]godo.Droplet, error) {
	// create a list to hold our droplets
	list := []godo.Droplet{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := client.Droplets.List(ctx, opt)
		if err != nil {
			return nil, err
		}

		// append the current page's droplets to our list
		list = append(list, droplets...)

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}

// filterDrops returns the droplets that have a specified tag
func FilterDrops(drops []godo.Droplet, tag string) (out []godo.Droplet, err error) {
	for _, drop := range drops {
		if hasTag(drop.Tags, tag) {
			out = append(out, drop)
		}
	}
	return
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
