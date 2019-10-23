package api_v1

import (
	"testing"

	"github.com/ilius/is"
)

func TestIsGoTest(t *testing.T) {
	is := is.New(t)
	is.True(isGoTest())
}
