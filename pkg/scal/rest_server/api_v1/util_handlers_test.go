package api_v1

import (
	"testing"

	"github.com/ilius/is"
	. "github.com/ilius/ripo"
)

func TestGetApiVersion(t *testing.T) {
	is := is.New(t)
	var req ExtendedRequest = nil
	// req argument is not used, so should work with req=nil, no need to mock it
	res, err := GetApiVersion(req)
	is.NotNil(res)
	is.NotErr(err)
	is.Equal(res.Data, "1")
	is.Equal(res.Header.Get("Content-Type"), "text/plain; charset=UTF-8")
}
