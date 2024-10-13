package elastic_connect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/satjan/gokit"
	"net/http"
	"strings"
)

type esSearch struct {
	BaseUrl string
}

type EsSearch interface {
	Search(ctx *fiber.Ctx, indexParam string) (*ListResponse, error)
}

var instance EsSearch

func GetInstance(baseUrl string) EsSearch {
	if instance == nil {
		instance = &esSearch{
			BaseUrl: baseUrl,
		}
	}
	return instance
}

func (e *esSearch) Search(ctx *fiber.Ctx, indexParam string) (*ListResponse, error) {
	var query map[string]interface{}
	_ = ctx.BodyParser(&query)

	elasticQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"__deleted": "false",
						},
					},
				},
			},
		},
	}

	if mustQueries, exists := query["must"]; exists {
		if mustArray, ok := mustQueries.([]interface{}); ok {
			for _, mustQuery := range mustArray {
				elasticQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
					elasticQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]interface{}),
					mustQuery,
				)
			}
		}
	}

	sourceFilter := make(map[string]interface{})
	if includes := ctx.Query("includes"); includes != "" {
		sourceFilter["includes"] = strings.Split(includes, ",")
	}

	excludesArr := []string{"__deleted"}
	if excludes := ctx.Query("excludes"); excludes != "" {
		excludesArr = append(excludesArr, strings.Split(excludes, ",")...)
	}
	sourceFilter["excludes"] = excludesArr

	if len(sourceFilter) > 0 {
		elasticQuery["_source"] = sourceFilter
	}

	size := ctx.Query("size", "100")
	if gokit.StrToInt(size) > 5000 {
		size = "100"
	}

	from := ctx.Query("from", "0")
	elasticURL := fmt.Sprintf("%s/%s/_search?pretty&size=%s&from=%s", e.BaseUrl, indexParam, size, from)
	body, _ := json.Marshal(elasticQuery)
	req, err := http.NewRequest("GET", elasticURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.New("unable to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("request failed")
	}
	defer resp.Body.Close()

	var elasticResp EsResponse
	if errDecode := json.NewDecoder(resp.Body).Decode(&elasticResp); errDecode != nil {
		return nil, errors.New("failed to decode response")
	}

	customResp := ListResponse{
		Total: elasticResp.Hits.Total.Value,
		Items: make([]map[string]interface{}, 0),
	}

	for _, hit := range elasticResp.Hits.Hits {
		var sourceData map[string]interface{}
		if errSrc := json.Unmarshal(hit.Source, &sourceData); errSrc != nil {
			return nil, errors.New("failed to decode response")
		}
		customResp.Items = append(customResp.Items, sourceData)
	}

	return &customResp, nil
}
