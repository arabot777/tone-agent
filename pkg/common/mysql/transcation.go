package mysql

import "context"

type Transaction interface {
	ExecTx(context.Context, func(ctx context.Context) error) error
}
