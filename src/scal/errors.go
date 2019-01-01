package scal

import (
	"github.com/ilius/ripo"
)

var ErrorCodeByName = map[string]ripo.Code{
	"Canceled":           ripo.Canceled,
	"Unknown":            ripo.Unknown,
	"InvalidArgument":    ripo.InvalidArgument,
	"DeadlineExceeded":   ripo.DeadlineExceeded,
	"NotFound":           ripo.NotFound,
	"AlreadyExists":      ripo.AlreadyExists,
	"PermissionDenied":   ripo.PermissionDenied,
	"Unauthenticated":    ripo.Unauthenticated,
	"ResourceExhausted":  ripo.ResourceExhausted,
	"FailedPrecondition": ripo.FailedPrecondition,
	"Aborted":            ripo.Aborted,
	"OutOfRange":         ripo.OutOfRange,
	"Unimplemented":      ripo.Unimplemented,
	"Internal":           ripo.Internal,
	"Unavailable":        ripo.Unavailable,
	"DataLoss":           ripo.DataLoss,
	"MissingArgument":    ripo.MissingArgument, // extra code
	"ResourceLocked":     ripo.ResourceLocked,  // extra code
}
