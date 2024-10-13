package elastic_connect

import "encoding/json"

type EsResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type ListResponse struct {
	Total int                      `json:"total"`
	Items []map[string]interface{} `json:"items"`
}
