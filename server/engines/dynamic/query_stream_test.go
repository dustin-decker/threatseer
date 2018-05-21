package main

import (
	"encoding/json"
	"testing"

	"github.com/caibirdme/yql"
)

var jsonString = `
  {
    "name": "enny",
    "gender": "f",
    "age": 36,
    "hobby": null,
    "skills": [
      "IC",
      "Electric design",
      "Verification"
    ]
  }
`

func BenchmarkYQL(b *testing.B) {
	rawYQL := `name='curl'`

	for i := 0; i < b.N; i++ {

		var temp map[string]interface{}

		json.Unmarshal([]byte(jsonString), &temp)

		yql.Match(rawYQL, temp)
		// result, _ := yql.Match(rawYQL, temp)
		// fmt.Println(result)
	}
}
