package store

import (
    "github.com/reshane/gorest/types"
)

type Store interface {
    // generic over DataType interface
    Get(string, int) (types.DataType, error)
    Create(types.DataType) (types.DataType, error)
    Update(types.DataType) (types.DataType, error)
    Delete(string, int) (error)
}
