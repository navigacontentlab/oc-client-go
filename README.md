# oc-client-go

A project for simplifying when working with Open Content API.

## Example

    import "github.com/navigacontentlab/oc-client-go"

    client, err := oc.New(oc.Options{
		BaseURL: "https://host:8443/opencontent",
		Auth:    oc.BearerAuth("token"),
	})

	req := oc.SearchRequest{
		Properties: "uuid,updated",
		Query:      "contenttype:Image",
		Sort: []oc.SearchSort{{
			IndexField: "updated",
			Descending: true,
		}},
	}

	resp, err := client.Search(context.Background(), req)


## Metrics

This library provides metrics in the form of a prometheus collector.

The following metrics are supported:

* statusCodes
* responseTimes