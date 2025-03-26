package types

import (
    "fmt"
    "strings"
    "strconv"
)

// Query types

// Id query
type ById struct {
    field string
    ids []int64
}

func ByIdFieldFromQueryParam(field string, queryParams []string) (Query, error) {
    ids := make([]int64, 0)
    for _, queryParam := range queryParams {
        idStrings := strings.Split(queryParam, "|")
        for _, idString := range idStrings {
            id, err := strconv.ParseInt(idString, 10, 64)
            if err != nil {
                return nil, err
            }
            ids = append(ids, id)
        }
    }
    return &ById { field: field, ids: ids }, nil
}

func (q *ById) Sql() (string, map[string]any) {
    clauses := make([]string, 0)
    args := make(map[string]any)
    for i := 0; i < len(q.ids); i++ {
        clauses = append(clauses, fmt.Sprintf("%s = @%sId%d", q.field, q.field, i))
        args[fmt.Sprintf("%sId%d", q.field, i)] = q.ids[i]
    }
    return strings.Join(clauses, " or "), args
}

// Contents Contains Query
type ByContains struct {
    contains string
    field string
}

func ByContainsFromQueryParam(field string, queryParam []string) (Query, error) {
    if len(queryParam) > 1 || len(queryParam) == 0 {
        return nil, fmt.Errorf("Only 1 parameter of type string allowed")
    }
    return &ByContains{ contains: queryParam[0], field: field }, nil
}

func (q *ByContains) Sql() (string, map[string]any) {
    args := map[string]any{ "contains": "%" + q.contains + "%" }
    clause := "contents like @contains"
    return clause, args
}


