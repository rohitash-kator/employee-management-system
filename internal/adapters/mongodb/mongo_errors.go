package mongodb

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

func isDuplicateKey(err error) bool {
	var writeExc mongo.WriteException
	if errors.As(err, &writeExc) {
		for _, we := range writeExc.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}

	var cmdExc mongo.CommandError
	if errors.As(err, &cmdExc) {
		// Some servers/drivers return duplicate key via CommandError
		if cmdExc.Code == 11000 {
			return true
		}
	}

	return false
}
