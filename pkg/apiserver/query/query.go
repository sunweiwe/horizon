package query

import (
	"strconv"

	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/horizon/pkg/utils/slice"
)

const (
	ParameterName          = "name"
	ParameterLabelSelector = "labelSelector"
	ParameterFieldSelector = "fieldSelector"
	ParameterPage          = "page"
	ParameterLimit         = "limit"
	ParameterOrderBy       = "sortBy"
	ParameterAscending     = "ascending"
)

type Query struct {
	Pagination *Pagination

	SortBy Field

	Ascending bool

	Filters map[Field]Value

	LabelSelector string
}

type Pagination struct {
	Limit int

	Offset int
}

var NoPagination = newPagination(-1, 0)

func newPagination(limit int, offset int) *Pagination {
	return &Pagination{
		Limit:  limit,
		Offset: offset,
	}
}

func New() *Query {
	return &Query{
		Pagination: NoPagination,
		SortBy:     "",
		Ascending:  false,
		Filters:    map[Field]Value{},
	}
}

func ParseQueryParameter(request *restful.Request) *Query {
	query := New()

	limit, err := strconv.Atoi(request.QueryParameter(ParameterLimit))

	if err != nil {
		limit = -1
	}

	page, err := strconv.Atoi(request.QueryParameter(ParameterPage))
	if err != nil {
		page = 1
	}

	query.Pagination = newPagination(limit, (page-1)*limit)

	query.SortBy = Field(defaultString(request.QueryParameter(ParameterOrderBy), FieldCreationTimeStamp))

	ascending, err := strconv.ParseBool(defaultString(request.QueryParameter(ParameterAscending), "false"))
	if err != nil {
		query.Ascending = false
	} else {
		query.Ascending = ascending
	}

	query.LabelSelector = request.QueryParameter(ParameterFieldSelector)

	for key, values := range request.Request.URL.Query() {
		if !slice.HasString([]string{ParameterPage, ParameterLimit, ParameterOrderBy, ParameterAscending, ParameterLabelSelector}, key) {
			for _, vaule := range values {
				query.Filters[Field(key)] = Value(vaule)
			}
		}
	}

	return query
}

func defaultString(value, defaultValue string) string {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
