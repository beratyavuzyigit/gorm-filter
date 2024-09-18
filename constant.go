package gormfilter

import "errors"

type SearchType string

var (
	EQUAL                 SearchType = "e"
	NOT_EQUAL             SearchType = "ne"
	GREATER_THAN          SearchType = "gt"
	GREATER_THAN_OR_EQUAL SearchType = "gte"
	LESS_THAN             SearchType = "lt"
	LESS_THAN_OR_EQUAL    SearchType = "lte"
	IN                    SearchType = "in"
	NOT_IN                SearchType = "nin"
	CONTAINS              SearchType = "contains"

	SEPERATOR       = "__"
	DEFAULT_PK      = "id"
	DEFAULT_PK_TYPE = "int"

	ErrFieldAppearsToBeUndefined = errors.New("field appears to be undefined")
	ErrBoolTypeNeedsTrueOrFalse  = errors.New("bool type needs true or false")
	ErrUndefinedSearchType       = errors.New("undefined search type")
)

var SearchTypes = []string{
	string(EQUAL),
	string(NOT_EQUAL),
	string(GREATER_THAN),
	string(GREATER_THAN_OR_EQUAL),
	string(LESS_THAN),
	string(LESS_THAN_OR_EQUAL),
	string(IN),
	string(CONTAINS),
}
