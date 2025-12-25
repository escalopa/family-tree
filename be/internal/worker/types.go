package worker

import "context"

type Cleaner interface {
	CleanExpired(ctx context.Context) error
}
